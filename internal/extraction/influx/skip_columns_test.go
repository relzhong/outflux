package influx

import (
	"testing"

	"github.com/timescale/outflux/internal/idrf"
)

func TestFilterSkippedColumns(t *testing.T) {
	timeColumn, _ := idrf.NewColumn("time", idrf.IDRFTimestamptz)
	tagColumn, _ := idrf.NewColumn("tag1", idrf.IDRFString)
	fieldColumn, _ := idrf.NewColumn("field1", idrf.IDRFString)
	dataSet, _ := idrf.NewDataSet("cpu", []*idrf.Column{timeColumn, tagColumn, fieldColumn}, "time")

	filtered, err := filterSkippedColumns(dataSet, []string{"tag1", "missing"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filtered.ColumnNamed("tag1") != nil {
		t.Fatal("expected skipped tag to be removed")
	}
	if filtered.ColumnNamed("field1") == nil {
		t.Fatal("expected unskipped field to remain")
	}
}

func TestFilterSkippedColumnsRejectsTimeOnlyDataSet(t *testing.T) {
	timeColumn, _ := idrf.NewColumn("time", idrf.IDRFTimestamptz)
	fieldColumn, _ := idrf.NewColumn("field1", idrf.IDRFString)
	dataSet, _ := idrf.NewDataSet("cpu", []*idrf.Column{timeColumn, fieldColumn}, "time")

	_, err := filterSkippedColumns(dataSet, []string{"field1"})
	if err == nil {
		t.Fatal("expected error when all non-time columns are skipped")
	}
}
