package service

import (
	"errors"
	"testing"
	"time"

	"github.com/zeromicro/go-zero/core/collection"
)

type lifecycleStartableStub struct {
	startCalls int
}

func (s *lifecycleStartableStub) Start() {
	s.startCalls++
}

func TestStartBackgroundService_StartsAndReturnsSameInstance(t *testing.T) {
	svc := &lifecycleStartableStub{}
	returned := startBackgroundService(svc)

	if returned != svc {
		t.Fatalf("期望返回同一个实例")
	}
	if svc.startCalls != 1 {
		t.Fatalf("期望 Start 被调用 1 次，实际: %d", svc.startCalls)
	}
}

func TestProvideTimingWheelService_ReturnsError(t *testing.T) {
	original := newTimingWheel
	t.Cleanup(func() { newTimingWheel = original })

	newTimingWheel = func(_ time.Duration, _ int, _ collection.Execute) (*collection.TimingWheel, error) {
		return nil, errors.New("boom")
	}

	svc, err := ProvideTimingWheelService(NewLifecycleRegistry())
	if err == nil {
		t.Fatalf("期望返回 error，但得到 nil")
	}
	if svc != nil {
		t.Fatalf("期望返回 nil svc，但得到非空")
	}
}

func TestProvideTimingWheelService_Success(t *testing.T) {
	registry := NewLifecycleRegistry()
	svc, err := ProvideTimingWheelService(registry)
	if err != nil {
		t.Fatalf("期望 err 为 nil，但得到: %v", err)
	}
	if svc == nil {
		t.Fatalf("期望 svc 非空，但得到 nil")
	}
	entries := registry.Entries()
	if len(entries) != 1 || entries[0].Name != "TimingWheelService" {
		t.Fatalf("unexpected lifecycle entries: %+v", entries)
	}
	entries[0].Stop()
}
