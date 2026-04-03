package service

import (
	"github.com/senran-N/sub2api/internal/domain"
	"github.com/senran-N/sub2api/internal/ports"
)

type IdempotencyRecord = domain.IdempotencyRecord
type IdempotencyRepository = ports.IdempotencyRepository

const (
	IdempotencyStatusProcessing      = domain.IdempotencyStatusProcessing
	IdempotencyStatusSucceeded       = domain.IdempotencyStatusSucceeded
	IdempotencyStatusFailedRetryable = domain.IdempotencyStatusFailedRetryable
)
