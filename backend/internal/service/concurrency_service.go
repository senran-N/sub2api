package service

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/logger"
)

var (
	requestIDPrefix  = initRequestIDPrefix()
	requestIDCounter atomic.Uint64
)

func initRequestIDPrefix() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err == nil {
		return "r" + strconv.FormatUint(binary.BigEndian.Uint64(b), 36)
	}
	fallback := uint64(time.Now().UnixNano()) ^ (uint64(os.Getpid()) << 16)
	return "r" + strconv.FormatUint(fallback, 36)
}

func RequestIDPrefix() string {
	return requestIDPrefix
}

func generateRequestID() string {
	seq := requestIDCounter.Add(1)
	return requestIDPrefix + "-" + strconv.FormatUint(seq, 36)
}

func (s *ConcurrencyService) CleanupStaleProcessSlots(ctx context.Context) error {
	if s == nil || s.cache == nil {
		return nil
	}
	return s.cache.CleanupStaleProcessSlots(ctx, RequestIDPrefix())
}

const (
	// Default extra wait slots beyond concurrency limit
	defaultExtraWaitSlots        = 20
	accountSlotCleanupTimeout    = 5 * time.Second
	accountSweeperCleanupTimeout = 15 * time.Second
	accountSweeperListTimeout    = 5 * time.Second
	accountSweeperPerItemTimeout = 2 * time.Second
)

func newCleanupContext(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	base := context.Background()
	if ctx != nil {
		base = context.WithoutCancel(ctx)
	}
	return context.WithTimeout(base, timeout)
}

// ConcurrencyService manages concurrent request limiting for accounts and users
type ConcurrencyService struct {
	cache ConcurrencyCache
}

type accountSlotCleanupSweeper interface {
	CleanupExpiredAccountSlotsAll(ctx context.Context) error
}

// NewConcurrencyService creates a new ConcurrencyService
func NewConcurrencyService(cache ConcurrencyCache) *ConcurrencyService {
	return &ConcurrencyService{cache: cache}
}

// AcquireResult represents the result of acquiring a concurrency slot
type AcquireResult struct {
	Acquired    bool
	ReleaseFunc func() // Must be called when done (typically via defer)
}

type AcquireOrQueueResult struct {
	Acquired     bool
	QueueAllowed bool
	WaitCounted  bool
	ReleaseFunc  func()
}

// AcquireAccountSlot attempts to acquire a concurrency slot for an account.
// If the account is at max concurrency, it waits until a slot is available or timeout.
// Returns a release function that MUST be called when the request completes.
func (s *ConcurrencyService) AcquireAccountSlot(ctx context.Context, accountID int64, maxConcurrency int) (*AcquireResult, error) {
	// If maxConcurrency is 0 or negative, no limit
	if maxConcurrency <= 0 {
		return &AcquireResult{
			Acquired:    true,
			ReleaseFunc: func() {}, // no-op
		}, nil
	}

	// Generate unique request ID for this slot
	requestID := generateRequestID()

	acquired, err := s.cache.AcquireAccountSlot(ctx, accountID, maxConcurrency, requestID)
	if err != nil {
		return nil, err
	}

	if acquired {
		return &AcquireResult{
			Acquired: true,
			ReleaseFunc: func() {
				bgCtx, cancel := newCleanupContext(ctx, accountSlotCleanupTimeout)
				defer cancel()
				if err := s.cache.ReleaseAccountSlot(bgCtx, accountID, requestID); err != nil {
					logger.LegacyPrintf("service.concurrency", "Warning: failed to release account slot for %d (req=%s): %v", accountID, requestID, err)
				}
			},
		}, nil
	}

	return &AcquireResult{
		Acquired:    false,
		ReleaseFunc: nil,
	}, nil
}

// AcquireUserSlot attempts to acquire a concurrency slot for a user.
// If the user is at max concurrency, it waits until a slot is available or timeout.
// Returns a release function that MUST be called when the request completes.
func (s *ConcurrencyService) AcquireUserSlot(ctx context.Context, userID int64, maxConcurrency int) (*AcquireResult, error) {
	// If maxConcurrency is 0 or negative, no limit
	if maxConcurrency <= 0 {
		return &AcquireResult{
			Acquired:    true,
			ReleaseFunc: func() {}, // no-op
		}, nil
	}

	// Generate unique request ID for this slot
	requestID := generateRequestID()

	acquired, err := s.cache.AcquireUserSlot(ctx, userID, maxConcurrency, requestID)
	if err != nil {
		return nil, err
	}

	if acquired {
		return &AcquireResult{
			Acquired: true,
			ReleaseFunc: func() {
				bgCtx, cancel := newCleanupContext(ctx, accountSlotCleanupTimeout)
				defer cancel()
				if err := s.cache.ReleaseUserSlot(bgCtx, userID, requestID); err != nil {
					logger.LegacyPrintf("service.concurrency", "Warning: failed to release user slot for %d (req=%s): %v", userID, requestID, err)
				}
			},
		}, nil
	}

	return &AcquireResult{
		Acquired:    false,
		ReleaseFunc: nil,
	}, nil
}

