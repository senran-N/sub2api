package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"sync/atomic"
	"time"

	dbent "github.com/senran-N/sub2api/ent"
	"github.com/senran-N/sub2api/internal/service"
)

const (
	usageLogCreateBatchMaxSize  = 64
	usageLogCreateBatchWindow   = 3 * time.Millisecond
	usageLogCreateBatchQueueCap = 4096
	usageLogCreateCancelWait    = 2 * time.Second

	usageLogBestEffortBatchMaxSize  = 256
	usageLogBestEffortBatchWindow   = 20 * time.Millisecond
	usageLogBestEffortBatchQueueCap = 32768
	usageLogBestEffortRecentTTL     = 30 * time.Second
)

type usageLogCreateRequest struct {
	log      *service.UsageLog
	prepared usageLogInsertPrepared
	shared   *usageLogCreateShared
	resultCh chan usageLogCreateResult
}

type usageLogCreateResult struct {
	inserted bool
	err      error
}

type usageLogBestEffortRequest struct {
	prepared usageLogInsertPrepared
	apiKeyID int64
	resultCh chan error
}

type usageLogCreateShared struct {
	state atomic.Int32
}

const (
	usageLogCreateStateQueued int32 = iota
	usageLogCreateStateProcessing
	usageLogCreateStateCompleted
	usageLogCreateStateCanceled
)

func (r *usageLogRepository) Create(ctx context.Context, log *service.UsageLog) (bool, error) {
	if log == nil {
		return false, nil
	}

	if tx := dbent.TxFromContext(ctx); tx != nil {
		return r.createSingle(ctx, tx.Client(), log)
	}
	requestID := strings.TrimSpace(log.RequestID)
	if requestID == "" {
		return r.createSingle(ctx, r.sql, log)
	}
	log.RequestID = requestID
	return r.createBatched(ctx, log)
}

func (r *usageLogRepository) CreateBestEffort(ctx context.Context, log *service.UsageLog) error {
	if log == nil {
		return nil
	}

	if tx := dbent.TxFromContext(ctx); tx != nil {
		_, err := r.createSingle(ctx, tx.Client(), log)
		return err
	}
	if r.db == nil {
		_, err := r.createSingle(ctx, r.sql, log)
		return err
	}

	r.ensureBestEffortBatcher()
	if r.bestEffortBatchCh == nil {
		_, err := r.createSingle(ctx, r.sql, log)
		return err
	}

	req := usageLogBestEffortRequest{
		prepared: prepareUsageLogInsert(log),
		apiKeyID: log.APIKeyID,
		resultCh: make(chan error, 1),
	}
	if key, ok := r.bestEffortRecentKey(req.prepared.requestID, req.apiKeyID); ok {
		if _, exists := r.bestEffortRecent.Get(key); exists {
			return nil
		}
	}

	select {
	case r.bestEffortBatchCh <- req:
	case <-ctx.Done():
		return service.MarkUsageLogCreateDropped(ctx.Err())
	default:
		return service.MarkUsageLogCreateDropped(errors.New("usage log best-effort queue full"))
	}

	select {
	case err := <-req.resultCh:
		return err
	case <-ctx.Done():
		return service.MarkUsageLogCreateDropped(ctx.Err())
	}
}

func (r *usageLogRepository) createSingle(ctx context.Context, sqlq sqlExecutor, log *service.UsageLog) (bool, error) {
	prepared := prepareUsageLogInsert(log)
	if sqlq == nil {
		sqlq = r.sql
	}
	if ctx != nil && ctx.Err() != nil {
		return false, service.MarkUsageLogCreateNotPersisted(ctx.Err())
	}

	if err := scanSingleRow(ctx, sqlq, usageLogSingleInsertReturningQuery, prepared.args, &log.ID, &log.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) && prepared.requestID != "" {
			selectQuery := "SELECT id, created_at FROM usage_logs WHERE request_id = $1 AND api_key_id = $2"
			if err := scanSingleRow(ctx, sqlq, selectQuery, []any{prepared.requestID, log.APIKeyID}, &log.ID, &log.CreatedAt); err != nil {
				return false, err
			}
			log.RateMultiplier = prepared.rateMultiplier
			return false, nil
		}
		return false, err
	}
	log.RateMultiplier = prepared.rateMultiplier
	return true, nil
}

func (r *usageLogRepository) createBatched(ctx context.Context, log *service.UsageLog) (bool, error) {
	if r.db == nil {
		return r.createSingle(ctx, r.sql, log)
	}
	r.ensureCreateBatcher()
	if r.createBatchCh == nil {
		return r.createSingle(ctx, r.sql, log)
	}

	req := usageLogCreateRequest{
		log:      log,
		prepared: prepareUsageLogInsert(log),
		shared:   &usageLogCreateShared{},
		resultCh: make(chan usageLogCreateResult, 1),
	}

	select {
	case r.createBatchCh <- req:
	case <-ctx.Done():
		return false, service.MarkUsageLogCreateNotPersisted(ctx.Err())
	default:
		return false, service.MarkUsageLogCreateNotPersisted(errors.New("usage log create batch queue full"))
	}

	select {
	case res := <-req.resultCh:
		return res.inserted, res.err
	case <-ctx.Done():
		if req.shared != nil && req.shared.state.CompareAndSwap(usageLogCreateStateQueued, usageLogCreateStateCanceled) {
			return false, service.MarkUsageLogCreateNotPersisted(ctx.Err())
		}
		timer := time.NewTimer(usageLogCreateCancelWait)
		defer timer.Stop()
		select {
		case res := <-req.resultCh:
			return res.inserted, res.err
		case <-timer.C:
			return false, ctx.Err()
		}
	}
}

func (r *usageLogRepository) ensureCreateBatcher() {
	if r == nil || r.db == nil || r.createBatchCh != nil {
		return
	}
	r.createBatchOnce.Do(func() {
		r.createBatchCh = make(chan usageLogCreateRequest, usageLogCreateBatchQueueCap)
		go r.runCreateBatcher(r.db)
	})
}

func (r *usageLogRepository) ensureBestEffortBatcher() {
	if r == nil || r.db == nil || r.bestEffortBatchCh != nil {
		return
	}
	r.bestEffortBatchOnce.Do(func() {
		r.bestEffortBatchCh = make(chan usageLogBestEffortRequest, usageLogBestEffortBatchQueueCap)
		go r.runBestEffortBatcher(r.db)
	})
}
