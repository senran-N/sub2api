package service

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/senran-N/sub2api/internal/config"
)

// IPRiskResult holds the aggregated IP risk assessment from all enabled checks.
type IPRiskResult struct {
	IPType         string // residential, datacenter, mobile, vpn, tor
	FraudScore     int    // 0-100 composite risk score
	IsProxy        bool
	IsHosting      bool
	IsMobile       bool
	ISP            string
	Org            string
	AS             string
	AbuseScore     int    // 0-100 from AbuseIPDB (0 if unavailable)
	AbuseReports   int    // total reports from AbuseIPDB
	DNSLeakRisk    string // none, possible, detected
	DNSResolverGeo string // country code of DNS resolver
	Items          []ProxyQualityCheckItem
}

// Known datacenter/cloud AS name patterns.
var datacenterASKeywords = []string{
	"amazon", "aws", "google cloud", "google llc", "microsoft",
	"digitalocean", "ovh", "hetzner", "vultr", "linode", "akamai",
	"cloudflare", "oracle cloud", "alibaba", "tencent", "huawei cloud",
	"contabo", "scaleway", "upcloud", "kamatera", "hostinger",
}

const (
	ipRiskCheckTimeout  = 8 * time.Second
	abuseIPDBEndpoint   = "https://api.abuseipdb.com/api/v2/check"
	cloudflareTraceURL  = "https://1.1.1.1/cdn-cgi/trace"
	dnsLeakMaxBodyBytes = 2048
	abuseMaxBodyBytes   = 8192
)

// IPRiskService assesses the risk profile of a proxy's exit IP.
type IPRiskService struct {
	abuseIPDBKey       string
	enableIPTypeCheck  bool
	enableAbuseCheck   bool
	enableDNSLeakCheck bool
}

// NewIPRiskService creates an IPRiskService from config.
func NewIPRiskService(cfg *config.Config) *IPRiskService {
	s := &IPRiskService{
		enableIPTypeCheck:  true,
		enableDNSLeakCheck: true,
	}
	if cfg != nil {
		s.abuseIPDBKey = cfg.IPRisk.AbuseIPDBAPIKey
		s.enableIPTypeCheck = cfg.IPRisk.EnableIPTypeCheck
		s.enableAbuseCheck = cfg.IPRisk.EnableAbuseCheck
		s.enableDNSLeakCheck = cfg.IPRisk.EnableDNSLeakCheck
	}
	// Auto-enable abuse check when API key is provided
	if s.abuseIPDBKey != "" {
		s.enableAbuseCheck = true
	}
	return s
}

// AssessIP runs all enabled IP risk checks and returns the aggregated result.
// proxyClient is the HTTP client configured to use the proxy (for DNS leak check).
// exitInfo provides the ip-api.com extended data from the probe step.
func (s *IPRiskService) AssessIP(ctx context.Context, proxyClient *http.Client, exitInfo *ProxyExitInfo) *IPRiskResult {
	result := &IPRiskResult{
		IPType:      "unknown",
		DNSLeakRisk: "none",
	}
	if exitInfo == nil {
		return result
	}

	result.ISP = exitInfo.ISP
	result.Org = exitInfo.Org
	result.AS = exitInfo.AS
	result.IsProxy = exitInfo.Proxy
	result.IsHosting = exitInfo.Hosting
	result.IsMobile = exitInfo.Mobile

	var mu sync.Mutex
	var wg sync.WaitGroup

	appendItem := func(item ProxyQualityCheckItem) {
		mu.Lock()
		result.Items = append(result.Items, item)
		mu.Unlock()
	}

	// IP type classification
	if s.enableIPTypeCheck {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ipType, item := s.classifyIPType(exitInfo)
			mu.Lock()
			result.IPType = ipType
			mu.Unlock()
			appendItem(item)
		}()
	}

	// AbuseIPDB check
	if s.enableAbuseCheck && s.abuseIPDBKey != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			checkCtx, cancel := context.WithTimeout(ctx, ipRiskCheckTimeout)
			defer cancel()
			abuseScore, abuseReports, item := s.checkAbuseIPDB(checkCtx, exitInfo.IP)
			mu.Lock()
			result.AbuseScore = abuseScore
			result.AbuseReports = abuseReports
			mu.Unlock()
			appendItem(item)
		}()
	}

	// DNS leak check
	if s.enableDNSLeakCheck && proxyClient != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			checkCtx, cancel := context.WithTimeout(ctx, ipRiskCheckTimeout)
			defer cancel()
			leakRisk, resolverGeo, item := s.checkDNSLeak(checkCtx, proxyClient, exitInfo.CountryCode)
			mu.Lock()
			result.DNSLeakRisk = leakRisk
			result.DNSResolverGeo = resolverGeo
			mu.Unlock()
			appendItem(item)
		}()
	}

	wg.Wait()

	result.FraudScore = s.computeFraudScore(result)
	return result
}

