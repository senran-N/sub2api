package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/logger"
	"github.com/senran-N/sub2api/internal/pkg/pagination"
)

func (s *adminServiceImpl) ListAccounts(ctx context.Context, page, pageSize int, platform, accountType, status, search string, groupID int64, privacyMode string) ([]Account, int64, error) {
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}
	accounts, result, err := s.accountRepo.ListWithFilters(ctx, params, platform, accountType, status, search, groupID, privacyMode)
	if err != nil {
		return nil, 0, err
	}
	now := time.Now()
	for i := range accounts {
		syncOpenAICodexRateLimitFromExtra(ctx, s.accountRepo, &accounts[i], now)
	}
	return accounts, result.Total, nil
}

func (s *adminServiceImpl) GetAccount(ctx context.Context, id int64) (*Account, error) {
	return s.accountRepo.GetByID(ctx, id)
}

func (s *adminServiceImpl) GetAccountsByIDs(ctx context.Context, ids []int64) ([]*Account, error) {
	if len(ids) == 0 {
		return []*Account{}, nil
	}

	accounts, err := s.accountRepo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts by IDs: %w", err)
	}

	return accounts, nil
}

func (s *adminServiceImpl) CreateAccount(ctx context.Context, input *CreateAccountInput) (*Account, error) {
	groupIDs, err := s.resolveCreateAccountGroupIDs(ctx, input)
	if err != nil {
		return nil, err
	}
	if err := validateAccountGroupBindings(ctx, s.groupRepo, input.Type, groupIDs); err != nil {
		return nil, err
	}

	account, err := s.buildAccountForCreate(input)
	if err != nil {
		return nil, err
	}
	if err := s.accountRepo.Create(ctx, account); err != nil {
		return nil, err
	}

	if account.Platform == PlatformSora && s.soraAccountRepo != nil {
		soraUpdates := map[string]any{
			"access_token":  account.GetCredential("access_token"),
			"refresh_token": account.GetCredential("refresh_token"),
		}
		if err := s.soraAccountRepo.Upsert(ctx, account.ID, soraUpdates); err != nil {
			logger.LegacyPrintf("service.admin", "[AdminService] 创建 sora_accounts 记录失败: account_id=%d err=%v", account.ID, err)
		}
	}

	if len(groupIDs) > 0 {
		if err := s.accountRepo.BindGroups(ctx, account.ID, groupIDs); err != nil {
			return nil, err
		}
	}

	if account.Type == AccountTypeOAuth {
		switch account.Platform {
		case PlatformOpenAI:
			go func() {
				defer func() {
					if r := recover(); r != nil {
						slog.Error("create_account_openai_privacy_panic", "account_id", account.ID, "recover", r)
					}
				}()
				s.EnsureOpenAIPrivacy(context.Background(), account)
			}()
		case PlatformAntigravity:
			go func() {
				defer func() {
					if r := recover(); r != nil {
						slog.Error("create_account_antigravity_privacy_panic", "account_id", account.ID, "recover", r)
					}
				}()
				s.EnsureAntigravityPrivacy(context.Background(), account)
			}()
		}
	}

	return account, nil
}

func (s *adminServiceImpl) UpdateAccount(ctx context.Context, id int64, input *UpdateAccountInput) (*Account, error) {
	account, err := s.accountRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	wasOveragesEnabled := account.IsOveragesEnabled()

	if err := s.applyAccountUpdateInput(ctx, account, id, input, wasOveragesEnabled); err != nil {
		return nil, err
	}

	if err := s.accountRepo.Update(ctx, account); err != nil {
		return nil, err
	}

	if input.GroupIDs != nil {
		if err := s.accountRepo.BindGroups(ctx, account.ID, *input.GroupIDs); err != nil {
			return nil, err
		}
	}

	updated, err := s.accountRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *adminServiceImpl) BulkUpdateAccounts(ctx context.Context, input *BulkUpdateAccountsInput) (*BulkUpdateAccountsResult, error) {
	result := &BulkUpdateAccountsResult{
		SuccessIDs: make([]int64, 0, len(input.AccountIDs)),
		FailedIDs:  make([]int64, 0, len(input.AccountIDs)),
		Results:    make([]BulkUpdateAccountResult, 0, len(input.AccountIDs)),
	}

	if len(input.AccountIDs) == 0 {
		return result, nil
	}
	if err := validateAccountIDList(input.AccountIDs); err != nil {
		return nil, err
	}
	if err := s.validateBulkAccountGroupChange(ctx, input); err != nil {
		return nil, err
	}

	repoUpdates, err := buildAccountBulkUpdate(input)
	if err != nil {
		return nil, err
	}

	if _, err := s.accountRepo.BulkUpdate(ctx, input.AccountIDs, repoUpdates); err != nil {
		return nil, err
	}

	for _, accountID := range input.AccountIDs {
		entry := BulkUpdateAccountResult{AccountID: accountID}

		if input.GroupIDs != nil {
			if err := s.accountRepo.BindGroups(ctx, accountID, *input.GroupIDs); err != nil {
				entry.Success = false
				entry.Error = err.Error()
				result.Failed++
				result.FailedIDs = append(result.FailedIDs, accountID)
				result.Results = append(result.Results, entry)
				continue
			}
		}

		entry.Success = true
		result.Success++
		result.SuccessIDs = append(result.SuccessIDs, accountID)
		result.Results = append(result.Results, entry)
	}

	return result, nil
}

func (s *adminServiceImpl) DeleteAccount(ctx context.Context, id int64) error {
	if err := s.accountRepo.Delete(ctx, id); err != nil {
		return err
	}
	return nil
}

func (s *adminServiceImpl) RefreshAccountCredentials(ctx context.Context, id int64) (*Account, error) {
	account, err := s.accountRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (s *adminServiceImpl) ClearAccountError(ctx context.Context, id int64) (*Account, error) {
	if err := s.accountRepo.ClearError(ctx, id); err != nil {
		return nil, err
	}
	if err := s.accountRepo.ClearRateLimit(ctx, id); err != nil {
		return nil, err
	}
	if err := s.accountRepo.ClearAntigravityQuotaScopes(ctx, id); err != nil {
		return nil, err
	}
	if err := s.accountRepo.ClearModelRateLimits(ctx, id); err != nil {
		return nil, err
	}
	if err := s.accountRepo.ClearTempUnschedulable(ctx, id); err != nil {
		return nil, err
	}
	return s.accountRepo.GetByID(ctx, id)
}

func (s *adminServiceImpl) SetAccountError(ctx context.Context, id int64, errorMsg string) error {
	return s.accountRepo.SetError(ctx, id, errorMsg)
}

func (s *adminServiceImpl) SetAccountSchedulable(ctx context.Context, id int64, schedulable bool) (*Account, error) {
	if err := s.accountRepo.SetSchedulable(ctx, id, schedulable); err != nil {
		return nil, err
	}
	updated, err := s.accountRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return updated, nil
}
