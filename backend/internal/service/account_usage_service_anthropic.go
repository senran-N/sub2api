package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// fetchOAuthUsageRaw 从 Anthropic API 获取原始响应（不构建 UsageInfo）。
// 如果账号开启了 TLS 指纹，则使用 TLS 指纹伪装；
// 如果有缓存的 Fingerprint，则复用缓存的 User-Agent 等信息。
func (s *AccountUsageService) fetchOAuthUsageRaw(ctx context.Context, account *Account) (*ClaudeUsageResponse, error) {
	accessToken := account.GetCredential("access_token")
	if accessToken == "" {
		return nil, fmt.Errorf("no access token available")
	}

	var proxyURL string
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}

	opts := &ClaudeUsageFetchOptions{
		AccessToken: accessToken,
		ProxyURL:    proxyURL,
		AccountID:   account.ID,
		TLSProfile:  s.tlsFPProfileService.ResolveTLSProfile(account),
	}

	if s.identityCache != nil {
		if fp, err := s.identityCache.GetFingerprint(ctx, account.ID); err == nil && fp != nil {
			opts.Fingerprint = fp
		}
	}

	return s.usageFetcher.FetchUsageWithOptions(ctx, opts)
}

func parseTime(value string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
	}
	for _, format := range formats {
		if parsed, err := time.Parse(format, value); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse time: %s", value)
}

// buildUsageInfo 构建 Anthropic usage 响应的统一领域模型。
func (s *AccountUsageService) buildUsageInfo(resp *ClaudeUsageResponse, updatedAt *time.Time) *UsageInfo {
	info := &UsageInfo{UpdatedAt: updatedAt}

	info.FiveHour = &UsageProgress{
		Utilization: resp.FiveHour.Utilization,
	}
	if resp.FiveHour.ResetsAt != "" {
		if resetAt, err := parseTime(resp.FiveHour.ResetsAt); err == nil {
			info.FiveHour.ResetsAt = &resetAt
			info.FiveHour.RemainingSeconds = int(time.Until(resetAt).Seconds())
		} else {
			log.Printf("Failed to parse FiveHour.ResetsAt: %s, error: %v", resp.FiveHour.ResetsAt, err)
		}
	}

	if resp.SevenDay.ResetsAt != "" {
		if resetAt, err := parseTime(resp.SevenDay.ResetsAt); err == nil {
			info.SevenDay = &UsageProgress{
				Utilization:      resp.SevenDay.Utilization,
				ResetsAt:         &resetAt,
				RemainingSeconds: int(time.Until(resetAt).Seconds()),
			}
		} else {
			log.Printf("Failed to parse SevenDay.ResetsAt: %s, error: %v", resp.SevenDay.ResetsAt, err)
			info.SevenDay = &UsageProgress{
				Utilization: resp.SevenDay.Utilization,
			}
		}
	}

	if resp.SevenDaySonnet.ResetsAt != "" {
		if resetAt, err := parseTime(resp.SevenDaySonnet.ResetsAt); err == nil {
			info.SevenDaySonnet = &UsageProgress{
				Utilization:      resp.SevenDaySonnet.Utilization,
				ResetsAt:         &resetAt,
				RemainingSeconds: int(time.Until(resetAt).Seconds()),
			}
		} else {
			log.Printf("Failed to parse SevenDaySonnet.ResetsAt: %s, error: %v", resp.SevenDaySonnet.ResetsAt, err)
			info.SevenDaySonnet = &UsageProgress{
				Utilization: resp.SevenDaySonnet.Utilization,
			}
		}
	}

	return info
}

// estimateSetupTokenUsage 根据 session_window 推算 Setup Token 账号的使用量。
func (s *AccountUsageService) estimateSetupTokenUsage(account *Account) *UsageInfo {
	info := &UsageInfo{}

	if account.SessionWindowEnd != nil {
		remaining := int(time.Until(*account.SessionWindowEnd).Seconds())
		if remaining < 0 {
			remaining = 0
		}

		var utilization float64
		var found bool
		if stored, ok := account.Extra["session_window_utilization"]; ok {
			switch value := stored.(type) {
			case float64:
				utilization = value * 100
				found = true
			case json.Number:
				if parsed, err := value.Float64(); err == nil {
					utilization = parsed * 100
					found = true
				}
			}
		}

		if !found {
			switch account.SessionWindowStatus {
			case "rejected":
				utilization = 100.0
			case "allowed_warning":
				utilization = 80.0
			}
		}

		info.FiveHour = &UsageProgress{
			Utilization:      utilization,
			ResetsAt:         account.SessionWindowEnd,
			RemainingSeconds: remaining,
		}
	} else {
		info.FiveHour = &UsageProgress{
			Utilization:      0,
			RemainingSeconds: 0,
		}
	}

	return info
}
