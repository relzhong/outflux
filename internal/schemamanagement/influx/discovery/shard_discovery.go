package discovery

import (
	"fmt"
	"time"

	influx "github.com/influxdata/influxdb/client/v2"
	"github.com/timescale/outflux/internal/schemamanagement/influx/influxqueries"
)

const showShardsQuery = "SHOW SHARDS"

type ShardGroup struct {
	Start time.Time
	End   time.Time
}

type ShardExplorer interface {
	FetchShardGroups(influxClient influx.Client, database, retentionPolicy string) ([]ShardGroup, error)
}

type defaultShardExplorer struct {
	queryService influxqueries.InfluxQueryService
}

func NewShardExplorer(queryService influxqueries.InfluxQueryService) ShardExplorer {
	return &defaultShardExplorer{queryService: queryService}
}

func (e *defaultShardExplorer) FetchShardGroups(influxClient influx.Client, database, retentionPolicy string) ([]ShardGroup, error) {
	results, err := e.queryService.ExecuteQuery(influxClient, database, showShardsQuery)
	if err != nil {
		return nil, err
	}

	groups := []ShardGroup{}
	for _, result := range results {
		for _, series := range result.Series {
			if series.Name != database {
				continue
			}
			indexes := make(map[string]int)
			for index, column := range series.Columns {
				indexes[column] = index
			}
			for _, row := range series.Values {
				if retentionPolicy != "" && fmt.Sprint(row[indexes["retention_policy"]]) != retentionPolicy {
					continue
				}
				start, err := time.Parse(time.RFC3339, fmt.Sprint(row[indexes["start_time"]]))
				if err != nil {
					return nil, err
				}
				end, err := time.Parse(time.RFC3339, fmt.Sprint(row[indexes["end_time"]]))
				if err != nil {
					return nil, err
				}
				groups = append(groups, ShardGroup{Start: start, End: end})
			}
		}
	}
	return groups, nil
}
