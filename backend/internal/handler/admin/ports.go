package admin

import (
	"context"

	"github.com/senran-N/sub2api/internal/service"
)

// Narrow handler ports keep each admin handler coupled only to the operations
// it actually consumes, instead of the full AdminService surface.

type accountAdminService interface {
	ListAccounts(ctx context.Context, page, pageSize int, platform, accountType, status, search string, groupID int64, privacyMode string) ([]service.Account, int64, error)
	GetAccount(ctx context.Context, id int64) (*service.Account, error)
	GetAccountsByIDs(ctx context.Context, ids []int64) ([]*service.Account, error)
	CreateAccount(ctx context.Context, input *service.CreateAccountInput) (*service.Account, error)
	UpdateAccount(ctx context.Context, id int64, input *service.UpdateAccountInput) (*service.Account, error)
	DeleteAccount(ctx context.Context, id int64) error
	GetProxy(ctx context.Context, id int64) (*service.Proxy, error)
	CreateProxy(ctx context.Context, input *service.CreateProxyInput) (*service.Proxy, error)
	UpdateProxy(ctx context.Context, id int64, input *service.UpdateProxyInput) (*service.Proxy, error)
	ListProxies(ctx context.Context, page, pageSize int, protocol, status, search string) ([]service.Proxy, int64, error)
	GetProxiesByIDs(ctx context.Context, ids []int64) ([]service.Proxy, error)
	ClearAccountError(ctx context.Context, id int64) (*service.Account, error)
	EnsureOpenAIPrivacy(ctx context.Context, account *service.Account) string
	EnsureAntigravityPrivacy(ctx context.Context, account *service.Account) string
	ForceOpenAIPrivacy(ctx context.Context, account *service.Account) string
	ForceAntigravityPrivacy(ctx context.Context, account *service.Account) string
	SetAccountSchedulable(ctx context.Context, id int64, schedulable bool) (*service.Account, error)
	BulkUpdateAccounts(ctx context.Context, input *service.BulkUpdateAccountsInput) (*service.BulkUpdateAccountsResult, error)
	CheckMixedChannelRisk(ctx context.Context, currentAccountID int64, currentAccountPlatform string, groupIDs []int64) error
	ResetAccountQuota(ctx context.Context, id int64) error
}

type groupAdminService interface {
	ListGroups(ctx context.Context, page, pageSize int, platform, status, search string, isExclusive *bool) ([]service.Group, int64, error)
	GetAllGroups(ctx context.Context) ([]service.Group, error)
	GetAllGroupsByPlatform(ctx context.Context, platform string) ([]service.Group, error)
	GetGroup(ctx context.Context, id int64) (*service.Group, error)
	CreateGroup(ctx context.Context, input *service.CreateGroupInput) (*service.Group, error)
	UpdateGroup(ctx context.Context, id int64, input *service.UpdateGroupInput) (*service.Group, error)
	DeleteGroup(ctx context.Context, id int64) error
	GetGroupAPIKeys(ctx context.Context, groupID int64, page, pageSize int) ([]service.APIKey, int64, error)
	GetGroupRateMultipliers(ctx context.Context, groupID int64) ([]service.UserGroupRateEntry, error)
	ClearGroupRateMultipliers(ctx context.Context, groupID int64) error
	BatchSetGroupRateMultipliers(ctx context.Context, groupID int64, entries []service.GroupRateMultiplierInput) error
	UpdateGroupSortOrders(ctx context.Context, updates []service.GroupSortOrderUpdate) error
}

type proxyAdminService interface {
	ListProxies(ctx context.Context, page, pageSize int, protocol, status, search string) ([]service.Proxy, int64, error)
	ListProxiesWithAccountCount(ctx context.Context, page, pageSize int, protocol, status, search string) ([]service.ProxyWithAccountCount, int64, error)
	GetAllProxies(ctx context.Context) ([]service.Proxy, error)
	GetAllProxiesWithAccountCount(ctx context.Context) ([]service.ProxyWithAccountCount, error)
	GetProxy(ctx context.Context, id int64) (*service.Proxy, error)
	GetProxiesByIDs(ctx context.Context, ids []int64) ([]service.Proxy, error)
	CreateProxy(ctx context.Context, input *service.CreateProxyInput) (*service.Proxy, error)
	UpdateProxy(ctx context.Context, id int64, input *service.UpdateProxyInput) (*service.Proxy, error)
	DeleteProxy(ctx context.Context, id int64) error
	BatchDeleteProxies(ctx context.Context, ids []int64) (*service.ProxyBatchDeleteResult, error)
	GetProxyAccounts(ctx context.Context, proxyID int64) ([]service.ProxyAccountSummary, error)
	CheckProxyExists(ctx context.Context, host string, port int, username, password string) (bool, error)
	TestProxy(ctx context.Context, id int64) (*service.ProxyTestResult, error)
	CheckProxyQuality(ctx context.Context, id int64) (*service.ProxyQualityCheckResult, error)
}

type userAdminService interface {
	ListUsers(ctx context.Context, page, pageSize int, filters service.UserListFilters) ([]service.User, int64, error)
	GetUser(ctx context.Context, id int64) (*service.User, error)
	CreateUser(ctx context.Context, input *service.CreateUserInput) (*service.User, error)
	UpdateUser(ctx context.Context, id int64, input *service.UpdateUserInput) (*service.User, error)
	DeleteUser(ctx context.Context, id int64) error
	UpdateUserBalance(ctx context.Context, userID int64, balance float64, operation string, notes string) (*service.User, error)
	GetUserAPIKeys(ctx context.Context, userID int64, page, pageSize int) ([]service.APIKey, int64, error)
	GetUserUsageStats(ctx context.Context, userID int64, period string) (any, error)
	GetUserBalanceHistory(ctx context.Context, userID int64, page, pageSize int, codeType string) ([]service.RedeemCode, int64, float64, error)
	ReplaceUserGroup(ctx context.Context, userID, oldGroupID, newGroupID int64) (*service.ReplaceUserGroupResult, error)
}

type usageUserSearchService interface {
	ListUsers(ctx context.Context, page, pageSize int, filters service.UserListFilters) ([]service.User, int64, error)
}

type adminAPIKeyService interface {
	AdminUpdateAPIKeyGroupID(ctx context.Context, keyID int64, groupID *int64) (*service.AdminUpdateAPIKeyGroupIDResult, error)
}

type openAIOAuthAdminService interface {
	GetProxy(ctx context.Context, id int64) (*service.Proxy, error)
	GetAccount(ctx context.Context, id int64) (*service.Account, error)
	UpdateAccount(ctx context.Context, id int64, input *service.UpdateAccountInput) (*service.Account, error)
	CreateAccount(ctx context.Context, input *service.CreateAccountInput) (*service.Account, error)
}

type redeemCodeAdminService interface {
	ListRedeemCodes(ctx context.Context, page, pageSize int, codeType, status, search string) ([]service.RedeemCode, int64, error)
	GetRedeemCode(ctx context.Context, id int64) (*service.RedeemCode, error)
	GenerateRedeemCodes(ctx context.Context, input *service.GenerateRedeemCodesInput) ([]service.RedeemCode, error)
	DeleteRedeemCode(ctx context.Context, id int64) error
	BatchDeleteRedeemCodes(ctx context.Context, ids []int64) (int64, error)
	ExpireRedeemCode(ctx context.Context, id int64) (*service.RedeemCode, error)
}
