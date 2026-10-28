package livestore

import (
	"testing"
	"time"

	"github.com/grafana/dskit/services"
	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kgo"
)

func TestLiveStoreBlockTimeBoundsOutsideSlack(t *testing.T) {
	now := time.Unix(1700000000, 0)
	tests := []struct {
		name                           string
		slack                          time.Duration
		start, end, wantStart, wantEnd time.Time
	}{
		{"within window", 2 * time.Minute, now.Add(-time.Minute), now, now.Add(-time.Minute), now},
		{"entirely old", 2 * time.Minute, now.Add(-2 * time.Hour), now.Add(-time.Hour), now.Add(-2 * time.Minute), now.Add(-2 * time.Minute)},
		{"entirely future", 2 * time.Minute, now.Add(time.Hour), now.Add(2 * time.Hour), now.Add(2 * time.Minute), now.Add(2 * time.Minute)},
		{"crossing window", 2 * time.Minute, now.Add(-5 * time.Minute), now.Add(5 * time.Minute), now.Add(-2 * time.Minute), now.Add(2 * time.Minute)},
		{"historical within extended slack", 14 * 24 * time.Hour, now.Add(-2 * time.Hour), now.Add(-time.Hour), now.Add(-2 * time.Hour), now.Add(-time.Hour)},
		{"inverted input", 2 * time.Minute, now.Add(time.Minute), now.Add(-time.Minute), now.Add(time.Minute), now.Add(time.Minute)},
		{"exact lower boundary", 2 * time.Minute, now.Add(-2 * time.Minute), now.Add(-2 * time.Minute), now.Add(-2 * time.Minute), now.Add(-2 * time.Minute)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := defaultConfig(t, t.TempDir())
			cfg.WAL.IngestionSlack = tt.slack
			ls, err := liveStoreWithConfig(t, cfg)
			require.NoError(t, err)
			defer func() { require.NoError(t, services.StopAndAwaitTerminated(t.Context(), ls)) }()
			records := []*kgo.Record{{
				Key: []byte(testTenantID), Timestamp: now,
				Value: createValidPushRequestStartEnd(t, tt.start, tt.end),
			}}
			_, err = ls.consume(t.Context(), createRecordIter(records), now)
			require.NoError(t, err)
			inst, err := ls.getOrCreateInstance(testTenantID)
			require.NoError(t, err)
			drained, err := inst.cutIdleTraces(t.Context(), true)
			require.NoError(t, err)
			require.True(t, drained)
			meta := inst.blocks.Load().headBlock.BlockMeta()
			require.True(t, tt.wantStart.Equal(meta.StartTime), "WAL start: %v", meta.StartTime)
			require.True(t, tt.wantEnd.Equal(meta.EndTime), "WAL end: %v", meta.EndTime)
			require.False(t, meta.EndTime.Before(meta.StartTime))
			id, err := inst.cutBlocks(t.Context(), true)
			require.NoError(t, err)
			_, err = inst.completeBlock(t.Context(), id)
			require.NoError(t, err)
			meta = inst.blocks.Load().completeBlocks[id].BlockMeta()
			require.True(t, tt.wantStart.Equal(meta.StartTime), "completed start: %v", meta.StartTime)
			require.True(t, tt.wantEnd.Equal(meta.EndTime), "completed end: %v", meta.EndTime)
		})
	}
}
