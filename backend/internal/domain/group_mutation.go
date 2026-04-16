package domain

// GroupSortOrderUpdate describes one group sort-order mutation.
type GroupSortOrderUpdate struct {
	ID        int64 `json:"id"`
	SortOrder int   `json:"sort_order"`
}

type CreateGroupInput struct {
	Name                            string
	Description                     string
	Platform                        string
	RateMultiplier                  float64
	IsExclusive                     bool
	SubscriptionType                string
	DailyLimitUSD                   *float64
	WeeklyLimitUSD                  *float64
	MonthlyLimitUSD                 *float64
	ImagePrice1K                    *float64
	ImagePrice2K                    *float64
	ImagePrice4K                    *float64
	ClaudeCodeOnly                  bool
	FallbackGroupID                 *int64
	FallbackGroupIDOnInvalidRequest *int64
	ModelRouting                    map[string][]int64
	ModelRoutingEnabled             bool
	MCPXMLInject                    *bool
	SupportedModelScopes            []string
	AllowMessagesDispatch           bool
	DefaultMappedModel              string
	MessagesDispatchModelConfig     OpenAIMessagesDispatchModelConfig
	RequireOAuthOnly                bool
	RequirePrivacySet               bool
	CopyAccountsFromGroupIDs        []int64
}

type CreateGroupRequest struct {
	Name           string
	Description    string
	RateMultiplier float64
	IsExclusive    bool
}

type UpdateGroupInput struct {
	Name                            string
	Description                     string
	Platform                        string
	RateMultiplier                  *float64
	IsExclusive                     *bool
	Status                          string
	SubscriptionType                string
	DailyLimitUSD                   *float64
	WeeklyLimitUSD                  *float64
	MonthlyLimitUSD                 *float64
	ImagePrice1K                    *float64
	ImagePrice2K                    *float64
	ImagePrice4K                    *float64
	ClaudeCodeOnly                  *bool
	FallbackGroupID                 *int64
	FallbackGroupIDOnInvalidRequest *int64
	ModelRouting                    map[string][]int64
	ModelRoutingEnabled             *bool
	MCPXMLInject                    *bool
	SupportedModelScopes            *[]string
	AllowMessagesDispatch           *bool
	DefaultMappedModel              *string
	MessagesDispatchModelConfig     *OpenAIMessagesDispatchModelConfig
	RequireOAuthOnly                *bool
	RequirePrivacySet               *bool
	CopyAccountsFromGroupIDs        []int64
}

type UpdateGroupRequest struct {
	Name           *string
	Description    *string
	RateMultiplier *float64
	IsExclusive    *bool
	Status         *string
}
