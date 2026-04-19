package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/senran-N/sub2api/internal/service"
)

type grokMediaAssetRepository struct {
	db *sql.DB
}

func NewGrokMediaAssetRepository(db *sql.DB) service.GrokMediaAssetRepository {
	return &grokMediaAssetRepository{db: db}
}

func (r *grokMediaAssetRepository) Upsert(ctx context.Context, record service.GrokMediaAssetRecord) error {
	if r == nil || r.db == nil {
		return nil
	}
	if strings.TrimSpace(record.AssetID) == "" || record.AccountID <= 0 || strings.TrimSpace(record.UpstreamURL) == "" {
		return nil
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO grok_media_assets (
			asset_id,
			account_id,
			job_id,
			requested_model,
			canonical_model,
			asset_type,
			upstream_url,
			local_path,
			content_hash,
			mime_type,
			size_bytes,
			status,
			expires_at,
			last_access_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (asset_id) DO UPDATE SET
			account_id = EXCLUDED.account_id,
			job_id = EXCLUDED.job_id,
			requested_model = EXCLUDED.requested_model,
			canonical_model = EXCLUDED.canonical_model,
			asset_type = EXCLUDED.asset_type,
			upstream_url = EXCLUDED.upstream_url,
			local_path = COALESCE(NULLIF(EXCLUDED.local_path, ''), grok_media_assets.local_path),
			content_hash = COALESCE(NULLIF(EXCLUDED.content_hash, ''), grok_media_assets.content_hash),
			mime_type = COALESCE(NULLIF(EXCLUDED.mime_type, ''), grok_media_assets.mime_type),
			size_bytes = CASE WHEN EXCLUDED.size_bytes > 0 THEN EXCLUDED.size_bytes ELSE grok_media_assets.size_bytes END,
			status = COALESCE(NULLIF(EXCLUDED.status, ''), grok_media_assets.status),
			expires_at = COALESCE(EXCLUDED.expires_at, grok_media_assets.expires_at),
			last_access_at = COALESCE(EXCLUDED.last_access_at, grok_media_assets.last_access_at),
			updated_at = NOW()
	`,
		record.AssetID,
		record.AccountID,
		nullableTrimmedString(record.JobID),
		strings.TrimSpace(record.RequestedModel),
		strings.TrimSpace(record.CanonicalModel),
		strings.TrimSpace(record.AssetType),
		strings.TrimSpace(record.UpstreamURL),
		strings.TrimSpace(record.LocalPath),
		strings.TrimSpace(record.ContentHash),
		strings.TrimSpace(record.MimeType),
		record.SizeBytes,
		strings.TrimSpace(record.Status),
		record.ExpiresAt,
		record.LastAccessAt,
	)
	return err
}

func (r *grokMediaAssetRepository) GetByAssetID(ctx context.Context, assetID string) (*service.GrokMediaAssetRecord, error) {
	if r == nil || r.db == nil {
		return nil, service.ErrGrokMediaAssetNotFound
	}
	assetID = strings.TrimSpace(assetID)
	if assetID == "" {
		return nil, service.ErrGrokMediaAssetNotFound
	}

	row := r.db.QueryRowContext(ctx, `
		SELECT
			asset_id,
			account_id,
			COALESCE(job_id, ''),
			requested_model,
			canonical_model,
			asset_type,
			upstream_url,
			COALESCE(local_path, ''),
			COALESCE(content_hash, ''),
			COALESCE(mime_type, ''),
			size_bytes,
			status,
			expires_at,
			last_access_at,
			created_at,
			updated_at
		FROM grok_media_assets
		WHERE asset_id = $1
	`, assetID)

	var (
		record       service.GrokMediaAssetRecord
		expiresAt    sql.NullTime
		lastAccessAt sql.NullTime
	)
	if err := row.Scan(
		&record.AssetID,
		&record.AccountID,
		&record.JobID,
		&record.RequestedModel,
		&record.CanonicalModel,
		&record.AssetType,
		&record.UpstreamURL,
		&record.LocalPath,
		&record.ContentHash,
		&record.MimeType,
		&record.SizeBytes,
		&record.Status,
		&expiresAt,
		&lastAccessAt,
		&record.CreatedAt,
		&record.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrGrokMediaAssetNotFound
		}
		return nil, err
	}
	if expiresAt.Valid {
		ts := expiresAt.Time
		record.ExpiresAt = &ts
	}
	if lastAccessAt.Valid {
		ts := lastAccessAt.Time
		record.LastAccessAt = &ts
	}
	return &record, nil
}

func (r *grokMediaAssetRepository) FindCachedByHash(ctx context.Context, contentHash string) (*service.GrokMediaAssetRecord, error) {
	if r == nil || r.db == nil {
		return nil, nil
	}
	contentHash = strings.TrimSpace(contentHash)
	if contentHash == "" {
		return nil, nil
	}

	row := r.db.QueryRowContext(ctx, `
		SELECT
			asset_id,
			account_id,
			COALESCE(job_id, ''),
			requested_model,
			canonical_model,
			asset_type,
			upstream_url,
			COALESCE(local_path, ''),
			COALESCE(content_hash, ''),
			COALESCE(mime_type, ''),
			size_bytes,
			status,
			expires_at,
			last_access_at,
			created_at,
			updated_at
		FROM grok_media_assets
		WHERE
			content_hash = $1
			AND COALESCE(local_path, '') <> ''
			AND status = 'ready'
			AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY updated_at DESC
		LIMIT 1
	`, contentHash)

	var (
		record       service.GrokMediaAssetRecord
		expiresAt    sql.NullTime
		lastAccessAt sql.NullTime
	)
	if err := row.Scan(
		&record.AssetID,
		&record.AccountID,
		&record.JobID,
		&record.RequestedModel,
		&record.CanonicalModel,
		&record.AssetType,
		&record.UpstreamURL,
		&record.LocalPath,
		&record.ContentHash,
		&record.MimeType,
		&record.SizeBytes,
		&record.Status,
		&expiresAt,
		&lastAccessAt,
		&record.CreatedAt,
		&record.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if expiresAt.Valid {
		ts := expiresAt.Time
		record.ExpiresAt = &ts
	}
	if lastAccessAt.Valid {
		ts := lastAccessAt.Time
		record.LastAccessAt = &ts
	}
	return &record, nil
}

func (r *grokMediaAssetRepository) UpdateCacheState(ctx context.Context, patch service.GrokMediaAssetCachePatch) error {
	if r == nil || r.db == nil {
		return nil
	}
	assetID := strings.TrimSpace(patch.AssetID)
	if assetID == "" {
		return nil
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE grok_media_assets
		SET
			local_path = $2,
			content_hash = $3,
			mime_type = $4,
			size_bytes = $5,
			status = $6,
			expires_at = $7,
			last_access_at = $8,
			updated_at = NOW()
		WHERE asset_id = $1
	`,
		assetID,
		strings.TrimSpace(patch.LocalPath),
		strings.TrimSpace(patch.ContentHash),
		strings.TrimSpace(patch.MimeType),
		patch.SizeBytes,
		strings.TrimSpace(patch.Status),
		patch.ExpiresAt,
		patch.LastAccessAt,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return service.ErrGrokMediaAssetNotFound
	}
	return nil
}

func (r *grokMediaAssetRepository) MarkAccessed(ctx context.Context, assetID string, accessAt time.Time, expiresAt *time.Time) error {
	if r == nil || r.db == nil {
		return nil
	}
	assetID = strings.TrimSpace(assetID)
	if assetID == "" {
		return nil
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE grok_media_assets
		SET
			last_access_at = $2,
			expires_at = COALESCE($3, expires_at),
			updated_at = NOW()
		WHERE asset_id = $1
	`, assetID, accessAt, expiresAt)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return service.ErrGrokMediaAssetNotFound
	}
	return nil
}

func (r *grokMediaAssetRepository) DeleteExpired(ctx context.Context, expiresBefore time.Time, limit int) ([]service.GrokMediaAssetRecord, error) {
	if r == nil || r.db == nil || limit <= 0 {
		return nil, nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	rows, err := tx.QueryContext(ctx, `
		SELECT
			asset_id,
			account_id,
			COALESCE(job_id, ''),
			requested_model,
			canonical_model,
			asset_type,
			upstream_url,
			COALESCE(local_path, ''),
			COALESCE(content_hash, ''),
			COALESCE(mime_type, ''),
			size_bytes,
			status,
			expires_at,
			last_access_at,
			created_at,
			updated_at
		FROM grok_media_assets
		WHERE expires_at IS NOT NULL AND expires_at <= $1
		ORDER BY expires_at ASC
		LIMIT $2
		FOR UPDATE SKIP LOCKED
	`, expiresBefore, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	records := make([]service.GrokMediaAssetRecord, 0, limit)
	for rows.Next() {
		var (
			record       service.GrokMediaAssetRecord
			expiresAt    sql.NullTime
			lastAccessAt sql.NullTime
		)
		if err := rows.Scan(
			&record.AssetID,
			&record.AccountID,
			&record.JobID,
			&record.RequestedModel,
			&record.CanonicalModel,
			&record.AssetType,
			&record.UpstreamURL,
			&record.LocalPath,
			&record.ContentHash,
			&record.MimeType,
			&record.SizeBytes,
			&record.Status,
			&expiresAt,
			&lastAccessAt,
			&record.CreatedAt,
			&record.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if expiresAt.Valid {
			ts := expiresAt.Time
			record.ExpiresAt = &ts
		}
		if lastAccessAt.Valid {
			ts := lastAccessAt.Time
			record.LastAccessAt = &ts
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(records) == 0 {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return nil, nil
	}

	args := make([]any, 0, len(records))
	placeholders := make([]string, 0, len(records))
	for i, record := range records {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
		args = append(args, record.AssetID)
	}
	query := `DELETE FROM grok_media_assets WHERE asset_id IN (` + strings.Join(placeholders, ", ") + `)`
	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return records, nil
}

func (r *grokMediaAssetRepository) CountByLocalPath(ctx context.Context, localPath string) (int, error) {
	if r == nil || r.db == nil {
		return 0, nil
	}
	localPath = strings.TrimSpace(localPath)
	if localPath == "" {
		return 0, nil
	}

	var count int
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM grok_media_assets
		WHERE local_path = $1
	`, localPath).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func nullableTrimmedString(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}
