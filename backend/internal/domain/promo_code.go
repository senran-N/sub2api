package domain

import "time"

type CreatePromoCodeInput struct {
	Code        string
	BonusAmount float64
	MaxUses     int
	ExpiresAt   *time.Time
	Notes       string
}

type UpdatePromoCodeInput struct {
	Code        *string
	BonusAmount *float64
	MaxUses     *int
	Status      *string
	ExpiresAt   *time.Time
	Notes       *string
}
