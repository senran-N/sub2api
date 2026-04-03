package service

import (
	"context"
	"errors"
	"testing"
)

type usageLogWriterStub struct {
	UsageLogRepository
	createErr           error
	bestEffortErr       error
	createCalls         int
	bestEffortCalls     int
	lastUsageLogRequest *UsageLog
}

func (s *usageLogWriterStub) Create(ctx context.Context, usageLog *UsageLog) (bool, error) {
	s.createCalls++
	s.lastUsageLogRequest = usageLog
	if s.createErr != nil {
		return false, s.createErr
	}
	return true, nil
}

func (s *usageLogWriterStub) CreateBestEffort(ctx context.Context, usageLog *UsageLog) error {
	s.bestEffortCalls++
	s.lastUsageLogRequest = usageLog
	return s.bestEffortErr
}

func TestTruncateForLog(t *testing.T) {
	got := truncateForLog([]byte("hello\nworld\r\n"), 8)
	if got != "hello\\nwo" {
		t.Fatalf("truncate=%q want=%q", got, "hello\\nwo")
	}
}

func TestDetachStreamUpstreamContext_StreamDetachesCancellation(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	child, release := detachStreamUpstreamContext(parent, true)
	cancel()
	defer release()

	if child.Err() != nil {
		t.Fatalf("detached child should not inherit cancellation")
	}
}

func TestWriteUsageLogBestEffort_FallsBackToCreate(t *testing.T) {
	repo := &usageLogWriterStub{
		bestEffortErr: errors.New("best effort failed"),
	}
	usageLog := &UsageLog{Model: "claude-sonnet-4-5"}

	writeUsageLogBestEffort(context.Background(), repo, usageLog, "service.gateway")

	if repo.bestEffortCalls != 1 {
		t.Fatalf("bestEffortCalls=%d want=1", repo.bestEffortCalls)
	}
	if repo.createCalls != 1 {
		t.Fatalf("createCalls=%d want=1", repo.createCalls)
	}
	if repo.lastUsageLogRequest != usageLog {
		t.Fatalf("usage log pointer mismatch")
	}
}