func (s *ConcurrencyService) AcquireUserSlotOrQueue(ctx context.Context, userID int64, maxConcurrency int, maxWait int) (*AcquireOrQueueResult, error) {
	if maxConcurrency <= 0 {
		return &AcquireOrQueueResult{
			Acquired:     true,
			QueueAllowed: true,
			ReleaseFunc:  func() {},
		}, nil
	}
	if s.cache == nil {
		return &AcquireOrQueueResult{
			Acquired:     true,
			QueueAllowed: true,
			ReleaseFunc:  func() {},
		}, nil
	}

	requestID := generateRequestID()
	acquired, enqueued, err := s.cache.AcquireUserSlotOrEnqueueWait(ctx, userID, maxConcurrency, maxWait, requestID)
	if err != nil {
		return s.acquireUserSlotOrQueueFallback(ctx, userID, maxConcurrency, maxWait, requestID, err)
	}
	if acquired {
		return &AcquireOrQueueResult{
			Acquired:     true,
			QueueAllowed: true,
			ReleaseFunc: func() {
				bgCtx, cancel := newCleanupContext(ctx, accountSlotCleanupTimeout)
				defer cancel()
				if err := s.cache.ReleaseUserSlot(bgCtx, userID, requestID); err != nil {
					logger.LegacyPrintf("service.concurrency", "Warning: failed to release user slot for %d (req=%s): %v", userID, requestID, err)
				}
			},
		}, nil
	}
	return &AcquireOrQueueResult{
		QueueAllowed: enqueued,
		WaitCounted:  enqueued,
	}, nil
}

func (s *ConcurrencyService) acquireUserSlotOrQueueFallback(ctx context.Context, userID int64, maxConcurrency int, maxWait int, requestID string, combinedErr error) (*AcquireOrQueueResult, error) {
	logger.LegacyPrintf("service.concurrency", "Warning: acquire-or-queue user script failed for %d: %v", userID, combinedErr)

	acquired, err := s.cache.AcquireUserSlot(ctx, userID, maxConcurrency, requestID)
	if err != nil {
		return nil, err
	}
	if acquired {
		return &AcquireOrQueueResult{
			Acquired:     true,
			QueueAllowed: true,
			ReleaseFunc: func() {
				bgCtx, cancel := newCleanupContext(ctx, accountSlotCleanupTimeout)
				defer cancel()
				if err := s.cache.ReleaseUserSlot(bgCtx, userID, requestID); err != nil {
					logger.LegacyPrintf("service.concurrency", "Warning: failed to release user slot for %d (req=%s): %v", userID, requestID, err)
				}
			},
		}, nil
	}

	enqueued, err := s.cache.IncrementWaitCount(ctx, userID, maxWait)
	if err != nil {
		logger.LegacyPrintf("service.concurrency", "Warning: increment wait count fallback failed for user %d: %v", userID, err)
		return &AcquireOrQueueResult{QueueAllowed: true}, nil
	}
	return &AcquireOrQueueResult{
		QueueAllowed: enqueued,
		WaitCounted:  enqueued,
	}, nil
}

func (s *ConcurrencyService) AcquireAccountSlotOrQueue(ctx context.Context, accountID int64, maxConcurrency int, maxWait int) (*AcquireOrQueueResult, error) {
	if maxConcurrency <= 0 {
		return &AcquireOrQueueResult{
			Acquired:     true,
			QueueAllowed: true,
			ReleaseFunc:  func() {},
		}, nil
	}
	if s.cache == nil {
		return &AcquireOrQueueResult{
			Acquired:     true,
			QueueAllowed: true,
			ReleaseFunc:  func() {},
		}, nil
	}

	requestID := generateRequestID()
	acquired, enqueued, err := s.cache.AcquireAccountSlotOrEnqueueWait(ctx, accountID, maxConcurrency, maxWait, requestID)
	if err != nil {
		return s.acquireAccountSlotOrQueueFallback(ctx, accountID, maxConcurrency, maxWait, requestID, err)
	}
	if acquired {
		return &AcquireOrQueueResult{
			Acquired:     true,
			QueueAllowed: true,
			ReleaseFunc: func() {
				bgCtx, cancel := newCleanupContext(ctx, accountSlotCleanupTimeout)
				defer cancel()
				if err := s.cache.ReleaseAccountSlot(bgCtx, accountID, requestID); err != nil {
					logger.LegacyPrintf("service.concurrency", "Warning: failed to release account slot for %d (req=%s): %v", accountID, requestID, err)
				}
			},
		}, nil
	}
	return &AcquireOrQueueResult{
		QueueAllowed: enqueued,
		WaitCounted:  enqueued,
	}, nil
}

