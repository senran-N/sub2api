package domain

import (
	"net"
	"net/url"
	"strconv"
	"time"
)

type Proxy struct {
	ID        int64
	Name      string
	Protocol  string
	Host      string
	Port      int
	Username  string
	Password  string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (p *Proxy) IsActive() bool {
	return p.Status == "active"
}

func (p *Proxy) URL() string {
	u := &url.URL{
		Scheme: p.Protocol,
		Host:   net.JoinHostPort(p.Host, strconv.Itoa(p.Port)),
	}
	if p.Username != "" && p.Password != "" {
		u.User = url.UserPassword(p.Username, p.Password)
	}
	return u.String()
}

type ProxyWithAccountCount struct {
	Proxy
	AccountCount   int64
	LatencyMs      *int64
	LatencyStatus  string
	LatencyMessage string
	IPAddress      string
	Country        string
	CountryCode    string
	Region         string
	City           string
	QualityStatus  string
	QualityScore   *int
	QualityGrade   string
	QualitySummary string
	QualityChecked *int64
	IPType         string
	IPRiskScore    *int
	ISP            string
}

type ProxyAccountSummary struct {
	ID       int64
	Name     string
	Platform string
	Type     string
	Notes    *string
}

type CreateProxyInput struct {
	Name     string
	Protocol string
	Host     string
	Port     int
	Username string
	Password string
}

type CreateProxyRequest struct {
	Name     string
	Protocol string
	Host     string
	Port     int
	Username string
	Password string
}

type UpdateProxyInput struct {
	Name     string
	Protocol string
	Host     string
	Port     int
	Username string
	Password string
	Status   string
}

type UpdateProxyRequest struct {
	Name     *string
	Protocol *string
	Host     *string
	Port     *int
	Username *string
	Password *string
	Status   *string
}

type ProxyBatchDeleteResult struct {
	DeletedIDs []int64                   `json:"deleted_ids"`
	Skipped    []ProxyBatchDeleteSkipped `json:"skipped"`
}

type ProxyBatchDeleteSkipped struct {
	ID     int64  `json:"id"`
	Reason string `json:"reason"`
}

type ProxyTestResult struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	LatencyMs   int64  `json:"latency_ms,omitempty"`
	IPAddress   string `json:"ip_address,omitempty"`
	City        string `json:"city,omitempty"`
	Region      string `json:"region,omitempty"`
	Country     string `json:"country,omitempty"`
	CountryCode string `json:"country_code,omitempty"`
}

type ProxyQualityCheckResult struct {
	ProxyID        int64                   `json:"proxy_id"`
	Score          int                     `json:"score"`
	Grade          string                  `json:"grade"`
	Summary        string                  `json:"summary"`
	ExitIP         string                  `json:"exit_ip,omitempty"`
	Country        string                  `json:"country,omitempty"`
	CountryCode    string                  `json:"country_code,omitempty"`
	BaseLatencyMs  int64                   `json:"base_latency_ms,omitempty"`
	PassedCount    int                     `json:"passed_count"`
	WarnCount      int                     `json:"warn_count"`
	FailedCount    int                     `json:"failed_count"`
	ChallengeCount int                     `json:"challenge_count"`
	CheckedAt      int64                   `json:"checked_at"`
	Items          []ProxyQualityCheckItem `json:"items"`
	IPType         string                  `json:"ip_type,omitempty"`
	IPRiskScore    int                     `json:"ip_risk_score,omitempty"`
	ISP            string                  `json:"isp,omitempty"`
	AS             string                  `json:"as,omitempty"`
	AbuseScore     int                     `json:"abuse_score,omitempty"`
	DNSLeakRisk    string                  `json:"dns_leak_risk,omitempty"`
	CategoryScores *QualityCategoryScores  `json:"category_scores,omitempty"`
}

type QualityCategoryScores struct {
	Reachability int `json:"reachability"`
	IPRisk       int `json:"ip_risk"`
	IPType       int `json:"ip_type"`
	AbuseHistory int `json:"abuse_history"`
	Latency      int `json:"latency"`
}

type ProxyQualityCheckItem struct {
	Target     string `json:"target"`
	Status     string `json:"status"`
	HTTPStatus int    `json:"http_status,omitempty"`
	LatencyMs  int64  `json:"latency_ms,omitempty"`
	Message    string `json:"message,omitempty"`
	CFRay      string `json:"cf_ray,omitempty"`
	Category   string `json:"category,omitempty"`
}

type ProxyExitInfo struct {
	IP          string
	City        string
	Region      string
	Country     string
	CountryCode string
	ISP         string
	Org         string
	AS          string
	Hosting     bool
	Proxy       bool
	Mobile      bool
}
