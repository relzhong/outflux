package flagparsers

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
)

func parseTableMappings(flags *pflag.FlagSet) (map[string]string, error) {
	rawMappings, err := flags.GetStringArray(TableMapFlag)
	if err != nil {
		return nil, err
	}

	mappings := make(map[string]string, len(rawMappings))
	targets := make(map[string]string, len(rawMappings))
	for _, rawMapping := range rawMappings {
		parts := strings.SplitN(rawMapping, "=", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, fmt.Errorf("invalid value for '%s': '%s'. Expected source=target", TableMapFlag, rawMapping)
		}

		source := parts[0]
		target := parts[1]
		if strings.Contains(target, ".") {
			return nil, fmt.Errorf("invalid target table in '%s': '%s'. Use '%s' for schema selection", TableMapFlag, target, OutputSchemaFlag)
		}
		if _, exists := mappings[source]; exists {
			return nil, fmt.Errorf("duplicate source measurement in '%s': '%s'", TableMapFlag, source)
		}
		if existingSource, exists := targets[target]; exists {
			return nil, fmt.Errorf("target table '%s' is mapped from both '%s' and '%s'", target, existingSource, source)
		}

		mappings[source] = target
		targets[target] = source
	}

	return mappings, nil
}
