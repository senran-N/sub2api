package service

import (
	"context"
	"fmt"
	"github.com/senran-N/sub2api/internal/pkg/pagination"
)

func (s *adminServiceImpl) ListGroups(ctx context.Context, page, pageSize int, platform, status, search string, isExclusive *bool) ([]Group, int64, error) {
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}
	groups, result, err := s.groupRepo.ListWithFilters(ctx, params, platform, status, search, isExclusive)
	if err != nil {
		return nil, 0, err
	}
	return groups, result.Total, nil
}

func (s *adminServiceImpl) GetAllGroups(ctx context.Context) ([]Group, error) {
	return s.groupRepo.ListActive(ctx)
}

func (s *adminServiceImpl) GetAllGroupsByPlatform(ctx context.Context, platform string) ([]Group, error) {
	return s.groupRepo.ListActiveByPlatform(ctx, platform)
}

func (s *adminServiceImpl) GetGroup(ctx context.Context, id int64) (*Group, error) {
	return s.groupRepo.GetByID(ctx, id)
}

func (s *adminServiceImpl) resolveGroupCopyAccountIDs(ctx context.Context, currentGroupID int64, platform string, sourceGroupIDs []int64, requireOAuthOnly bool) ([]int64, error) {
	if len(sourceGroupIDs) == 0 {
		return nil, nil
	}

	uniqueSourceGroupIDs, err := s.uniqueCopySourceGroupIDs(ctx, currentGroupID, platform, sourceGroupIDs)
	if err != nil {
		return nil, err
	}

	accountIDs, err := s.groupRepo.GetAccountIDsByGroupIDs(ctx, uniqueSourceGroupIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts from source groups: %w", err)
	}

	accountIDs, err = s.filterOAuthOnlyGroupAccountIDs(ctx, platform, requireOAuthOnly, accountIDs)
	if err != nil {
		return nil, err
	}
	return accountIDs, nil
}

func (s *adminServiceImpl) uniqueCopySourceGroupIDs(ctx context.Context, currentGroupID int64, platform string, sourceGroupIDs []int64) ([]int64, error) {
	seen := make(map[int64]struct{}, len(sourceGroupIDs))
	uniqueSourceGroupIDs := make([]int64, 0, len(sourceGroupIDs))
	for _, srcGroupID := range sourceGroupIDs {
		if currentGroupID > 0 && srcGroupID == currentGroupID {
			return nil, fmt.Errorf("cannot copy accounts from self")
		}
		if _, exists := seen[srcGroupID]; exists {
			continue
		}
		seen[srcGroupID] = struct{}{}
		uniqueSourceGroupIDs = append(uniqueSourceGroupIDs, srcGroupID)
	}

	for _, srcGroupID := range uniqueSourceGroupIDs {
		srcGroup, err := s.groupRepo.GetByIDLite(ctx, srcGroupID)
		if err != nil {
			return nil, fmt.Errorf("source group %d not found: %w", srcGroupID, err)
		}
		if srcGroup.Platform != platform {
			return nil, fmt.Errorf("source group %d platform mismatch: expected %s, got %s", srcGroupID, platform, srcGroup.Platform)
		}
	}

	return uniqueSourceGroupIDs, nil
}

func (s *adminServiceImpl) filterOAuthOnlyGroupAccountIDs(ctx context.Context, platform string, requireOAuthOnly bool, accountIDs []int64) ([]int64, error) {
	if !requireOAuthOnly || !groupPlatformSupportsOAuthOnly(platform) || len(accountIDs) == 0 {
		return accountIDs, nil
	}

	accounts, err := s.accountRepo.GetByIDs(ctx, accountIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch accounts for oauth filter: %w", err)
	}

	oauthIDs := make(map[int64]struct{}, len(accounts))
	for _, acc := range accounts {
		if acc.Type != AccountTypeAPIKey {
			oauthIDs[acc.ID] = struct{}{}
		}
	}

	filtered := make([]int64, 0, len(accountIDs))
	for _, accountID := range accountIDs {
		if _, ok := oauthIDs[accountID]; ok {
			filtered = append(filtered, accountID)
		}
	}
	return filtered, nil
}

func groupPlatformSupportsOAuthOnly(platform string) bool {
	switch platform {
	case PlatformOpenAI, PlatformAntigravity, PlatformAnthropic, PlatformGemini:
		return true
	default:
		return false
	}
}

func (s *adminServiceImpl) CreateGroup(ctx context.Context, input *CreateGroupInput) (*Group, error) {
	group, platform, err := s.buildGroupForCreate(ctx, input)
	if err != nil {
		return nil, err
	}

	accountIDsToCopy, err := s.resolveGroupCopyAccountIDs(ctx, 0, platform, input.CopyAccountsFromGroupIDs, group.RequireOAuthOnly)
	if err != nil {
		return nil, err
	}
	if err := s.groupRepo.Create(ctx, group); err != nil {
		return nil, err
	}

	if len(accountIDsToCopy) > 0 {
		if err := s.groupRepo.BindAccountsToGroup(ctx, group.ID, accountIDsToCopy); err != nil {
			return nil, fmt.Errorf("failed to bind accounts to new group: %w", err)
		}
		group.AccountCount = int64(len(accountIDsToCopy))
	}

	return group, nil
}

func (s *adminServiceImpl) UpdateGroup(ctx context.Context, id int64, input *UpdateGroupInput) (*Group, error) {
	group, err := s.groupRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.applyUpdateGroupInput(ctx, id, group, input); err != nil {
		return nil, err
	}

	if err := s.groupRepo.Update(ctx, group); err != nil {
		return nil, err
	}

	if len(input.CopyAccountsFromGroupIDs) > 0 {
		accountIDsToCopy, err := s.resolveGroupCopyAccountIDs(ctx, id, group.Platform, input.CopyAccountsFromGroupIDs, group.RequireOAuthOnly)
		if err != nil {
			return nil, err
		}

		if _, err := s.groupRepo.DeleteAccountGroupsByGroupID(ctx, id); err != nil {
			return nil, fmt.Errorf("failed to clear existing account bindings: %w", err)
		}

		if len(accountIDsToCopy) > 0 {
			if err := s.groupRepo.BindAccountsToGroup(ctx, id, accountIDsToCopy); err != nil {
				return nil, fmt.Errorf("failed to bind accounts to group: %w", err)
			}
		}
	}

	if s.authCacheInvalidator != nil {
		s.authCacheInvalidator.InvalidateAuthCacheByGroupID(ctx, id)
	}
	return group, nil
}

func (s *adminServiceImpl) DeleteGroup(ctx context.Context, id int64) error {
	groupKeys := s.loadGroupAuthCacheKeys(ctx, id)

	affectedUserIDs, err := s.groupRepo.DeleteCascade(ctx, id)
	if err != nil {
		return err
	}

	s.invalidateDeletedGroupSubscriptions(affectedUserIDs, id)
	s.invalidateDeletedGroupAuthCache(ctx, groupKeys)

	return nil
}
