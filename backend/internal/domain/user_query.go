package domain

// UserListFilters contains all filter options for listing users.
type UserListFilters struct {
	Status     string           // User status filter
	Role       string           // User role filter
	Search     string           // Search in email, username
	GroupName  string           // Filter by allowed group name (fuzzy match)
	Attributes map[int64]string // Custom attribute filters: attributeID -> value
	// IncludeSubscriptions controls whether ListWithFilters should load active subscriptions.
	// For large datasets this can be expensive; admin list pages should enable it on demand.
	// nil means not specified (default: load subscriptions for backward compatibility).
	IncludeSubscriptions *bool
}
