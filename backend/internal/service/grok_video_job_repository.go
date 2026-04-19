package service

import (
	"context"
	"errors"
	"time"
)

var ErrGrokVideoJobNotFound = errors.New("grok video job not found")

type GrokVideoJobRecord struct {
	JobID                  string
	AccountID              int64
	GroupID                *int64
	RequestedModel         string
	CanonicalModel         string
	OutputAssetID          string
	RequestPayloadSnapshot []byte
	UpstreamStatus         string
	NormalizedStatus       string
	PollAfter              *time.Time
	ErrorCode              string
	ErrorMessage           string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type GrokVideoJobStatusPatch struct {
	JobID            string
	UpstreamStatus   string
	NormalizedStatus string
	PollAfter        *time.Time
	ErrorCode        string
	ErrorMessage     string
	OutputAssetID    string
}

type GrokVideoJobRepository interface {
	Upsert(context.Context, GrokVideoJobRecord) error
	GetByJobID(context.Context, string) (*GrokVideoJobRecord, error)
	UpdateStatus(context.Context, GrokVideoJobStatusPatch) error
}
