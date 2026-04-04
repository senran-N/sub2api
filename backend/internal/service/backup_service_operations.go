package service

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"

	infraerrors "github.com/senran-N/sub2api/internal/pkg/errors"
	"github.com/senran-N/sub2api/internal/pkg/logger"
)

// CreateBackup 创建全量数据库备份并上传到 S3（流式处理）
func (s *BackupService) CreateBackup(ctx context.Context, triggeredBy string, expireDays int) (*BackupRecord, error) {
	if s.shuttingDown.Load() {
		return nil, infraerrors.ServiceUnavailable("SERVER_SHUTTING_DOWN", "server is shutting down")
	}

	s.opMu.Lock()
	if s.backingUp {
		s.opMu.Unlock()
		return nil, ErrBackupInProgress
	}
	s.backingUp = true
	s.opMu.Unlock()
	defer s.finishBackupOperation()

	record, objectStore, err := s.prepareBackupOperation(ctx, triggeredBy, expireDays)
	if err != nil {
		return nil, err
	}
	if err := s.runBackupPipeline(ctx, record, objectStore); err != nil {
		return record, err
	}
	if err := s.saveRecord(ctx, record); err != nil {
		logger.LegacyPrintf("service.backup", "[Backup] 保存备份记录失败: %v", err)
	}
	return record, nil
}

// StartBackup 异步创建备份，立即返回 running 状态的记录
func (s *BackupService) StartBackup(ctx context.Context, triggeredBy string, expireDays int) (*BackupRecord, error) {
	if s.shuttingDown.Load() {
		return nil, infraerrors.ServiceUnavailable("SERVER_SHUTTING_DOWN", "server is shutting down")
	}

	s.opMu.Lock()
	if s.backingUp {
		s.opMu.Unlock()
		return nil, ErrBackupInProgress
	}
	s.backingUp = true
	s.opMu.Unlock()

	launched := false
	defer func() {
		if !launched {
			s.finishBackupOperation()
		}
	}()

	record, objectStore, err := s.prepareBackupOperation(ctx, triggeredBy, expireDays)
	if err != nil {
		return nil, err
	}
	if err := s.saveRecord(ctx, record); err != nil {
		return nil, fmt.Errorf("save initial record: %w", err)
	}

	launched = true
	result := *record

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer s.finishBackupOperation()
		defer s.recoverBackupPanic(record)
		s.executeBackup(record, objectStore)
	}()

	return &result, nil
}

// executeBackup 后台执行备份（独立于 HTTP context）
func (s *BackupService) executeBackup(record *BackupRecord, objectStore BackupObjectStore) {
	ctx, cancel := context.WithTimeout(s.bgCtx, 30*time.Minute)
	defer cancel()

	if err := s.runBackupPipeline(ctx, record, objectStore); err != nil {
		return
	}
	if err := s.saveRecord(context.Background(), record); err != nil {
		logger.LegacyPrintf("service.backup", "[Backup] 保存备份记录失败: %v", err)
	}
}

// RestoreBackup 从 S3 下载备份并流式恢复到数据库
func (s *BackupService) RestoreBackup(ctx context.Context, backupID string) error {
	s.opMu.Lock()
	if s.restoring {
		s.opMu.Unlock()
		return ErrRestoreInProgress
	}
	s.restoring = true
	s.opMu.Unlock()
	defer s.finishRestoreOperation()

	record, objectStore, err := s.prepareRestoreOperation(ctx, backupID)
	if err != nil {
		return err
	}
	return s.restoreBackupData(ctx, record, objectStore)
}

// StartRestore 异步恢复备份，立即返回
func (s *BackupService) StartRestore(ctx context.Context, backupID string) (*BackupRecord, error) {
	if s.shuttingDown.Load() {
		return nil, infraerrors.ServiceUnavailable("SERVER_SHUTTING_DOWN", "server is shutting down")
	}

	s.opMu.Lock()
	if s.restoring {
		s.opMu.Unlock()
		return nil, ErrRestoreInProgress
	}
	s.restoring = true
	s.opMu.Unlock()

	launched := false
	defer func() {
		if !launched {
			s.finishRestoreOperation()
		}
	}()

	record, objectStore, err := s.prepareRestoreOperation(ctx, backupID)
	if err != nil {
		return nil, err
	}
	record.RestoreStatus = "running"
	_ = s.saveRecord(ctx, record)

	launched = true
	result := *record

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer s.finishRestoreOperation()
		defer s.recoverRestorePanic(record)
		s.executeRestore(record, objectStore)
	}()

	return &result, nil
}

