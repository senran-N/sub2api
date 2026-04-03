package domain

type AssignSubscriptionInput struct {
	UserID       int64
	GroupID      int64
	ValidityDays int
	AssignedBy   int64
	Notes        string
}

type BulkAssignSubscriptionInput struct {
	UserIDs      []int64
	GroupID      int64
	ValidityDays int
	AssignedBy   int64
	Notes        string
}
