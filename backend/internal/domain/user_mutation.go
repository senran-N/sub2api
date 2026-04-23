package domain

type CreateUserInput struct {
	Email         string
	Password      string
	Username      string
	Notes         string
	Balance       float64
	Concurrency   int
	AllowedGroups []int64
}

type UpdateUserInput struct {
	Email         string
	Password      string
	Username      *string
	Notes         *string
	Balance       *float64
	Concurrency   *int
	Status        string
	AllowedGroups *[]int64
	GroupRates    map[int64]*float64
}

type ReplaceUserGroupResult struct {
	MigratedKeys int64
}

// UpdateProfileRequest describes a user profile mutation.
type UpdateProfileRequest struct {
	Email                      *string  `json:"email"`
	Username                   *string  `json:"username"`
	Concurrency                *int     `json:"concurrency"`
	BalanceNotifyEnabled       *bool    `json:"balance_notify_enabled"`
	BalanceNotifyThreshold     *float64 `json:"balance_notify_threshold"`
	BalanceNotifyThresholdType *string  `json:"balance_notify_threshold_type"`
}

// ChangePasswordRequest describes a user password change request.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}