// executeRestore 后台执行恢复
func (s *BackupService) executeRestore(record *BackupRecord, objectStore BackupObjectStore) {
	ctx, cancel := context.WithTimeout(s.bgCtx, 30*time.Minute)
	defer cancel()

	if err := s.restoreBackupData(ctx, record, objectStore); err != nil {
		_ = s.saveRecord(context.Background(), record)
		return
	}
	if err := s.saveRecord(context.Background(), record); err != nil {
		logger.LegacyPrintf("service.backup", "[Backup] 保存恢复记录失败: %v", err)
	}
}

func (s *BackupService) prepareBackupOperation(ctx context.Context, triggeredBy string, expireDays int) (*BackupRecord, BackupObjectStore, error) {
	s3Cfg, err := s.loadS3Config(ctx)
	if err != nil {
		return nil, nil, err
	}
	if s3Cfg == nil || !s3Cfg.IsConfigured() {
		return nil, nil, ErrBackupS3NotConfigured
	}

	objectStore, err := s.getOrCreateStore(ctx, s3Cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("init object store: %w", err)
	}

	now := time.Now()
	backupID := uuid.New().String()[:8]
	fileName := fmt.Sprintf("%s_%s.sql.gz", s.dbCfg.DBName, now.Format("20060102_150405"))
	s3Key := s.buildS3Key(s3Cfg, fileName)

	var expiresAt string
	if expireDays > 0 {
		expiresAt = now.AddDate(0, 0, expireDays).Format(time.RFC3339)
	}

	return &BackupRecord{
		ID:          backupID,
		Status:      "running",
		BackupType:  "postgres",
		FileName:    fileName,
		S3Key:       s3Key,
		TriggeredBy: triggeredBy,
		StartedAt:   now.Format(time.RFC3339),
		ExpiresAt:   expiresAt,
		Progress:    "pending",
	}, objectStore, nil
}

func (s *BackupService) runBackupPipeline(ctx context.Context, record *BackupRecord, objectStore BackupObjectStore) error {
	record.Progress = "dumping"
	_ = s.saveRecord(ctx, record)

	dumpReader, err := s.dumper.Dump(ctx)
	if err != nil {
		record.Status = "failed"
		record.ErrorMsg = fmt.Sprintf("pg_dump failed: %v", err)
		record.Progress = ""
		record.FinishedAt = time.Now().Format(time.RFC3339)
		_ = s.saveRecord(context.Background(), record)
		return fmt.Errorf("pg_dump: %w", err)
	}

	record.Progress = "uploading"
	_ = s.saveRecord(ctx, record)

	sizeBytes, uploadErr := streamGzipUpload(dumpReader, func(uploadBody io.Reader) (int64, error) {
		return objectStore.Upload(ctx, record.S3Key, uploadBody, "application/gzip")
	})
	if uploadErr != nil {
		record.Status = "failed"
		record.ErrorMsg = uploadErr.Error()
		record.Progress = ""
		record.FinishedAt = time.Now().Format(time.RFC3339)
		_ = s.saveRecord(context.Background(), record)
		return uploadErr
	}

	record.SizeBytes = sizeBytes
	record.Status = "completed"
	record.Progress = ""
	record.FinishedAt = time.Now().Format(time.RFC3339)
	return nil
}