func (s *ConcurrencyService) acquireAccountSlotOrQueueFallback(ctx context.Context, accountID int64, maxConcurrency int, maxWait int, requestID string, combinedErr error) (*AcquireOrQueueResult, error) {
	logger.LegacyPrintf("service.concurrency", "Warning: acquire-or-queue account script failed for %d: %v", accountID, combinedErr)

	acquired, err := s.cache.AcquireAccountSlot(ctx, accountID, maxConcurrency, requestID)
	if err != nil {
		return nil, err
	}
	if acquired {
		return &AcquireOrQueueResult{
			Acquired:     true,
			QueueAllowed: true,
			ReleaseFunc: func() {
				bgCtx, cancel := newCleanupContext(ctx, accountSlotCleanupTimeout)
				defer cancel()
				if err := s.cache.ReleaseAccountSlot(bgCtx, accountID, requestID); err != nil {
					logger.LegacyPrintf("service.concurrency", "Warning: failed to release account slot for %d (req=%s): %v", accountID, requestID, err)
				}
			},
		}, nil
	}

	enqueued, err := s.cache.IncrementAccountWaitCount(ctx, accountID, maxWait)
	if err != nil {
		logger.LegacyPrintf("service.concurrency", "Warning: increment wait count fallback failed for account %d: %v", accountID, err)
		return &AcquireOrQueueResult{QueueAllowed: true}, nil
	}
	return &AcquireOrQueueResult{
		QueueAllowed: enqueued,
		WaitCounted:  enqueued,
	}, nil
}

// ============================================
// Wait Queue Count Methods
// ============================================

// IncrementWaitCount attempts to increment the wait queue counter for a user.
// Returns true if successful, false if the wait queue is full.
// maxWait should be user.Concurrency + defaultExtraWaitSlots
func (s *ConcurrencyService) IncrementWaitCount(ctx context.Context, userID int64, maxWait int) (bool, error) {
	if s.cache == nil {
		// Redis not available, allow request
		return true, nil
	}

	result, err := s.cache.IncrementWaitCount(ctx, userID, maxWait)
	if err != nil {
		// On error, allow the request to proceed (fail open)
		logger.LegacyPrintf("service.concurrency", "Warning: increment wait count failed for user %d: %v", userID, err)
		return true, nil
	}
	return result, nil
}

// DecrementWaitCount decrements the wait queue counter for a user.
// Should be called when a request completes or exits the wait queue.
func (s *ConcurrencyService) DecrementWaitCount(ctx context.Context, userID int64) {
	if s.cache == nil {
		return
	}

	// Use background context to ensure decrement even if original context is cancelled
	bgCtx, cancel := newCleanupContext(ctx, accountSlotCleanupTimeout)
	defer cancel()

	if err := s.cache.DecrementWaitCount(bgCtx, userID); err != nil {
		logger.LegacyPrintf("service.concurrency", "Warning: decrement wait count failed for user %d: %v", userID, err)
	}
}

// IncrementAccountWaitCount increments the wait queue counter for an account.
func (s *ConcurrencyService) IncrementAccountWaitCount(ctx context.Context, accountID int64, maxWait int) (bool, error) {
	if s.cache == nil {
		return true, nil
	}

	result, err := s.cache.IncrementAccountWaitCount(ctx, accountID, maxWait)
	if err != nil {
		logger.LegacyPrintf("service.concurrency", "Warning: increment wait count failed for account %d: %v", accountID, err)
		return true, nil
	}
	return result, nil
}

// DecrementAccountWaitCount decrements the wait queue counter for an account.
func (s *ConcurrencyService) DecrementAccountWaitCount(ctx context.Context, accountID int64) {
	if s.cache == nil {
		return
	}

	bgCtx, cancel := newCleanupContext(ctx, accountSlotCleanupTimeout)
	defer cancel()

	if err := s.cache.DecrementAccountWaitCount(bgCtx, accountID); err != nil {
		logger.LegacyPrintf("service.concurrency", "Warning: decrement wait count failed for account %d: %v", accountID, err)
	}
}

// GetAccountWaitingCount gets current wait queue count for an account.
func (s *ConcurrencyService) GetAccountWaitingCount(ctx context.Context, accountID int64) (int, error) {
	if s.cache == nil {
		return 0, nil
	}
	return s.cache.GetAccountWaitingCount(ctx, accountID)
}

// CalculateMaxWait calculates the maximum wait queue size for a user
// maxWait = userConcurrency + defaultExtraWaitSlots
func CalculateMaxWait(userConcurrency int) int {
	if userConcurrency <= 0 {
		userConcurrency = 1
	}
	return userConcurrency + defaultExtraWaitSlots
}

