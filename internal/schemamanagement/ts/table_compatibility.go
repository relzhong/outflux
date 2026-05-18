package ts

import (
	"fmt"

	"github.com/timescale/outflux/internal/idrf"
)

func isExistingTableCompatible(existingColumns []*columnDesc, requiredColumns []*idrf.Column, timeCol string) error {
	_, err := validateExistingTableCompatibility(existingColumns, requiredColumns, timeCol, false)
	return err
}

func validateExistingTableCompatibility(existingColumns []*columnDesc, requiredColumns []*idrf.Column, timeCol string, allowNonTimeNotNull bool) ([]string, error) {
	columnsByName := make(map[string]*columnDesc)
	for _, column := range existingColumns {
		columnsByName[column.columnName] = column
	}

	nonTimeNotNullColumns := []string{}
	for _, reqColumn := range requiredColumns {
		colName := reqColumn.Name
		var existingCol *columnDesc
		var ok bool
		if existingCol, ok = columnsByName[colName]; !ok {
			return nil, fmt.Errorf("Required column %s not found in existing table", colName)
		}

		existingType := pgTypeToIdrf(existingCol.dataType)
		if !existingType.CanFitInto(reqColumn.DataType) {
			return nil, fmt.Errorf(
				"Required column %s of type %s is not compatible with existing type %s",
				colName, reqColumn.DataType, existingType)
		}

		// Only time column is allowed to have a NOT NULL constraint
		if !existingCol.isColumnNullable() && existingCol.columnName != timeCol {
			if !allowNonTimeNotNull {
				return nil, fmt.Errorf("Existing column %s is not nullable. Can't guarantee data transfer", existingCol.columnName)
			}
			nonTimeNotNullColumns = append(nonTimeNotNullColumns, existingCol.columnName)
		}
	}

	return nonTimeNotNullColumns, nil
}
