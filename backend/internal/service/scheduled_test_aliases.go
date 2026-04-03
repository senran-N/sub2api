package service

import (
	"github.com/senran-N/sub2api/internal/domain"
	"github.com/senran-N/sub2api/internal/ports"
)

type ScheduledTestPlan = domain.ScheduledTestPlan
type ScheduledTestResult = domain.ScheduledTestResult

type ScheduledTestPlanRepository = ports.ScheduledTestPlanRepository
type ScheduledTestResultRepository = ports.ScheduledTestResultRepository
