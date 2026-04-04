package service

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/senran-N/sub2api/internal/config"
)

const (
	opsAlertEvaluatorJobName = "ops_alert_evaluator"

	opsAlertEvaluatorTimeout         = 45 * time.Second
	opsAlertEvaluatorLeaderLockKey   = "ops:alert:evaluator:leader"
	opsAlertEvaluatorLeaderLockTTL   = 90 * time.Second
	opsAlertEvaluatorSkipLogInterval = 1 * time.Minute
)

var opsAlertEvaluatorReleaseScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
  return redis.call("DEL", KEYS[1])
end
return 0
`)

type OpsAlertEvaluatorService struct {
	opsService   *OpsService
	opsRepo      OpsRepository
	emailService *EmailService

	redisClient *redis.Client
	cfg         *config.Config
	instanceID  string

	stopCh    chan struct{}
	startOnce sync.Once
	stopOnce  sync.Once
	wg        sync.WaitGroup

	mu         sync.Mutex
	ruleStates map[int64]*opsAlertRuleState

	emailLimiter *slidingWindowLimiter

	skipLogMu sync.Mutex
	skipLogAt time.Time

	warnNoRedisOnce sync.Once
}

type opsAlertRuleState struct {
	LastEvaluatedAt     time.Time
	ConsecutiveBreaches int
}

func NewOpsAlertEvaluatorService(
	opsService *OpsService,
	opsRepo OpsRepository,
	emailService *EmailService,
	redisClient *redis.Client,
	cfg *config.Config,
) *OpsAlertEvaluatorService {
	return &OpsAlertEvaluatorService{
		opsService:   opsService,
		opsRepo:      opsRepo,
		emailService: emailService,
		redisClient:  redisClient,
		cfg:          cfg,
		instanceID:   uuid.NewString(),
		ruleStates:   map[int64]*opsAlertRuleState{},
		emailLimiter: newSlidingWindowLimiter(0, time.Hour),
	}
}
