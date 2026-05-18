package discovery

import (
	"testing"

	influx "github.com/influxdata/influxdb/client/v2"
	"github.com/influxdata/influxdb/models"
	"github.com/timescale/outflux/internal/schemamanagement/influx/influxqueries"
)

func TestFetchShardGroups(t *testing.T) {
	explorer := NewShardExplorer(&shardQueryService{
		results: []influx.Result{{
			Series: []models.Row{{
				Name:    "db",
				Columns: []string{"id", "database", "retention_policy", "start_time", "end_time"},
				Values: [][]interface{}{
					{"1", "db", "autogen", "2026-05-01T00:00:00Z", "2026-05-08T00:00:00Z"},
					{"2", "db", "other", "2026-05-08T00:00:00Z", "2026-05-15T00:00:00Z"},
				},
			}},
		}},
	})
	groups, err := explorer.FetchShardGroups(nil, "db", "autogen")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected one shard group, got %d", len(groups))
	}
}

type shardQueryService struct {
	results []influx.Result
}

func (s *shardQueryService) ExecuteQuery(influx.Client, string, string) ([]influx.Result, error) {
	return s.results, nil
}
func (s *shardQueryService) ExecuteShowQuery(influx.Client, string, string) (*influxqueries.InfluxShowResult, error) {
	return nil, nil
}
