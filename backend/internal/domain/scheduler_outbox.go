package domain

import "time"

const (
	SchedulerOutboxEventAccountChanged       = "account_changed"
	SchedulerOutboxEventAccountGroupsChanged = "account_groups_changed"
	SchedulerOutboxEventAccountBulkChanged   = "account_bulk_changed"
	SchedulerOutboxEventAccountLastUsed      = "account_last_used"
	SchedulerOutboxEventGroupChanged         = "group_changed"
	SchedulerOutboxEventFullRebuild          = "full_rebuild"
)

// SchedulerOutboxEvent represents a scheduler cache invalidation event.
type SchedulerOutboxEvent struct {
	ID        int64
	EventType string
	AccountID *int64
	GroupID   *int64
	Payload   map[string]any
	CreatedAt time.Time
}
