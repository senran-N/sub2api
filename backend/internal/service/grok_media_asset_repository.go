package service

import (
	"context"
	"errors"
	"time"
)

var ErrGrokMediaAssetNotFound = errors.New("grok media asset not found")

type GrokMediaAssetRecord struct {
	AssetID        string
	AccountID      int64
	JobID          string
	RequestedModel string
	CanonicalModel string
	AssetType      string
	UpstreamURL    string
	LocalPath      string
	ContentHash    string
	MimeType       string
	SizeBytes      int64
	Status         string
	ExpiresAt      *time.Time
	LastAccessAt   *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type GrokMediaAssetCachePatch struct {
	AssetID      string
	LocalPath    string
	ContentHash  string
	MimeType     string
	SizeBytes    int64
	Status       string
	ExpiresAt    *time.Time
	LastAccessAt *time.Time
}

type GrokMediaAssetRepository interface {
	Upsert(context.Context, GrokMediaAssetRecord) error
	GetByAssetID(context.Context, string) (*GrokMediaAssetRecord, error)
	FindCachedByHash(context.Context, string) (*GrokMediaAssetRecord, error)
	UpdateCacheState(context.Context, GrokMediaAssetCachePatch) error
	MarkAccessed(context.Context, string, time.Time, *time.Time) error
	DeleteExpired(context.Context, time.Time, int) ([]GrokMediaAssetRecord, error)
	CountByLocalPath(context.Context, string) (int, error)
}
