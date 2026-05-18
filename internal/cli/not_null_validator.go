package cli

import (
	"fmt"
	"sync"

	influx "github.com/influxdata/influxdb/client/v2"
	"github.com/timescale/outflux/internal/extraction"
	"github.com/timescale/outflux/internal/idrf"
	"github.com/timescale/outflux/internal/transformation"
)

type NotNullValidationSession interface {
	DataSet() *idrf.DataSet
	Validate(columns []string) error
}

type NotNullValidator interface {
	Prepare(infConn influx.Client, measure, inputDb string, conf *MigrationConfig) (NotNullValidationSession, error)
}

type notNullValidator struct {
	extractorService   extraction.ExtractorService
	transformerService TransformerService
	extractionConfCreator
}

func NewNotNullValidator(extractorService extraction.ExtractorService, transformerService TransformerService) NotNullValidator {
	return &notNullValidator{
		extractorService:      extractorService,
		transformerService:    transformerService,
		extractionConfCreator: &defaultExtractionConfCreator{},
	}
}

func (v *notNullValidator) Prepare(infConn influx.Client, measure, inputDb string, conf *MigrationConfig) (NotNullValidationSession, error) {
	pipeID := fmt.Sprintf("preflight_%s", measure)
	extractionConf := v.extractionConfCreator.create(pipeID, inputDb, measure, conf)
	extractor, err := v.extractorService.InfluxExtractor(infConn, extractionConf)
	if err != nil {
		return nil, err
	}

	transformers, err := createTransformers(v.transformerService, pipeID, infConn, measure, inputDb, conf)
	if err != nil {
		return nil, err
	}

	bundle, err := extractor.Prepare()
	if err != nil {
		return nil, err
	}
	for _, transformer := range transformers {
		bundle, err = transformer.Prepare(bundle)
		if err != nil {
			return nil, err
		}
	}

	return &notNullValidationSession{
		extractor:    extractor,
		transformers: transformers,
		bundle:       bundle,
	}, nil
}

type notNullValidationSession struct {
	extractor    extraction.Extractor
	transformers []transformation.Transformer
	bundle       *idrf.Bundle
}

func (s *notNullValidationSession) DataSet() *idrf.DataSet {
	return s.bundle.DataDef
}

func (s *notNullValidationSession) Validate(columns []string) error {
	requiredColumnIndexes := make(map[int]string, len(columns))
	for _, column := range columns {
		found := false
		for i, candidate := range s.bundle.DataDef.Columns {
			if candidate.Name == column {
				requiredColumnIndexes[i] = column
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("required output column %s not found during non-null validation", column)
		}
	}

	externalErrors := make(chan error)
	componentErrors := make(chan error, 1+len(s.transformers))
	var wg sync.WaitGroup
	wg.Add(1 + len(s.transformers))
	go func() {
		defer wg.Done()
		componentErrors <- s.extractor.Start(externalErrors)
	}()
	for _, transformer := range s.transformers {
		transformer := transformer
		go func() {
			defer wg.Done()
			componentErrors <- transformer.Start(externalErrors)
		}()
	}

	var validationErr error
	for row := range s.bundle.DataChan {
		for index, column := range requiredColumnIndexes {
			if row[index] == nil {
				if validationErr == nil {
					validationErr = fmt.Errorf("source data contains null for target NOT NULL column %s", column)
				}
				break
			}
		}
	}

	wg.Wait()
	close(componentErrors)
	for err := range componentErrors {
		if err != nil && validationErr == nil {
			validationErr = err
		}
	}
	return validationErr
}