// classifyIPType determines the IP type based on ip-api.com extended fields and AS name heuristics.
func (s *IPRiskService) classifyIPType(exitInfo *ProxyExitInfo) (string, ProxyQualityCheckItem) {
	item := ProxyQualityCheckItem{
		Target:   "ip_type",
		Category: "ip_risk",
	}

	ipType := "residential"
	switch {
	case exitInfo.Proxy:
		ipType = "vpn"
	case exitInfo.Hosting:
		ipType = "datacenter"
	case exitInfo.Mobile:
		ipType = "mobile"
	default:
		// Heuristic: check AS name for known datacenter providers
		asLower := strings.ToLower(exitInfo.AS)
		orgLower := strings.ToLower(exitInfo.Org)
		for _, keyword := range datacenterASKeywords {
			if strings.Contains(asLower, keyword) || strings.Contains(orgLower, keyword) {
				ipType = "datacenter"
				break
			}
		}
	}

	switch ipType {
	case "residential":
		item.Status = "pass"
		item.Message = fmt.Sprintf("家庭宽带 IP (ISP: %s)", exitInfo.ISP)
	case "mobile":
		item.Status = "pass"
		item.Message = fmt.Sprintf("移动网络 IP (ISP: %s)", exitInfo.ISP)
	case "datacenter":
		item.Status = "warn"
		item.Message = fmt.Sprintf("数据中心 IP (AS: %s)", exitInfo.AS)
	case "vpn":
		item.Status = "warn"
		item.Message = fmt.Sprintf("VPN/代理 IP (ISP: %s)", exitInfo.ISP)
	default:
		item.Status = "fail"
		item.Message = "Tor 出口节点"
	}

	return ipType, item
}

