package service

import (
	"context"
	"time"
)

const (
	windowCostReservationTTL          = 45 * time.Second
	windowCostEstimateSafetyFactor    = 1.15
	defaultEstimatedOutputTokens      = 1024
	thinkingEstimatedOutputTokenFloor = 2048
)

type WindowCostReservation struct {
	AccountID      int64
	WindowStart    time.Time
	ReservationID  string
	EstimatedCost  float64
	ReservationTTL time.Duration
}

func windowCostReservationLimit(account *Account, allowStickyOverflow bool) float64 {
	if account == nil {
		return 0
	}
	limit := account.GetWindowCostLimit()
	if limit <= 0 {
		return 0
	}
	if allowStickyOverflow {
		limit += account.GetWindowCostStickyReserve()
	}
	return limit
}

func (s *GatewayService) TryReserveAccountWindowCost(
	ctx context.Context,
	account *Account,
	apiKey *APIKey,
	parsed *ParsedRequest,
	billingModel string,
	allowStickyOverflow bool,
) (*WindowCostReservation, bool, float64, error) {
	if account == nil || apiKey == nil || parsed == nil || s.sessionLimitCache == nil || s.billingService == nil {
		return nil, true, 0, nil
	}
	if !account.IsAnthropicOAuthOrSetupToken() {
		return nil, true, 0, nil
	}

	limit := windowCostReservationLimit(account, allowStickyOverflow)
	if limit <= 0 {
		return nil, true, 0, nil
	}

	estimatedCost, err := s.estimateAdmissionWindowCost(ctx, apiKey, parsed, billingModel)
	if err != nil {
		return nil, true, 0, err
	}
	if estimatedCost <= 0 {
		return nil, true, 0, nil
	}

	reservation := &WindowCostReservation{
		AccountID:      account.ID,
		WindowStart:    account.GetCurrentWindowStartTime(),
		ReservationID:  generateRequestID(),
		EstimatedCost:  estimatedCost,
		ReservationTTL: windowCostReservationTTL,
	}
	allowed, total, err := s.sessionLimitCache.ReserveWindowCost(
		ctx,
		account.ID,
		reservation.WindowStart,
		reservation.ReservationID,
		reservation.EstimatedCost,
		limit,
		reservation.ReservationTTL,
	)
	if err != nil {
		return nil, true, 0, err
	}
	if !allowed {
		return nil, false, total, nil
	}
	return reservation, true, total, nil
}

func (s *GatewayService) ReleaseWindowCostReservation(ctx context.Context, reservation *WindowCostReservation) error {
	if reservation == nil || s == nil || s.sessionLimitCache == nil {
		return nil
	}
	return s.sessionLimitCache.ReleaseWindowCost(ctx, reservation.AccountID, reservation.WindowStart, reservation.ReservationID)
}

func (s *GatewayService) estimateAdmissionWindowCost(
	ctx context.Context,
	apiKey *APIKey,
	parsed *ParsedRequest,
	billingModel string,
) (float64, error) {
	if parsed == nil || apiKey == nil || s.billingService == nil {
		return 0, nil
	}
	if billingModel == "" {
		billingModel = parsed.Model
	}
	if billingModel == "" {
		return 0, nil
	}

	tokens := UsageTokens{
		InputTokens:  s.estimateParsedRequestInputTokens(parsed),
		OutputTokens: estimatedOutputTokensForAdmission(parsed),
	}
	if tokens.InputTokens <= 0 && tokens.OutputTokens <= 0 {
		return 0, nil
	}

	var (
		cost *CostBreakdown
		err  error
	)
	if resolved := resolveChannelPricing(ctx, s.resolver, apiKey, billingModel); resolved != nil {
		groupID := apiKey.Group.ID
		cost, err = s.billingService.CalculateCostUnified(CostInput{
			Ctx:            ctx,
			Model:          billingModel,
			GroupID:        &groupID,
			Tokens:         tokens,
			RequestCount:   1,
			RateMultiplier: 1.0,
			Resolver:       s.resolver,
			Resolved:       resolved,
		})
	} else {
		cost, err = s.billingService.CalculateCost(billingModel, tokens, 1.0)
	}
	if err != nil || cost == nil || cost.TotalCost <= 0 {
		return 0, err
	}

	return cost.TotalCost * windowCostEstimateSafetyFactor, nil
}

func estimatedOutputTokensForAdmission(parsed *ParsedRequest) int {
	if parsed == nil {
		return 0
	}
	if parsed.MaxTokens > 0 {
		return parsed.MaxTokens
	}
	if parsed.ThinkingEnabled {
		return thinkingEstimatedOutputTokenFloor
	}
	return defaultEstimatedOutputTokens
}

func (s *GatewayService) estimateParsedRequestInputTokens(parsed *ParsedRequest) int {
	if s == nil || parsed == nil {
		return 0
	}

	total := estimateTokensForText(s.extractTextFromSystem(parsed.System))
	for _, msg := range parsed.Messages {
		msgMap, ok := msg.(map[string]any)
		if !ok {
			continue
		}
		if content, exists := msgMap["content"]; exists {
			total += estimateTokensForText(s.extractTextFromContent(content))
			continue
		}
		if parts, ok := msgMap["parts"].([]any); ok {
			for _, part := range parts {
				partMap, ok := part.(map[string]any)
				if !ok {
					continue
				}
				if text, ok := partMap["text"].(string); ok {
					total += estimateTokensForText(text)
				}
			}
		}
	}
	if total < 0 {
		return 0
	}
	return total
}
