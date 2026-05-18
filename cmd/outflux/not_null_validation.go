package main

import (
	"fmt"
	"time"

	influx "github.com/influxdata/influxdb/client/v2"
	"github.com/timescale/outflux/internal/cli"
	"github.com/timescale/outflux/internal/idrf"
	"github.com/timescale/outflux/internal/schemamanagement/schemaconfig"
)

type existingNonNullableColumnFinder interface {
	ExistingNonNullableColumns(*idrf.DataSet, schemaconfig.SchemaStrategy) ([]string, error)
}

func validateNotNullSourceData(app *appContext, connArgs *cli.ConnectionConfig, args *cli.MigrationConfig, sharedWindows []migrationWindow) error {
	pause, err := parseOptionalDuration(args.WindowPause)
	if err != nil {
		return fmt.Errorf("invalid window pause: %v", err)
	}

	for _, measure := range connArgs.InputMeasures {
		infConn, pgConn, err := openConnections(app, connArgs)
		if err != nil {
			return fmt.Errorf("could not open connections for non-null validation\n%v", err)
		}

		session, err := app.notNullValidator.Prepare(infConn, measure, connArgs.InputDb, args)
		if err != nil {
			infConn.Close()
			pgConn.Close()
			return fmt.Errorf("could not prepare non-null validation for measurement '%s'\n%v", measure, err)
		}

		dataSet := session.DataSet()
		targetDataSet, err := renameDataSet(dataSet, targetTableForMeasure(measure, args.TableMappings))
		if err != nil {
			infConn.Close()
			pgConn.Close()
			return err
		}

		tsManager := app.schemaManagerService.TimeScale(pgConn, args.OutputSchema, args.ChunkTimeInterval)
		finder, ok := tsManager.(existingNonNullableColumnFinder)
		if !ok {
			infConn.Close()
			pgConn.Close()
			return fmt.Errorf("timescale schema manager does not support non-null validation")
		}
		columns, err := finder.ExistingNonNullableColumns(targetDataSet, args.OutputSchemaStrategy)
		if err == nil && len(columns) > 0 {
			windows := sharedWindows
			var windowErr error
			if windows == nil {
				windows, windowErr = buildMigrationWindows(app, infConn, connArgs.InputDb, args)
			}
			if windowErr != nil {
				err = windowErr
			} else if len(windows) == 0 {
				err = session.Validate(columns)
			} else {
				err = validateWindows(app, infConn, measure, connArgs.InputDb, args, columns, windows, pause)
			}
		}
		infConn.Close()
		pgConn.Close()
		if err != nil {
			return fmt.Errorf("could not validate non-null source data for measurement '%s'\n%v", measure, err)
		}
	}
	return nil
}

func validateWindows(app *appContext, infConn influx.Client, measure, inputDb string, args *cli.MigrationConfig, columns []string, windows []migrationWindow, pause time.Duration) error {
	for index, window := range windows {
		session, err := app.notNullValidator.Prepare(infConn, measure, inputDb, applyWindow(args, window))
		if err != nil {
			return err
		}
		if err := session.Validate(columns); err != nil {
			return err
		}
		if pause > 0 && index < len(windows)-1 {
			time.Sleep(pause)
		}
	}
	return nil
}

func renameDataSet(dataSet *idrf.DataSet, targetTable string) (*idrf.DataSet, error) {
	if targetTable == "" || targetTable == dataSet.DataSetName {
		return dataSet, nil
	}
	return idrf.NewDataSet(targetTable, dataSet.Columns, dataSet.TimeColumn)
}
