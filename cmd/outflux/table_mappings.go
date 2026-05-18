package main

import "fmt"

func targetTableForMeasure(measure string, mappings map[string]string) string {
	if target, exists := mappings[measure]; exists {
		return target
	}
	return measure
}

func validateTableMappings(measures []string, mappings map[string]string) error {
	selectedMeasures := make(map[string]bool, len(measures))
	outputTables := make(map[string]string, len(measures))

	for _, measure := range measures {
		selectedMeasures[measure] = true
		targetTable := targetTableForMeasure(measure, mappings)
		if existingMeasure, exists := outputTables[targetTable]; exists {
			return fmt.Errorf("measurements '%s' and '%s' both target output table '%s'", existingMeasure, measure, targetTable)
		}
		outputTables[targetTable] = measure
	}

	for measure := range mappings {
		if !selectedMeasures[measure] {
			return fmt.Errorf("table mapping provided for measurement '%s', but it is not selected for this run", measure)
		}
	}

	return nil
}
