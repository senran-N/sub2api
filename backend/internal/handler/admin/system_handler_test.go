package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

type testSystemLockRepo struct {
	mu     sync.Mutex
	nextID int64
	record *service.IdempotencyRecord
}

func newTestSystemLockRepo() *testSystemLockRepo {
	return &testSystemLockRepo{nextID: 1}
}

func (r *testSystemLockRepo) CreateProcessing(_ context.Context, record *service.IdempotencyRecord) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.record != nil {
		return false, nil
	}
	cp := *record
	cp.ID = r.nextID
	r.nextID++
	r.record = &cp
	record.ID = cp.ID
	return true, nil
}

func (r *testSystemLockRepo) GetByScopeAndKeyHash(_ context.Context, _, _ string) (*service.IdempotencyRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return cloneSystemLockRecord(r.record), nil
}

func (r *testSystemLockRepo) TryReclaim(_ context.Context, id int64, status string, _ time.Time, lockedUntil, expiresAt time.Time) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.record == nil || r.record.ID != id || r.record.Status != status {
		return false, nil
	}
	r.record.Status = service.IdempotencyStatusProcessing
	r.record.LockedUntil = &lockedUntil
	r.record.ExpiresAt = expiresAt
	return true, nil
}

func (r *testSystemLockRepo) ExtendProcessingLock(_ context.Context, id int64, requestFingerprint string, newLockedUntil, newExpiresAt time.Time) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.record == nil || r.record.ID != id || r.record.RequestFingerprint != requestFingerprint || r.record.Status != service.IdempotencyStatusProcessing {
		return false, nil
	}
	r.record.LockedUntil = &newLockedUntil
	r.record.ExpiresAt = newExpiresAt
	return true, nil
}

func (r *testSystemLockRepo) MarkSucceeded(_ context.Context, id int64, _ int, responseBody string, expiresAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.record == nil || r.record.ID != id {
		return nil
	}
	r.record.Status = service.IdempotencyStatusSucceeded
	r.record.ResponseBody = &responseBody
	r.record.LockedUntil = nil
	r.record.ExpiresAt = expiresAt
	return nil
}

func (r *testSystemLockRepo) MarkFailedRetryable(_ context.Context, id int64, _ string, lockedUntil, expiresAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.record == nil || r.record.ID != id {
		return nil
	}
	r.record.Status = service.IdempotencyStatusFailedRetryable
	r.record.LockedUntil = &lockedUntil
	r.record.ExpiresAt = expiresAt
	return nil
}

func (r *testSystemLockRepo) DeleteExpired(context.Context, time.Time, int) (int64, error) {
	return 0, nil
}

func cloneSystemLockRecord(record *service.IdempotencyRecord) *service.IdempotencyRecord {
	if record == nil {
		return nil
	}
	cp := *record
	if record.LockedUntil != nil {
		t := *record.LockedUntil
		cp.LockedUntil = &t
	}
	return &cp
}

func TestRestartService_KeepsLockUntilRestartTriggerRuns(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service.SetDefaultIdempotencyCoordinator(nil)
	t.Cleanup(func() {
		service.SetDefaultIdempotencyCoordinator(nil)
	})

	oldDelay := restartServiceDelay
	oldRestart := restartServiceAsync
	restartServiceDelay = 50 * time.Millisecond
	restartCalled := make(chan struct{}, 1)
	restartServiceAsync = func() {
		select {
		case restartCalled <- struct{}{}:
		default:
		}
	}
	t.Cleanup(func() {
		restartServiceDelay = oldDelay
		restartServiceAsync = oldRestart
	})

	lockSvc := service.NewSystemOperationLockService(newTestSystemLockRepo(), service.IdempotencyConfig{
		SystemOperationTTL: 5 * time.Second,
		ProcessingTimeout:  500 * time.Millisecond,
	})
	handler := &SystemHandler{lockSvc: lockSvc}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/system/restart", nil)

	handler.RestartService(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)

	_, err := lockSvc.Acquire(context.Background(), "other-op")
	require.Error(t, err, "lock should remain held until restart is actually triggered")

	select {
	case <-restartCalled:
	case <-time.After(time.Second):
		t.Fatal("restart callback was not invoked")
	}

	require.Eventually(t, func() bool {
		lock, acquireErr := lockSvc.Acquire(context.Background(), "other-op")
		if acquireErr != nil {
			return false
		}
		_ = lockSvc.Release(context.Background(), lock, true, "")
		return true
	}, time.Second, 10*time.Millisecond)
}
