package main

import (
	"fmt"
	"sort"
	"time"

	influx "github.com/influxdata/influxdb/client/v2"
	"github.com/timescale/outflux/internal/cli"
	"github.com/timescale/outflux/internal/idrf"
	"github.com/timescale/outflux/internal/schemamanagement/schemaconfig"
)

type existingNonNullableColumnFinder interface {
	ExistingNonNullableColumns(*idrf.DataSet, schemaconfig.SchemaStrategy) ([]string, error)
}

func validateNotNullSourceData(app *appContext, connArgs *cli.ConnectionConfig, args *cli.MigrationConfig) error {
	pause, err := parseOptionalDuration(args.PreflightShardPause)
	if err != nil {
		return fmt.Errorf("invalid preflight shard pause: %v", err)
	}
	maxWindow, err := parseOptionalDuration(args.PreflightMaxWindow)
	if err != nil {
		return fmt.Errorf("invalid preflight max window: %v", err)
	}

	for _, measure := range connArgs.InputMeasures {
		infConn, pgConn, err := openConnections(app, connArgs)
		if err != nil {
			return fmt.Errorf("could not open connections for non-null validation\n%v", err)
		}

		session, err := app.notNullValidator.Prepare(infConn, measure, connArgs.InputDb, args)
		if err != nil {
			infConn.Close()
			pgConn.Close()
			return fmt.Errorf("could not prepare non-null validation for measurement '%s'\n%v", measure, err)
		}

		dataSet := session.DataSet()
		targetDataSet, err := renameDataSet(dataSet, targetTableForMeasure(measure, args.TableMappings))
		if err != nil {
			infConn.Close()
			pgConn.Close()
			return err
		}

		tsManager := app.schemaManagerService.TimeScale(pgConn, args.OutputSchema, args.ChunkTimeInterval)
		finder, ok := tsManager.(existingNonNullableColumnFinder)
		if !ok {
			infConn.Close()
			pgConn.Close()
			return fmt.Errorf("timescale schema manager does not support non-null validation")
		}
		columns, err := finder.ExistingNonNullableColumns(targetDataSet, args.OutputSchemaStrategy)
		if err == nil && len(columns) > 0 {
			windows, windowErr := buildPreflightWindows(app, infConn, connArgs.InputDb, args, maxWindow)
			if windowErr != nil {
				err = windowErr
			} else if len(windows) == 0 {
				err = session.Validate(columns)
			} else {
				err = validateWindows(app, infConn, measure, connArgs.InputDb, args, columns, windows, pause)
			}
		}
		infConn.Close()
		pgConn.Close()
		if err != nil {
			return fmt.Errorf("could not validate non-null source data for measurement '%s'\n%v", measure, err)
		}
	}
	return nil
}

type preflightWindow struct {
	From time.Time
	To   time.Time
}

func parseOptionalDuration(raw string) (time.Duration, error) {
	if raw == "" {
		return 0, nil
	}
	return time.ParseDuration(raw)
}

func buildPreflightWindows(app *appContext, infConn influx.Client, db string, args *cli.MigrationConfig, maxWindow time.Duration) ([]preflightWindow, error) {
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

	windows := []preflightWindow{}
	for _, group := range groups {
		from := group.Start
		to := group.End
		if requestedFrom != nil && requestedFrom.After(from) {
			from = *requestedFrom
		}
		if requestedTo != nil && requestedTo.Before(to) {
			to = *requestedTo
		}
		if !from.Before(to) {
			continue
		}
		windows = append(windows, splitPreflightWindow(preflightWindow{From: from, To: to}, maxWindow)...)
	}
	sort.Slice(windows, func(i, j int) bool { return windows[i].From.Before(windows[j].From) })
	return windows, nil
}

func splitPreflightWindow(window preflightWindow, maxWindow time.Duration) []preflightWindow {
	if maxWindow <= 0 || window.To.Sub(window.From) <= maxWindow {
		return []preflightWindow{window}
	}
	windows := []preflightWindow{}
	for start := window.From; start.Before(window.To); {
		end := start.Add(maxWindow)
		if end.After(window.To) {
			end = window.To
		}
		windows = append(windows, preflightWindow{From: start, To: end})
		start = end
	}
	return windows
}

func validateWindows(app *appContext, infConn influx.Client, measure, inputDb string, args *cli.MigrationConfig, columns []string, windows []preflightWindow, pause time.Duration) error {
	for index, window := range windows {
		windowArgs := *args
		windowArgs.From = window.From.Format(time.RFC3339)
		windowArgs.To = window.To.Add(-time.Nanosecond).Format(time.RFC3339Nano)
		windowArgs.Limit = 0
		session, err := app.notNullValidator.Prepare(infConn, measure, inputDb, &windowArgs)
		if err != nil {
			return err
		}
		if err := session.Validate(columns); err != nil {
			return err
		}
		if pause > 0 && index < len(windows)-1 {
			time.Sleep(pause)
		}
	}
	return nil
}

func renameDataSet(dataSet *idrf.DataSet, targetTable string) (*idrf.DataSet, error) {
	if targetTable == "" || targetTable == dataSet.DataSetName {
		return dataSet, nil
	}
	return idrf.NewDataSet(targetTable, dataSet.Columns, dataSet.TimeColumn)
}
