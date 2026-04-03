package service

import (
	"context"
	"errors"
	"fmt"
)

type groupExistenceBatchChecker interface {
	ExistsByIDs(ctx context.Context, ids []int64) (map[int64]bool, error)
}

func validateAccountGroupBindings(ctx context.Context, groupRepo GroupRepository, accountType string, groupIDs []int64) error {
	if err := validateGroupIDsExist(ctx, groupRepo, groupIDs); err != nil {
		return err
	}
	return validateAPIKeyGroupCompatibility(ctx, groupRepo, accountType, groupIDs)
}

func validateGroupIDsExist(ctx context.Context, groupRepo GroupRepository, groupIDs []int64) error {
	if len(groupIDs) == 0 {
		return nil
	}
	if groupRepo == nil {
		return errors.New("group repository not configured")
	}

	if batchChecker, ok := groupRepo.(groupExistenceBatchChecker); ok {
		existsByID, err := batchChecker.ExistsByIDs(ctx, groupIDs)
		if err != nil {
			return fmt.Errorf("check groups exists: %w", err)
		}
		for _, groupID := range groupIDs {
			if groupID <= 0 || !existsByID[groupID] {
				return fmt.Errorf("get group: %w", ErrGroupNotFound)
			}
		}
		return nil
	}

	for _, groupID := range groupIDs {
		if groupID <= 0 {
			return fmt.Errorf("get group: %w", ErrGroupNotFound)
		}
		group, err := groupRepo.GetByIDLite(ctx, groupID)
		if err != nil {
			return fmt.Errorf("get group: %w", err)
		}
		if group == nil {
			return fmt.Errorf("get group: %w", ErrGroupNotFound)
		}
	}
	return nil
}

func validateAPIKeyGroupCompatibility(ctx context.Context, groupRepo GroupRepository, accountType string, groupIDs []int64) error {
	if accountType != AccountTypeAPIKey || len(groupIDs) == 0 {
		return nil
	}
	if groupRepo == nil {
		return errors.New("group repository not configured")
	}

	for _, groupID := range groupIDs {
		group, err := groupRepo.GetByIDLite(ctx, groupID)
		if err != nil {
			return fmt.Errorf("get group: %w", err)
		}
		if group == nil {
			return fmt.Errorf("get group: %w", ErrGroupNotFound)
		}
		if group.RequireOAuthOnly && groupPlatformSupportsOAuthOnly(group.Platform) {
			return fmt.Errorf("分组 [%s] 仅允许 OAuth 账号，apikey 类型账号无法加入", group.Name)
		}
	}
	return nil
}
