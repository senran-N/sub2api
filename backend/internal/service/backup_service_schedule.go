package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"

	infraerrors "github.com/senran-N/sub2api/internal/pkg/errors"
	"github.com/senran-N/sub2api/internal/pkg/logger"
)

// Start 启动定时备份调度器并清理孤立记录
func (s *BackupService) Start() {
	s.cronSched = cron.New()
	s.cronSched.Start()

	s.recoverStaleRecords()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	schedule, err := s.GetSchedule(ctx)
	if err != nil {
		logger.LegacyPrintf("service.backup", "[Backup] 加载定时备份配置失败: %v", err)
		return
	}
	if schedule.Enabled && schedule.CronExpr != "" {
		if err := s.applyCronSchedule(schedule); err != nil {
			logger.LegacyPrintf("service.backup", "[Backup] 应用定时备份配置失败: %v", err)
		}
	}
}

// recoverStaleRecords 启动时将孤立的 running 记录标记为 failed
func (s *BackupService) recoverStaleRecords() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	records, err := s.loadRecords(ctx)
	if err != nil {
		return
	}
	for i := range records {
		if records[i].Status == "running" {
			records[i].Status = "failed"
			records[i].ErrorMsg = "interrupted by server restart"
			records[i].Progress = ""
			records[i].FinishedAt = time.Now().Format(time.RFC3339)
			_ = s.saveRecord(ctx, &records[i])
			logger.LegacyPrintf("service.backup", "[Backup] recovered stale running record: %s", records[i].ID)
		}
		if records[i].RestoreStatus == "running" {
			records[i].RestoreStatus = "failed"
			records[i].RestoreError = "interrupted by server restart"
			_ = s.saveRecord(ctx, &records[i])
			logger.LegacyPrintf("service.backup", "[Backup] recovered stale restoring record: %s", records[i].ID)
		}
	}
}

// Stop 停止定时备份并等待活跃操作完成
func (s *BackupService) Stop() {
	s.shuttingDown.Store(true)

	s.cronMu.Lock()
	if s.cronSched != nil {
		s.cronSched.Stop()
	}
	s.cronMu.Unlock()

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		logger.LegacyPrintf("service.backup", "[Backup] all active operations finished")
	case <-time.After(5 * time.Minute):
		logger.LegacyPrintf("service.backup", "[Backup] shutdown timeout after 5min, cancelling active operations")
		if s.bgCancel != nil {
			s.bgCancel()
		}
		select {
		case <-done:
			logger.LegacyPrintf("service.backup", "[Backup] active operations cancelled and cleaned up")
		case <-time.After(10 * time.Second):
			logger.LegacyPrintf("service.backup", "[Backup] goroutine cleanup timed out")
		}
	}
}

func (s *BackupService) GetSchedule(ctx context.Context) (*BackupScheduleConfig, error) {
	raw, err := s.settingRepo.GetValue(ctx, settingKeyBackupSchedule)
	if err != nil || raw == "" {
		return &BackupScheduleConfig{}, nil
	}
	var cfg BackupScheduleConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return &BackupScheduleConfig{}, nil
	}
	return &cfg, nil
}

func (s *BackupService) UpdateSchedule(ctx context.Context, cfg BackupScheduleConfig) (*BackupScheduleConfig, error) {
	if cfg.Enabled && cfg.CronExpr == "" {
		return nil, infraerrors.BadRequest("INVALID_CRON", "cron expression is required when schedule is enabled")
	}
	if cfg.CronExpr != "" {
		parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		if _, err := parser.Parse(cfg.CronExpr); err != nil {
			return nil, infraerrors.BadRequest("INVALID_CRON", fmt.Sprintf("invalid cron expression: %v", err))
		}
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshal schedule config: %w", err)
	}
	if err := s.settingRepo.Set(ctx, settingKeyBackupSchedule, string(data)); err != nil {
		return nil, fmt.Errorf("save schedule config: %w", err)
	}

	if cfg.Enabled {
		if err := s.applyCronSchedule(&cfg); err != nil {
			return nil, err
		}
	} else {
		s.removeCronSchedule()
	}

	return &cfg, nil
}

func (s *BackupService) applyCronSchedule(cfg *BackupScheduleConfig) error {
	s.cronMu.Lock()
	defer s.cronMu.Unlock()

	if s.cronSched == nil {
		return fmt.Errorf("cron scheduler not initialized")
	}
	if s.cronEntryID != 0 {
		s.cronSched.Remove(s.cronEntryID)
		s.cronEntryID = 0
	}

	entryID, err := s.cronSched.AddFunc(cfg.CronExpr, func() {
		s.runScheduledBackup()
	})
	if err != nil {
		return infraerrors.BadRequest("INVALID_CRON", fmt.Sprintf("failed to schedule: %v", err))
	}
	s.cronEntryID = entryID
	logger.LegacyPrintf("service.backup", "[Backup] 定时备份已启用: %s", cfg.CronExpr)
	return nil
}

func (s *BackupService) removeCronSchedule() {
	s.cronMu.Lock()
	defer s.cronMu.Unlock()
	if s.cronSched != nil && s.cronEntryID != 0 {
		s.cronSched.Remove(s.cronEntryID)
		s.cronEntryID = 0
		logger.LegacyPrintf("service.backup", "[Backup] 定时备份已停用")
	}
}

func (s *BackupService) runScheduledBackup() {
	s.wg.Add(1)
	defer s.wg.Done()

	ctx, cancel := context.WithTimeout(s.bgCtx, 30*time.Minute)
	defer cancel()

	schedule, _ := s.GetSchedule(ctx)
	expireDays := 14
	if schedule != nil && schedule.RetainDays > 0 {
		expireDays = schedule.RetainDays
	}

	logger.LegacyPrintf("service.backup", "[Backup] 开始执行定时备份, 过期天数: %d", expireDays)
	record, err := s.CreateBackup(ctx, "scheduled", expireDays)
	if err != nil {
		if errors.Is(err, ErrBackupInProgress) {
			logger.LegacyPrintf("service.backup", "[Backup] 定时备份跳过: 已有备份正在进行中")
		} else {
			logger.LegacyPrintf("service.backup", "[Backup] 定时备份失败: %v", err)
		}
		return
	}
	logger.LegacyPrintf("service.backup", "[Backup] 定时备份完成: id=%s size=%d", record.ID, record.SizeBytes)

	if schedule == nil {
		return
	}
	if err := s.cleanupOldBackups(ctx, schedule); err != nil {
		logger.LegacyPrintf("service.backup", "[Backup] 清理过期备份失败: %v", err)
	}
}
