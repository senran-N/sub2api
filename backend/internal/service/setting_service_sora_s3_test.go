//go:build unit

package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

type settingSoraS3RepoStub struct {
	values   map[string]string
	updates  map[string]string
	getValue error
}

func (s *settingSoraS3RepoStub) Get(ctx context.Context, key string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *settingSoraS3RepoStub) GetValue(ctx context.Context, key string) (string, error) {
	if s.getValue != nil {
		return "", s.getValue
	}
	value, ok := s.values[key]
	if !ok {
		return "", ErrSettingNotFound
	}
	return value, nil
}

func (s *settingSoraS3RepoStub) Set(ctx context.Context, key, value string) error {
	panic("unexpected Set call")
}

func (s *settingSoraS3RepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	result := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := s.values[key]; ok {
			result[key] = value
		}
	}
	return result, nil
}

func (s *settingSoraS3RepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	s.updates = make(map[string]string, len(settings))
	for key, value := range settings {
		s.updates[key] = value
	}
	return nil
}

func (s *settingSoraS3RepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *settingSoraS3RepoStub) Delete(ctx context.Context, key string) error {
	panic("unexpected Delete call")
}

func TestSettingService_ListSoraS3Profiles_BackfillsLegacySettings(t *testing.T) {
	repo := &settingSoraS3RepoStub{
		values: map[string]string{
			SettingKeySoraS3Enabled:                "true",
			SettingKeySoraS3Endpoint:               " https://s3.example.com ",
			SettingKeySoraS3Region:                 " us-east-1 ",
			SettingKeySoraS3Bucket:                 " media-bucket ",
			SettingKeySoraS3AccessKeyID:            " AKID ",
			SettingKeySoraS3SecretAccessKey:        "SECRET",
			SettingKeySoraS3Prefix:                 " sora/ ",
			SettingKeySoraS3ForcePathStyle:         "true",
			SettingKeySoraS3CDNURL:                 " https://cdn.example.com ",
			SettingKeySoraDefaultStorageQuotaBytes: "2048",
		},
	}
	svc := NewSettingService(repo, nil)

	profiles, err := svc.ListSoraS3Profiles(context.Background())
	require.NoError(t, err)
	require.Equal(t, "default", profiles.ActiveProfileID)
	require.Len(t, profiles.Items, 1)

	profile := profiles.Items[0]
	require.Equal(t, "default", profile.ProfileID)
	require.True(t, profile.IsActive)
	require.True(t, profile.Enabled)
	require.Equal(t, "https://s3.example.com", profile.Endpoint)
	require.Equal(t, "us-east-1", profile.Region)
	require.Equal(t, "media-bucket", profile.Bucket)
	require.Equal(t, "AKID", profile.AccessKeyID)
	require.True(t, profile.SecretAccessKeyConfigured)
	require.Equal(t, "sora/", profile.Prefix)
	require.True(t, profile.ForcePathStyle)
	require.Equal(t, "https://cdn.example.com", profile.CDNURL)
	require.Equal(t, int64(2048), profile.DefaultStorageQuotaBytes)
}

func TestSettingService_SetSoraS3Settings_CreatesDefaultProfileAndSyncsLegacyKeys(t *testing.T) {
	repo := &settingSoraS3RepoStub{
		values: map[string]string{
			SettingKeySoraS3Profiles: "",
		},
	}
	svc := NewSettingService(repo, nil)
	updateCalls := 0
	s3UpdateCalls := 0
	svc.SetOnUpdateCallback(func() { updateCalls++ })
	svc.SetOnS3UpdateCallback(func() { s3UpdateCalls++ })

	err := svc.SetSoraS3Settings(context.Background(), &SoraS3Settings{
		Enabled:                  true,
		Endpoint:                 " https://s3.example.com ",
		Region:                   " us-west-1 ",
		Bucket:                   " uploads ",
		AccessKeyID:              " AKID ",
		SecretAccessKey:          "SECRET",
		Prefix:                   " sora-assets/ ",
		ForcePathStyle:           true,
		CDNURL:                   " https://cdn.example.com ",
		DefaultStorageQuotaBytes: 4096,
	})
	require.NoError(t, err)
	require.Equal(t, 1, updateCalls)
	require.Equal(t, 1, s3UpdateCalls)

	require.Equal(t, "true", repo.updates[SettingKeySoraS3Enabled])
	require.Equal(t, "https://s3.example.com", repo.updates[SettingKeySoraS3Endpoint])
	require.Equal(t, "us-west-1", repo.updates[SettingKeySoraS3Region])
	require.Equal(t, "uploads", repo.updates[SettingKeySoraS3Bucket])
	require.Equal(t, "AKID", repo.updates[SettingKeySoraS3AccessKeyID])
	require.Equal(t, "SECRET", repo.updates[SettingKeySoraS3SecretAccessKey])
	require.Equal(t, "sora-assets/", repo.updates[SettingKeySoraS3Prefix])
	require.Equal(t, "true", repo.updates[SettingKeySoraS3ForcePathStyle])
	require.Equal(t, "https://cdn.example.com", repo.updates[SettingKeySoraS3CDNURL])
	require.Equal(t, "4096", repo.updates[SettingKeySoraDefaultStorageQuotaBytes])

	var store soraS3ProfilesStore
	require.NoError(t, json.Unmarshal([]byte(repo.updates[SettingKeySoraS3Profiles]), &store))
	require.Equal(t, "default", store.ActiveProfileID)
	require.Len(t, store.Items, 1)
	require.Equal(t, "Default", store.Items[0].Name)
	require.Equal(t, "uploads", store.Items[0].Bucket)
	require.Equal(t, int64(4096), store.Items[0].DefaultStorageQuotaBytes)
}
