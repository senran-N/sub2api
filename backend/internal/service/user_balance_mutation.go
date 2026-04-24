package service

type UserBalanceMutationInput struct {
	UserID    int64
	Amount    float64
	Operation string
}

type UserBalanceMutationResult struct {
	User        *User
	BalanceDiff float64
}
