package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func resetOpenAIStickyBindingMetricsForTest() {
	defaultOpenAIStickyBindingMetrics.stickySoftMissTotal.Store(0)
	defaultOpenAIStickyBindingMetrics.stickyHardInvalidateTotal.Store(0)
	defaultOpenAIStickyBindingMetrics.stickyLookupMissTotal.Store(0)
	defaultOpenAIStickyBindingMetrics.stickyTransportSoftMissTotal.Store(0)
	defaultOpenAIStickyBindingMetrics.stickyTemporarySoftMissTotal.Store(0)
	defaultOpenAIStickyBindingMetrics.stickyModelInvalidateTotal.Store(0)
	defaultOpenAIStickyBindingMetrics.previousSoftMissTotal.Store(0)
	defaultOpenAIStickyBindingMetrics.previousHardInvalidateTotal.Store(0)
	defaultOpenAIStickyBindingMetrics.previousLookupMissTotal.Store(0)
	defaultOpenAIStickyBindingMetrics.previousTransportSoftMissTotal.Store(0)
}

func TestSnapshotOpenAIStickyBindingMetrics(t *testing.T) {
	resetOpenAIStickyBindingMetricsForTest()
	t.Cleanup(resetOpenAIStickyBindingMetricsForTest)

	ctx := context.Background()
	recordOpenAIStickyBindingDisposition(ctx, stickyBindingKindSession, newStickyBindingSoftMiss("lookup_miss"), 101, "session-a", "")
	recordOpenAIStickyBindingDisposition(ctx, stickyBindingKindSession, newStickyBindingSoftMiss("model_rate_limited"), 101, "session-a", "")
	recordOpenAIStickyBindingDisposition(ctx, stickyBindingKindSession, newStickyBindingHardInvalidate("model_unsupported"), 101, "session-a", "")
	recordOpenAIStickyBindingDisposition(ctx, stickyBindingKindPreviousResponse, newStickyBindingSoftMiss("transport_cooling"), 202, "", "resp-1")
	recordOpenAIStickyBindingDisposition(ctx, stickyBindingKindPreviousResponse, newStickyBindingHardInvalidate("oauth_expired"), 202, "", "resp-1")

	snapshot := snapshotOpenAIStickyBindingMetrics()
	require.Equal(t, int64(2), snapshot.StickySoftMissTotal)
	require.Equal(t, int64(1), snapshot.StickyHardInvalidateTotal)
	require.Equal(t, int64(1), snapshot.StickyLookupMissTotal)
	require.Equal(t, int64(1), snapshot.StickyTemporarySoftMissTotal)
	require.Equal(t, int64(1), snapshot.StickyModelInvalidateTotal)
	require.Equal(t, int64(1), snapshot.PreviousSoftMissTotal)
	require.Equal(t, int64(1), snapshot.PreviousHardInvalidateTotal)
	require.Equal(t, int64(1), snapshot.PreviousTransportSoftMissTotal)
}