// GetAccountsLoadBatch returns load info for multiple accounts.
func (s *ConcurrencyService) GetAccountsLoadBatch(ctx context.Context, accounts []AccountWithConcurrency) (map[int64]*AccountLoadInfo, error) {
	if s.cache == nil {
		return map[int64]*AccountLoadInfo{}, nil
	}
	if len(accounts) == 0 {
		return map[int64]*AccountLoadInfo{}, nil
	}

	requestCache := requestAccountLoadCacheFromContext(ctx)
	now := time.Now()
	if requestCache == nil {
		return s.cache.GetAccountsLoadBatch(ctx, accounts)
	}

	cachedLoadMap, missing := requestCache.get(accounts, now)
	if len(missing) == 0 {
		return cachedLoadMap, nil
	}

	freshLoadMap, err := s.cache.GetAccountsLoadBatch(ctx, missing)
	if err != nil {
		return nil, err
	}
	requestCache.store(freshLoadMap, now)

	result := make(map[int64]*AccountLoadInfo, len(accounts))
	for accountID, info := range cachedLoadMap {
		result[accountID] = info
	}
	for _, account := range missing {
		if info, ok := freshLoadMap[account.ID]; ok && info != nil {
			result[account.ID] = buildAccountLoadInfo(account, info.CurrentConcurrency, info.WaitingCount)
			continue
		}
		result[account.ID] = &AccountLoadInfo{AccountID: account.ID}
	}

	return result, nil
}

// GetUsersLoadBatch returns load info for multiple users.
func (s *ConcurrencyService) GetUsersLoadBatch(ctx context.Context, users []UserWithConcurrency) (map[int64]*UserLoadInfo, error) {
	if s.cache == nil {
		return map[int64]*UserLoadInfo{}, nil
	}
	return s.cache.GetUsersLoadBatch(ctx, users)
}

// CleanupExpiredAccountSlots removes expired slots for one account (background task).
func (s *ConcurrencyService) CleanupExpiredAccountSlots(ctx context.Context, accountID int64) error {
	if s.cache == nil {
		return nil
	}
	return s.cache.CleanupExpiredAccountSlots(ctx, accountID)
}

// StartSlotCleanupWorker starts a background cleanup worker for expired account slots.
func (s *ConcurrencyService) StartSlotCleanupWorker(accountRepo AccountRepository, interval time.Duration) {
	if s == nil || s.cache == nil || interval <= 0 {
		return
	}
	if accountRepo == nil {
		if _, ok := s.cache.(accountSlotCleanupSweeper); !ok {
			return
		}
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		s.runSlotCleanupWorkerOnce(accountRepo)
		for range ticker.C {
			s.runSlotCleanupWorkerOnce(accountRepo)
		}
	}()
}

func (s *ConcurrencyService) runSlotCleanupWorkerOnce(accountRepo AccountRepository) {
	if s == nil || s.cache == nil {
		return
	}
	if sweeper, ok := s.cache.(accountSlotCleanupSweeper); ok {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), accountSweeperCleanupTimeout)
		err := sweeper.CleanupExpiredAccountSlotsAll(cleanupCtx)
		cancel()
		if err != nil {
			logger.LegacyPrintf("service.concurrency", "Warning: cleanup expired account slots failed: %v", err)
		}
		return
	}

	if accountRepo == nil {
		return
	}

	listCtx, cancel := context.WithTimeout(context.Background(), accountSweeperListTimeout)
	accounts, err := accountRepo.ListSchedulable(listCtx)
	cancel()
	if err != nil {
		logger.LegacyPrintf("service.concurrency", "Warning: list schedulable accounts failed: %v", err)
		return
	}
	for _, account := range accounts {
		accountCtx, accountCancel := context.WithTimeout(context.Background(), accountSweeperPerItemTimeout)
		err := s.cache.CleanupExpiredAccountSlots(accountCtx, account.ID)
		accountCancel()
		if err != nil {
			logger.LegacyPrintf("service.concurrency", "Warning: cleanup expired slots failed for account %d: %v", account.ID, err)
		}
	}
}

// GetAccountConcurrencyBatch gets current concurrency counts for multiple accounts
// Returns a map of accountID -> current concurrency count
func (s *ConcurrencyService) GetAccountConcurrencyBatch(ctx context.Context, accountIDs []int64) (map[int64]int, error) {
	if len(accountIDs) == 0 {
		return map[int64]int{}, nil
	}
	if s.cache == nil {
		result := make(map[int64]int, len(accountIDs))
		for _, accountID := range accountIDs {
			result[accountID] = 0
		}
		return result, nil
	}
	return s.cache.GetAccountConcurrencyBatch(ctx, accountIDs)
}
