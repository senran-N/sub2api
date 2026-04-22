package service

import (
	"context"
	"testing"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type grokSessionBatchImportAccountRepoStub struct {
	existingAccounts []Account
	createdAccounts  []*Account
	boundGroups      [][]int64
	nextID           int64
}

func (s *grokSessionBatchImportAccountRepoStub) Create(_ context.Context, account *Account) error {
	s.nextID++
	account.ID = s.nextID
	s.createdAccounts = append(s.createdAccounts, account)
	return nil
}

func (s *grokSessionBatchImportAccountRepoStub) GetByID(context.Context, int64) (*Account, error) {
	panic("unexpected GetByID call")
}

func (s *grokSessionBatchImportAccountRepoStub) GetByIDs(context.Context, []int64) ([]*Account, error) {
	panic("unexpected GetByIDs call")
}

func (s *grokSessionBatchImportAccountRepoStub) ExistsByID(context.Context, int64) (bool, error) {
	panic("unexpected ExistsByID call")
}

func (s *grokSessionBatchImportAccountRepoStub) GetByCRSAccountID(context.Context, string) (*Account, error) {
	panic("unexpected GetByCRSAccountID call")
}

func (s *grokSessionBatchImportAccountRepoStub) ListCRSAccountIDs(context.Context) (map[string]int64, error) {
	panic("unexpected ListCRSAccountIDs call")
}

func (s *grokSessionBatchImportAccountRepoStub) Update(context.Context, *Account) error {
	panic("unexpected Update call")
}

func (s *grokSessionBatchImportAccountRepoStub) Delete(context.Context, int64) error {
	panic("unexpected Delete call")
}

func (s *grokSessionBatchImportAccountRepoStub) List(context.Context, pagination.PaginationParams) ([]Account, *pagination.PaginationResult, error) {
	panic("unexpected List call")
}

func (s *grokSessionBatchImportAccountRepoStub) ListWithFilters(context.Context, pagination.PaginationParams, string, string, string, string, int64, string) ([]Account, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters call")
}

func (s *grokSessionBatchImportAccountRepoStub) ListByGroup(context.Context, int64) ([]Account, error) {
	panic("unexpected ListByGroup call")
}

func (s *grokSessionBatchImportAccountRepoStub) ListActive(context.Context) ([]Account, error) {
	panic("unexpected ListActive call")
}

func (s *grokSessionBatchImportAccountRepoStub) ListByPlatform(_ context.Context, platform string) ([]Account, error) {
	if platform != PlatformGrok {
		return nil, nil
	}
	return append([]Account(nil), s.existingAccounts...), nil
}

func (s *grokSessionBatchImportAccountRepoStub) UpdateLastUsed(context.Context, int64) error {
	panic("unexpected UpdateLastUsed call")
}

func (s *grokSessionBatchImportAccountRepoStub) BatchUpdateLastUsed(context.Context, map[int64]time.Time) error {
	panic("unexpected BatchUpdateLastUsed call")
}

func (s *grokSessionBatchImportAccountRepoStub) SetError(context.Context, int64, string) error {
	panic("unexpected SetError call")
}

func (s *grokSessionBatchImportAccountRepoStub) ClearError(context.Context, int64) error {
	panic("unexpected ClearError call")
}

func (s *grokSessionBatchImportAccountRepoStub) SetSchedulable(context.Context, int64, bool) error {
	panic("unexpected SetSchedulable call")
}

func (s *grokSessionBatchImportAccountRepoStub) AutoPauseExpiredAccounts(context.Context, time.Time) (int64, error) {
	panic("unexpected AutoPauseExpiredAccounts call")
}

func (s *grokSessionBatchImportAccountRepoStub) BindGroups(_ context.Context, _ int64, groupIDs []int64) error {
	s.boundGroups = append(s.boundGroups, append([]int64(nil), groupIDs...))
	return nil
}

func (s *grokSessionBatchImportAccountRepoStub) ListSchedulable(context.Context) ([]Account, error) {
	panic("unexpected ListSchedulable call")
}

func (s *grokSessionBatchImportAccountRepoStub) ListSchedulableByGroupID(context.Context, int64) ([]Account, error) {
	panic("unexpected ListSchedulableByGroupID call")
}

func (s *grokSessionBatchImportAccountRepoStub) ListSchedulableByPlatform(context.Context, string) ([]Account, error) {
	panic("unexpected ListSchedulableByPlatform call")
}

func (s *grokSessionBatchImportAccountRepoStub) ListSchedulableByGroupIDAndPlatform(context.Context, int64, string) ([]Account, error) {
	panic("unexpected ListSchedulableByGroupIDAndPlatform call")
}

func (s *grokSessionBatchImportAccountRepoStub) ListSchedulableByPlatforms(context.Context, []string) ([]Account, error) {
	panic("unexpected ListSchedulableByPlatforms call")
}

func (s *grokSessionBatchImportAccountRepoStub) ListSchedulableByGroupIDAndPlatforms(context.Context, int64, []string) ([]Account, error) {
	panic("unexpected ListSchedulableByGroupIDAndPlatforms call")
}

func (s *grokSessionBatchImportAccountRepoStub) ListSchedulableUngroupedByPlatform(context.Context, string) ([]Account, error) {
	panic("unexpected ListSchedulableUngroupedByPlatform call")
}

func (s *grokSessionBatchImportAccountRepoStub) ListSchedulableUngroupedByPlatforms(context.Context, []string) ([]Account, error) {
	panic("unexpected ListSchedulableUngroupedByPlatforms call")
}

func (s *grokSessionBatchImportAccountRepoStub) SetRateLimited(context.Context, int64, time.Time) error {
	panic("unexpected SetRateLimited call")
}

func (s *grokSessionBatchImportAccountRepoStub) SetModelRateLimit(context.Context, int64, string, time.Time) error {
	panic("unexpected SetModelRateLimit call")
}

func (s *grokSessionBatchImportAccountRepoStub) SetOverloaded(context.Context, int64, time.Time) error {
	panic("unexpected SetOverloaded call")
}

func (s *grokSessionBatchImportAccountRepoStub) SetTempUnschedulable(context.Context, int64, time.Time, string) error {
	panic("unexpected SetTempUnschedulable call")
}

func (s *grokSessionBatchImportAccountRepoStub) ClearTempUnschedulable(context.Context, int64) error {
	panic("unexpected ClearTempUnschedulable call")
}

func (s *grokSessionBatchImportAccountRepoStub) ClearRateLimit(context.Context, int64) error {
	panic("unexpected ClearRateLimit call")
}

func (s *grokSessionBatchImportAccountRepoStub) ClearAntigravityQuotaScopes(context.Context, int64) error {
	panic("unexpected ClearAntigravityQuotaScopes call")
}

func (s *grokSessionBatchImportAccountRepoStub) ClearModelRateLimits(context.Context, int64) error {
	panic("unexpected ClearModelRateLimits call")
}

func (s *grokSessionBatchImportAccountRepoStub) UpdateSessionWindow(context.Context, int64, *time.Time, *time.Time, string) error {
	panic("unexpected UpdateSessionWindow call")
}

func (s *grokSessionBatchImportAccountRepoStub) UpdateExtra(context.Context, int64, map[string]any) error {
	panic("unexpected UpdateExtra call")
}

func (s *grokSessionBatchImportAccountRepoStub) BulkUpdate(context.Context, []int64, AccountBulkUpdate) (int64, error) {
	panic("unexpected BulkUpdate call")
}

func (s *grokSessionBatchImportAccountRepoStub) IncrementQuotaUsed(context.Context, int64, float64) error {
	panic("unexpected IncrementQuotaUsed call")
}

func (s *grokSessionBatchImportAccountRepoStub) ResetQuotaUsed(context.Context, int64) error {
	panic("unexpected ResetQuotaUsed call")
}

type grokSessionBatchImportGroupRepoStub struct{}

type grokSessionBatchImportSyncerStub struct {
	syncedAccountIDs []int64
}

func (s *grokSessionBatchImportSyncerStub) SyncAccount(_ context.Context, account *Account) error {
	if account != nil {
		s.syncedAccountIDs = append(s.syncedAccountIDs, account.ID)
	}
	return nil
}

func (g grokSessionBatchImportGroupRepoStub) Create(context.Context, *Group) error {
	panic("unexpected Create call")
}

func (g grokSessionBatchImportGroupRepoStub) GetByID(context.Context, int64) (*Group, error) {
	panic("unexpected GetByID call")
}

func (g grokSessionBatchImportGroupRepoStub) GetByIDLite(context.Context, int64) (*Group, error) {
	panic("unexpected GetByIDLite call")
}

func (g grokSessionBatchImportGroupRepoStub) Update(context.Context, *Group) error {
	panic("unexpected Update call")
}

func (g grokSessionBatchImportGroupRepoStub) Delete(context.Context, int64) error {
	panic("unexpected Delete call")
}

func (g grokSessionBatchImportGroupRepoStub) DeleteCascade(context.Context, int64) ([]int64, error) {
	panic("unexpected DeleteCascade call")
}

func (g grokSessionBatchImportGroupRepoStub) List(context.Context, pagination.PaginationParams) ([]Group, *pagination.PaginationResult, error) {
	panic("unexpected List call")
}

func (g grokSessionBatchImportGroupRepoStub) ListWithFilters(context.Context, pagination.PaginationParams, string, string, string, *bool) ([]Group, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters call")
}

func (g grokSessionBatchImportGroupRepoStub) ListActive(context.Context) ([]Group, error) {
	panic("unexpected ListActive call")
}

func (g grokSessionBatchImportGroupRepoStub) ListActiveByPlatform(context.Context, string) ([]Group, error) {
	return nil, nil
}

func (g grokSessionBatchImportGroupRepoStub) ExistsByName(context.Context, string) (bool, error) {
	panic("unexpected ExistsByName call")
}

func (g grokSessionBatchImportGroupRepoStub) GetAccountCount(context.Context, int64) (int64, int64, error) {
	panic("unexpected GetAccountCount call")
}

func (g grokSessionBatchImportGroupRepoStub) DeleteAccountGroupsByGroupID(context.Context, int64) (int64, error) {
	panic("unexpected DeleteAccountGroupsByGroupID call")
}

func (g grokSessionBatchImportGroupRepoStub) GetAccountIDsByGroupIDs(context.Context, []int64) ([]int64, error) {
	panic("unexpected GetAccountIDsByGroupIDs call")
}

func (g grokSessionBatchImportGroupRepoStub) BindAccountsToGroup(context.Context, int64, []int64) error {
	panic("unexpected BindAccountsToGroup call")
}

func (g grokSessionBatchImportGroupRepoStub) UpdateSortOrders(context.Context, []GroupSortOrderUpdate) error {
	panic("unexpected UpdateSortOrders call")
}

func TestAdminServiceBatchImportGrokSessionAccounts_CreatesNormalizedAccounts(t *testing.T) {
	accountRepo := &grokSessionBatchImportAccountRepoStub{}
	svc := &adminServiceImpl{
		accountRepo: accountRepo,
		groupRepo:   grokSessionBatchImportGroupRepoStub{},
	}
	rawTokenA := "groksessiontoken1234567890abcd"
	rawTokenB := "abcdefghijklmnopqrstuvwxyz123456"
	rawTokenBRW := "mnopqrstuvwxyzabcdef123456"

	result, err := svc.BatchImportGrokSessionAccounts(context.Background(), &GrokSessionBatchImportInput{
		RawInput:    rawTokenA + "\nCookie: sso=" + rawTokenB + "; sso-rw=" + rawTokenBRW + "\n",
		Concurrency: 1,
	})

	require.NoError(t, err)
	require.Equal(t, 2, result.Total)
	require.Equal(t, 2, result.Created)
	require.Equal(t, 0, result.Skipped)
	require.Equal(t, 0, result.Invalid)
	require.Len(t, accountRepo.createdAccounts, 2)
	require.Equal(t, PlatformGrok, accountRepo.createdAccounts[0].Platform)
	require.Equal(t, AccountTypeSession, accountRepo.createdAccounts[0].Type)
	require.Equal(t, "sso="+rawTokenA+"; sso-rw="+rawTokenA, accountRepo.createdAccounts[0].GetCredential("session_token"))
	require.Equal(t, "sso="+rawTokenB+"; sso-rw="+rawTokenBRW, accountRepo.createdAccounts[1].GetCredential("session_token"))
	require.NotEmpty(t, getStringFromMaps(accountRepo.createdAccounts[0].grokExtraMap(), nil, "auth_fingerprint"))
	require.NotContains(t, result.Results[0].Fingerprint, rawTokenA)
	require.Equal(t, "grok-sso-001", result.Results[0].Name)
	require.Equal(t, "grok-sso-002", result.Results[1].Name)
}

func TestAdminServiceBatchImportGrokSessionAccounts_DedupesExistingAndRejectsMissingSSO(t *testing.T) {
	existingToken := "abcdefghijklmnopqrstuvwxyz123456"
	existingCookie := "sso=" + existingToken
	accountRepo := &grokSessionBatchImportAccountRepoStub{
		existingAccounts: []Account{
			{
				ID:          41,
				Platform:    PlatformGrok,
				Type:        AccountTypeSession,
				Credentials: map[string]any{"session_token": existingCookie},
			},
		},
	}
	svc := &adminServiceImpl{
		accountRepo: accountRepo,
		groupRepo:   grokSessionBatchImportGroupRepoStub{},
	}

	result, err := svc.BatchImportGrokSessionAccounts(context.Background(), &GrokSessionBatchImportInput{
		RawInput: existingToken + "\nx-anonuserid=anon-only\n" + existingToken + "\n",
	})

	require.NoError(t, err)
	require.Equal(t, 3, result.Total)
	require.Equal(t, 0, result.Created)
	require.Equal(t, 2, result.Skipped)
	require.Equal(t, 1, result.Invalid)
	require.Len(t, accountRepo.createdAccounts, 0)
	require.Equal(t, "missing sso cookie", result.Results[1].Reason)
	require.NotContains(t, result.Results[0].Reason, existingToken)
}

func TestAdminServiceBatchImportGrokSessionAccounts_SyncsCreatedSessionAccount(t *testing.T) {
	accountRepo := &grokSessionBatchImportAccountRepoStub{}
	syncer := &grokSessionBatchImportSyncerStub{}
	svc := &adminServiceImpl{
		accountRepo:     accountRepo,
		groupRepo:       grokSessionBatchImportGroupRepoStub{},
		grokQuotaSyncer: syncer,
	}

	result, err := svc.BatchImportGrokSessionAccounts(context.Background(), &GrokSessionBatchImportInput{
		RawInput:    "groksessiontoken1234567890abcd\n",
		Concurrency: 1,
	})

	require.NoError(t, err)
	require.Equal(t, 1, result.Created)
	require.Equal(t, []int64{1}, syncer.syncedAccountIDs)
}
