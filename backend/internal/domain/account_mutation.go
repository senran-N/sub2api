package domain

import "time"

// AccountBulkUpdate describes the fields that can be updated in a bulk operation.
// Nil pointers mean "do not change".
type AccountBulkUpdate struct {
	Name           *string
	ProxyID        *int64
	Concurrency    *int
	Priority       *int
	RateMultiplier *float64
	LoadFactor     *int
	Status         *string
	Schedulable    *bool
	Credentials    map[string]any
	Extra          map[string]any
}

type CreateAccountRequest struct {
	Name               string
	Notes              *string
	Platform           string
	Type               string
	Credentials        map[string]any
	Extra              map[string]any
	ProxyID            *int64
	Concurrency        int
	Priority           int
	GroupIDs           []int64
	ExpiresAt          *time.Time
	AutoPauseOnExpired *bool
}

type UpdateAccountRequest struct {
	Name               *string
	Notes              *string
	Credentials        *map[string]any
	Extra              *map[string]any
	ProxyID            *int64
	Concurrency        *int
	Priority           *int
	Status             *string
	GroupIDs           *[]int64
	ExpiresAt          *time.Time
	AutoPauseOnExpired *bool
}

type CreateAccountInput struct {
	Name                  string
	Notes                 *string
	Platform              string
	Type                  string
	Credentials           map[string]any
	Extra                 map[string]any
	ProxyID               *int64
	Concurrency           int
	Priority              int
	RateMultiplier        *float64
	LoadFactor            *int
	GroupIDs              []int64
	ExpiresAt             *int64
	AutoPauseOnExpired    *bool
	SkipDefaultGroupBind  bool
	SkipMixedChannelCheck bool
}

type UpdateAccountInput struct {
	Name                  string
	Notes                 *string
	Type                  string
	Credentials           map[string]any
	Extra                 map[string]any
	ProxyID               *int64
	Concurrency           *int
	Priority              *int
	RateMultiplier        *float64
	LoadFactor            *int
	Status                string
	GroupIDs              *[]int64
	ExpiresAt             *int64
	AutoPauseOnExpired    *bool
	SkipMixedChannelCheck bool
}

type BulkUpdateAccountsInput struct {
	AccountIDs            []int64
	Name                  string
	ProxyID               *int64
	Concurrency           *int
	Priority              *int
	RateMultiplier        *float64
	LoadFactor            *int
	Status                string
	Schedulable           *bool
	GroupIDs              *[]int64
	Credentials           map[string]any
	Extra                 map[string]any
	SkipMixedChannelCheck bool
}

type BulkUpdateAccountResult struct {
	AccountID int64  `json:"account_id"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
}

type BulkUpdateAccountsResult struct {
	Success    int                       `json:"success"`
	Failed     int                       `json:"failed"`
	SuccessIDs []int64                   `json:"success_ids"`
	FailedIDs  []int64                   `json:"failed_ids"`
	Results    []BulkUpdateAccountResult `json:"results"`
}
