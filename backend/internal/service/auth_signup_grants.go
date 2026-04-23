package service

import (
	"context"
	"strings"

	"github.com/senran-N/sub2api/internal/pkg/logger"
)

type signupGrantPlan struct {
	Balance       float64
	Concurrency   int
	Subscriptions []DefaultSubscriptionSetting
}

func (s *AuthService) assignSubscriptions(ctx context.Context, userID int64, items []DefaultSubscriptionSetting, notes string) {
	if s == nil || s.settingService == nil || s.defaultSubAssigner == nil || userID <= 0 {
		return
	}
	for _, item := range items {
		if _, _, err := s.defaultSubAssigner.AssignOrExtendSubscription(ctx, &AssignSubscriptionInput{
			UserID:       userID,
			GroupID:      item.GroupID,
			ValidityDays: item.ValidityDays,
			Notes:        notes,
		}); err != nil {
			logger.LegacyPrintf("service.auth", "[Auth] Failed to assign default subscription: user_id=%d group_id=%d err=%v", userID, item.GroupID, err)
		}
	}
}

func (s *AuthService) resolveSignupGrantPlan(ctx context.Context, signupSource string) signupGrantPlan {
	plan := signupGrantPlan{}
	if s != nil && s.cfg != nil {
		plan.Balance = s.cfg.Default.UserBalance
		plan.Concurrency = s.cfg.Default.UserConcurrency
	}
	if s == nil || s.settingService == nil {
		return plan
	}

	plan.Balance = s.settingService.GetDefaultBalance(ctx)
	plan.Concurrency = s.settingService.GetDefaultConcurrency(ctx)
	plan.Subscriptions = s.settingService.GetDefaultSubscriptions(ctx)

	resolved, enabled, err := s.settingService.ResolveAuthSourceGrantSettings(ctx, signupSource, false)
	if err != nil {
		logger.LegacyPrintf("service.auth", "[Auth] Failed to load auth source signup defaults for %s: %v", signupSource, err)
		return plan
	}
	if !enabled {
		return plan
	}
	plan.Balance = resolved.Balance
	plan.Concurrency = resolved.Concurrency
	plan.Subscriptions = resolved.Subscriptions
	return plan
}

func authSourceSignupSettings(defaults *AuthSourceDefaultSettings, signupSource string) (ProviderDefaultGrantSettings, bool) {
	if defaults == nil {
		return ProviderDefaultGrantSettings{}, false
	}
	switch strings.ToLower(strings.TrimSpace(signupSource)) {
	case "email":
		return defaults.Email, true
	case "linuxdo":
		return defaults.LinuxDo, true
	case "oidc":
		return defaults.OIDC, true
	case "wechat":
		return defaults.WeChat, true
	default:
		return ProviderDefaultGrantSettings{}, false
	}
}

func inferLegacySignupSource(email string) string {
	normalized := strings.ToLower(strings.TrimSpace(email))
	switch {
	case strings.HasSuffix(normalized, LinuxDoConnectSyntheticEmailDomain):
		return "linuxdo"
	case strings.HasSuffix(normalized, OIDCConnectSyntheticEmailDomain):
		return "oidc"
	case strings.HasSuffix(normalized, WeChatConnectSyntheticEmailDomain):
		return "wechat"
	default:
		return "email"
	}
}
