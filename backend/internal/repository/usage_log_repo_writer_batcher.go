package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/logger"
	"github.com/senran-N/sub2api/internal/service"
)

func (r *usageLogRepository) runCreateBatcher(db *sql.DB) {
	for {
		first, ok := <-r.createBatchCh
		if !ok {
			return
		}

		batch := make([]usageLogCreateRequest, 0, usageLogCreateBatchMaxSize)
		batch = append(batch, first)

		timer := time.NewTimer(usageLogCreateBatchWindow)
	batchLoop:
		for len(batch) < usageLogCreateBatchMaxSize {
			select {
			case req, ok := <-r.createBatchCh:
				if !ok {
					break batchLoop
				}
				batch = append(batch, req)
			case <-timer.C:
				break batchLoop
			}
		}
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}

		r.flushCreateBatch(db, batch)
	}
}

func (r *usageLogRepository) runBestEffortBatcher(db *sql.DB) {
	for {
		first, ok := <-r.bestEffortBatchCh
		if !ok {
			return
		}

		batch := make([]usageLogBestEffortRequest, 0, usageLogBestEffortBatchMaxSize)
		batch = append(batch, first)

		timer := time.NewTimer(usageLogBestEffortBatchWindow)
	bestEffortLoop:
		for len(batch) < usageLogBestEffortBatchMaxSize {
			select {
			case req, ok := <-r.bestEffortBatchCh:
				if !ok {
					break bestEffortLoop
				}
				batch = append(batch, req)
			case <-timer.C:
				break bestEffortLoop
			}
		}
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}

		r.flushBestEffortBatch(db, batch)
	}
}

func (r *usageLogRepository) flushCreateBatch(db *sql.DB, batch []usageLogCreateRequest) {
	if len(batch) == 0 {
		return
	}

	uniqueOrder := make([]string, 0, len(batch))
	preparedByKey := make(map[string]usageLogInsertPrepared, len(batch))
	requestsByKey := make(map[string][]usageLogCreateRequest, len(batch))
	fallback := make([]usageLogCreateRequest, 0)

	for _, req := range batch {
		if req.log == nil {
			completeUsageLogCreateRequest(req, usageLogCreateResult{inserted: false, err: nil})
			continue
		}
		if req.shared != nil && !req.shared.state.CompareAndSwap(usageLogCreateStateQueued, usageLogCreateStateProcessing) {
			if req.shared.state.Load() == usageLogCreateStateCanceled {
				completeUsageLogCreateRequest(req, usageLogCreateResult{
					inserted: false,
					err:      service.MarkUsageLogCreateNotPersisted(context.Canceled),
				})
				continue
			}
		}

		prepared := req.prepared
		if prepared.requestID == "" {
			fallback = append(fallback, req)
			continue
		}

		key := usageLogBatchKey(prepared.requestID, req.log.APIKeyID)
		if _, exists := requestsByKey[key]; !exists {
			uniqueOrder = append(uniqueOrder, key)
			preparedByKey[key] = prepared
		}
		requestsByKey[key] = append(requestsByKey[key], req)
	}

	if len(uniqueOrder) > 0 {
		insertedMap, stateMap, safeFallback, err := r.batchInsertUsageLogs(db, uniqueOrder, preparedByKey)
		if err != nil {
			if safeFallback {
				for _, key := range uniqueOrder {
					fallback = append(fallback, requestsByKey[key]...)
				}
			} else {
				for _, key := range uniqueOrder {
					reqs := requestsByKey[key]
					state, hasState := stateMap[key]
					inserted := insertedMap[key]
					for idx, req := range reqs {
						req.log.RateMultiplier = preparedByKey[key].rateMultiplier
						if hasState {
							req.log.ID = state.ID
							req.log.CreatedAt = state.CreatedAt
						}
						switch {
						case inserted && idx == 0:
							completeUsageLogCreateRequest(req, usageLogCreateResult{inserted: true, err: nil})
						case inserted:
							completeUsageLogCreateRequest(req, usageLogCreateResult{inserted: false, err: nil})
						case hasState:
							completeUsageLogCreateRequest(req, usageLogCreateResult{inserted: false, err: nil})
						case idx == 0:
							completeUsageLogCreateRequest(req, usageLogCreateResult{inserted: false, err: err})
						default:
							completeUsageLogCreateRequest(req, usageLogCreateResult{inserted: false, err: nil})
						}
					}
				}
			}
		} else {
			for _, key := range uniqueOrder {
				reqs := requestsByKey[key]
				state, ok := stateMap[key]
				if !ok {
					for _, req := range reqs {
						completeUsageLogCreateRequest(req, usageLogCreateResult{
							inserted: false,
							err:      fmt.Errorf("usage log batch state missing for key=%s", key),
						})
					}
					continue
				}
				for idx, req := range reqs {
					req.log.ID = state.ID
					req.log.CreatedAt = state.CreatedAt
					req.log.RateMultiplier = preparedByKey[key].rateMultiplier
					completeUsageLogCreateRequest(req, usageLogCreateResult{
						inserted: idx == 0 && insertedMap[key],
						err:      nil,
					})
				}
			}
		}
	}

	if len(fallback) == 0 {
		return
	}

	for _, req := range fallback {
		fallbackCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		inserted, err := r.createSingle(fallbackCtx, db, req.log)
		cancel()
		completeUsageLogCreateRequest(req, usageLogCreateResult{inserted: inserted, err: err})
	}
}

