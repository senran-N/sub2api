package service

import (
	"github.com/senran-N/sub2api/internal/domain"
	"github.com/senran-N/sub2api/internal/ports"
)

type APIKeyAuthSnapshot = domain.APIKeyAuthSnapshot
type APIKeyAuthUserSnapshot = domain.APIKeyAuthUserSnapshot
type APIKeyAuthGroupSnapshot = domain.APIKeyAuthGroupSnapshot
type APIKeyAuthCacheEntry = domain.APIKeyAuthCacheEntry

type APIKeyCache = ports.APIKeyCache
