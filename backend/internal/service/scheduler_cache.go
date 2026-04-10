package service

import (
	"context"
	"time"
)

// SchedulerCache 负责调度快照与账号快照的缓存读写。
type SchedulerCache interface {
	// GetSnapshot 读取快照并返回命中与否（ready + active + 数据完整）。
	GetSnapshot(ctx context.Context, bucket SchedulerBucket) ([]*Account, bool, error)
	// SetSnapshot 写入快照并切换激活版本。
	SetSnapshot(ctx context.Context, bucket SchedulerBucket, accounts []Account) error
	// GetAccount 获取单账号快照。
	GetAccount(ctx context.Context, accountID int64) (*Account, error)
	// SetAccount 写入单账号快照（包含不可调度状态）。
	SetAccount(ctx context.Context, account *Account) error
	// DeleteAccount 删除单账号快照。
	DeleteAccount(ctx context.Context, accountID int64) error
	// UpdateLastUsed 批量更新账号的最后使用时间。
	UpdateLastUsed(ctx context.Context, updates map[int64]time.Time) error
	// TryLockBucket 尝试获取分桶重建锁。
	TryLockBucket(ctx context.Context, bucket SchedulerBucket, ttl time.Duration) (bool, error)
	// ListBuckets 返回已注册的分桶集合。
	ListBuckets(ctx context.Context) ([]SchedulerBucket, error)
	// GetOutboxWatermark 读取 outbox 水位。
	GetOutboxWatermark(ctx context.Context) (int64, error)
	// SetOutboxWatermark 保存 outbox 水位。
	SetOutboxWatermark(ctx context.Context, id int64) error
}

// SchedulerCachePager 提供按页读取快照的能力，避免调度热路径整桶扫描。
//
// 这是一个可选扩展接口，旧实现无需立即修改；调用方应通过 type assertion 使用。
type SchedulerCachePager interface {
	// GetSnapshotPage 返回指定 offset/limit 的快照页。
	// hit 表示快照 ready + active + 数据完整；hasMore 表示后续仍有更多账号。
	GetSnapshotPage(ctx context.Context, bucket SchedulerBucket, offset, limit int) (accounts []*Account, hit bool, hasMore bool, err error)
}

type SchedulerCapabilityIndexKind string

const (
	SchedulerCapabilityIndexAll          SchedulerCapabilityIndexKind = "all"
	SchedulerCapabilityIndexPrivacySet   SchedulerCapabilityIndexKind = "privacy_set"
	SchedulerCapabilityIndexOpenAIWS     SchedulerCapabilityIndexKind = "openai_ws"
	SchedulerCapabilityIndexModelAny     SchedulerCapabilityIndexKind = "model_any"
	SchedulerCapabilityIndexModelExact   SchedulerCapabilityIndexKind = "model_exact"
	SchedulerCapabilityIndexModelPattern SchedulerCapabilityIndexKind = "model_pattern"
)

type SchedulerCapabilityIndex struct {
	Kind  SchedulerCapabilityIndexKind
	Value string
}

// SchedulerCacheIndexed 提供按能力索引读取与成员校验能力。
//
// 这是用于高规模候选源裁剪的可选扩展接口。
type SchedulerCacheIndexed interface {
	// GetCapabilityIndexPage 返回指定能力索引下的一页账号。
	GetCapabilityIndexPage(ctx context.Context, bucket SchedulerBucket, index SchedulerCapabilityIndex, offset, limit int) (accounts []*Account, hit bool, hasMore bool, err error)
	// HasCapabilityIndexMembers 检查账号是否属于某个能力索引。
	HasCapabilityIndexMembers(ctx context.Context, bucket SchedulerBucket, index SchedulerCapabilityIndex, accountIDs []int64) (matches map[int64]bool, hit bool, err error)
	// ListCapabilityIndexValues 返回该能力索引 kind 已注册的值列表。
	ListCapabilityIndexValues(ctx context.Context, bucket SchedulerBucket, kind SchedulerCapabilityIndexKind) (values []string, hit bool, err error)
}
