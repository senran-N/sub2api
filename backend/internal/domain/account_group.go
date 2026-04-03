package domain

import "time"

// AccountGroupLink is the persisted relationship between one account and one group.
type AccountGroupLink struct {
	AccountID int64
	GroupID   int64
	Priority  int
	CreatedAt time.Time
}
