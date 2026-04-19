package service

import (
	"context"
	"testing"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestSchedulerSnapshotService_DefaultBucketsIncludeGrok(t *testing.T) {
	svc := &SchedulerSnapshotService{
		cfg: &config.Config{},
	}

	buckets, err := svc.defaultBuckets(context.Background())
	require.NoError(t, err)

	var hasGrokSingle bool
	var hasGrokForced bool
	for _, bucket := range buckets {
		if bucket.GroupID != 0 || bucket.Platform != PlatformGrok {
			continue
		}
		switch bucket.Mode {
		case SchedulerModeSingle:
			hasGrokSingle = true
		case SchedulerModeForced:
			hasGrokForced = true
		}
	}

	require.True(t, hasGrokSingle, "expected grok single bucket in default snapshot bootstrap")
	require.True(t, hasGrokForced, "expected grok forced bucket in default snapshot bootstrap")
}
