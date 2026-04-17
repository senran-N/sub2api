package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type proxyRepoGetByIDStub struct {
	proxyRepoStub
	proxy  *Proxy
	getErr error
	lastID int64
}

func (s *proxyRepoGetByIDStub) GetByID(ctx context.Context, id int64) (*Proxy, error) {
	s.lastID = id
	if s.getErr != nil {
		return nil, s.getErr
	}
	return s.proxy, nil
}

type proxyLatencyCacheStub struct {
	entries map[int64]*ProxyLatencyInfo
}

func (s *proxyLatencyCacheStub) GetProxyLatencies(ctx context.Context, proxyIDs []int64) (map[int64]*ProxyLatencyInfo, error) {
	if s.entries == nil {
		s.entries = make(map[int64]*ProxyLatencyInfo)
	}
	result := make(map[int64]*ProxyLatencyInfo, len(proxyIDs))
	for _, id := range proxyIDs {
		if info, ok := s.entries[id]; ok && info != nil {
			cloned := *info
			result[id] = &cloned
		}
	}
	return result, nil
}

func (s *proxyLatencyCacheStub) SetProxyLatency(ctx context.Context, proxyID int64, info *ProxyLatencyInfo) error {
	if s.entries == nil {
		s.entries = make(map[int64]*ProxyLatencyInfo)
	}
	cloned := *info
	s.entries[proxyID] = &cloned
	return nil
}

type proxyExitInfoProberStub struct {
	exitInfo *ProxyExitInfo
	latency  int64
	err      error
	lastURL  string
}

func (s *proxyExitInfoProberStub) ProbeProxy(ctx context.Context, proxyURL string) (*ProxyExitInfo, int64, error) {
	s.lastURL = proxyURL
	return s.exitInfo, s.latency, s.err
}

func TestAdminServiceTestProxy_WithoutProberReturnsFailedResult(t *testing.T) {
	t.Parallel()

	repo := &proxyRepoGetByIDStub{
		proxy: &Proxy{
			ID:       42,
			Protocol: "http",
			Host:     "proxy.example.com",
			Port:     8080,
		},
	}
	cache := &proxyLatencyCacheStub{}
	svc := &adminServiceImpl{
		proxyRepo:         repo,
		proxyLatencyCache: cache,
	}

	result, err := svc.TestProxy(context.Background(), 42)
	require.NoError(t, err)
	require.Equal(t, int64(42), repo.lastID)
	require.False(t, result.Success)
	require.Equal(t, "代理探测服务未配置", result.Message)

	stored := cache.entries[42]
	require.NotNil(t, stored)
	require.False(t, stored.Success)
	require.Equal(t, "代理探测服务未配置", stored.Message)
	require.False(t, stored.UpdatedAt.IsZero())
}

func TestAdminServiceTestProxy_SuccessWritesLatencySnapshot(t *testing.T) {
	t.Parallel()

	repo := &proxyRepoGetByIDStub{
		proxy: &Proxy{
			ID:       7,
			Protocol: "http",
			Host:     "proxy.example.com",
			Port:     8080,
		},
	}
	cache := &proxyLatencyCacheStub{}
	prober := &proxyExitInfoProberStub{
		exitInfo: &ProxyExitInfo{
			IP:          "1.2.3.4",
			City:        "San Francisco",
			Region:      "California",
			Country:     "United States",
			CountryCode: "US",
		},
		latency: 187,
	}
	svc := &adminServiceImpl{
		proxyRepo:         repo,
		proxyProber:       prober,
		proxyLatencyCache: cache,
	}

	result, err := svc.TestProxy(context.Background(), 7)
	require.NoError(t, err)
	require.True(t, result.Success)
	require.Equal(t, "Proxy is accessible", result.Message)
	require.Equal(t, int64(187), result.LatencyMs)
	require.Equal(t, "1.2.3.4", result.IPAddress)
	require.Equal(t, "San Francisco", result.City)
	require.Equal(t, "California", result.Region)
	require.Equal(t, "United States", result.Country)
	require.Equal(t, "US", result.CountryCode)
	require.Equal(t, "http://proxy.example.com:8080", prober.lastURL)

	stored := cache.entries[7]
	require.NotNil(t, stored)
	require.True(t, stored.Success)
	require.NotNil(t, stored.LatencyMs)
	require.Equal(t, int64(187), *stored.LatencyMs)
	require.Equal(t, "Proxy is accessible", stored.Message)
	require.Equal(t, "1.2.3.4", stored.IPAddress)
	require.Equal(t, "United States", stored.Country)
	require.Equal(t, "US", stored.CountryCode)
	require.False(t, stored.UpdatedAt.IsZero())
}
