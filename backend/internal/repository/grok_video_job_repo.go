package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/senran-N/sub2api/internal/service"
)

type grokVideoJobRepository struct {
	db *sql.DB
}

func NewGrokVideoJobRepository(db *sql.DB) service.GrokVideoJobRepository {
	return &grokVideoJobRepository{db: db}
}

func (r *grokVideoJobRepository) Upsert(ctx context.Context, record service.GrokVideoJobRecord) error {
	if r == nil || r.db == nil {
		return nil
	}
	if strings.TrimSpace(record.JobID) == "" || record.AccountID <= 0 {
		return nil
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO grok_video_jobs (
			job_id,
			account_id,
			group_id,
			requested_model,
			canonical_model,
			output_asset_id,
			request_payload_snapshot,
			upstream_status,
			normalized_status,
			poll_after,
			error_code,
			error_message
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8, $9, $10, $11, $12)
		ON CONFLICT (job_id) DO UPDATE SET
			account_id = EXCLUDED.account_id,
			group_id = EXCLUDED.group_id,
			requested_model = EXCLUDED.requested_model,
			canonical_model = EXCLUDED.canonical_model,
			output_asset_id = COALESCE(NULLIF(EXCLUDED.output_asset_id, ''), grok_video_jobs.output_asset_id),
			request_payload_snapshot = COALESCE(EXCLUDED.request_payload_snapshot, grok_video_jobs.request_payload_snapshot),
			upstream_status = EXCLUDED.upstream_status,
			normalized_status = EXCLUDED.normalized_status,
			poll_after = EXCLUDED.poll_after,
			error_code = EXCLUDED.error_code,
			error_message = EXCLUDED.error_message,
			updated_at = NOW()
	`,
		record.JobID,
		record.AccountID,
		record.GroupID,
		strings.TrimSpace(record.RequestedModel),
		strings.TrimSpace(record.CanonicalModel),
		strings.TrimSpace(record.OutputAssetID),
		nilIfEmptyBytes(record.RequestPayloadSnapshot),
		strings.TrimSpace(record.UpstreamStatus),
		strings.TrimSpace(record.NormalizedStatus),
		record.PollAfter,
		strings.TrimSpace(record.ErrorCode),
		strings.TrimSpace(record.ErrorMessage),
	)
	return err
}

func (r *grokVideoJobRepository) GetByJobID(ctx context.Context, jobID string) (*service.GrokVideoJobRecord, error) {
	if r == nil || r.db == nil {
		return nil, service.ErrGrokVideoJobNotFound
	}

	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return nil, service.ErrGrokVideoJobNotFound
	}

	row := r.db.QueryRowContext(ctx, `
		SELECT
			job_id,
			account_id,
			group_id,
			requested_model,
			canonical_model,
			COALESCE(output_asset_id, ''),
			COALESCE(request_payload_snapshot::text, ''),
			upstream_status,
			normalized_status,
			poll_after,
			error_code,
			error_message,
			created_at,
			updated_at
		FROM grok_video_jobs
		WHERE job_id = $1
	`, jobID)

	var (
		record                 service.GrokVideoJobRecord
		groupID                sql.NullInt64
		requestPayloadSnapshot string
		pollAfter              sql.NullTime
	)
	if err := row.Scan(
		&record.JobID,
		&record.AccountID,
		&groupID,
		&record.RequestedModel,
		&record.CanonicalModel,
		&record.OutputAssetID,
		&requestPayloadSnapshot,
		&record.UpstreamStatus,
		&record.NormalizedStatus,
		&pollAfter,
		&record.ErrorCode,
		&record.ErrorMessage,
		&record.CreatedAt,
		&record.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrGrokVideoJobNotFound
		}
		return nil, err
	}
	if groupID.Valid {
		record.GroupID = &groupID.Int64
	}
	if requestPayloadSnapshot != "" {
		record.RequestPayloadSnapshot = []byte(requestPayloadSnapshot)
	}
	if pollAfter.Valid {
		ts := pollAfter.Time
		record.PollAfter = &ts
	}
	return &record, nil
}

func (r *grokVideoJobRepository) UpdateStatus(ctx context.Context, patch service.GrokVideoJobStatusPatch) error {
	if r == nil || r.db == nil {
		return nil
	}

	jobID := strings.TrimSpace(patch.JobID)
	if jobID == "" {
		return nil
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE grok_video_jobs
		SET
			upstream_status = $2,
			normalized_status = $3,
			poll_after = $4,
			error_code = $5,
			error_message = $6,
			output_asset_id = CASE WHEN $7 = '' THEN output_asset_id ELSE $7 END,
			updated_at = NOW()
		WHERE job_id = $1
	`,
		jobID,
		strings.TrimSpace(patch.UpstreamStatus),
		strings.TrimSpace(patch.NormalizedStatus),
		patch.PollAfter,
		strings.TrimSpace(patch.ErrorCode),
		strings.TrimSpace(patch.ErrorMessage),
		strings.TrimSpace(patch.OutputAssetID),
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return service.ErrGrokVideoJobNotFound
	}
	return nil
}

func nilIfEmptyBytes(raw []byte) any {
	if len(raw) == 0 {
		return nil
	}
	return raw
}
