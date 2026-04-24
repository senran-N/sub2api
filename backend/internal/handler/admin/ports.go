package admin

import (
	"context"

	"github.com/senran-N/sub2api/internal/service"
)

// Narrow handler ports keep each admin handler coupled only to the operations
// it actually consumes, instead of the full AdminService surface.

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