func streamGzipUpload(dumpReader io.ReadCloser, upload func(io.Reader) (int64, error)) (int64, error) {
	pr, pw := io.Pipe()
	gzipDone := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				pw.CloseWithError(fmt.Errorf("gzip goroutine panic: %v", r)) //nolint:errcheck
				gzipDone <- fmt.Errorf("gzip goroutine panic: %v", r)
			}
		}()
		gzWriter := gzip.NewWriter(pw)
		var gzErr error
		_, gzErr = io.Copy(gzWriter, dumpReader)
		if closeErr := gzWriter.Close(); closeErr != nil && gzErr == nil {
			gzErr = closeErr
		}
		if closeErr := dumpReader.Close(); closeErr != nil && gzErr == nil {
			gzErr = closeErr
		}
		if gzErr != nil {
			_ = pw.CloseWithError(gzErr)
		} else {
			_ = pw.Close()
		}
		gzipDone <- gzErr
	}()

	sizeBytes, err := upload(pr)
	if err != nil {
		_ = pr.CloseWithError(err)
		gzErr := <-gzipDone
		if gzErr != nil {
			return 0, fmt.Errorf("gzip/dump failed: %v", gzErr)
		}
		return 0, fmt.Errorf("S3 upload failed: %v", err)
	}
	if gzErr := <-gzipDone; gzErr != nil {
		return 0, fmt.Errorf("gzip/dump failed: %v", gzErr)
	}
	return sizeBytes, nil
}

func (s *BackupService) prepareRestoreOperation(ctx context.Context, backupID string) (*BackupRecord, BackupObjectStore, error) {
	record, err := s.GetBackupRecord(ctx, backupID)
	if err != nil {
		return nil, nil, err
	}
	if record.Status != "completed" {
		return nil, nil, infraerrors.BadRequest("BACKUP_NOT_COMPLETED", "can only restore from a completed backup")
	}

	s3Cfg, err := s.loadS3Config(ctx)
	if err != nil {
		return nil, nil, err
	}
	objectStore, err := s.getOrCreateStore(ctx, s3Cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("init object store: %w", err)
	}
	return record, objectStore, nil
}

func (s *BackupService) restoreBackupData(ctx context.Context, record *BackupRecord, objectStore BackupObjectStore) error {
	body, err := objectStore.Download(ctx, record.S3Key)
	if err != nil {
		record.RestoreStatus = "failed"
		record.RestoreError = fmt.Sprintf("S3 download failed: %v", err)
		return fmt.Errorf("S3 download failed: %w", err)
	}
	defer func() { _ = body.Close() }()

	gzReader, err := gzip.NewReader(body)
	if err != nil {
		record.RestoreStatus = "failed"
		record.RestoreError = fmt.Sprintf("gzip reader: %v", err)
		return fmt.Errorf("gzip reader: %w", err)
	}
	defer func() { _ = gzReader.Close() }()

	if err := s.dumper.Restore(ctx, gzReader); err != nil {
		record.RestoreStatus = "failed"
		record.RestoreError = fmt.Sprintf("pg restore: %v", err)
		return fmt.Errorf("pg restore: %w", err)
	}

	record.RestoreStatus = "completed"
	record.RestoredAt = time.Now().Format(time.RFC3339)
	return nil
}

func (s *BackupService) finishBackupOperation() {
	s.opMu.Lock()
	s.backingUp = false
	s.opMu.Unlock()
}

func (s *BackupService) finishRestoreOperation() {
	s.opMu.Lock()
	s.restoring = false
	s.opMu.Unlock()
}

func (s *BackupService) recoverBackupPanic(record *BackupRecord) {
	if r := recover(); r != nil {
		logger.LegacyPrintf("service.backup", "[Backup] panic recovered: %v", r)
		record.Status = "failed"
		record.ErrorMsg = fmt.Sprintf("internal panic: %v", r)
		record.Progress = ""
		record.FinishedAt = time.Now().Format(time.RFC3339)
		_ = s.saveRecord(context.Background(), record)
	}
}

func (s *BackupService) recoverRestorePanic(record *BackupRecord) {
	if r := recover(); r != nil {
		logger.LegacyPrintf("service.backup", "[Backup] restore panic recovered: %v", r)
		record.RestoreStatus = "failed"
		record.RestoreError = fmt.Sprintf("internal panic: %v", r)
		_ = s.saveRecord(context.Background(), record)
	}
}
