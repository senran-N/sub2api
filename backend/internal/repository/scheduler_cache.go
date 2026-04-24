package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/senran-N/sub2api/internal/service"
)

const (
	schedulerBucketSetKey        = "sched:buckets"
	schedulerOutboxWatermarkKey  = "sched:outbox:watermark"
	schedulerAccountPrefix       = "sched:acc:"
	schedulerAccountMetaPrefix   = "sched:meta:"
	schedulerActivePrefix        = "sched:active:"
	schedulerIndexPrefix         = "sched:idx:"
	schedulerIndexRegistryPrefix = "sched:idxreg:"
	schedulerIndexValuePrefix    = "sched:idxvals:"
	schedulerReadyPrefix         = "sched:ready:"
	schedulerVersionPrefix       = "sched:ver:"
	schedulerSnapshotPrefix      = "sched:"
	schedulerLockPrefix          = "sched:lock:"

	defaultSchedulerSnapshotMGetChunkSize  = 128
	defaultSchedulerSnapshotWriteChunkSize = 256
)

type schedulerCache struct {
	rdb            *redis.Client
	mgetChunkSize  int
	writeChunkSize int
}

func NewSchedulerCache(rdb *redis.Client) service.SchedulerCache {
	return newSchedulerCacheWithChunkSizes(rdb, defaultSchedulerSnapshotMGetChunkSize, defaultSchedulerSnapshotWriteChunkSize)
}

func newSchedulerCacheWithChunkSizes(rdb *redis.Client, mgetChunkSize, writeChunkSize int) service.SchedulerCache {
	if mgetChunkSize <= 0 {
		mgetChunkSize = defaultSchedulerSnapshotMGetChunkSize
	}
	if writeChunkSize <= 0 {
		writeChunkSize = defaultSchedulerSnapshotWriteChunkSize
	}
	return &schedulerCache{
		rdb:            rdb,
		mgetChunkSize:  mgetChunkSize,
		writeChunkSize: writeChunkSize,
	}
}

func (c *schedulerCache) GetSnapshot(ctx context.Context, bucket service.SchedulerBucket) ([]*service.Account, bool, error) {
	_, snapshotKey, hit, err := c.resolveActiveSnapshotKey(ctx, bucket)
	if err != nil || !hit {
		return nil, hit, err
	}

	ids, err := c.rdb.ZRange(ctx, snapshotKey, 0, -1).Result()
	if err != nil {
		return nil, false, err
	}
	if len(ids) == 0 {
		// 空快照视为缓存未命中，触发数据库回退查询
		// 这解决了新分组创建后立即绑定账号时的竞态条件问题
		return nil, false, nil
	}

	accounts, complete, err := c.loadSnapshotAccounts(ctx, ids)
	if err != nil {
		return nil, false, err
	}
	if !complete {
		return nil, false, nil
	}
	return accounts, true, nil
}

func (c *schedulerCache) GetSnapshotPage(
	ctx context.Context,
	bucket service.SchedulerBucket,
	offset, limit int,
) ([]*service.Account, bool, bool, error) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = defaultSchedulerSnapshotMGetChunkSize
	}

	_, snapshotKey, hit, err := c.resolveActiveSnapshotKey(ctx, bucket)
	if err != nil || !hit {
		return nil, hit, false, err
	}

	stop := int64(offset + limit)
	ids, err := c.rdb.ZRange(ctx, snapshotKey, int64(offset), stop).Result()
	if err != nil {
		return nil, false, false, err
	}
	if len(ids) == 0 {
		if offset == 0 {
			return nil, false, false, nil
		}
		return []*service.Account{}, true, false, nil
	}

	hasMore := len(ids) > limit
	if hasMore {
		ids = ids[:limit]
	}
	accounts, complete, err := c.loadSnapshotAccounts(ctx, ids)
	if err != nil {
		return nil, false, false, err
	}
	if !complete {
		return nil, false, false, nil
	}
	return accounts, true, hasMore, nil
}

