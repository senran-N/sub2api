package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSortAccountsByPriorityLoadAndLastUsed_OrdersByPriorityThenLoadThenLRU_Runtime(t *testing.T) {
	now := time.Now()
	earlier := now.Add(-time.Minute)
	later := now.Add(-time.Second)

	accounts := []accountWithLoad{
		{account: &Account{ID: 1, Priority: 2, LastUsedAt: &now}, loadInfo: &AccountLoadInfo{AccountID: 1, LoadRate: 10}},
		{account: &Account{ID: 2, Priority: 1, LastUsedAt: &later}, loadInfo: &AccountLoadInfo{AccountID: 2, LoadRate: 20}},
		{account: &Account{ID: 3, Priority: 1, LastUsedAt: &earlier}, loadInfo: &AccountLoadInfo{AccountID: 3, LoadRate: 10}},
	}

	sortAccountsByPriorityLoadAndLastUsed(accounts, false)

	require.Equal(t, int64(3), accounts[0].account.ID)
	require.Equal(t, int64(2), accounts[1].account.ID)
	require.Equal(t, int64(1), accounts[2].account.ID)
}

func TestSortAccountsByPriorityLoadAndLastUsed_PrefersOAuthWithinSameGroup_Runtime(t *testing.T) {
	accounts := []accountWithLoad{
		{account: &Account{ID: 1, Priority: 1, Type: AccountTypeAPIKey}, loadInfo: &AccountLoadInfo{AccountID: 1, LoadRate: 10}},
		{account: &Account{ID: 2, Priority: 1, Type: AccountTypeOAuth}, loadInfo: &AccountLoadInfo{AccountID: 2, LoadRate: 10}},
	}

	sortAccountsByPriorityLoadAndLastUsed(accounts, true)

	require.Equal(t, int64(2), accounts[0].account.ID)
	require.Equal(t, int64(1), accounts[1].account.ID)
}
