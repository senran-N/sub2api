//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type identityBindingUserRepoStub struct {
	user               *User
	identities         []UserAuthIdentityRecord
	unboundProviders   []string
	updatedConcurrency int
}

func (s *identityBindingUserRepoStub) Create(context.Context, *User) error { return nil }
func (s *identityBindingUserRepoStub) GetByID(context.Context, int64) (*User, error) {
	if s.user == nil {
		return nil, ErrUserNotFound
	}
	copyUser := *s.user
	return &copyUser, nil
}
func (s *identityBindingUserRepoStub) GetByEmail(context.Context, string) (*User, error) {
	return nil, ErrUserNotFound
}
func (s *identityBindingUserRepoStub) GetFirstAdmin(context.Context) (*User, error) {
	return nil, ErrUserNotFound
}
func (s *identityBindingUserRepoStub) Update(context.Context, *User) error { return nil }
func (s *identityBindingUserRepoStub) Delete(context.Context, int64) error { return nil }
func (s *identityBindingUserRepoStub) List(context.Context, pagination.PaginationParams) ([]User, *pagination.PaginationResult, error) {
	return nil, nil, nil
}
func (s *identityBindingUserRepoStub) ListWithFilters(context.Context, pagination.PaginationParams, UserListFilters) ([]User, *pagination.PaginationResult, error) {
	return nil, nil, nil
}
func (s *identityBindingUserRepoStub) UpdateBalance(context.Context, int64, float64) error {
	return nil
}
func (s *identityBindingUserRepoStub) DeductBalance(context.Context, int64, float64) error {
	return nil
}
func (s *identityBindingUserRepoStub) UpdateConcurrency(context.Context, int64, int) error {
	return nil
}
func (s *identityBindingUserRepoStub) ExistsByEmail(context.Context, string) (bool, error) {
	return false, nil
}
func (s *identityBindingUserRepoStub) RemoveGroupFromAllowedGroups(context.Context, int64) (int64, error) {
	return 0, nil
}
func (s *identityBindingUserRepoStub) AddGroupToAllowedGroups(context.Context, int64, int64) error {
	return nil
}
func (s *identityBindingUserRepoStub) RemoveGroupFromUserAllowedGroups(context.Context, int64, int64) error {
	return nil
}
func (s *identityBindingUserRepoStub) ListUserAuthIdentities(context.Context, int64) ([]UserAuthIdentityRecord, error) {
	out := make([]UserAuthIdentityRecord, len(s.identities))
	copy(out, s.identities)
	return out, nil
}
func (s *identityBindingUserRepoStub) UnbindUserAuthProvider(_ context.Context, _ int64, provider string) error {
	s.unboundProviders = append(s.unboundProviders, provider)
	filtered := make([]UserAuthIdentityRecord, 0, len(s.identities))
	for _, record := range s.identities {
		if normalizeBoundIdentityProvider(record.ProviderType) == normalizeBoundIdentityProvider(provider) {
			continue
		}
		filtered = append(filtered, record)
	}
	s.identities = filtered
	return nil
}
func (s *identityBindingUserRepoStub) UpdateUserLastActiveAt(context.Context, int64, time.Time) error {
	return nil
}
func (s *identityBindingUserRepoStub) UpdateTotpSecret(context.Context, int64, *string) error {
	return nil
}
func (s *identityBindingUserRepoStub) EnableTotp(context.Context, int64) error  { return nil }
func (s *identityBindingUserRepoStub) DisableTotp(context.Context, int64) error { return nil }

func TestUnbindIdentity_RejectsLastSignInMethod(t *testing.T) {
	repo := &identityBindingUserRepoStub{
		user:       &User{ID: 1, Email: "only" + LinuxDoConnectSyntheticEmailDomain, EmailBound: false},
		identities: []UserAuthIdentityRecord{{ProviderType: "linuxdo", ProviderSubject: "linuxdo-user-1"}},
	}
	svc := NewUserService(repo, nil, nil)

	updated, err := svc.UnbindIdentity(context.Background(), 1, "linuxdo")
	require.Nil(t, updated)
	require.Error(t, err)
	require.Contains(t, err.Error(), "bind another sign-in method")
	require.Empty(t, repo.unboundProviders)
}

func TestUnbindIdentity_AllowsWhenRealEmailCanSignIn(t *testing.T) {
	repo := &identityBindingUserRepoStub{
		user: &User{ID: 2, Email: "user@example.com", EmailBound: true, LinuxDoBound: true},
		identities: []UserAuthIdentityRecord{
			{ProviderType: "email", ProviderSubject: "user@example.com"},
			{ProviderType: "linuxdo", ProviderSubject: "linuxdo-user-2"},
		},
	}
	svc := NewUserService(repo, nil, nil)

	updated, err := svc.UnbindIdentity(context.Background(), 2, "linuxdo")
	require.NoError(t, err)
	require.NotNil(t, updated)
	require.Equal(t, []string{"linuxdo"}, repo.unboundProviders)
}

func TestUnbindIdentity_AllowsWhenAnotherOAuthProviderRemains(t *testing.T) {
	repo := &identityBindingUserRepoStub{
		user: &User{ID: 3, Email: "oauth" + LinuxDoConnectSyntheticEmailDomain, EmailBound: false},
		identities: []UserAuthIdentityRecord{
			{ProviderType: "linuxdo", ProviderSubject: "linuxdo-user-3"},
			{ProviderType: "oidc", ProviderSubject: "oidc-user-3"},
		},
	}
	svc := NewUserService(repo, nil, nil)

	updated, err := svc.UnbindIdentity(context.Background(), 3, "linuxdo")
	require.NoError(t, err)
	require.NotNil(t, updated)
	require.Equal(t, []string{"linuxdo"}, repo.unboundProviders)
}
