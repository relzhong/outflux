package ts

import (
	"testing"

	"github.com/timescale/outflux/internal/idrf"
	"github.com/timescale/outflux/internal/ingestion/config"
	"github.com/timescale/outflux/internal/schemamanagement/schemaconfig"
)

func TestPrepareUsesTargetTable(t *testing.T) {
	timeColumn, _ := idrf.NewColumn("time", idrf.IDRFTimestamptz)
	dataSet, _ := idrf.NewDataSet("cpu", []*idrf.Column{timeColumn}, "time")
	schemaManager := &recordingSchemaManager{}
	ingestor := &TSIngestor{
		Config:        &config.IngestorConfig{TargetTable: "computer_cpu"},
		SchemaManager: schemaManager,
	}

	err := ingestor.Prepare(&idrf.Bundle{DataDef: dataSet, DataChan: make(chan idrf.Row)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schemaManager.preparedDataSet.DataSetName != "computer_cpu" {
		t.Fatalf("expected target table name, got %s", schemaManager.preparedDataSet.DataSetName)
	}
	if ingestor.cachedBundle.DataDef.DataSetName != "computer_cpu" {
		t.Fatalf("expected cached target table name, got %s", ingestor.cachedBundle.DataDef.DataSetName)
	}
	if dataSet.DataSetName != "cpu" {
		t.Fatalf("expected source dataset name to remain unchanged, got %s", dataSet.DataSetName)
	}
}

type recordingSchemaManager struct {
	preparedDataSet *idrf.DataSet
}

func (r *recordingSchemaManager) DiscoverDataSets() ([]string, error) { return nil, nil }
func (r *recordingSchemaManager) FetchDataSet(string) (*idrf.DataSet, error) {
	return nil, nil
}
func (r *recordingSchemaManager) PrepareDataSet(dataSet *idrf.DataSet, _ schemaconfig.SchemaStrategy) error {
	r.preparedDataSet = dataSet
	return nil
}
