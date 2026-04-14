package service

import (
	"context"
	"errors"
	"testing"
	"time"
)

type rpmCacheRuntimeLimitsStub struct {
	batchCounts  map[int64]int
	getCount     int
	getErr       error
	reserveOK    bool
	reserveCount int
	reserveErr   error
	reserveCall  struct {
		accountID int64
		limit     int
	}
	incremented  int64
	incrementErr error
}

func (s *rpmCacheRuntimeLimitsStub) ReserveRPM(ctx context.Context, accountID int64, limit int) (bool, int, error) {
	s.reserveCall.accountID = accountID
	s.reserveCall.limit = limit
	if s.reserveErr != nil {
		return false, 0, s.reserveErr
	}
	return s.reserveOK, s.reserveCount, nil
}

func (s *rpmCacheRuntimeLimitsStub) IncrementRPM(ctx context.Context, accountID int64) (int, error) {
	s.incremented = accountID
	if s.incrementErr != nil {
		return 0, s.incrementErr
	}
	return 1, nil
}

func (s *rpmCacheRuntimeLimitsStub) GetRPM(ctx context.Context, accountID int64) (int, error) {
	if s.getErr != nil {
		return 0, s.getErr
	}
	return s.getCount, nil
}

func (s *rpmCacheRuntimeLimitsStub) GetRPMBatch(ctx context.Context, accountIDs []int64) (map[int64]int, error) {
	out := make(map[int64]int, len(accountIDs))
	for _, id := range accountIDs {
		if count, ok := s.batchCounts[id]; ok {
			out[id] = count
		}
	}
	return out, nil
}

type sessionLimitRuntimeStub struct {
	windowCost   float64
	windowHit    bool
	windowErr    error
	windowStart  time.Time
	reserveOK    bool
	reserveTotal float64
	reserveErr   error
	reserveCall  struct {
		accountID     int64
		windowStart   time.Time
		reservationID string
		cost          float64
		limit         float64
		ttl           time.Duration
	}
	registerOK   bool
	registerErr  error
	registerCall struct {
		accountID   int64
		sessionID   string
		maxSessions int
		idleTimeout time.Duration
	}
}

func (s *sessionLimitRuntimeStub) RegisterSession(ctx context.Context, accountID int64, sessionUUID string, maxSessions int, idleTimeout time.Duration) (bool, error) {
	s.registerCall.accountID = accountID
	s.registerCall.sessionID = sessionUUID
	s.registerCall.maxSessions = maxSessions
	s.registerCall.idleTimeout = idleTimeout
	if s.registerErr != nil {
		return false, s.registerErr
	}
	return s.registerOK, nil
}

func (s *sessionLimitRuntimeStub) RefreshSession(ctx context.Context, accountID int64, sessionUUID string, idleTimeout time.Duration) error {
	return nil
}

func (s *sessionLimitRuntimeStub) GetActiveSessionCount(ctx context.Context, accountID int64) (int, error) {
	return 0, nil
}

func (s *sessionLimitRuntimeStub) GetActiveSessionCountBatch(ctx context.Context, accountIDs []int64, idleTimeouts map[int64]time.Duration) (map[int64]int, error) {
	return nil, nil
}

func (s *sessionLimitRuntimeStub) IsSessionActive(ctx context.Context, accountID int64, sessionUUID string) (bool, error) {
	return false, nil
}

func (s *sessionLimitRuntimeStub) GetWindowCost(ctx context.Context, accountID int64, windowStart time.Time) (float64, bool, error) {
	s.windowStart = windowStart
	return s.windowCost, s.windowHit, s.windowErr
}

func (s *sessionLimitRuntimeStub) SetWindowCost(ctx context.Context, accountID int64, windowStart time.Time, cost float64) error {
	s.windowCost = cost
	s.windowHit = true
	s.windowStart = windowStart
	return nil
}

func (s *sessionLimitRuntimeStub) GetWindowCostBatch(ctx context.Context, accountWindows map[int64]time.Time) (map[int64]float64, error) {
	return nil, nil
}

