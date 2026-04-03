package service

import (
	"github.com/senran-N/sub2api/internal/domain"
	"github.com/senran-N/sub2api/internal/ports"
)

type SoraGeneration = domain.SoraGeneration
type SoraGenerationListParams = domain.SoraGenerationListParams
type SoraGenerationRepository = ports.SoraGenerationRepository

const (
	SoraGenStatusPending    = domain.SoraGenStatusPending
	SoraGenStatusGenerating = domain.SoraGenStatusGenerating
	SoraGenStatusCompleted  = domain.SoraGenStatusCompleted
	SoraGenStatusFailed     = domain.SoraGenStatusFailed
	SoraGenStatusCancelled  = domain.SoraGenStatusCancelled
)

const (
	SoraStorageTypeS3       = domain.SoraStorageTypeS3
	SoraStorageTypeLocal    = domain.SoraStorageTypeLocal
	SoraStorageTypeUpstream = domain.SoraStorageTypeUpstream
	SoraStorageTypeNone     = domain.SoraStorageTypeNone
)

var (
	ErrSoraGenerationConcurrencyLimit = domain.ErrSoraGenerationConcurrencyLimit
	ErrSoraGenerationStateConflict    = domain.ErrSoraGenerationStateConflict
	ErrSoraGenerationNotActive        = domain.ErrSoraGenerationNotActive
)
