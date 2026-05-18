package cli

import (
	"testing"

	"github.com/timescale/outflux/internal/idrf"
)

func TestNotNullValidationSession(t *testing.T) {
	timeColumn, _ := idrf.NewColumn("time", idrf.IDRFTimestamptz)
	valueColumn, _ := idrf.NewColumn("value", idrf.IDRFString)
	dataSet, _ := idrf.NewDataSet("cpu", []*idrf.Column{timeColumn, valueColumn}, "time")

	testCases := []struct {
		name        string
		rows        []idrf.Row
		expectError bool
	}{
		{name: "all values present", rows: []idrf.Row{{"time", "ok"}}},
		{name: "empty slice"},
		{name: "null value", rows: []idrf.Row{{"time", nil}}, expectError: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			bundle := &idrf.Bundle{DataDef: dataSet, DataChan: make(chan idrf.Row, len(testCase.rows))}
			session := &notNullValidationSession{
				extractor: &validationExtractor{
					bundle: bundle,
					rows:   testCase.rows,
				},
				bundle: bundle,
			}
			err := session.Validate([]string{"value"})
			if testCase.expectError && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !testCase.expectError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

type validationExtractor struct {
	bundle *idrf.Bundle
	rows   []idrf.Row
}

func (v *validationExtractor) ID() string { return "validation" }
func (v *validationExtractor) Prepare() (*idrf.Bundle, error) {
	return v.bundle, nil
}
func (v *validationExtractor) Start(chan error) error {
	defer close(v.bundle.DataChan)
	for _, row := range v.rows {
		v.bundle.DataChan <- row
	}
	return nil
}