func (s *sessionLimitRuntimeStub) ReserveWindowCost(ctx context.Context, accountID int64, windowStart time.Time, reservationID string, cost float64, limit float64, ttl time.Duration) (bool, float64, error) {
	s.reserveCall.accountID = accountID
	s.reserveCall.windowStart = windowStart
	s.reserveCall.reservationID = reservationID
	s.reserveCall.cost = cost
	s.reserveCall.limit = limit
	s.reserveCall.ttl = ttl
	if s.reserveErr != nil {
		return false, 0, s.reserveErr
	}
	return s.reserveOK, s.reserveTotal, nil
}

func (s *sessionLimitRuntimeStub) ReleaseWindowCost(ctx context.Context, accountID int64, windowStart time.Time, reservationID string) error {
	return nil
}

func TestWithRPMPrefetch_StoresEligibleCounts(t *testing.T) {
	svc := &GatewayService{
		rpmCache: &rpmCacheRuntimeLimitsStub{
			batchCounts: map[int64]int{
				1: 7,
			},
		},
	}

	accounts := []Account{
		{
			ID:       1,
			Platform: PlatformAnthropic,
			Type:     AccountTypeOAuth,
			Extra: map[string]any{
				"base_rpm": 120,
			},
		},
		{
			ID:       2,
			Platform: PlatformOpenAI,
			Type:     AccountTypeOAuth,
			Extra: map[string]any{
				"base_rpm": 120,
			},
		},
		{
			ID:       3,
			Platform: PlatformAnthropic,
			Type:     AccountTypeOAuth,
		},
	}

	outCtx := svc.withRPMPrefetch(context.Background(), accounts)

	count, ok := rpmFromPrefetchContext(outCtx, 1)
	if !ok || count != 7 {
		t.Fatalf("rpm account1 ok=%v count=%d want ok=true count=7", ok, count)
	}
	if _, ok := rpmFromPrefetchContext(outCtx, 2); ok {
		t.Fatalf("unexpected rpm entry for non-anthropic account")
	}
	if _, ok := rpmFromPrefetchContext(outCtx, 3); ok {
		t.Fatalf("unexpected rpm entry for account without base_rpm")
	}
}

func TestIsAccountSchedulableForRPM_StickyOnlyDependsOnStickyFlag(t *testing.T) {
	svc := &GatewayService{
		rpmCache: &rpmCacheRuntimeLimitsStub{
			getCount: 10,
		},
	}
	account := &Account{
		ID:       9,
		Platform: PlatformAnthropic,
		Type:     AccountTypeOAuth,
		Extra: map[string]any{
			"base_rpm": 10,
		},
	}

	if svc.isAccountSchedulableForRPM(context.Background(), account, false) {
		t.Fatalf("expected non-sticky request to be blocked in sticky-only RPM zone")
	}
	if !svc.isAccountSchedulableForRPM(context.Background(), account, true) {
		t.Fatalf("expected sticky request to pass in sticky-only RPM zone")
	}
}

func TestCheckAndRegisterSession_UsesConfiguredLimitAndIdleTimeout(t *testing.T) {
	cache := &sessionLimitRuntimeStub{registerOK: false}
	svc := &GatewayService{sessionLimitCache: cache}
	account := &Account{
		ID:       12,
		Platform: PlatformAnthropic,
		Type:     AccountTypeOAuth,
		Extra: map[string]any{
			"max_sessions":                 3,
			"session_idle_timeout_minutes": 7,
		},
	}

	allowed := svc.checkAndRegisterSession(context.Background(), account, "session-1")
	if allowed {
		t.Fatalf("expected register result to be propagated")
	}
	if cache.registerCall.accountID != 12 {
		t.Fatalf("accountID=%d want=12", cache.registerCall.accountID)
	}
	if cache.registerCall.sessionID != "session-1" {
		t.Fatalf("sessionID=%s want=session-1", cache.registerCall.sessionID)
	}
	if cache.registerCall.maxSessions != 3 {
		t.Fatalf("maxSessions=%d want=3", cache.registerCall.maxSessions)
	}
	if cache.registerCall.idleTimeout != 7*time.Minute {
		t.Fatalf("idleTimeout=%s want=7m0s", cache.registerCall.idleTimeout)
	}
}

