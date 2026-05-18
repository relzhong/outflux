package main

import (
	"fmt"
	"sort"
	"time"

	influx "github.com/influxdata/influxdb/client/v2"
	"github.com/timescale/outflux/internal/cli"
)

type migrationWindow struct {
	From time.Time
	To   time.Time
}

func parseOptionalDuration(raw string) (time.Duration, error) {
	if raw == "" {
		return 0, nil
	}
	return time.ParseDuration(raw)
}

func buildMigrationWindows(app *appContext, infConn influx.Client, db string, args *cli.MigrationConfig) ([]migrationWindow, error) {
	maxWindow, err := parseOptionalDuration(args.MaxWindow)
	if err != nil {
		return nil, fmt.Errorf("invalid max window: %v", err)
	}
	groups, err := app.influxShardExplorer.FetchShardGroups(infConn, db, args.RetentionPolicy)
	if err != nil {
		return nil, err
	}
	var requestedFrom, requestedTo *time.Time
	if args.From != "" {
		parsed, err := time.Parse(time.RFC3339, args.From)
		if err != nil {
			return nil, err
		}
		requestedFrom = &parsed
	}
	if args.To != "" {
		parsed, err := time.Parse(time.RFC3339, args.To)
		if err != nil {
			return nil, err
		}
		requestedTo = &parsed
	}

	windows := []migrationWindow{}
	for _, group := range groups {
		from := group.Start
		to := group.End
		if requestedFrom != nil && requestedFrom.After(from) {
			from = *requestedFrom
		}
		if requestedTo != nil && requestedTo.Before(to) {
			to = requestedTo.Add(time.Nanosecond)
		}
		if !from.Before(to) {
			continue
		}
		windows = append(windows, splitMigrationWindow(migrationWindow{From: from, To: to}, maxWindow)...)
	}
	sort.Slice(windows, func(i, j int) bool { return windows[i].From.Before(windows[j].From) })
	return windows, nil
}

func splitMigrationWindow(window migrationWindow, maxWindow time.Duration) []migrationWindow {
	if maxWindow <= 0 || window.To.Sub(window.From) <= maxWindow {
		return []migrationWindow{window}
	}
	windows := []migrationWindow{}
	for start := window.From; start.Before(window.To); {
		end := start.Add(maxWindow)
		if end.After(window.To) {
			end = window.To
		}
		windows = append(windows, migrationWindow{From: start, To: end})
		start = end
	}
	return windows
}

func applyWindow(args *cli.MigrationConfig, window migrationWindow) *cli.MigrationConfig {
	windowArgs := *args
	windowArgs.From = window.From.Format(time.RFC3339Nano)
	windowArgs.To = window.To.Add(-time.Nanosecond).Format(time.RFC3339Nano)
	windowArgs.Limit = 0
	return &windowArgs
}