func (c *schedulerCache) GetCapabilityIndexPage(
	ctx context.Context,
	bucket service.SchedulerBucket,
	index service.SchedulerCapabilityIndex,
	offset, limit int,
) ([]*service.Account, bool, bool, error) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = defaultSchedulerSnapshotMGetChunkSize
	}

	version, indexKey, hit, err := c.resolveCapabilityIndexKey(ctx, bucket, index)
	if err != nil || !hit {
		return nil, hit, false, err
	}
	if index.Kind == service.SchedulerCapabilityIndexAll {
		_ = version
		return c.GetSnapshotPage(ctx, bucket, offset, limit)
	}

	stop := int64(offset + limit)
	ids, err := c.rdb.ZRange(ctx, indexKey, int64(offset), stop).Result()
	if err != nil {
		return nil, false, false, err
	}
	if len(ids) == 0 {
		return []*service.Account{}, true, false, nil
	}

	hasMore := len(ids) > limit
	if hasMore {
		ids = ids[:limit]
	}
	accounts, complete, err := c.loadSnapshotAccounts(ctx, ids)
	if err != nil {
		return nil, false, false, err
	}
	if !complete {
		return nil, false, false, nil
	}
	return accounts, true, hasMore, nil
}

func (c *schedulerCache) HasCapabilityIndexMembers(
	ctx context.Context,
	bucket service.SchedulerBucket,
	index service.SchedulerCapabilityIndex,
	accountIDs []int64,
) (map[int64]bool, bool, error) {
	matches := make(map[int64]bool, len(accountIDs))
	if len(accountIDs) == 0 {
		return matches, true, nil
	}

	version, indexKey, hit, err := c.resolveCapabilityIndexKey(ctx, bucket, index)
	if err != nil || !hit {
		return nil, hit, err
	}
	if index.Kind == service.SchedulerCapabilityIndexAll {
		_ = version
		for _, id := range accountIDs {
			if id > 0 {
				matches[id] = true
			}
		}
		return matches, true, nil
	}

	members := make([]string, 0, len(accountIDs))
	for _, id := range accountIDs {
		members = append(members, strconv.FormatInt(id, 10))
	}
	pipe := c.rdb.Pipeline()
	cmds := make([]*redis.FloatCmd, 0, len(members))
	for _, member := range members {
		cmds = append(cmds, pipe.ZScore(ctx, indexKey, member))
	}
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, false, err
	}
	for i, cmd := range cmds {
		if _, err := cmd.Result(); err == nil {
			matches[accountIDs[i]] = true
		}
	}
	return matches, true, nil
}

func (c *schedulerCache) ListCapabilityIndexValues(
	ctx context.Context,
	bucket service.SchedulerBucket,
	kind service.SchedulerCapabilityIndexKind,
) ([]string, bool, error) {
	version, hit, err := c.resolveActiveSnapshotVersion(ctx, bucket)
	if err != nil || !hit {
		return nil, hit, err
	}
	valuesKey := schedulerIndexValuesKey(bucket, version, kind)
	values, err := c.rdb.SMembers(ctx, valuesKey).Result()
	if err == redis.Nil {
		return []string{}, true, nil
	}
	if err != nil {
		return nil, false, err
	}
	sort.Strings(values)
	return values, true, nil
}

