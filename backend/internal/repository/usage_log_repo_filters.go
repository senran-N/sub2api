package repository

import (
	"fmt"
	"time"
)

type usageLogScopedFilters struct {
	UserID      int64
	APIKeyID    int64
	AccountID   int64
	GroupID     int64
	Model       string
	RequestType *int16
	Stream      *bool
	BillingType *int8
	StartTime   *time.Time
	EndTime     *time.Time
}

type usageLogScopedFilterOptions struct {
	TableAlias   string
	IncludeModel bool
	IncludeTime  bool
}

func usageLogScopedFiltersFromUsageLogFilters(filters UsageLogFilters) usageLogScopedFilters {
	return usageLogScopedFilters{
		UserID:      filters.UserID,
		APIKeyID:    filters.APIKeyID,
		AccountID:   filters.AccountID,
		GroupID:     filters.GroupID,
		Model:       filters.Model,
		RequestType: filters.RequestType,
		Stream:      filters.Stream,
		BillingType: filters.BillingType,
		StartTime:   filters.StartTime,
		EndTime:     filters.EndTime,
	}
}

func appendUsageLogScopedConditions(conditions []string, args []any, filters usageLogScopedFilters, options usageLogScopedFilterOptions) ([]string, []any) {
	column := func(name string) string {
		if options.TableAlias == "" {
			return name
		}
		return options.TableAlias + "." + name
	}

	if filters.UserID > 0 {
		conditions = append(conditions, fmt.Sprintf("%s = $%d", column("user_id"), len(args)+1))
		args = append(args, filters.UserID)
	}
	if filters.APIKeyID > 0 {
		conditions = append(conditions, fmt.Sprintf("%s = $%d", column("api_key_id"), len(args)+1))
		args = append(args, filters.APIKeyID)
	}
	if filters.AccountID > 0 {
		conditions = append(conditions, fmt.Sprintf("%s = $%d", column("account_id"), len(args)+1))
		args = append(args, filters.AccountID)
	}
	if filters.GroupID > 0 {
		conditions = append(conditions, fmt.Sprintf("%s = $%d", column("group_id"), len(args)+1))
		args = append(args, filters.GroupID)
	}
	if options.IncludeModel {
		conditions, args = appendRawUsageLogModelWhereConditionWithColumn(conditions, args, column(rawUsageLogModelColumn), filters.Model)
	}
	conditions, args = appendRequestTypeOrStreamWhereConditionWithColumns(
		conditions,
		args,
		column("request_type"),
		column("stream"),
		column("openai_ws_mode"),
		filters.RequestType,
		filters.Stream,
	)
	if filters.BillingType != nil {
		conditions = append(conditions, fmt.Sprintf("%s = $%d", column("billing_type"), len(args)+1))
		args = append(args, int16(*filters.BillingType))
	}
	if options.IncludeTime {
		if filters.StartTime != nil {
			conditions = append(conditions, fmt.Sprintf("%s >= $%d", column("created_at"), len(args)+1))
			args = append(args, *filters.StartTime)
		}
		if filters.EndTime != nil {
			conditions = append(conditions, fmt.Sprintf("%s < $%d", column("created_at"), len(args)+1))
			args = append(args, *filters.EndTime)
		}
	}
	return conditions, args
}

func appendConditionsToQuery(query string, conditions []string) string {
	for _, condition := range conditions {
		query += " AND " + condition
	}
	return query
}
