package main

import (
	"testing"
	"time"

	influx "github.com/influxdata/influxdb/client/v2"
	"github.com/timescale/outflux/internal/cli"
	"github.com/timescale/outflux/internal/schemamanagement/influx/discovery"
)

func TestSplitMigrationWindow(t *testing.T) {
	start := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	windows := splitMigrationWindow(migrationWindow{From: start, To: start.Add(48 * time.Hour)}, 24*time.Hour)
	if len(windows) != 2 {
		t.Fatalf("expected 2 windows, got %d", len(windows))
	}
	if !windows[1].From.Equal(start.Add(24 * time.Hour)) {
		t.Fatalf("unexpected second window start: %s", windows[1].From)
	}
}

func TestApplyWindowUsesNonOverlappingInclusiveBounds(t *testing.T) {
	start := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	args := applyWindow(nilConfig(), migrationWindow{From: start, To: start.Add(time.Hour)})
	if args.From != "2026-05-01T00:00:00Z" {
		t.Fatalf("unexpected from: %s", args.From)
	}
	if args.To != "2026-05-01T00:59:59.999999999Z" {
		t.Fatalf("unexpected to: %s", args.To)
	}
}

func nilConfig() *cli.MigrationConfig { return &cli.MigrationConfig{} }

func TestBuildMigrationWindowsIntersectsAndOrdersShardGroups(t *testing.T) {
	app := &appContext{influxShardExplorer: &windowShardExplorer{groups: []discovery.ShardGroup{
		{Start: time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)},
		{Start: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC)},
	}}}
	windows, err := buildMigrationWindows(app, nil, "db", &cli.MigrationConfig{
		From:      "2026-05-03T00:00:00Z",
		To:        "2026-05-10T00:00:00Z",
		MaxWindow: "48h",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(windows) != 5 {
		t.Fatalf("expected 5 windows, got %d", len(windows))
	}
	if !windows[0].From.Equal(time.Date(2026, 5, 3, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected first window: %+v", windows[0])
	}
	if !windows[len(windows)-1].To.Equal(time.Date(2026, 5, 10, 0, 0, 0, 1, time.UTC)) {
		t.Fatalf("unexpected final bound: %+v", windows[len(windows)-1])
	}
}

type windowShardExplorer struct {
	groups []discovery.ShardGroup
}

func (w *windowShardExplorer) FetchShardGroups(influx.Client, string, string) ([]discovery.ShardGroup, error) {
	return w.groups, nil
}