func (c *schedulerCache) SetSnapshot(ctx context.Context, bucket service.SchedulerBucket, accounts []service.Account) error {
	activeKey := schedulerBucketKey(schedulerActivePrefix, bucket)
	oldActive, _ := c.rdb.Get(ctx, activeKey).Result()

	versionKey := schedulerBucketKey(schedulerVersionPrefix, bucket)
	version, err := c.rdb.Incr(ctx, versionKey).Result()
	if err != nil {
		return err
	}

	versionStr := strconv.FormatInt(version, 10)
	snapshotKey := schedulerSnapshotKey(bucket, versionStr)
	indexKeysRegistryKey := schedulerIndexRegistryKey(bucket, versionStr)

	if err := c.writeAccounts(ctx, accounts); err != nil {
		return err
	}

	indexBuild := buildSchedulerCapabilityIndices(bucket, versionStr, accounts)

	pipe := c.rdb.Pipeline()
	if len(accounts) > 0 {
		// 使用序号作为 score，保持数据库返回的排序语义。
		members := make([]redis.Z, 0, len(accounts))
		for idx, account := range accounts {
			members = append(members, redis.Z{
				Score:  float64(idx),
				Member: strconv.FormatInt(account.ID, 10),
			})
		}
		for start := 0; start < len(members); start += c.writeChunkSize {
			end := start + c.writeChunkSize
			if end > len(members) {
				end = len(members)
			}
			pipe.ZAdd(ctx, snapshotKey, members[start:end]...)
		}
	} else {
		pipe.Del(ctx, snapshotKey)
	}
	for indexKey, members := range indexBuild.zsets {
		if len(members) == 0 {
			pipe.Del(ctx, indexKey)
			continue
		}
		for start := 0; start < len(members); start += c.writeChunkSize {
			end := start + c.writeChunkSize
			if end > len(members) {
				end = len(members)
			}
			pipe.ZAdd(ctx, indexKey, members[start:end]...)
		}
	}
	for valuesKey, values := range indexBuild.values {
		pipe.Del(ctx, valuesKey)
		if len(values) > 0 {
			valueMembers := make([]any, 0, len(values))
			for _, value := range values {
				valueMembers = append(valueMembers, value)
			}
			pipe.SAdd(ctx, valuesKey, valueMembers...)
		}
	}
	pipe.Del(ctx, indexKeysRegistryKey)
	if len(indexBuild.allKeys) > 0 {
		registryMembers := make([]any, 0, len(indexBuild.allKeys))
		for _, key := range indexBuild.allKeys {
			registryMembers = append(registryMembers, key)
		}
		pipe.SAdd(ctx, indexKeysRegistryKey, registryMembers...)
	}
	pipe.Set(ctx, activeKey, versionStr, 0)
	pipe.Set(ctx, schedulerBucketKey(schedulerReadyPrefix, bucket), "1", 0)
	pipe.SAdd(ctx, schedulerBucketSetKey, bucket.String())
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	if oldActive != "" && oldActive != versionStr {
		c.cleanupSnapshotVersion(ctx, bucket, oldActive)
	}

	return nil
}

func (c *schedulerCache) GetAccount(ctx context.Context, accountID int64) (*service.Account, error) {
	key := schedulerAccountKey(strconv.FormatInt(accountID, 10))
	val, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return decodeCachedAccount(val)
}

func (c *schedulerCache) SetAccount(ctx context.Context, account *service.Account) error {
	if account == nil || account.ID <= 0 {
		return nil
	}
	return c.writeAccounts(ctx, []service.Account{*account})
}

func (c *schedulerCache) DeleteAccount(ctx context.Context, accountID int64) error {
	if accountID <= 0 {
		return nil
	}
	id := strconv.FormatInt(accountID, 10)
	return c.rdb.Del(ctx, schedulerAccountKey(id), schedulerAccountMetaKey(id)).Err()
}

func (c *schedulerCache) UpdateLastUsed(ctx context.Context, updates map[int64]time.Time) error {
	if len(updates) == 0 {
		return nil
	}

	keys := make([]string, 0, len(updates))
	ids := make([]int64, 0, len(updates))
	for id := range updates {
		keys = append(keys, schedulerAccountKey(strconv.FormatInt(id, 10)))
		ids = append(ids, id)
	}

	values, err := c.mgetChunked(ctx, keys)
	if err != nil {
		return err
	}

	pipe := c.rdb.Pipeline()
	for i, val := range values {
		if val == nil {
			continue
		}
		account, err := decodeCachedAccount(val)
		if err != nil {
			return err
		}
		account.LastUsedAt = ptrTime(updates[ids[i]])
		updated, err := json.Marshal(account)
		if err != nil {
			return err
		}
		metaPayload, err := json.Marshal(buildSchedulerMetadataAccount(*account))
		if err != nil {
			return err
		}
		pipe.Set(ctx, keys[i], updated, 0)
		pipe.Set(ctx, schedulerAccountMetaKey(strconv.FormatInt(ids[i], 10)), metaPayload, 0)
	}
	_, err = pipe.Exec(ctx)
	return err
}

