package service

import (
	"context"
	"sync"
	"time"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/pkg/logger"
	"github.com/robfig/cron/v3"
)

const (
	scheduledTestDefaultMaxWorkers = 10
	scheduledTestDefaultStartDelay = 10 * time.Second
	scheduledTestDefaultRunTimeout = 5 * time.Minute
	scheduledTestDefaultStopWait   = 3 * time.Second
)

// ScheduledTestRunnerService periodically scans due test plans and executes them.
type ScheduledTestRunnerService struct {
	planRepo       ScheduledTestPlanRepository
	scheduledSvc   *ScheduledTestService
	accountTestSvc *AccountTestService
	rateLimitSvc   *RateLimitService
	cfg            *config.Config

	startDelay time.Duration
	runTimeout time.Duration
	stopWait   time.Duration
	maxWorkers int

	sleep          func(time.Duration)
	now            func() time.Time
	listDue        func(context.Context, time.Time) ([]*ScheduledTestPlan, error)
	runTest        func(context.Context, int64, string) (*ScheduledTestResult, error)
	saveResult     func(context.Context, int64, int, *ScheduledTestResult) error
	updateAfterRun func(context.Context, int64, time.Time, time.Time) error
	recoverAccount func(context.Context, int64) (*SuccessfulTestRecoveryResult, error)

	cron      *cron.Cron
	startOnce sync.Once
	stopOnce  sync.Once
}

// NewScheduledTestRunnerService creates a new runner.
func NewScheduledTestRunnerService(
	planRepo ScheduledTestPlanRepository,
	scheduledSvc *ScheduledTestService,
	accountTestSvc *AccountTestService,
	rateLimitSvc *RateLimitService,
	cfg *config.Config,
) *ScheduledTestRunnerService {
	svc := &ScheduledTestRunnerService{
		planRepo:       planRepo,
		scheduledSvc:   scheduledSvc,
		accountTestSvc: accountTestSvc,
		rateLimitSvc:   rateLimitSvc,
		cfg:            cfg,
		startDelay:     scheduledTestDefaultStartDelay,
		runTimeout:     scheduledTestDefaultRunTimeout,
		stopWait:       scheduledTestDefaultStopWait,
		maxWorkers:     scheduledTestDefaultMaxWorkers,
		sleep:          time.Sleep,
		now:            time.Now,
	}
	svc.listDue = func(ctx context.Context, now time.Time) ([]*ScheduledTestPlan, error) {
		return svc.planRepo.ListDue(ctx, now)
	}
	svc.runTest = func(ctx context.Context, accountID int64, modelID string) (*ScheduledTestResult, error) {
		return svc.accountTestSvc.RunTestBackground(ctx, accountID, modelID)
	}
	svc.saveResult = func(ctx context.Context, planID int64, maxResults int, result *ScheduledTestResult) error {
		return svc.scheduledSvc.SaveResult(ctx, planID, maxResults, result)
	}
	svc.updateAfterRun = func(ctx context.Context, id int64, lastRunAt time.Time, nextRunAt time.Time) error {
		return svc.planRepo.UpdateAfterRun(ctx, id, lastRunAt, nextRunAt)
	}
	svc.recoverAccount = func(ctx context.Context, accountID int64) (*SuccessfulTestRecoveryResult, error) {
		if svc.rateLimitSvc == nil {
			return nil, nil
		}
		return svc.rateLimitSvc.RecoverAccountAfterSuccessfulTest(ctx, accountID)
	}
	return svc
}

// Start begins the cron ticker (every minute).
func (s *ScheduledTestRunnerService) Start() {
	if s == nil {
		return
	}
	s.startOnce.Do(func() {
		loc := time.Local
		if s.cfg != nil {
			if parsed, err := time.LoadLocation(s.cfg.Timezone); err == nil && parsed != nil {
				loc = parsed
			}
		}

		c := cron.New(cron.WithParser(scheduledTestCronParser), cron.WithLocation(loc))
		_, err := c.AddFunc("* * * * *", func() { s.runScheduled() })
		if err != nil {
			logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] not started (invalid schedule): %v", err)
			return
		}
		s.cron = c
		s.cron.Start()
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] started (tick=every minute)")
	})
}

// Stop gracefully shuts down the cron scheduler.
func (s *ScheduledTestRunnerService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		if s.cron != nil {
			ctx := s.cron.Stop()
			select {
			case <-ctx.Done():
			case <-time.After(s.stopWait):
				logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] cron stop timed out")
			}
		}
	})
}

func (s *ScheduledTestRunnerService) runScheduled() {
	// Delay 10s so execution lands at ~:10 of each minute instead of :00.
	s.sleep(s.startDelay)

	ctx, cancel := context.WithTimeout(context.Background(), s.runTimeout)
	defer cancel()

	now := s.now()
	plans, err := s.listDue(ctx, now)
	if err != nil {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] ListDue error: %v", err)
		return
	}
	if len(plans) == 0 {
		return
	}

	logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] found %d due plans", len(plans))

	sem := make(chan struct{}, s.maxWorkers)
	var wg sync.WaitGroup

	for _, plan := range plans {
		sem <- struct{}{}
		wg.Add(1)
		go func(p *ScheduledTestPlan) {
			defer wg.Done()
			defer func() { <-sem }()
			s.runOnePlan(ctx, p)
		}(plan)
	}

	wg.Wait()
}

func (s *ScheduledTestRunnerService) runOnePlan(ctx context.Context, plan *ScheduledTestPlan) {
	result, err := s.runTest(ctx, plan.AccountID, plan.ModelID)
	if err != nil {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] plan=%d RunTestBackground error: %v", plan.ID, err)
		return
	}

	if err := s.saveResult(ctx, plan.ID, plan.MaxResults, result); err != nil {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] plan=%d SaveResult error: %v", plan.ID, err)
	}

	// Auto-recover account if test succeeded and auto_recover is enabled.
	if result.Status == "success" && plan.AutoRecover {
		s.tryRecoverAccount(ctx, plan.AccountID, plan.ID)
	}

	nextRun, err := computeNextRun(plan.CronExpression, time.Now())
	if err != nil {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] plan=%d computeNextRun error: %v", plan.ID, err)
		return
	}

	if err := s.updateAfterRun(ctx, plan.ID, s.now(), nextRun); err != nil {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] plan=%d UpdateAfterRun error: %v", plan.ID, err)
	}
}

// tryRecoverAccount attempts to recover an account from recoverable runtime state.
func (s *ScheduledTestRunnerService) tryRecoverAccount(ctx context.Context, accountID int64, planID int64) {
	if s.rateLimitSvc == nil {
		return
	}

	recovery, err := s.recoverAccount(ctx, accountID)
	if err != nil {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] plan=%d auto-recover failed: %v", planID, err)
		return
	}
	if recovery == nil {
		return
	}

	if recovery.ClearedError {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] plan=%d auto-recover: account=%d recovered from error status", planID, accountID)
	}
	if recovery.ClearedRateLimit {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] plan=%d auto-recover: account=%d cleared rate-limit/runtime state", planID, accountID)
	}
}
