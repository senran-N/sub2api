package admin

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/timezone"
	"github.com/senran-N/sub2api/internal/pkg/usagestats"
	"github.com/senran-N/sub2api/internal/service"
)

type usageQueryParamError struct {
	param string
	cause error
}

func (e *usageQueryParamError) Error() string {
	return fmt.Sprintf("invalid %s: %v", e.param, e.cause)
}

func (e *usageQueryParamError) Unwrap() error {
	return e.cause
}

func newUsageQueryParamError(param string, cause error) error {
	return &usageQueryParamError{param: param, cause: cause}
}

func parseOptionalInt64Query(c *gin.Context, key string) (int64, bool, error) {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return 0, false, nil
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, false, newUsageQueryParamError(key, err)
	}

	return parsed, true, nil
}

func parseOptionalInt8Query(c *gin.Context, key string) (*int8, error) {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return nil, nil
	}

	parsed, err := strconv.ParseInt(value, 10, 8)
	if err != nil {
		return nil, newUsageQueryParamError(key, err)
	}

	result := int8(parsed)
	return &result, nil
}

func parseUsageRequestModeQuery(c *gin.Context) (*int16, *bool, error) {
	requestTypeRaw := strings.TrimSpace(c.Query("request_type"))
	if requestTypeRaw != "" {
		parsed, err := service.ParseUsageRequestType(requestTypeRaw)
		if err != nil {
			return nil, nil, newUsageQueryParamError("request_type", err)
		}

		value := int16(parsed)
		return &value, nil, nil
	}

	streamRaw := strings.TrimSpace(c.Query("stream"))
	if streamRaw == "" {
		return nil, nil, nil
	}

	parsed, err := strconv.ParseBool(streamRaw)
	if err != nil {
		return nil, nil, newUsageQueryParamError("stream", err)
	}

	return nil, &parsed, nil
}

func parseUsageLogFiltersFromQuery(c *gin.Context) (usagestats.UsageLogFilters, error) {
	userID, _, err := parseOptionalInt64Query(c, "user_id")
	if err != nil {
		return usagestats.UsageLogFilters{}, err
	}

	apiKeyID, _, err := parseOptionalInt64Query(c, "api_key_id")
	if err != nil {
		return usagestats.UsageLogFilters{}, err
	}

	accountID, _, err := parseOptionalInt64Query(c, "account_id")
	if err != nil {
		return usagestats.UsageLogFilters{}, err
	}

	groupID, _, err := parseOptionalInt64Query(c, "group_id")
	if err != nil {
		return usagestats.UsageLogFilters{}, err
	}

	requestType, stream, err := parseUsageRequestModeQuery(c)
	if err != nil {
		return usagestats.UsageLogFilters{}, err
	}

	billingType, err := parseOptionalInt8Query(c, "billing_type")
	if err != nil {
		return usagestats.UsageLogFilters{}, err
	}

	return usagestats.UsageLogFilters{
		UserID:      userID,
		APIKeyID:    apiKeyID,
		AccountID:   accountID,
		GroupID:     groupID,
		Model:       strings.TrimSpace(c.Query("model")),
		RequestType: requestType,
		Stream:      stream,
		BillingType: billingType,
	}, nil
}

func parseUsageDateParam(raw, key, timezoneParam string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}

	parsed, err := timezone.ParseInUserLocation("2006-01-02", raw, timezoneParam)
	if err != nil {
		return nil, newUsageQueryParamError(key, err)
	}

	return &parsed, nil
}

func parseUsageListDateRange(c *gin.Context) (*time.Time, *time.Time, error) {
	userTZ := strings.TrimSpace(c.Query("timezone"))

	startDateStr := strings.TrimSpace(c.Query("start_date"))
	startTime, err := parseUsageDateParam(startDateStr, "start_date", userTZ)
	if err != nil {
		return nil, nil, err
	}

	endDateStr := strings.TrimSpace(c.Query("end_date"))
	endTime, err := parseUsageDateParam(endDateStr, "end_date", userTZ)
	if err != nil {
		return nil, nil, err
	}
	if endTime != nil {
		adjusted := endTime.AddDate(0, 0, 1)
		endTime = &adjusted
	}

	return startTime, endTime, nil
}

func parseUsageStatsDateRange(c *gin.Context) (time.Time, time.Time, error) {
	userTZ := strings.TrimSpace(c.Query("timezone"))
	now := timezone.NowInUserLocation(userTZ)

	startDateStr := strings.TrimSpace(c.Query("start_date"))
	endDateStr := strings.TrimSpace(c.Query("end_date"))
	if startDateStr != "" && endDateStr != "" {
		startTime, err := parseUsageDateParam(startDateStr, "start_date", userTZ)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}

		endTime, err := parseUsageDateParam(endDateStr, "end_date", userTZ)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}

		return *startTime, endTime.AddDate(0, 0, 1), nil
	}

	startTime := timezone.StartOfDayInUserLocation(now, userTZ)
	switch strings.TrimSpace(c.DefaultQuery("period", "today")) {
	case "week":
		startTime = now.AddDate(0, 0, -7)
	case "month":
		startTime = now.AddDate(0, -1, 0)
	}

	return startTime, now, nil
}
