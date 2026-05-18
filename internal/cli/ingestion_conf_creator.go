package cli

import (
	"fmt"

	"github.com/timescale/outflux/internal/ingestion/config"
)

const (
	ingestorIDTemplate = "%s_ing"
)

type ingestionConfCreator interface {
	create(pipeID, targetTable string, conf *MigrationConfig) *config.IngestorConfig
}

type defaultIngestionConfCreator struct {
}

func (s *defaultIngestionConfCreator) create(pipeID, targetTable string, conf *MigrationConfig) *config.IngestorConfig {
	return &config.IngestorConfig{
		IngestorID:              fmt.Sprintf(ingestorIDTemplate, pipeID),
		BatchSize:               conf.BatchSize,
		RollbackOnExternalError: conf.RollbackAllMeasureExtractionsOnError,
		CommitStrategy:          conf.CommitStrategy,
		SchemaStrategy:          conf.OutputSchemaStrategy,
		Schema:                  conf.OutputSchema,
		ChunkTimeInterval:       conf.ChunkTimeInterval,
		TargetTable:             targetTable,
		ValidatedNotNullColumns: conf.ValidateNotNullSourceData,
	}
}
