package service

import "context"

func (s *GatewayService) hydrateSelectedAccount(ctx context.Context, account *Account) (*Account, error) {
	if account == nil || s.schedulerSnapshot == nil {
		return account, nil
	}
	hydrated, err := s.schedulerSnapshot.GetAccount(ctx, account.ID)
	if err != nil {
		return nil, err
	}
	if hydrated == nil {
		return nil, nil
	}
	return hydrated, nil
}

func (s *GatewayService) hydrateSelectedAccountOrNil(ctx context.Context, account *Account) *Account {
	hydrated, err := s.hydrateSelectedAccount(ctx, account)
	if err != nil {
		return nil
	}
	return hydrated
}
