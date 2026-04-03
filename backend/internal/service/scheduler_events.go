package service

import "github.com/senran-N/sub2api/internal/domain"

const (
	SchedulerOutboxEventAccountChanged       = domain.SchedulerOutboxEventAccountChanged
	SchedulerOutboxEventAccountGroupsChanged = domain.SchedulerOutboxEventAccountGroupsChanged
	SchedulerOutboxEventAccountBulkChanged   = domain.SchedulerOutboxEventAccountBulkChanged
	SchedulerOutboxEventAccountLastUsed      = domain.SchedulerOutboxEventAccountLastUsed
	SchedulerOutboxEventGroupChanged         = domain.SchedulerOutboxEventGroupChanged
	SchedulerOutboxEventFullRebuild          = domain.SchedulerOutboxEventFullRebuild
)
