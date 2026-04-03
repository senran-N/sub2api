package service

import (
	"context"

	"github.com/senran-N/sub2api/internal/domain"
)

type DataManagementPostgresConfig = domain.DataManagementPostgresConfig
type DataManagementRedisConfig = domain.DataManagementRedisConfig
type DataManagementS3Config = domain.DataManagementS3Config
type DataManagementConfig = domain.DataManagementConfig
type DataManagementTestS3Result = domain.DataManagementTestS3Result
type DataManagementCreateBackupJobInput = domain.DataManagementCreateBackupJobInput
type DataManagementListBackupJobsInput = domain.DataManagementListBackupJobsInput
type DataManagementArtifactInfo = domain.DataManagementArtifactInfo
type DataManagementS3ObjectInfo = domain.DataManagementS3ObjectInfo
type DataManagementBackupJob = domain.DataManagementBackupJob
type DataManagementSourceProfile = domain.DataManagementSourceProfile
type DataManagementSourceConfig = domain.DataManagementSourceConfig
type DataManagementCreateSourceProfileInput = domain.DataManagementCreateSourceProfileInput
type DataManagementUpdateSourceProfileInput = domain.DataManagementUpdateSourceProfileInput
type DataManagementS3Profile = domain.DataManagementS3Profile
type DataManagementCreateS3ProfileInput = domain.DataManagementCreateS3ProfileInput
type DataManagementUpdateS3ProfileInput = domain.DataManagementUpdateS3ProfileInput
type DataManagementListBackupJobsResult = domain.DataManagementListBackupJobsResult

func (s *DataManagementService) GetConfig(ctx context.Context) (DataManagementConfig, error) {
	_ = ctx
	return DataManagementConfig{}, s.deprecatedError()
}

func (s *DataManagementService) UpdateConfig(ctx context.Context, cfg DataManagementConfig) (DataManagementConfig, error) {
	_, _ = ctx, cfg
	return DataManagementConfig{}, s.deprecatedError()
}

func (s *DataManagementService) ListSourceProfiles(ctx context.Context, sourceType string) ([]DataManagementSourceProfile, error) {
	_, _ = ctx, sourceType
	return nil, s.deprecatedError()
}

func (s *DataManagementService) CreateSourceProfile(ctx context.Context, input DataManagementCreateSourceProfileInput) (DataManagementSourceProfile, error) {
	_, _ = ctx, input
	return DataManagementSourceProfile{}, s.deprecatedError()
}

func (s *DataManagementService) UpdateSourceProfile(ctx context.Context, input DataManagementUpdateSourceProfileInput) (DataManagementSourceProfile, error) {
	_, _ = ctx, input
	return DataManagementSourceProfile{}, s.deprecatedError()
}

func (s *DataManagementService) DeleteSourceProfile(ctx context.Context, sourceType, profileID string) error {
	_, _, _ = ctx, sourceType, profileID
	return s.deprecatedError()
}

func (s *DataManagementService) SetActiveSourceProfile(ctx context.Context, sourceType, profileID string) (DataManagementSourceProfile, error) {
	_, _, _ = ctx, sourceType, profileID
	return DataManagementSourceProfile{}, s.deprecatedError()
}

func (s *DataManagementService) ValidateS3(ctx context.Context, cfg DataManagementS3Config) (DataManagementTestS3Result, error) {
	_, _ = ctx, cfg
	return DataManagementTestS3Result{}, s.deprecatedError()
}

func (s *DataManagementService) ListS3Profiles(ctx context.Context) ([]DataManagementS3Profile, error) {
	_ = ctx
	return nil, s.deprecatedError()
}

func (s *DataManagementService) CreateS3Profile(ctx context.Context, input DataManagementCreateS3ProfileInput) (DataManagementS3Profile, error) {
	_, _ = ctx, input
	return DataManagementS3Profile{}, s.deprecatedError()
}

func (s *DataManagementService) UpdateS3Profile(ctx context.Context, input DataManagementUpdateS3ProfileInput) (DataManagementS3Profile, error) {
	_, _ = ctx, input
	return DataManagementS3Profile{}, s.deprecatedError()
}

func (s *DataManagementService) DeleteS3Profile(ctx context.Context, profileID string) error {
	_, _ = ctx, profileID
	return s.deprecatedError()
}

func (s *DataManagementService) SetActiveS3Profile(ctx context.Context, profileID string) (DataManagementS3Profile, error) {
	_, _ = ctx, profileID
	return DataManagementS3Profile{}, s.deprecatedError()
}

func (s *DataManagementService) CreateBackupJob(ctx context.Context, input DataManagementCreateBackupJobInput) (DataManagementBackupJob, error) {
	_, _ = ctx, input
	return DataManagementBackupJob{}, s.deprecatedError()
}

func (s *DataManagementService) ListBackupJobs(ctx context.Context, input DataManagementListBackupJobsInput) (DataManagementListBackupJobsResult, error) {
	_, _ = ctx, input
	return DataManagementListBackupJobsResult{}, s.deprecatedError()
}

func (s *DataManagementService) GetBackupJob(ctx context.Context, jobID string) (DataManagementBackupJob, error) {
	_, _ = ctx, jobID
	return DataManagementBackupJob{}, s.deprecatedError()
}

func (s *DataManagementService) deprecatedError() error {
	return ErrDataManagementDeprecated.WithMetadata(map[string]string{"socket_path": s.SocketPath()})
}
