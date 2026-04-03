package service

import (
	"github.com/senran-N/sub2api/internal/ports"
)

// GeminiTokenCache stores short-lived access tokens and coordinates refresh to avoid stampedes.
type GeminiTokenCache = ports.GeminiTokenCache
