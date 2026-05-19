package influx

import (
	"fmt"
	"log"

	influx "github.com/influxdata/influxdb/client/v2"
	"github.com/timescale/outflux/internal/extraction/config"
	"github.com/timescale/outflux/internal/extraction/influx/idrfconversion"
	"github.com/timescale/outflux/internal/idrf"
	"github.com/timescale/outflux/internal/schemamanagement"
)

// Extractor is an implementation of the extraction.Extractor interface for
// pulling data out of InfluxDB
type Extractor struct {
	Config            *config.ExtractionConfig
	SM                schemamanagement.SchemaManager
	cachedElementData *idrf.Bundle
	DataProducer      DataProducer
}

// ID of the extractor, useful for logging and error reporting
func (e *Extractor) ID() string {
	return e.Config.ExtractorID
}

// Prepare discovers the data set schema for the measure in the config
func (e *Extractor) Prepare() (*idrf.Bundle, error) {
	measureName := e.Config.MeasureExtraction.Measure
	log.Printf("Discovering influx schema for measurement: %s", measureName)

	discoveredDataSet, err := e.SM.FetchDataSet(measureName)
	if err != nil {
		return nil, fmt.Errorf("%s: could not fetch data set definition for measure: %s\n%v", e.ID(), measureName, err)
	}

	log.Printf("Discovered: %s", discoveredDataSet.String())
	discoveredDataSet, err = filterSkippedColumns(discoveredDataSet, e.Config.MeasureExtraction.SkipColumns)
	if err != nil {
		return nil, err
	}
	e.cachedElementData = &idrf.Bundle{
		DataDef:  discoveredDataSet,
		DataChan: make(chan idrf.Row, e.Config.DataBufferSize),
	}

	return e.cachedElementData, nil
}

func filterSkippedColumns(dataSet *idrf.DataSet, skipColumns []string) (*idrf.DataSet, error) {
	if len(skipColumns) == 0 {
		return dataSet, nil
	}
	skip := make(map[string]bool, len(skipColumns))
	for _, column := range skipColumns {
		skip[column] = true
	}
	columns := make([]*idrf.Column, 0, len(dataSet.Columns))
	for _, column := range dataSet.Columns {
		if column.Name == dataSet.TimeColumn || !skip[column.Name] {
			columns = append(columns, column)
		}
	}
	if len(columns) <= 1 {
		return nil, fmt.Errorf("all non-time columns skipped for measure '%s'", dataSet.DataSetName)
	}
	return idrf.NewDataSet(dataSet.DataSetName, columns, dataSet.TimeColumn)
}

// Start pulls the data from an InfluxDB measure and feeds it to a data channel
// Peridicly (between chunks) checks for external errors and quits if it detects them
func (e *Extractor) Start(errChan chan error) error {
	if e.cachedElementData == nil {
		return fmt.Errorf("%s: Prepare not called before start", e.ID())
	}

	id := e.Config.ExtractorID
	dataDef := e.cachedElementData.DataDef
	measureConf := e.Config.MeasureExtraction

	log.Printf("Starting extractor '%s' for measure: %s\n", id, dataDef.DataSetName)
	intChunkSize := int(measureConf.ChunkSize)

	query := &influx.Query{
		Command:         buildSelectCommand(measureConf, dataDef.Columns),
		Database:        measureConf.Database,
		RetentionPolicy: measureConf.RetentionPolicy,
		Chunked:         true,
		ChunkSize:       intChunkSize,
	}

	log.Printf("%s: Extracting data from database '%s'\n", id, query.Database)
	log.Printf("%s: %s\n", id, query.Command)
	log.Printf("%s:Pulling chunks with size %d\n", id, intChunkSize)

	idrfConverter := idrfconversion.NewIdrfConverter(dataDef)
	producerArgs := &producerArgs{
		dataChannel: e.cachedElementData.DataChan,
		errChannel:  errChan,
		query:       query,
		converter:   idrfConverter,
	}

	return e.DataProducer.Fetch(producerArgs)
}
