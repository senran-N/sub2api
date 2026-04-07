package service

import (
	"context"
	"testing"
	"time"
)

func TestRunDetachedTaskExecutesFunction(t *testing.T) {
	doneSignal := make(chan struct{})

	done := runDetachedTask("test_execute", func(ctx context.Context) {
		if ctx == nil {
			t.Fatal("expected background context")
		}
		close(doneSignal)
	})

	select {
	case <-doneSignal:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for detached task execution")
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for detached task completion")
	}
}

func TestRunDetachedTaskRecoversPanic(t *testing.T) {
	done := runDetachedTask("test_panic", func(context.Context) {
		panic("boom")
	})

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for detached panic task completion")
	}
}
