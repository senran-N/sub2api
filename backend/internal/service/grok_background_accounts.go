package service

import "context"

type grokBackgroundAccountRepo interface {
	ListByPlatform(ctx context.Context, platform string) ([]Account, error)
}

type grokBackgroundAccountStatusRepo interface {
	ListByPlatformStatuses(ctx context.Context, platform string, statuses []string) ([]Account, error)
}

func listGrokBackgroundAccounts(ctx context.Context, repo grokBackgroundAccountRepo) ([]Account, error) {
	if repo == nil {
		return nil, nil
	}
	if statusRepo, ok := repo.(grokBackgroundAccountStatusRepo); ok {
		return statusRepo.ListByPlatformStatuses(ctx, PlatformGrok, []string{StatusActive, StatusError})
	}
	return repo.ListByPlatform(ctx, PlatformGrok)
}
