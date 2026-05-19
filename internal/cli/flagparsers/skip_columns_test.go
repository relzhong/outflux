package flagparsers

import (
	"testing"

	"github.com/spf13/pflag"
)

func TestParseSkipColumns(t *testing.T) {
	testCases := []struct {
		name        string
		values      []string
		expectError bool
	}{
		{name: "empty"},
		{name: "valid repeated values", values: []string{"field1", "tag1"}},
		{name: "empty value", values: []string{" "}, expectError: true},
		{name: "time is rejected", values: []string{"time"}, expectError: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
			flags.StringArray(SkipColumnFlag, nil, "")
			for _, value := range testCase.values {
				if err := flags.Set(SkipColumnFlag, value); err != nil {
					t.Fatalf("could not set flag: %v", err)
				}
			}

			columns, err := parseSkipColumns(flags)
			if testCase.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(columns) != len(testCase.values) {
				t.Fatalf("expected %d columns, got %d", len(testCase.values), len(columns))
			}
		})
	}
}
