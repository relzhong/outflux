package flagparsers

import (
	"testing"

	"github.com/spf13/pflag"
)

func TestParseTableMappings(t *testing.T) {
	testCases := []struct {
		name        string
		values      []string
		expectError bool
	}{
		{name: "empty"},
		{name: "valid repeated values", values: []string{"cpu=computer_cpu", "mem=computer_mem"}},
		{name: "missing separator", values: []string{"cpu"}, expectError: true},
		{name: "empty source", values: []string{"=computer_cpu"}, expectError: true},
		{name: "empty target", values: []string{"cpu="}, expectError: true},
		{name: "schema qualified target", values: []string{"cpu=analytics.computer_cpu"}, expectError: true},
		{name: "duplicate source", values: []string{"cpu=computer_cpu", "cpu=other_cpu"}, expectError: true},
		{name: "duplicate target", values: []string{"cpu=metrics", "mem=metrics"}, expectError: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
			flags.StringArray(TableMapFlag, nil, "")
			for _, value := range testCase.values {
				if err := flags.Set(TableMapFlag, value); err != nil {
					t.Fatalf("could not set flag: %v", err)
				}
			}

			mappings, err := parseTableMappings(flags)
			if testCase.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(mappings) != len(testCase.values) {
				t.Fatalf("expected %d mappings, got %d", len(testCase.values), len(mappings))
			}
		})
	}
}
