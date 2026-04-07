package admin

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/response"
)

func TestParseUsageLogFiltersFromQuery(t *testing.T) {
	t.Parallel()

	ctx := newUsageQueryTestContext(url.Values{
		"user_id":      []string{"1"},
		"api_key_id":   []string{"2"},
		"account_id":   []string{"3"},
		"group_id":     []string{"4"},
		"model":        []string{"gpt-5"},
		"stream":       []string{"true"},
		"billing_type": []string{"7"},
	})

	filters, err := parseUsageLogFiltersFromQuery(ctx)
	if err != nil {
		t.Fatalf("parseUsageLogFiltersFromQuery() error = %v", err)
	}

	if filters.UserID != 1 || filters.APIKeyID != 2 || filters.AccountID != 3 || filters.GroupID != 4 {
		t.Fatalf("unexpected id filters: %+v", filters)
	}
	if filters.Model != "gpt-5" {
		t.Fatalf("expected model gpt-5, got %q", filters.Model)
	}
	if filters.RequestType != nil {
		t.Fatalf("expected request type to be nil, got %v", *filters.RequestType)
	}
	if filters.Stream == nil || !*filters.Stream {
		t.Fatalf("expected stream=true, got %+v", filters.Stream)
	}
	if filters.BillingType == nil || *filters.BillingType != 7 {
		t.Fatalf("expected billing type 7, got %+v", filters.BillingType)
	}
}

func TestParseUsageLogFiltersFromQueryInvalidStream(t *testing.T) {
	t.Parallel()

	ctx := newUsageQueryTestContext(url.Values{
		"stream": []string{"definitely-not-bool"},
	})

	_, err := parseUsageLogFiltersFromQuery(ctx)
	if err == nil {
		t.Fatal("expected stream parse error")
	}

	var queryErr *usageQueryParamError
	if !errors.As(err, &queryErr) {
		t.Fatalf("expected usageQueryParamError, got %T", err)
	}
	if queryErr.param != "stream" {
		t.Fatalf("expected stream param, got %q", queryErr.param)
	}
}

func TestParseUsageListDateRangeUsesHalfOpenEnd(t *testing.T) {
	t.Parallel()

	ctx := newUsageQueryTestContext(url.Values{
		"timezone":   []string{"UTC"},
		"start_date": []string{"2026-04-07"},
		"end_date":   []string{"2026-04-08"},
	})

	startTime, endTime, err := parseUsageListDateRange(ctx)
	if err != nil {
		t.Fatalf("parseUsageListDateRange() error = %v", err)
	}

	assertTimeEqual(t, startTime, time.Date(2026, 4, 7, 0, 0, 0, 0, time.UTC))
	assertTimeEqual(t, endTime, time.Date(2026, 4, 9, 0, 0, 0, 0, time.UTC))
}

func TestParseUsageStatsDateRangeUsesHalfOpenEnd(t *testing.T) {
	t.Parallel()

	ctx := newUsageQueryTestContext(url.Values{
		"timezone":   []string{"UTC"},
		"start_date": []string{"2026-04-01"},
		"end_date":   []string{"2026-04-07"},
	})

	startTime, endTime, err := parseUsageStatsDateRange(ctx)
	if err != nil {
		t.Fatalf("parseUsageStatsDateRange() error = %v", err)
	}

	if !startTime.Equal(time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected start time: %s", startTime)
	}
	if !endTime.Equal(time.Date(2026, 4, 8, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected end time: %s", endTime)
	}
}

func TestRespondUsageQueryParamErrorUsesOverrideMessage(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	handled := respondUsageQueryParamError(
		ctx,
		newUsageQueryParamError("stream", errors.New("bad bool")),
		usageLogFilterParamMessages,
	)
	if !handled {
		t.Fatal("expected error to be handled")
	}

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}

	var payload response.Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !strings.Contains(payload.Message, "Invalid stream value") {
		t.Fatalf("unexpected message: %q", payload.Message)
	}
}

func newUsageQueryTestContext(values url.Values) *gin.Context {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/admin/usage?"+values.Encode(), nil)
	return ctx
}

func assertTimeEqual(t *testing.T, actual *time.Time, expected time.Time) {
	t.Helper()

	if actual == nil {
		t.Fatal("expected non-nil time")
	}
	if !actual.Equal(expected) {
		t.Fatalf("unexpected time: got %s want %s", actual, expected)
	}
}
