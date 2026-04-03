package service

import "github.com/senran-N/sub2api/internal/domain"

const (
	SchedulerModeSingle = domain.SchedulerModeSingle
	SchedulerModeMixed  = domain.SchedulerModeMixed
	SchedulerModeForced = domain.SchedulerModeForced
)

type SchedulerBucket = domain.SchedulerBucket

func ParseSchedulerBucket(raw string) (SchedulerBucket, bool) {
	return domain.ParseSchedulerBucket(raw)
}