func (r *usageLogRepository) flushBestEffortBatch(db *sql.DB, batch []usageLogBestEffortRequest) {
	if len(batch) == 0 {
		return
	}

	type bestEffortGroup struct {
		prepared usageLogInsertPrepared
		apiKeyID int64
		key      string
		reqs     []usageLogBestEffortRequest
	}

	groupsByKey := make(map[string]*bestEffortGroup, len(batch))
	groupOrder := make([]*bestEffortGroup, 0, len(batch))
	preparedList := make([]usageLogInsertPrepared, 0, len(batch))

	for idx, req := range batch {
		prepared := req.prepared
		key := fmt.Sprintf("__best_effort_%d", idx)
		if prepared.requestID != "" {
			key = usageLogBatchKey(prepared.requestID, req.apiKeyID)
		}
		group, exists := groupsByKey[key]
		if !exists {
			group = &bestEffortGroup{
				prepared: prepared,
				apiKeyID: req.apiKeyID,
				key:      key,
			}
			groupsByKey[key] = group
			groupOrder = append(groupOrder, group)
			preparedList = append(preparedList, prepared)
		}
		group.reqs = append(group.reqs, req)
	}

	if len(preparedList) == 0 {
		for _, req := range batch {
			sendUsageLogBestEffortResult(req.resultCh, nil)
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query, args := buildUsageLogBestEffortInsertQuery(preparedList)
	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		logger.LegacyPrintf("repository.usage_log", "best-effort batch insert failed: %v", err)
		for _, group := range groupOrder {
			singleErr := execUsageLogInsertNoResult(ctx, db, group.prepared)
			if singleErr != nil {
				logger.LegacyPrintf("repository.usage_log", "best-effort single fallback insert failed: %v", singleErr)
			} else if group.prepared.requestID != "" && r != nil && r.bestEffortRecent != nil {
				r.bestEffortRecent.SetDefault(group.key, struct{}{})
			}
			for _, req := range group.reqs {
				sendUsageLogBestEffortResult(req.resultCh, singleErr)
			}
		}
		return
	}

	for _, group := range groupOrder {
		if group.prepared.requestID != "" && r != nil && r.bestEffortRecent != nil {
			r.bestEffortRecent.SetDefault(group.key, struct{}{})
		}
		for _, req := range group.reqs {
			sendUsageLogBestEffortResult(req.resultCh, nil)
		}
	}
}

func sendUsageLogBestEffortResult(ch chan error, err error) {
	if ch == nil {
		return
	}
	select {
	case ch <- err:
	default:
	}
}

func completeUsageLogCreateRequest(req usageLogCreateRequest, res usageLogCreateResult) {
	if req.shared != nil {
		req.shared.state.Store(usageLogCreateStateCompleted)
	}
	sendUsageLogCreateResult(req.resultCh, res)
}

func (r *usageLogRepository) batchInsertUsageLogs(db *sql.DB, keys []string, preparedByKey map[string]usageLogInsertPrepared) (map[string]bool, map[string]usageLogBatchState, bool, error) {
	if len(keys) == 0 {
		return map[string]bool{}, map[string]usageLogBatchState{}, false, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query, args := buildUsageLogBatchInsertQuery(keys, preparedByKey)
	var payload []byte
	if err := db.QueryRowContext(ctx, query, args...).Scan(&payload); err != nil {
		return nil, nil, true, err
	}

	var rows []usageLogBatchRow
	if err := json.Unmarshal(payload, &rows); err != nil {
		return nil, nil, false, err
	}

	insertedMap := make(map[string]bool, len(keys))
	stateMap := make(map[string]usageLogBatchState, len(keys))
	for _, row := range rows {
		key := usageLogBatchKey(row.RequestID, row.APIKeyID)
		insertedMap[key] = row.Inserted
		stateMap[key] = usageLogBatchState{
			ID:        row.ID,
			CreatedAt: row.CreatedAt,
		}
	}
	if len(stateMap) != len(keys) {
		return insertedMap, stateMap, false, fmt.Errorf("usage log batch state count mismatch: got=%d want=%d", len(stateMap), len(keys))
	}
	return insertedMap, stateMap, false, nil
}

func sendUsageLogCreateResult(ch chan usageLogCreateResult, res usageLogCreateResult) {
	if ch == nil {
		return
	}
	select {
	case ch <- res:
	default:
	}
}

func (r *usageLogRepository) bestEffortRecentKey(requestID string, apiKeyID int64) (string, bool) {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" || r == nil || r.bestEffortRecent == nil {
		return "", false
	}
	return usageLogBatchKey(requestID, apiKeyID), true
}
