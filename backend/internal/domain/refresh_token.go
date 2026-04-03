package domain

import (
	"errors"
	"time"
)

// ErrRefreshTokenNotFound is returned when a refresh token is not found in cache.
var ErrRefreshTokenNotFound = errors.New("refresh token not found")

// RefreshTokenData is the persisted Redis payload for refresh token state.
type RefreshTokenData struct {
	UserID       int64     `json:"user_id"`
	TokenVersion int64     `json:"token_version"`
	FamilyID     string    `json:"family_id"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}
