package ts

import (
	"testing"

	"github.com/timescale/outflux/internal/idrf"
)

func TestExistingTableCompatible(t *testing.T) {
	testCases := []struct {
		existingColumns []*columnDesc
		reqColumns      []*idrf.Column
		timeCol         string
		desc            string
		errorExpected   bool
	}{
		{
			existingColumns: []*columnDesc{},
			reqColumns:      []*idrf.Column{{Name: "a"}},
			desc:            "required column not found in existing table",
			errorExpected:   true,
		}, {
			existingColumns: []*columnDesc{{columnName: "a", dataType: "text"}},
			reqColumns:      []*idrf.Column{{Name: "a", DataType: idrf.IDRFBoolean}},
			desc:            "required data type is incompatible with existing column type",
			errorExpected:   true,
		}, {
			existingColumns: []*columnDesc{
				{columnName: "a", dataType: "text"},
				{columnName: "b", dataType: "text", isNullable: "NO"}},
			reqColumns: []*idrf.Column{
				{Name: "a", DataType: idrf.IDRFString},
				{Name: "b", DataType: idrf.IDRFString}},
			timeCol:       "a",
			desc:          "only time column should be not-nullable",
			errorExpected: true,
		}, {
			existingColumns: []*columnDesc{
				{columnName: "a", dataType: "text"},
				{columnName: "b", dataType: "text", isNullable: "YES"}},
			reqColumns: []*idrf.Column{
				{Name: "a", DataType: idrf.IDRFString},
				{Name: "b", DataType: idrf.IDRFString}},
			timeCol:       "a",
			desc:          "all is good",
			errorExpected: false,
		},
	}

	for _, testCase := range testCases {
		err := isExistingTableCompatible(testCase.existingColumns, testCase.reqColumns, testCase.timeCol)
		if testCase.errorExpected && err == nil {
			t.Errorf("Tested: %s.\nExpected an error. None returned", testCase.desc)
		}

		if !testCase.errorExpected && err != nil {
			t.Errorf("Tested: %s.\nError wasn't expected, got:\n%v", testCase.desc, err)
		}
	}
}

func TestExistingTableCompatibilityCanCollectNotNullColumns(t *testing.T) {
	columns, err := validateExistingTableCompatibility(
		[]*columnDesc{
			{columnName: "time", dataType: "timestamp with time zone", isNullable: "NO"},
			{columnName: "value", dataType: "text", isNullable: "NO"},
		},
		[]*idrf.Column{
			{Name: "time", DataType: idrf.IDRFTimestamptz},
			{Name: "value", DataType: idrf.IDRFString},
		},
		"time",
		true,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(columns) != 1 || columns[0] != "value" {
		t.Fatalf("expected value column, got %v", columns)
	}
}
