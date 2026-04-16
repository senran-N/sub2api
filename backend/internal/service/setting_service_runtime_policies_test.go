package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type jsonSettingRepoStub struct {
	value string
}

func (s *jsonSettingRepoStub) Get(ctx context.Context, key string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *jsonSettingRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	if s.value == "" {
		return "", ErrSettingNotFound
	}
	return s.value, nil
}

func (s *jsonSettingRepoStub) Set(ctx context.Context, key, value string) error {
	s.value = value
	return nil
}

func (s *jsonSettingRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	panic("unexpected GetMultiple call")
}

func (s *jsonSettingRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *jsonSettingRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *jsonSettingRepoStub) Delete(ctx context.Context, key string) error {
	panic("unexpected Delete call")
}

func TestSettingService_SetBetaPolicySettings_NormalizesModelWhitelistAndFallback(t *testing.T) {
	repo := &jsonSettingRepoStub{}
	svc := NewSettingService(repo, nil)

	err := svc.SetBetaPolicySettings(context.Background(), &BetaPolicySettings{
		Rules: []BetaPolicyRule{
			{
				BetaToken:            " context-1m-2025-08-07 ",
				Action:               BetaPolicyActionFilter,
				Scope:                BetaPolicyScopeAll,
				ModelWhitelist:       []string{" claude-opus-* ", "", "claude-opus-4-1"},
				FallbackErrorMessage: " should-clear ",
			},
		},
	})
	require.NoError(t, err)

	settings, err := svc.GetBetaPolicySettings(context.Background())
	require.NoError(t, err)
	require.Len(t, settings.Rules, 1)
	require.Equal(t, "context-1m-2025-08-07", settings.Rules[0].BetaToken)
	require.Equal(t, []string{"claude-opus-*", "claude-opus-4-1"}, settings.Rules[0].ModelWhitelist)
	require.Equal(t, BetaPolicyActionPass, settings.Rules[0].FallbackAction)
	require.Empty(t, settings.Rules[0].FallbackErrorMessage)
}
