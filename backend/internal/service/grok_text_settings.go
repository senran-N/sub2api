package service

import "context"

type GrokTextSettings struct {
	ThinkingSummary   bool `json:"grok_thinking_summary"`
	ShowSearchSources bool `json:"grok_show_search_sources"`
}

func DefaultGrokTextSettings() GrokTextSettings {
	return GrokTextSettings{
		ThinkingSummary:   false,
		ShowSearchSources: false,
	}
}

func (s *SettingService) GetGrokTextSettings(ctx context.Context) GrokTextSettings {
	defaults := DefaultGrokTextSettings()
	if s == nil || s.settingRepo == nil {
		return defaults
	}

	values, err := s.settingRepo.GetMultiple(ctx, []string{
		SettingKeyGrokThinkingSummary,
		SettingKeyGrokShowSearchSources,
	})
	if err != nil {
		return defaults
	}

	return GrokTextSettings{
		ThinkingSummary:   values[SettingKeyGrokThinkingSummary] == "true",
		ShowSearchSources: values[SettingKeyGrokShowSearchSources] == "true",
	}
}
