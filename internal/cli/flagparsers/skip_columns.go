package flagparsers

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
)

func parseSkipColumns(flags *pflag.FlagSet) ([]string, error) {
	skipColumns, err := flags.GetStringArray(SkipColumnFlag)
	if err != nil {
		return nil, err
	}
	for i, column := range skipColumns {
		column = strings.TrimSpace(column)
		if column == "" {
			return nil, fmt.Errorf("value for '%s' can't be empty", SkipColumnFlag)
		}
		if column == "time" {
			return nil, fmt.Errorf("column 'time' cannot be skipped")
		}
		skipColumns[i] = column
	}
	return skipColumns, nil
}
