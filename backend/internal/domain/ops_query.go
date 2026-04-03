package domain

import (
	"strings"
	"time"
)

type OpsQueryMode string

const (
	OpsQueryModeAuto   OpsQueryMode = "auto"
	OpsQueryModeRaw    OpsQueryMode = "raw"
	OpsQueryModePreagg OpsQueryMode = "preagg"
)

func ParseOpsQueryMode(raw string) OpsQueryMode {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case string(OpsQueryModeRaw):
		return OpsQueryModeRaw
	case string(OpsQueryModePreagg):
		return OpsQueryModePreagg
	default:
		return OpsQueryModeAuto
	}
}

func (m OpsQueryMode) IsValid() bool {
	switch m {
	case OpsQueryModeAuto, OpsQueryModeRaw, OpsQueryModePreagg:
		return true
	default:
		return false
	}
}

type OpsRequestKind string

const (
	OpsRequestKindSuccess OpsRequestKind = "success"
	OpsRequestKindError   OpsRequestKind = "error"
)

type OpsRequestDetail struct {
	Kind      OpsRequestKind `json:"kind"`
	CreatedAt time.Time      `json:"created_at"`
	RequestID string         `json:"request_id"`

	Platform string `json:"platform,omitempty"`
	Model    string `json:"model,omitempty"`

	DurationMs *int `json:"duration_ms,omitempty"`
	StatusCode *int `json:"status_code,omitempty"`

	ErrorID *int64 `json:"error_id,omitempty"`

	Phase    string `json:"phase,omitempty"`
	Severity string `json:"severity,omitempty"`
	Message  string `json:"message,omitempty"`

	UserID    *int64 `json:"user_id,omitempty"`
	APIKeyID  *int64 `json:"api_key_id,omitempty"`
	AccountID *int64 `json:"account_id,omitempty"`
	GroupID   *int64 `json:"group_id,omitempty"`

	Stream bool `json:"stream"`
}

type OpsRequestDetailFilter struct {
	StartTime *time.Time
	EndTime   *time.Time

	Kind string

	Platform string
	GroupID  *int64

	UserID    *int64
	APIKeyID  *int64
	AccountID *int64

	Model     string
	RequestID string
	Query     string

	MinDurationMs *int
	MaxDurationMs *int

	Sort string

	Page     int
	PageSize int
}

func (f *OpsRequestDetailFilter) Normalize() (page, pageSize int, startTime, endTime time.Time) {
	page = 1
	pageSize = 50
	endTime = time.Now()
	startTime = endTime.Add(-1 * time.Hour)

	if f == nil {
		return page, pageSize, startTime, endTime
	}

	if f.Page > 0 {
		page = f.Page
	}
	if f.PageSize > 0 {
		pageSize = f.PageSize
	}
	if pageSize > 100 {
		pageSize = 100
	}

	if f.EndTime != nil {
		endTime = *f.EndTime
	}
	if f.StartTime != nil {
		startTime = *f.StartTime
	} else if f.EndTime != nil {
		startTime = endTime.Add(-1 * time.Hour)
	}

	if startTime.After(endTime) {
		startTime, endTime = endTime, startTime
	}

	return page, pageSize, startTime, endTime
}

type OpsRequestDetailList struct {
	Items    []*OpsRequestDetail `json:"items"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}

type OpsDashboardFilter struct {
	StartTime time.Time
	EndTime   time.Time

	Platform string
	GroupID  *int64

	QueryMode OpsQueryMode
}

type OpsOpenAITokenStatsFilter struct {
	TimeRange string
	StartTime time.Time
	EndTime   time.Time

	Platform string
	GroupID  *int64

	Page     int
	PageSize int

	TopN int
}

func (f *OpsOpenAITokenStatsFilter) IsTopNMode() bool {
	return f != nil && f.TopN > 0
}

type OpsOpenAITokenStatsItem struct {
	Model                  string   `json:"model"`
	RequestCount           int64    `json:"request_count"`
	AvgTokensPerSec        *float64 `json:"avg_tokens_per_sec"`
	AvgFirstTokenMs        *float64 `json:"avg_first_token_ms"`
	TotalOutputTokens      int64    `json:"total_output_tokens"`
	AvgDurationMs          int64    `json:"avg_duration_ms"`
	RequestsWithFirstToken int64    `json:"requests_with_first_token"`
}

type OpsOpenAITokenStatsResponse struct {
	TimeRange string    `json:"time_range"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`

	Platform string `json:"platform,omitempty"`
	GroupID  *int64 `json:"group_id,omitempty"`

	Items []*OpsOpenAITokenStatsItem `json:"items"`

	Total int64 `json:"total"`

	Page     int `json:"page,omitempty"`
	PageSize int `json:"page_size,omitempty"`

	TopN *int `json:"top_n,omitempty"`
}

type OpsAlertEventFilter struct {
	Limit int

	BeforeFiredAt *time.Time
	BeforeID      *int64

	Status    string
	Severity  string
	EmailSent *bool

	StartTime *time.Time
	EndTime   *time.Time

	Platform string
	GroupID  *int64
}

type OpsSystemLogFilter struct {
	StartTime *time.Time
	EndTime   *time.Time

	Level     string
	Component string

	RequestID       string
	ClientRequestID string
	UserID          *int64
	AccountID       *int64
	Platform        string
	Model           string
	Query           string

	Page     int
	PageSize int
}

type OpsSystemLogCleanupFilter struct {
	StartTime *time.Time
	EndTime   *time.Time

	Level     string
	Component string

	RequestID       string
	ClientRequestID string
	UserID          *int64
	AccountID       *int64
	Platform        string
	Model           string
	Query           string
}
