//go:build unit

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type accountRepoStubForBulkUpdate struct {
	accountRepoStub
	bulkUpdateErr    error
	bulkUpdateIDs    []int64
	bindGroupErrByID map[int64]error
	bindGroupsCalls  []int64
	getByIDsAccounts []*Account
	getByIDsErr      error
	getByIDsCalled   bool
	getByIDsIDs      []int64
	getByIDAccounts  map[int64]*Account
	getByIDErrByID   map[int64]error
	getByIDCalled    []int64
	updateErrByID    map[int64]error
	updatedAccounts  []*Account
	listByGroupData  map[int64][]Account
	listByGroupErr   map[int64]error
}

func (s *accountRepoStubForBulkUpdate) BulkUpdate(_ context.Context, ids []int64, _ AccountBulkUpdate) (int64, error) {
	s.bulkUpdateIDs = append([]int64{}, ids...)
	if s.bulkUpdateErr != nil {
		return 0, s.bulkUpdateErr
	}
	return int64(len(ids)), nil
}

func (s *accountRepoStubForBulkUpdate) BindGroups(_ context.Context, accountID int64, _ []int64) error {
	s.bindGroupsCalls = append(s.bindGroupsCalls, accountID)
	if err, ok := s.bindGroupErrByID[accountID]; ok {
		return err
	}
	return nil
}

func (s *accountRepoStubForBulkUpdate) GetByIDs(_ context.Context, ids []int64) ([]*Account, error) {
	s.getByIDsCalled = true
	s.getByIDsIDs = append([]int64{}, ids...)
	if s.getByIDsErr != nil {
		return nil, s.getByIDsErr
	}
	return s.getByIDsAccounts, nil
}

func (s *accountRepoStubForBulkUpdate) GetByID(_ context.Context, id int64) (*Account, error) {
	s.getByIDCalled = append(s.getByIDCalled, id)
	if err, ok := s.getByIDErrByID[id]; ok {
		return nil, err
	}
	if account, ok := s.getByIDAccounts[id]; ok {
		return account, nil
	}
	return nil, errors.New("account not found")
}

func (s *accountRepoStubForBulkUpdate) Update(_ context.Context, account *Account) error {
	if err, ok := s.updateErrByID[account.ID]; ok {
		return err
	}
	s.updatedAccounts = append(s.updatedAccounts, account)
	return nil
}

func (s *accountRepoStubForBulkUpdate) ListByGroup(_ context.Context, groupID int64) ([]Account, error) {
	if err, ok := s.listByGroupErr[groupID]; ok {
		return nil, err
	}
	if rows, ok := s.listByGroupData[groupID]; ok {
		return rows, nil
	}
	return nil, nil
}

// TestAdminService_BulkUpdateAccounts_AllSuccessIDs 验证批量更新成功时返回 success_ids/failed_ids。
func TestAdminService_BulkUpdateAccounts_AllSuccessIDs(t *testing.T) {
	repo := &accountRepoStubForBulkUpdate{}
	svc := &adminServiceImpl{accountRepo: repo}

	schedulable := true
	input := &BulkUpdateAccountsInput{
		AccountIDs:  []int64{1, 2, 3},
		Schedulable: &schedulable,
	}

	result, err := svc.BulkUpdateAccounts(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 3, result.Success)
	require.Equal(t, 0, result.Failed)
	require.ElementsMatch(t, []int64{1, 2, 3}, result.SuccessIDs)
	require.Empty(t, result.FailedIDs)
	require.Len(t, result.Results, 3)
}

// TestAdminService_BulkUpdateAccounts_PartialFailureIDs 验证部分失败时 success_ids/failed_ids 正确。
func TestAdminService_BulkUpdateAccounts_PartialFailureIDs(t *testing.T) {
	repo := &accountRepoStubForBulkUpdate{
		bindGroupErrByID: map[int64]error{
			2: errors.New("bind failed"),
		},
	}
	svc := &adminServiceImpl{
		accountRepo: repo,
		groupRepo:   &groupRepoStubForAdmin{getByID: &Group{ID: 10, Name: "g10"}},
	}

	groupIDs := []int64{10}
	schedulable := false
	input := &BulkUpdateAccountsInput{
		AccountIDs:            []int64{1, 2, 3},
		GroupIDs:              &groupIDs,
		Schedulable:           &schedulable,
		SkipMixedChannelCheck: true,
	}

	result, err := svc.BulkUpdateAccounts(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 2, result.Success)
	require.Equal(t, 1, result.Failed)
	require.ElementsMatch(t, []int64{1, 3}, result.SuccessIDs)
	require.ElementsMatch(t, []int64{2}, result.FailedIDs)
	require.Len(t, result.Results, 3)
}

func TestAdminService_BulkUpdateAccounts_NilGroupRepoReturnsError(t *testing.T) {
	repo := &accountRepoStubForBulkUpdate{}
	svc := &adminServiceImpl{accountRepo: repo}

	groupIDs := []int64{10}
	input := &BulkUpdateAccountsInput{
		AccountIDs: []int64{1},
		GroupIDs:   &groupIDs,
	}

	result, err := svc.BulkUpdateAccounts(context.Background(), input)
	require.Nil(t, result)
	require.Error(t, err)
	require.Contains(t, err.Error(), "group repository not configured")
}

