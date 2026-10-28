package frontend

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/google/uuid"
	"github.com/grafana/tempo/v3/modules/frontend/pipeline"
	"github.com/grafana/tempo/v3/pkg/tempopb"
	"github.com/grafana/tempo/v3/tempodb/backend"
	"github.com/stretchr/testify/require"
)

func TestBuildBackendRequestsRejectsInvalidBlockRange(t *testing.T) {
	now := time.Unix(1700000000, 0)
	tests := []struct {
		name       string
		start, end time.Time
		step       time.Duration
		want       int
	}{
		{"reversed bounds", now.Add(time.Hour), now, time.Minute, 0},
		{"reversed within coarse bucket", now.Add(time.Second), now, time.Hour, 0},
		{"valid bounds", now, now.Add(time.Hour), time.Minute, 1},
		{"equal metadata remains searchable", now, now, time.Minute, 1},
		{"no overlap", now.Add(-2 * time.Hour), now.Add(-time.Hour), time.Minute, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sharder := &queryRangeSharder{logger: log.NewNopLogger()}
			parent := pipeline.NewHTTPRequest(httptest.NewRequest("GET", "/api/metrics/query_range", nil))
			query := tempopb.QueryRangeRequest{
				Query: "{} | count_over_time()",
				Start: uint64(now.Add(-time.Minute).UnixNano()), End: uint64(now.Add(2 * time.Hour).UnixNano()), Step: uint64(tt.step),
			}
			meta := &backend.BlockMeta{
				BlockID: backend.MustParse(uuid.NewString()), Version: "vParquet4",
				StartTime: tt.start, EndTime: tt.end, TotalRecords: 1, Size_: 1000,
			}
			iter := backendJobsFunc([]*backend.BlockMeta{meta}, 100000, 1, uint32(query.End/uint64(time.Second)))
			out := make(chan pipeline.Request, 2)
			sharder.buildBackendRequests(t.Context(), "test-tenant", parent, query, 0, iter, out, func(*backend.BlockMeta) uint32 { return 0 })
			count := 0
			for range out {
				count++
			}
			require.Equal(t, tt.want, count)
		})
	}
}
