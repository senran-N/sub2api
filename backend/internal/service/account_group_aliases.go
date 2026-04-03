package service

import "github.com/senran-N/sub2api/internal/domain"

type AccountGroupLink = domain.AccountGroupLink

func NewAccountGroup(link AccountGroupLink, account *Account, group *Group) AccountGroup {
	return AccountGroup{
		AccountID: link.AccountID,
		GroupID:   link.GroupID,
		Priority:  link.Priority,
		CreatedAt: link.CreatedAt,
		Account:   account,
		Group:     group,
	}
}

func (ag AccountGroup) Link() AccountGroupLink {
	return AccountGroupLink{
		AccountID: ag.AccountID,
		GroupID:   ag.GroupID,
		Priority:  ag.Priority,
		CreatedAt: ag.CreatedAt,
	}
}
