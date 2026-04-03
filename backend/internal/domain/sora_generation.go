package domain

import (
	"errors"
	"time"
)

// SoraGeneration represents a persisted Sora client generation record.
type SoraGeneration struct {
	ID             int64      `json:"id"`
	UserID         int64      `json:"user_id"`
	APIKeyID       *int64     `json:"api_key_id,omitempty"`
	Model          string     `json:"model"`
	Prompt         string     `json:"prompt"`
	MediaType      string     `json:"media_type"`
	Status         string     `json:"status"`
	MediaURL       string     `json:"media_url"`
	MediaURLs      []string   `json:"media_urls"`
	FileSizeBytes  int64      `json:"file_size_bytes"`
	StorageType    string     `json:"storage_type"`
	S3ObjectKeys   []string   `json:"s3_object_keys"`
	UpstreamTaskID string     `json:"upstream_task_id"`
	ErrorMessage   string     `json:"error_message"`
	CreatedAt      time.Time  `json:"created_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
}

const (
	SoraGenStatusPending    = "pending"
	SoraGenStatusGenerating = "generating"
	SoraGenStatusCompleted  = "completed"
	SoraGenStatusFailed     = "failed"
	SoraGenStatusCancelled  = "cancelled"
)

const (
	SoraStorageTypeS3       = "s3"
	SoraStorageTypeLocal    = "local"
	SoraStorageTypeUpstream = "upstream"
	SoraStorageTypeNone     = "none"
)

// SoraGenerationListParams defines list filters for generation records.
type SoraGenerationListParams struct {
	UserID      int64
	Status      string
	StorageType string
	MediaType   string
	Page        int
	PageSize    int
}

var (
	ErrSoraGenerationConcurrencyLimit = errors.New("sora generation concurrent limit exceeded")
	ErrSoraGenerationStateConflict    = errors.New("sora generation state conflict")
	ErrSoraGenerationNotActive        = errors.New("sora generation is not active")
)