// TestAdminService_BulkUpdateAccounts_MixedChannelPreCheckBlocksOnExistingConflict verifies
// that the global pre-check detects a conflict with existing group members and returns an
// error before any DB write is performed.
func TestAdminService_BulkUpdateAccounts_MixedChannelPreCheckBlocksOnExistingConflict(t *testing.T) {
	repo := &accountRepoStubForBulkUpdate{
		getByIDsAccounts: []*Account{
			{ID: 1, Platform: PlatformAntigravity},
		},
		// Group 10 already contains an Anthropic account.
		listByGroupData: map[int64][]Account{
			10: {{ID: 99, Platform: PlatformAnthropic}},
		},
	}
	svc := &adminServiceImpl{
		accountRepo: repo,
		groupRepo:   &groupRepoStubForAdmin{getByID: &Group{ID: 10, Name: "target-group"}},
	}

	groupIDs := []int64{10}
	input := &BulkUpdateAccountsInput{
		AccountIDs: []int64{1},
		GroupIDs:   &groupIDs,
	}

	result, err := svc.BulkUpdateAccounts(context.Background(), input)
	require.Nil(t, result)
	require.Error(t, err)
	require.Contains(t, err.Error(), "mixed channel")
	// No BindGroups should have been called since the check runs before any write.
	require.Empty(t, repo.bindGroupsCalls)
}

func TestAdminService_BulkUpdateAccounts_RejectsMissingProxyBeforeBulkUpdate(t *testing.T) {
	repo := &accountRepoStubForBulkUpdate{}
	svc := &adminServiceImpl{
		accountRepo: repo,
		proxyRepo:   &accountProxyValidationProxyRepoStub{},
	}
	missingProxyID := int64(999)

	result, err := svc.BulkUpdateAccounts(context.Background(), &BulkUpdateAccountsInput{
		AccountIDs: []int64{1, 2},
		ProxyID:    &missingProxyID,
	})

	require.Nil(t, result)
	require.ErrorIs(t, err, ErrProxyNotFound)
	require.Empty(t, repo.bulkUpdateIDs)
}

func TestAdminService_BulkUpdateAccounts_ExtraUsesPerAccountDeepMerge(t *testing.T) {
	repo := &accountRepoStubForBulkUpdate{
		getByIDAccounts: map[int64]*Account{
			1: {
				ID:       1,
				Platform: PlatformGrok,
				Type:     AccountTypeSession,
				Extra: map[string]any{
					"grok": map[string]any{
						"sync_state": map[string]any{
							"last_sync_at": "2026-04-19T00:00:00Z",
						},
						"runtime_state": map[string]any{
							"last_success_at": "2026-04-20T00:00:00Z",
						},
					},
				},
			},
		},
	}
	svc := &adminServiceImpl{accountRepo: repo}

	result, err := svc.BulkUpdateAccounts(context.Background(), &BulkUpdateAccountsInput{
		AccountIDs: []int64{1},
		Extra: map[string]any{
			"grok": map[string]any{
				"tier": map[string]any{
					"normalized": "heavy",
					"source":     "manual",
				},
			},
		},
	})

	require.NoError(t, err)
	require.Equal(t, 1, result.Success)
	require.Empty(t, repo.bulkUpdateIDs)
	require.Len(t, repo.updatedAccounts, 1)
	grokExtra := grokExtraMap(repo.updatedAccounts[0].Extra)
	require.Equal(t, "heavy", getNestedGrokValue(grokExtra, "tier", "normalized"))
	require.Equal(t, "2026-04-19T00:00:00Z", getNestedGrokValue(grokExtra, "sync_state", "last_sync_at"))
	require.Equal(t, "2026-04-20T00:00:00Z", getNestedGrokValue(grokExtra, "runtime_state", "last_success_at"))
}

func TestAdminService_BulkUpdateAccounts_CredentialsPreserveExistingFields(t *testing.T) {
	repo := &accountRepoStubForBulkUpdate{
		getByIDAccounts: map[int64]*Account{
			1: {
				ID:       1,
				Platform: PlatformOpenAI,
				Type:     AccountTypeAPIKey,
				Credentials: map[string]any{
					"api_key":  "sk-existing",
					"base_url": "https://api.openai.com",
					"model_mapping": map[string]any{
						"gpt-4": "gpt-4",
					},
				},
			},
		},
	}
	svc := &adminServiceImpl{accountRepo: repo}

	result, err := svc.BulkUpdateAccounts(context.Background(), &BulkUpdateAccountsInput{
		AccountIDs: []int64{1},
		Credentials: map[string]any{
			"base_url":      "https://relay.example.com",
			"model_mapping": map[string]any{},
		},
	})

	require.NoError(t, err)
	require.Equal(t, 1, result.Success)
	require.Empty(t, repo.bulkUpdateIDs)
	require.Len(t, repo.updatedAccounts, 1)
	require.Equal(t, "sk-existing", repo.updatedAccounts[0].Credentials["api_key"])
	require.Equal(t, "https://relay.example.com", repo.updatedAccounts[0].Credentials["base_url"])
	require.Empty(t, repo.updatedAccounts[0].Credentials["model_mapping"])
}
