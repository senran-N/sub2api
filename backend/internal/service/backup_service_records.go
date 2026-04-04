package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	infraerrors "github.com/senran-N/sub2api/internal/pkg/errors"
	"github.com/senran-N/sub2api/internal/pkg/logger"
)

func (s *BackupService) ListBackups(ctx context.Context) ([]BackupRecord, error) {
	records, err := s.loadRecords(ctx)
	if err != nil {
		return nil, err
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].StartedAt > records[j].StartedAt
	})
	return records, nil
}

func (s *BackupService) GetBackupRecord(ctx context.Context, backupID string) (*BackupRecord, error) {
	records, err := s.loadRecords(ctx)
	if err != nil {
		return nil, err
	}
	for i := range records {
		if records[i].ID == backupID {
			return &records[i], nil
		}
	}
	return nil, ErrBackupNotFound
}

func (s *BackupService) DeleteBackup(ctx context.Context, backupID string) error {
	s.recordsMu.Lock()
	defer s.recordsMu.Unlock()

	records, err := s.loadRecordsLocked(ctx)
	if err != nil {
		return err
	}

	var found *BackupRecord
	var remaining []BackupRecord
	for i := range records {
		if records[i].ID == backupID {
			found = &records[i]
		} else {
			remaining = append(remaining, records[i])
		}
	}
	if found == nil {
		return ErrBackupNotFound
	}

	if found.S3Key != "" && found.Status == "completed" {
		s3Cfg, err := s.loadS3Config(ctx)
		if err == nil && s3Cfg != nil && s3Cfg.IsConfigured() {
			objectStore, err := s.getOrCreateStore(ctx, s3Cfg)
			if err == nil {
				_ = objectStore.Delete(ctx, found.S3Key)
			}
		}
	}

	return s.saveRecordsLocked(ctx, remaining)
}

// GetBackupDownloadURL 获取备份文件预签名下载 URL
func (s *BackupService) GetBackupDownloadURL(ctx context.Context, backupID string) (string, error) {
	record, err := s.GetBackupRecord(ctx, backupID)
	if err != nil {
		return "", err
	}
	if record.Status != "completed" {
		return "", infraerrors.BadRequest("BACKUP_NOT_COMPLETED", "backup is not completed")
	}

	s3Cfg, err := s.loadS3Config(ctx)
	if err != nil {
		return "", err
	}
	objectStore, err := s.getOrCreateStore(ctx, s3Cfg)
	if err != nil {
		return "", err
	}

	url, err := objectStore.PresignURL(ctx, record.S3Key, time.Hour)
	if err != nil {
		return "", fmt.Errorf("presign url: %w", err)
	}
	return url, nil
}

// loadRecords 加载备份记录，区分"无数据"和"数据损坏"
func (s *BackupService) loadRecords(ctx context.Context) ([]BackupRecord, error) {
	s.recordsMu.Lock()
	defer s.recordsMu.Unlock()
	return s.loadRecordsLocked(ctx)
}

// loadRecordsLocked 在已持有 recordsMu 锁的情况下加载记录
func (s *BackupService) loadRecordsLocked(ctx context.Context) ([]BackupRecord, error) {
	raw, err := s.settingRepo.GetValue(ctx, settingKeyBackupRecords)
	if err != nil || raw == "" {
		return nil, nil //nolint:nilnil
	}
	var records []BackupRecord
	if err := json.Unmarshal([]byte(raw), &records); err != nil {
		return nil, ErrBackupRecordsCorrupt
	}
	return records, nil
}

// saveRecordsLocked 在已持有 recordsMu 锁的情况下保存记录
func (s *BackupService) saveRecordsLocked(ctx context.Context, records []BackupRecord) error {
	data, err := json.Marshal(records)
	if err != nil {
		return err
	}
	return s.settingRepo.Set(ctx, settingKeyBackupRecords, string(data))
}

// saveRecord 保存单条记录（带互斥锁保护）
func (s *BackupService) saveRecord(ctx context.Context, record *BackupRecord) error {
	s.recordsMu.Lock()
	defer s.recordsMu.Unlock()

	records, _ := s.loadRecordsLocked(ctx)
	found := false
	for i := range records {
		if records[i].ID == record.ID {
			records[i] = *record
			found = true
			break
		}
	}
	if !found {
		records = append(records, *record)
	}

	if len(records) > maxBackupRecords {
		records = records[len(records)-maxBackupRecords:]
	}
	return s.saveRecordsLocked(ctx, records)
}

func (s *BackupService) cleanupOldBackups(ctx context.Context, schedule *BackupScheduleConfig) error {
	if schedule == nil {
		return nil
	}

	s.recordsMu.Lock()
	defer s.recordsMu.Unlock()

	records, err := s.loadRecordsLocked(ctx)
	if err != nil {
		return err
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].StartedAt > records[j].StartedAt
	})

	var toDelete []BackupRecord
	var toKeep []BackupRecord
	for i, record := range records {
		shouldDelete := false
		if schedule.RetainCount > 0 && i >= schedule.RetainCount {
			shouldDelete = true
		}
		if schedule.RetainDays > 0 && record.StartedAt != "" {
			startedAt, err := time.Parse(time.RFC3339, record.StartedAt)
			if err == nil && time.Since(startedAt) > time.Duration(schedule.RetainDays)*24*time.Hour {
				shouldDelete = true
			}
		}

		if shouldDelete && record.Status == "completed" {
			toDelete = append(toDelete, record)
		} else {
			toKeep = append(toKeep, record)
		}
	}

	for _, record := range toDelete {
		if record.S3Key != "" {
			_ = s.deleteS3Object(ctx, record.S3Key)
		}
	}

	if len(toDelete) > 0 {
		logger.LegacyPrintf("service.backup", "[Backup] 自动清理了 %d 个过期备份", len(toDelete))
		return s.saveRecordsLocked(ctx, toKeep)
	}
	return nil
}