// checkAbuseIPDB queries the AbuseIPDB API for abuse history.
func (s *IPRiskService) checkAbuseIPDB(ctx context.Context, exitIP string) (int, int, ProxyQualityCheckItem) {
	item := ProxyQualityCheckItem{
		Target:   "abuse_check",
		Category: "abuse",
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, abuseIPDBEndpoint+"?ipAddress="+exitIP+"&maxAgeInDays=90", nil)
	if err != nil {
		item.Status = "skip"
		item.Message = fmt.Sprintf("构建请求失败: %v", err)
		return 0, 0, item
	}
	req.Header.Set("Key", s.abuseIPDBKey)
	req.Header.Set("Accept", "application/json")

	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		item.Status = "skip"
		item.LatencyMs = time.Since(start).Milliseconds()
		item.Message = fmt.Sprintf("AbuseIPDB 请求失败: %v", err)
		return 0, 0, item
	}
	defer func() { _ = resp.Body.Close() }()
	item.LatencyMs = time.Since(start).Milliseconds()
	item.HTTPStatus = resp.StatusCode

	if resp.StatusCode != http.StatusOK {
		item.Status = "skip"
		item.Message = fmt.Sprintf("AbuseIPDB 返回 %d", resp.StatusCode)
		return 0, 0, item
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, abuseMaxBodyBytes))
	if err != nil {
		item.Status = "skip"
		item.Message = fmt.Sprintf("读取响应失败: %v", err)
		return 0, 0, item
	}

	var abuseResp struct {
		Data struct {
			AbuseConfidenceScore int  `json:"abuseConfidenceScore"`
			TotalReports         int  `json:"totalReports"`
			IsPublic             bool `json:"isPublic"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &abuseResp); err != nil {
		item.Status = "skip"
		item.Message = fmt.Sprintf("解析响应失败: %v", err)
		return 0, 0, item
	}

	score := abuseResp.Data.AbuseConfidenceScore
	reports := abuseResp.Data.TotalReports

	switch {
	case score == 0 && reports == 0:
		item.Status = "pass"
		item.Message = "无滥用记录"
	case score <= 25:
		item.Status = "pass"
		item.Message = fmt.Sprintf("低风险 (置信度 %d%%, %d 条举报)", score, reports)
	case score <= 50:
		item.Status = "warn"
		item.Message = fmt.Sprintf("中风险 (置信度 %d%%, %d 条举报)", score, reports)
	default:
		item.Status = "fail"
		item.Message = fmt.Sprintf("高风险 (置信度 %d%%, %d 条举报)", score, reports)
	}

	return score, reports, item
}

// checkDNSLeak detects DNS leaks by comparing the DNS resolver's location with the proxy exit location.
func (s *IPRiskService) checkDNSLeak(ctx context.Context, proxyClient *http.Client, exitCountryCode string) (string, string, ProxyQualityCheckItem) {
	item := ProxyQualityCheckItem{
		Target:   "dns_leak",
		Category: "dns_leak",
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cloudflareTraceURL, nil)
	if err != nil {
		item.Status = "skip"
		item.Message = fmt.Sprintf("构建请求失败: %v", err)
		return "none", "", item
	}

	start := time.Now()
	resp, err := proxyClient.Do(req)
	if err != nil {
		item.Status = "skip"
		item.LatencyMs = time.Since(start).Milliseconds()
		item.Message = fmt.Sprintf("DNS 泄漏检测请求失败: %v", err)
		return "none", "", item
	}
	defer func() { _ = resp.Body.Close() }()
	item.LatencyMs = time.Since(start).Milliseconds()
	item.HTTPStatus = resp.StatusCode

	if resp.StatusCode != http.StatusOK {
		item.Status = "skip"
		item.Message = fmt.Sprintf("Cloudflare trace 返回 %d", resp.StatusCode)
		return "none", "", item
	}

	// Parse key=value format from Cloudflare trace
	resolverGeo := ""
	scanner := bufio.NewScanner(io.LimitReader(resp.Body, dnsLeakMaxBodyBytes))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "loc=") {
			resolverGeo = strings.TrimPrefix(line, "loc=")
			break
		}
	}

	if resolverGeo == "" {
		item.Status = "skip"
		item.Message = "无法解析 DNS 解析器地理位置"
		return "none", "", item
	}

	exitUpper := strings.ToUpper(exitCountryCode)
	resolverUpper := strings.ToUpper(resolverGeo)

	if exitUpper == "" || resolverUpper == "" {
		item.Status = "skip"
		item.Message = "缺少地理位置数据，无法比对"
		return "none", resolverGeo, item
	}

	if exitUpper == resolverUpper {
		item.Status = "pass"
		item.Message = fmt.Sprintf("DNS 解析器与出口 IP 同地区 (%s)", resolverGeo)
		return "none", resolverGeo, item
	}

	item.Status = "warn"
	item.Message = fmt.Sprintf("DNS 解析器 (%s) 与出口 IP (%s) 地区不一致，可能存在 DNS 泄漏", resolverGeo, exitCountryCode)
	return "possible", resolverGeo, item
}

// computeFraudScore calculates a composite fraud score (0-100) from all risk signals.
func (s *IPRiskService) computeFraudScore(result *IPRiskResult) int {
	score := 0

	// IP type contribution
	switch result.IPType {
	case "datacenter":
		score += 30
	case "vpn":
		score += 35
	case "tor":
		score += 50
	case "mobile":
		score += 5
		// residential: +0
	}

	// Proxy flag from ip-api
	if result.IsProxy && result.IPType != "vpn" {
		score += 15
	}

	// AbuseIPDB contribution (scaled)
	if result.AbuseScore > 0 {
		abuseContrib := result.AbuseScore / 3 // max ~33
		if abuseContrib > 30 {
			abuseContrib = 30
		}
		score += abuseContrib
	}

	// DNS leak contribution
	switch result.DNSLeakRisk {
	case "possible":
		score += 10
	case "detected":
		score += 20
	}

	if score > 100 {
		score = 100
	}
	return score
}
