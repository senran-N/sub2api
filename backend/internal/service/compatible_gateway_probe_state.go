package service

import (
	"context"
	"net/http"
	"time"
)

func (s *AccountTestService) persistCompatibleGatewayProbeState(ctx context.Context, account *Account, modelID string, resp *http.Response, probeErr error, isOAuth bool) {
	if s == nil || s.accountRepo == nil || account == nil {
		return
	}
	if NormalizeCompatibleGatewayPlatform(account.Platform) == PlatformGrok {
		s.getGrokAccountStateService().PersistProbeResult(ctx, account, modelID, resp, probeErr)
		return
	}

	updates, err := buildCompatibleGatewayProbeStateUpdates(account, modelID, resp, probeErr, isOAuth, time.Now().UTC())
	if err != nil || len(updates) == 0 {
		return
	}

	_ = s.accountRepo.UpdateExtra(ctx, account.ID, updates)
	mergeAccountExtra(account, updates)
}

func (s *AccountTestService) getGrokAccountStateService() *GrokAccountStateService {
	if s == nil || s.accountRepo == nil {
		return nil
	}
	if s.grokAccountStateService == nil {
		s.grokAccountStateService = NewGrokAccountStateService(s.accountRepo)
	}
	return s.grokAccountStateService
}

func buildCompatibleGatewayProbeStateUpdates(account *Account, modelID string, resp *http.Response, probeErr error, isOAuth bool, now time.Time) (map[string]any, error) {
	if account == nil {
		return nil, nil
	}

	switch NormalizeCompatibleGatewayPlatform(account.Platform) {
	case PlatformOpenAI:
		if !isOAuth {
			return nil, nil
		}
		return extractOpenAICodexProbeUpdates(resp)
	case PlatformGrok:
		return buildGrokProbeStateExtraUpdates(account, modelID, resp, probeErr, now), nil
	default:
		return nil, nil
	}
}