func TestTryReserveAccountRPM_UsesBaseLimitForNonStickyRequests(t *testing.T) {
	cache := &rpmCacheRuntimeLimitsStub{
		reserveOK:    true,
		reserveCount: 9,
	}
	svc := &GatewayService{rpmCache: cache}
	account := &Account{
		ID:          21,
		Platform:    PlatformAnthropic,
		Type:        AccountTypeOAuth,
		Concurrency: 3,
		Extra: map[string]any{
			"base_rpm":     8,
			"max_sessions": 5,
		},
	}

	allowed, count, err := svc.TryReserveAccountRPM(context.Background(), account, false)
	if err != nil {
		t.Fatalf("TryReserveAccountRPM err=%v", err)
	}
	if !allowed || count != 9 {
		t.Fatalf("allowed=%v count=%d want allowed=true count=9", allowed, count)
	}
	if cache.reserveCall.accountID != 21 {
		t.Fatalf("accountID=%d want=21", cache.reserveCall.accountID)
	}
	if cache.reserveCall.limit != 8 {
		t.Fatalf("limit=%d want=8", cache.reserveCall.limit)
	}
}

func TestTryReserveAccountRPM_StickyBoundRequestsCanUseStickyBuffer(t *testing.T) {
	cache := &rpmCacheRuntimeLimitsStub{
		reserveOK: true,
	}
	svc := &GatewayService{rpmCache: cache}
	account := &Account{
		ID:          22,
		Platform:    PlatformAnthropic,
		Type:        AccountTypeOAuth,
		Concurrency: 2,
		Extra: map[string]any{
			"base_rpm":     10,
			"max_sessions": 5,
		},
	}

	_, _, err := svc.TryReserveAccountRPM(context.Background(), account, true)
	if err != nil {
		t.Fatalf("TryReserveAccountRPM err=%v", err)
	}
	if cache.reserveCall.limit != 17 {
		t.Fatalf("limit=%d want=17", cache.reserveCall.limit)
	}
}

func TestTryReserveAccountWindowCost_UsesEstimatedCostAndStickyReserve(t *testing.T) {
	cache := &sessionLimitRuntimeStub{
		reserveOK:    true,
		reserveTotal: 6.5,
	}
	svc := &GatewayService{
		sessionLimitCache: cache,
		billingService:    NewBillingService(nil, nil),
	}
	account := &Account{
		ID:       31,
		Platform: PlatformAnthropic,
		Type:     AccountTypeOAuth,
		Extra: map[string]any{
			"window_cost_limit":          5.0,
			"window_cost_sticky_reserve": 1.5,
		},
	}
	apiKey := &APIKey{
		Group: &Group{ID: 9},
	}
	parsed := &ParsedRequest{
		Model:     "claude-3-5-sonnet",
		MaxTokens: 512,
		System:    "You are helpful.",
		Messages: []any{
			map[string]any{"content": "hello world"},
		},
	}

	reservation, allowed, total, err := svc.TryReserveAccountWindowCost(context.Background(), account, apiKey, parsed, "", true)
	if err != nil {
		t.Fatalf("TryReserveAccountWindowCost err=%v", err)
	}
	if !allowed {
		t.Fatalf("expected reservation to be allowed")
	}
	if total != 6.5 {
		t.Fatalf("total=%v want=6.5", total)
	}
	if reservation == nil || reservation.ReservationID == "" {
		t.Fatalf("expected non-nil reservation with id")
	}
	if cache.reserveCall.accountID != 31 {
		t.Fatalf("accountID=%d want=31", cache.reserveCall.accountID)
	}
	if cache.reserveCall.limit != 6.5 {
		t.Fatalf("limit=%v want=6.5", cache.reserveCall.limit)
	}
	if cache.reserveCall.cost <= 0 {
		t.Fatalf("expected positive estimated reservation cost, got %v", cache.reserveCall.cost)
	}
	if cache.reserveCall.ttl != windowCostReservationTTL {
		t.Fatalf("ttl=%s want=%s", cache.reserveCall.ttl, windowCostReservationTTL)
	}
}

func TestIncrementAccountRPM_PropagatesCacheError(t *testing.T) {
	svc := &GatewayService{
		rpmCache: &rpmCacheRuntimeLimitsStub{
			incrementErr: errors.New("boom"),
		},
	}

	err := svc.IncrementAccountRPM(context.Background(), 21)
	if err == nil || err.Error() != "boom" {
		t.Fatalf("err=%v want boom", err)
	}
}
