package service

import (
	"testing"

	"github.com/senran-N/sub2api/internal/model"
	"github.com/stretchr/testify/require"
)

func TestTLSFingerprintProfileService_ResolveTLSProfile_StableForSameAccount(t *testing.T) {
	svc := &TLSFingerprintProfileService{
		localCache: map[int64]*model.TLSFingerprintProfile{
			3:  {ID: 3, Name: "profile-3"},
			7:  {ID: 7, Name: "profile-7"},
			11: {ID: 11, Name: "profile-11"},
		},
	}
	account := &Account{
		ID:       4242,
		Platform: PlatformAnthropic,
		Type:     AccountTypeOAuth,
		Extra: map[string]any{
			"enable_tls_fingerprint":     true,
			"tls_fingerprint_profile_id": -1,
		},
	}

	first := svc.ResolveTLSProfile(account)
	require.NotNil(t, first)
	require.NotEmpty(t, first.Name)
	for range 1000 {
		profile := svc.ResolveTLSProfile(account)
		require.NotNil(t, profile)
		require.Equal(t, first.Name, profile.Name)
	}
}

func TestTLSFingerprintProfileService_ResolveTLSProfile_StableAcrossLocalCacheOrder(t *testing.T) {
	account := &Account{
		ID:       5151,
		Platform: PlatformAnthropic,
		Type:     AccountTypeOAuth,
		Extra: map[string]any{
			"enable_tls_fingerprint":     true,
			"tls_fingerprint_profile_id": -1,
		},
	}
	profilesA := map[int64]*model.TLSFingerprintProfile{
		21: {ID: 21, Name: "profile-21"},
		4:  {ID: 4, Name: "profile-4"},
		9:  {ID: 9, Name: "profile-9"},
	}
	profilesB := map[int64]*model.TLSFingerprintProfile{
		9:  {ID: 9, Name: "profile-9"},
		21: {ID: 21, Name: "profile-21"},
		4:  {ID: 4, Name: "profile-4"},
	}

	svcA := &TLSFingerprintProfileService{localCache: profilesA}
	svcB := &TLSFingerprintProfileService{localCache: profilesB}

	profileA := svcA.ResolveTLSProfile(account)
	profileB := svcB.ResolveTLSProfile(account)
	require.NotNil(t, profileA)
	require.NotNil(t, profileB)
	require.Equal(t, profileA.Name, profileB.Name)
}