func (c *schedulerCache) TryLockBucket(ctx context.Context, bucket service.SchedulerBucket, ttl time.Duration) (bool, error) {
	key := schedulerBucketKey(schedulerLockPrefix, bucket)
	return c.rdb.SetNX(ctx, key, time.Now().UnixNano(), ttl).Result()
}

func (c *schedulerCache) ListBuckets(ctx context.Context) ([]service.SchedulerBucket, error) {
	raw, err := c.rdb.SMembers(ctx, schedulerBucketSetKey).Result()
	if err != nil {
		return nil, err
	}
	out := make([]service.SchedulerBucket, 0, len(raw))
	for _, entry := range raw {
		bucket, ok := service.ParseSchedulerBucket(entry)
		if !ok {
			continue
		}
		out = append(out, bucket)
	}
	return out, nil
}

func (c *schedulerCache) GetOutboxWatermark(ctx context.Context) (int64, error) {
	val, err := c.rdb.Get(ctx, schedulerOutboxWatermarkKey).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	id, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (c *schedulerCache) SetOutboxWatermark(ctx context.Context, id int64) error {
	return c.rdb.Set(ctx, schedulerOutboxWatermarkKey, strconv.FormatInt(id, 10), 0).Err()
}

func (c *schedulerCache) resolveActiveSnapshotKey(
	ctx context.Context,
	bucket service.SchedulerBucket,
) (string, string, bool, error) {
	version, hit, err := c.resolveActiveSnapshotVersion(ctx, bucket)
	if err != nil || !hit {
		return "", "", hit, err
	}
	return version, schedulerSnapshotKey(bucket, version), true, nil
}

func (c *schedulerCache) resolveActiveSnapshotVersion(
	ctx context.Context,
	bucket service.SchedulerBucket,
) (string, bool, error) {
	readyKey := schedulerBucketKey(schedulerReadyPrefix, bucket)
	readyVal, err := c.rdb.Get(ctx, readyKey).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	if readyVal != "1" {
		return "", false, nil
	}

	activeKey := schedulerBucketKey(schedulerActivePrefix, bucket)
	activeVal, err := c.rdb.Get(ctx, activeKey).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return activeVal, true, nil
}

func (c *schedulerCache) resolveCapabilityIndexKey(
	ctx context.Context,
	bucket service.SchedulerBucket,
	index service.SchedulerCapabilityIndex,
) (string, string, bool, error) {
	version, hit, err := c.resolveActiveSnapshotVersion(ctx, bucket)
	if err != nil || !hit {
		return "", "", hit, err
	}
	if index.Kind == service.SchedulerCapabilityIndexAll {
		return version, schedulerSnapshotKey(bucket, version), true, nil
	}
	return version, schedulerCapabilityIndexKey(bucket, version, index), true, nil
}

func (c *schedulerCache) cleanupSnapshotVersion(ctx context.Context, bucket service.SchedulerBucket, version string) {
	keys := []string{schedulerSnapshotKey(bucket, version), schedulerIndexRegistryKey(bucket, version)}
	registryMembers, err := c.rdb.SMembers(ctx, schedulerIndexRegistryKey(bucket, version)).Result()
	if err == nil && len(registryMembers) > 0 {
		keys = append(keys, registryMembers...)
	}
	_ = c.rdb.Del(ctx, keys...).Err()
}

func schedulerBucketKey(prefix string, bucket service.SchedulerBucket) string {
	return fmt.Sprintf("%s%d:%s:%s", prefix, bucket.GroupID, bucket.Platform, bucket.Mode)
}

func schedulerSnapshotKey(bucket service.SchedulerBucket, version string) string {
	return fmt.Sprintf("%s%d:%s:%s:v%s", schedulerSnapshotPrefix, bucket.GroupID, bucket.Platform, bucket.Mode, version)
}

func schedulerIndexRegistryKey(bucket service.SchedulerBucket, version string) string {
	return fmt.Sprintf("%s%d:%s:%s:v%s", schedulerIndexRegistryPrefix, bucket.GroupID, bucket.Platform, bucket.Mode, version)
}

func schedulerIndexValuesKey(bucket service.SchedulerBucket, version string, kind service.SchedulerCapabilityIndexKind) string {
	return fmt.Sprintf("%s%d:%s:%s:v%s:%s", schedulerIndexValuePrefix, bucket.GroupID, bucket.Platform, bucket.Mode, version, kind)
}

func schedulerCapabilityIndexKey(bucket service.SchedulerBucket, version string, index service.SchedulerCapabilityIndex) string {
	base := fmt.Sprintf("%s%d:%s:%s:v%s:%s", schedulerIndexPrefix, bucket.GroupID, bucket.Platform, bucket.Mode, version, index.Kind)
	value := strings.TrimSpace(index.Value)
	if value == "" {
		return base
	}
	return base + ":" + schedulerCapabilityValueToken(value)
}

func schedulerCapabilityValueToken(value string) string {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(strings.TrimSpace(strings.ToLower(value))))
	return strconv.FormatUint(hasher.Sum64(), 16)
}

