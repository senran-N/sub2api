package service

import (
	"github.com/senran-N/sub2api/internal/domain"
	"github.com/senran-N/sub2api/internal/ports"
)

type ConcurrencyCache = ports.ConcurrencyCache

type AccountWithConcurrency = domain.AccountWithConcurrency
type UserWithConcurrency = domain.UserWithConcurrency
type AccountLoadInfo = domain.AccountLoadInfo
type UserLoadInfo = domain.UserLoadInfo
