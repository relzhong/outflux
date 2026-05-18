package main

import "testing"

func TestTargetTableForMeasure(t *testing.T) {
	mappings := map[string]string{"cpu": "computer_cpu"}
	if got := targetTableForMeasure("cpu", mappings); got != "computer_cpu" {
		t.Fatalf("expected mapped table, got %s", got)
	}
	if got := targetTableForMeasure("mem", mappings); got != "mem" {
		t.Fatalf("expected original table name, got %s", got)
	}
}

func TestValidateTableMappings(t *testing.T) {
	testCases := []struct {
		name        string
		measures    []string
		mappings    map[string]string
		expectError bool
	}{
		{name: "mapped and unmapped measurements", measures: []string{"cpu", "mem"}, mappings: map[string]string{"cpu": "computer_cpu"}},
		{name: "unused mapping", measures: []string{"mem"}, mappings: map[string]string{"cpu": "computer_cpu"}, expectError: true},
		{name: "mapped target collides with unmapped measurement", measures: []string{"cpu", "mem"}, mappings: map[string]string{"cpu": "mem"}, expectError: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := validateTableMappings(testCase.measures, testCase.mappings)
			if testCase.expectError && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !testCase.expectError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
