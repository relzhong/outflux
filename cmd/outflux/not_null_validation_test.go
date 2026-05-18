package main

import (
	"testing"
	"time"
)

func TestSplitPreflightWindow(t *testing.T) {
	start := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	windows := splitPreflightWindow(preflightWindow{From: start, To: start.Add(48 * time.Hour)}, 24*time.Hour)
	if len(windows) != 2 {
		t.Fatalf("expected 2 windows, got %d", len(windows))
	}
	if !windows[1].From.Equal(start.Add(24 * time.Hour)) {
		t.Fatalf("unexpected second window start: %s", windows[1].From)
	}
}
