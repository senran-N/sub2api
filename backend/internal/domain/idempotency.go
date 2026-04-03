package domain

import "time"

const (
	IdempotencyStatusProcessing      = "processing"
	IdempotencyStatusSucceeded       = "succeeded"
	IdempotencyStatusFailedRetryable = "failed_retryable"
)

// IdempotencyRecord stores the persisted state for an idempotent request.
type IdempotencyRecord struct {
	ID                 int64
	Scope              string
	IdempotencyKeyHash string
	RequestFingerprint string
	Status             string
	ResponseStatus     *int
	ResponseBody       *string
	ErrorReason        *string
	LockedUntil        *time.Time
	ExpiresAt          time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