func schedulerAccountKey(id string) string {
	return schedulerAccountPrefix + id
}

func schedulerAccountMetaKey(id string) string {
	return schedulerAccountMetaPrefix + id
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

func decodeCachedAccount(val any) (*service.Account, error) {
	var payload []byte
	switch raw := val.(type) {
	case string:
		payload = []byte(raw)
	case []byte:
		payload = raw
	default:
		return nil, fmt.Errorf("unexpected account cache type: %T", val)
	}
	var account service.Account
	if err := json.Unmarshal(payload, &account); err != nil {
		return nil, err
	}
	return &account, nil
}

func (c *schedulerCache) writeAccounts(ctx context.Context, accounts []service.Account) error {
	if len(accounts) == 0 {
		return nil
	}

	pipe := c.rdb.Pipeline()
	pending := 0
	flush := func() error {
		if pending == 0 {
			return nil
		}
		if _, err := pipe.Exec(ctx); err != nil {
			return err
		}
		pipe = c.rdb.Pipeline()
		pending = 0
		return nil
	}

	for _, account := range accounts {
		fullPayload, err := json.Marshal(account)
		if err != nil {
			return err
		}
		metaPayload, err := json.Marshal(buildSchedulerMetadataAccount(account))
		if err != nil {
			return err
		}

		id := strconv.FormatInt(account.ID, 10)
		pipe.Set(ctx, schedulerAccountKey(id), fullPayload, 0)
		pipe.Set(ctx, schedulerAccountMetaKey(id), metaPayload, 0)
		pending++
		if pending >= c.writeChunkSize {
			if err := flush(); err != nil {
				return err
			}
		}
	}

	return flush()
}

func (c *schedulerCache) mgetChunked(ctx context.Context, keys []string) ([]any, error) {
	if len(keys) == 0 {
		return []any{}, nil
	}

	out := make([]any, 0, len(keys))
	chunkSize := c.mgetChunkSize
	if chunkSize <= 0 {
		chunkSize = defaultSchedulerSnapshotMGetChunkSize
	}
	for start := 0; start < len(keys); start += chunkSize {
		end := start + chunkSize
		if end > len(keys) {
			end = len(keys)
		}
		part, err := c.rdb.MGet(ctx, keys[start:end]...).Result()
		if err != nil {
			return nil, err
		}
		out = append(out, part...)
	}
	return out, nil
}

func (c *schedulerCache) loadSnapshotAccounts(ctx context.Context, ids []string) ([]*service.Account, bool, error) {
	if len(ids) == 0 {
		return []*service.Account{}, true, nil
	}

	keys := make([]string, 0, len(ids))
	for _, id := range ids {
		keys = append(keys, schedulerAccountMetaKey(id))
	}
	values, err := c.mgetChunked(ctx, keys)
	if err != nil {
		return nil, false, err
	}

	accounts := make([]*service.Account, 0, len(values))
	for _, val := range values {
		if val == nil {
			return nil, false, nil
		}
		account, err := decodeCachedAccount(val)
		if err != nil {
			return nil, false, err
		}
		accounts = append(accounts, account)
	}
	return accounts, true, nil
}

type schedulerCapabilityIndexBuild struct {
	zsets   map[string][]redis.Z
	values  map[string][]string
	allKeys []string
}

func buildSchedulerCapabilityIndices(
	bucket service.SchedulerBucket,
	version string,
	accounts []service.Account,
) schedulerCapabilityIndexBuild {
	build := schedulerCapabilityIndexBuild{
		zsets:  make(map[string][]redis.Z),
		values: make(map[string][]string),
	}
	seenKeys := make(map[string]struct{})
	seenValues := make(map[string]map[string]struct{})

	registerKey := func(key string) {
		if key == "" {
			return
		}
		if _, exists := seenKeys[key]; exists {
			return
		}
		seenKeys[key] = struct{}{}
		build.allKeys = append(build.allKeys, key)
	}
	registerValue := func(valuesKey, value string) {
		if valuesKey == "" || strings.TrimSpace(value) == "" {
			return
		}
		if _, ok := seenValues[valuesKey]; !ok {
			seenValues[valuesKey] = make(map[string]struct{})
		}
		if _, exists := seenValues[valuesKey][value]; exists {
			return
		}
		seenValues[valuesKey][value] = struct{}{}
		build.values[valuesKey] = append(build.values[valuesKey], value)
	}
	appendMember := func(index service.SchedulerCapabilityIndex, score float64, accountID int64) {
		indexKey := schedulerCapabilityIndexKey(bucket, version, index)
		build.zsets[indexKey] = append(build.zsets[indexKey], redis.Z{
			Score:  score,
			Member: strconv.FormatInt(accountID, 10),
		})
		registerKey(indexKey)
	}

	for idx, account := range accounts {
		score := float64(idx)
		if account.IsPrivacySet() {
			appendMember(service.SchedulerCapabilityIndex{Kind: service.SchedulerCapabilityIndexPrivacySet}, score, account.ID)
		}
		if isPotentialOpenAIWSAccount(&account) {
			appendMember(service.SchedulerCapabilityIndex{Kind: service.SchedulerCapabilityIndexOpenAIWS}, score, account.ID)
		}

		modelValues, unrestrictedModels := account.SchedulerModelCapabilityValues()
		if unrestrictedModels {
			appendMember(service.SchedulerCapabilityIndex{Kind: service.SchedulerCapabilityIndexModelAny}, score, account.ID)
			continue
		}

		patternValuesKey := schedulerIndexValuesKey(bucket, version, service.SchedulerCapabilityIndexModelPattern)
		for _, modelValue := range modelValues {
			trimmed := strings.TrimSpace(modelValue)
			if trimmed == "" {
				continue
			}
			if strings.Contains(trimmed, "*") {
				registerValue(patternValuesKey, trimmed)
				appendMember(service.SchedulerCapabilityIndex{Kind: service.SchedulerCapabilityIndexModelPattern, Value: trimmed}, score, account.ID)
				continue
			}
			appendMember(service.SchedulerCapabilityIndex{Kind: service.SchedulerCapabilityIndexModelExact, Value: trimmed}, score, account.ID)
		}
	}

	sort.Strings(build.allKeys)
	for key := range build.values {
		sort.Strings(build.values[key])
		registerKey(key)
	}
	registerKey(schedulerIndexRegistryKey(bucket, version))
	return build
}

func isPotentialOpenAIWSAccount(account *service.Account) bool {
	if account == nil || !account.IsOpenAI() {
		return false
	}
	if account.IsOpenAIWSForceHTTPEnabled() {
		return false
	}
	if !account.IsOpenAIOAuth() && !account.IsOpenAIApiKey() {
		return false
	}
	if account.Concurrency <= 0 {
		return false
	}
	mode := account.ResolveOpenAIResponsesWebSocketV2Mode(service.OpenAIWSIngressModeCtxPool)
	return mode != service.OpenAIWSIngressModeOff
}

func buildSchedulerMetadataAccount(account service.Account) service.Account {
	return service.Account{
		ID:                      account.ID,
		Name:                    account.Name,
		Platform:                account.Platform,
		Type:                    account.Type,
		Concurrency:             account.Concurrency,
		Priority:                account.Priority,
		RateMultiplier:          account.RateMultiplier,
		LoadFactor:              account.LoadFactor,
		Status:                  account.Status,
		LastUsedAt:              account.LastUsedAt,
		ExpiresAt:               account.ExpiresAt,
		AutoPauseOnExpired:      account.AutoPauseOnExpired,
		Schedulable:             account.Schedulable,
		RateLimitedAt:           account.RateLimitedAt,
		RateLimitResetAt:        account.RateLimitResetAt,
		OverloadUntil:           account.OverloadUntil,
		TempUnschedulableUntil:  account.TempUnschedulableUntil,
		TempUnschedulableReason: account.TempUnschedulableReason,
		SessionWindowStart:      account.SessionWindowStart,
		SessionWindowEnd:        account.SessionWindowEnd,
		SessionWindowStatus:     account.SessionWindowStatus,
		Credentials:             filterSchedulerCredentials(account.Credentials),
		Extra:                   filterSchedulerExtra(account.Extra),
	}
}

func filterSchedulerCredentials(credentials map[string]any) map[string]any {
	if len(credentials) == 0 {
		return nil
	}
	keys := []string{"model_mapping", "api_key", "project_id", "oauth_type"}
	filtered := make(map[string]any)
	for _, key := range keys {
		if value, ok := credentials[key]; ok && value != nil {
			filtered[key] = value
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return filtered
}

func filterSchedulerExtra(extra map[string]any) map[string]any {
	if len(extra) == 0 {
		return nil
	}
	keys := []string{
		"mixed_scheduling",
		"privacy_mode",
		"window_cost_limit",
		"window_cost_sticky_reserve",
		"max_sessions",
		"session_idle_timeout_minutes",
		"base_rpm",
		"rpm_strategy",
		"rpm_sticky_buffer",
		"allow_overages",
		"model_rate_limits",
		"quota_limit",
		"quota_used",
		"quota_daily_limit",
		"quota_daily_used",
		"quota_daily_start",
		"quota_daily_reset_mode",
		"quota_daily_reset_hour",
		"quota_weekly_limit",
		"quota_weekly_used",
		"quota_weekly_start",
		"quota_weekly_reset_mode",
		"quota_weekly_reset_day",
		"quota_weekly_reset_hour",
		"quota_reset_timezone",
		"openai_ws_enabled",
		"openai_ws_force_http",
		"openai_ws_allow_store_recovery",
	}
	prefixes := []string{
		"responses_websockets",
		"openai_oauth_responses_websockets",
		"openai_apikey_responses_websockets",
	}
	filtered := make(map[string]any)
	for _, key := range keys {
		if value, ok := extra[key]; ok && value != nil {
			filtered[key] = value
		}
	}
	for key, value := range extra {
		if value == nil {
			continue
		}
		for _, prefix := range prefixes {
			if key == prefix || len(key) > len(prefix) && key[:len(prefix)] == prefix {
				filtered[key] = value
				break
			}
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return filtered
}
