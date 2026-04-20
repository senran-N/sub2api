//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/pkg/grok"
	"github.com/stretchr/testify/require"
)

type grokFallbackSettingRepoStub struct {
	values map[string]string
}

func (s *grokFallbackSettingRepoStub) Get(ctx context.Context, key string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *grokFallbackSettingRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	return s.values[key], nil
}

func (s *grokFallbackSettingRepoStub) Set(ctx context.Context, key, value string) error {
	panic("unexpected Set call")
}

func (s *grokFallbackSettingRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	panic("unexpected GetMultiple call")
}

func (s *grokFallbackSettingRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *grokFallbackSettingRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	return s.values, nil
}

func (s *grokFallbackSettingRepoStub) Delete(ctx context.Context, key string) error {
	panic("unexpected Delete call")
}

func TestSettingServiceGetFallbackModel_UsesRegistryBackedGrokDefault(t *testing.T) {
	svc := NewSettingService(&grokFallbackSettingRepoStub{values: map[string]string{}}, &config.Config{})

	require.Equal(t, grok.DefaultFallbackModelID(), svc.GetFallbackModel(context.Background(), PlatformGrok))
}

func TestSettingServiceGetAllSettings_UsesRegistryBackedGrokDefault(t *testing.T) {
	svc := NewSettingService(&grokFallbackSettingRepoStub{values: map[string]string{}}, &config.Config{})

	settings, err := svc.GetAllSettings(context.Background())
	require.NoError(t, err)
	require.Equal(t, grok.DefaultFallbackModelID(), settings.FallbackModelGrok)
}
