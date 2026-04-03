package domain

// AccountWithConcurrency carries the runtime concurrency limit for an account.
type AccountWithConcurrency struct {
	ID             int64
	MaxConcurrency int
}

// UserWithConcurrency carries the runtime concurrency limit for a user.
type UserWithConcurrency struct {
	ID             int64
	MaxConcurrency int
}

// AccountLoadInfo is a snapshot of current account runtime load.
type AccountLoadInfo struct {
	AccountID          int64
	CurrentConcurrency int
	WaitingCount       int
	LoadRate           int
}

// UserLoadInfo is a snapshot of current user runtime load.
type UserLoadInfo struct {
	UserID             int64
	CurrentConcurrency int
	WaitingCount       int
	LoadRate           int
}
