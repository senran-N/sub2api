//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type groupBindingRuleGroupRepoStub struct {
	groupRepoStubForAdmin
	groups map[int64]*Group
}

func (s *groupBindingRuleGroupRepoStub) GetByID(_ context.Context, id int64) (*Group, error) {
	group, ok := s.groups[id]
	if !ok {
		return nil, ErrGroupNotFound
	}
	return group, nil
}

func (s *groupBindingRuleGroupRepoStub) GetByIDLite(_ context.Context, id int64) (*Group, error) {
	return s.GetByID(context.Background(), id)
}

func (s *groupBindingRuleGroupRepoStub) ExistsByIDs(_ context.Context, ids []int64) (map[int64]bool, error) {
	result := make(map[int64]bool, len(ids))
	for _, id := range ids {
		_, ok := s.groups[id]
		result[id] = ok
	}
	return result, nil
}

type groupBindingRuleAccountRepoStub struct {
	accountRepoStub
	createCalls    int
	updateCalls    int
	bindCalls      []int64
	bulkUpdateIDs  []int64
	getByIDAccount *Account
	getByIDsResult []*Account
}

func (s *groupBindingRuleAccountRepoStub) Create(_ context.Context, account *Account) error {
	s.createCalls++
	if account.ID == 0 {
		account.ID = int64(s.createCalls)
	}
	return nil
}

func (s *groupBindingRuleAccountRepoStub) GetByID(_ context.Context, _ int64) (*Account, error) {
	if s.getByIDAccount == nil {
		return nil, ErrAccountNotFound
	}
	return s.getByIDAccount, nil
}

func (s *groupBindingRuleAccountRepoStub) Update(_ context.Context, _ *Account) error {
	s.updateCalls++
	return nil
}

func (s *groupBindingRuleAccountRepoStub) BindGroups(_ context.Context, accountID int64, _ []int64) error {
	s.bindCalls = append(s.bindCalls, accountID)
	return nil
}

func (s *groupBindingRuleAccountRepoStub) BulkUpdate(_ context.Context, ids []int64, _ AccountBulkUpdate) (int64, error) {
	s.bulkUpdateIDs = append([]int64{}, ids...)
	return int64(len(ids)), nil
}

func (s *groupBindingRuleAccountRepoStub) GetByIDs(_ context.Context, _ []int64) ([]*Account, error) {
	return s.getByIDsResult, nil
}

func (s *groupBindingRuleAccountRepoStub) ListByGroup(_ context.Context, _ int64) ([]Account, error) {
	return nil, nil
}

func TestAccountServiceCreateRejectsAPIKeyOAuthOnlyGroupBeforeCreate(t *testing.T) {
	accountRepo := &groupBindingRuleAccountRepoStub{}
	groupRepo := &groupBindingRuleGroupRepoStub{
		groups: map[int64]*Group{
			1: {
				ID:               1,
				Name:             "restricted-openai",
				Platform:         PlatformOpenAI,
				RequireOAuthOnly: true,
			},
		},
	}
	svc := &AccountService{accountRepo: accountRepo, groupRepo: groupRepo}

	_, err := svc.Create(context.Background(), CreateAccountRequest{
		Name:     "apikey-account",
		Platform: PlatformOpenAI,
		Type:     AccountTypeAPIKey,
		GroupIDs: []int64{1},
	})

	require.ErrorContains(t, err, "仅允许 OAuth 账号")
	require.Zero(t, accountRepo.createCalls)
	require.Empty(t, accountRepo.bindCalls)
}

func TestAccountServiceUpdateRejectsAPIKeyOAuthOnlyGroupBeforeUpdate(t *testing.T) {
	accountRepo := &groupBindingRuleAccountRepoStub{
		getByIDAccount: &Account{
			ID:   7,
			Type: AccountTypeAPIKey,
		},
	}
	groupRepo := &groupBindingRuleGroupRepoStub{
		groups: map[int64]*Group{
			2: {
				ID:               2,
				Name:             "restricted-antigravity",
				Platform:         PlatformAntigravity,
				RequireOAuthOnly: true,
			},
		},
	}
	svc := &AccountService{accountRepo: accountRepo, groupRepo: groupRepo}
	groupIDs := []int64{2}

	_, err := svc.Update(context.Background(), 7, UpdateAccountRequest{
		GroupIDs: &groupIDs,
	})

	require.ErrorContains(t, err, "仅允许 OAuth 账号")
	require.Zero(t, accountRepo.updateCalls)
	require.Empty(t, accountRepo.bindCalls)
}

func TestAdminServiceCreateAccountRejectsAPIKeyOAuthOnlyGroupBeforeCreate(t *testing.T) {
	accountRepo := &groupBindingRuleAccountRepoStub{}
	groupRepo := &groupBindingRuleGroupRepoStub{
		groups: map[int64]*Group{
			3: {
				ID:               3,
				Name:             "restricted-gemini",
				Platform:         PlatformGemini,
				RequireOAuthOnly: true,
			},
		},
	}
	svc := &adminServiceImpl{
		accountRepo: accountRepo,
		groupRepo:   groupRepo,
	}

	_, err := svc.CreateAccount(context.Background(), &CreateAccountInput{
		Name:     "admin-apikey",
		Platform: PlatformGemini,
		Type:     AccountTypeAPIKey,
		GroupIDs: []int64{3},
	})

	require.ErrorContains(t, err, "仅允许 OAuth 账号")
	require.Zero(t, accountRepo.createCalls)
	require.Empty(t, accountRepo.bindCalls)
}

func TestAdminServiceBulkUpdateAccountsRejectsAPIKeyOAuthOnlyGroupBeforeBulkUpdate(t *testing.T) {
	accountRepo := &groupBindingRuleAccountRepoStub{
		getByIDsResult: []*Account{
			{
				ID:       9,
				Type:     AccountTypeAPIKey,
				Platform: PlatformOpenAI,
			},
		},
	}
	groupRepo := &groupBindingRuleGroupRepoStub{
		groups: map[int64]*Group{
			4: {
				ID:               4,
				Name:             "restricted-claude",
				Platform:         PlatformAnthropic,
				RequireOAuthOnly: true,
			},
		},
	}
	svc := &adminServiceImpl{
		accountRepo: accountRepo,
		groupRepo:   groupRepo,
	}
	groupIDs := []int64{4}

	result, err := svc.BulkUpdateAccounts(context.Background(), &BulkUpdateAccountsInput{
		AccountIDs:            []int64{9},
		GroupIDs:              &groupIDs,
		SkipMixedChannelCheck: true,
	})

	require.Nil(t, result)
	require.ErrorContains(t, err, "仅允许 OAuth 账号")
	require.Empty(t, accountRepo.bulkUpdateIDs)
	require.Empty(t, accountRepo.bindCalls)
}
