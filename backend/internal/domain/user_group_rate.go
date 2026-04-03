package domain

type GroupRateMultiplierInput struct {
	UserID         int64   `json:"user_id"`
	RateMultiplier float64 `json:"rate_multiplier"`
}
