package repository

import (
	"context"
	"database/sql"
	"time"

	dbaccount "github.com/senran-N/sub2api/ent/account"
	dbapikey "github.com/senran-N/sub2api/ent/apikey"
	dbgroup "github.com/senran-N/sub2api/ent/group"
	dbuser "github.com/senran-N/sub2api/ent/user"
	dbusersub "github.com/senran-N/sub2api/ent/usersubscription"
	"github.com/senran-N/sub2api/internal/service"
	"golang.org/x/sync/errgroup"
)

func (r *usageLogRepository) queryUsageLogs(ctx context.Context, query string, args ...any) (logs []service.UsageLog, err error) {
	rows, err := r.sql.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		// 保持主错误优先；仅在无错误时回传 Close 失败。
		// 同时清空返回值，避免误用不完整结果。
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
			logs = nil
		}
	}()

	logs = make([]service.UsageLog, 0)
	for rows.Next() {
		var log *service.UsageLog
		log, err = scanUsageLog(rows)
		if err != nil {
			return nil, err
		}
		logs = append(logs, *log)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *usageLogRepository) hydrateUsageLogAssociations(ctx context.Context, logs []service.UsageLog) error {
	// 关联数据使用 Ent 批量加载，避免把复杂 SQL 继续膨胀。
	if len(logs) == 0 {
		return nil
	}

	ids := collectUsageLogIDs(logs)
	users, apiKeys, accounts, groups, subs, err := r.loadUsageLogAssociations(ctx, ids)
	if err != nil {
		return err
	}

	for i := range logs {
		if user, ok := users[logs[i].UserID]; ok {
			logs[i].User = user
		}
		if key, ok := apiKeys[logs[i].APIKeyID]; ok {
			logs[i].APIKey = key
		}
		if acc, ok := accounts[logs[i].AccountID]; ok {
			logs[i].Account = acc
		}
		if logs[i].GroupID != nil {
			if group, ok := groups[*logs[i].GroupID]; ok {
				logs[i].Group = group
			}
		}
		if logs[i].SubscriptionID != nil {
			if sub, ok := subs[*logs[i].SubscriptionID]; ok {
				logs[i].Subscription = sub
			}
		}
	}
	return nil
}

func (r *usageLogRepository) loadUsageLogAssociations(ctx context.Context, ids usageLogIDs) (
	users map[int64]*service.User,
	apiKeys map[int64]*service.APIKey,
	accounts map[int64]*service.Account,
	groups map[int64]*service.Group,
	subs map[int64]*service.UserSubscription,
	err error,
) {
	if r == nil || r.client == nil {
		return map[int64]*service.User{}, map[int64]*service.APIKey{}, map[int64]*service.Account{}, map[int64]*service.Group{}, map[int64]*service.UserSubscription{}, nil
	}
	if r.db == nil {
		return r.loadUsageLogAssociationsSerial(ctx, ids)
	}
	return r.loadUsageLogAssociationsParallel(ctx, ids)
}

func (r *usageLogRepository) loadUsageLogAssociationsSerial(ctx context.Context, ids usageLogIDs) (
	users map[int64]*service.User,
	apiKeys map[int64]*service.APIKey,
	accounts map[int64]*service.Account,
	groups map[int64]*service.Group,
	subs map[int64]*service.UserSubscription,
	err error,
) {
	users, err = r.loadUsers(ctx, ids.userIDs)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	apiKeys, err = r.loadAPIKeys(ctx, ids.apiKeyIDs)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	accounts, err = r.loadAccounts(ctx, ids.accountIDs)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	groups, err = r.loadGroups(ctx, ids.groupIDs)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	subs, err = r.loadSubscriptions(ctx, ids.subscriptionIDs)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	return users, apiKeys, accounts, groups, subs, nil
}

func (r *usageLogRepository) loadUsageLogAssociationsParallel(ctx context.Context, ids usageLogIDs) (
	users map[int64]*service.User,
	apiKeys map[int64]*service.APIKey,
	accounts map[int64]*service.Account,
	groups map[int64]*service.Group,
	subs map[int64]*service.UserSubscription,
	err error,
) {
	users = make(map[int64]*service.User)
	apiKeys = make(map[int64]*service.APIKey)
	accounts = make(map[int64]*service.Account)
	groups = make(map[int64]*service.Group)
	subs = make(map[int64]*service.UserSubscription)

	queryGroup, queryCtx := errgroup.WithContext(ctx)
	queryGroup.Go(func() error {
		var loadErr error
		users, loadErr = r.loadUsers(queryCtx, ids.userIDs)
		return loadErr
	})
	queryGroup.Go(func() error {
		var loadErr error
		apiKeys, loadErr = r.loadAPIKeys(queryCtx, ids.apiKeyIDs)
		return loadErr
	})
	queryGroup.Go(func() error {
		var loadErr error
		accounts, loadErr = r.loadAccounts(queryCtx, ids.accountIDs)
		return loadErr
	})
	queryGroup.Go(func() error {
		var loadErr error
		groups, loadErr = r.loadGroups(queryCtx, ids.groupIDs)
		return loadErr
	})
	queryGroup.Go(func() error {
		var loadErr error
		subs, loadErr = r.loadSubscriptions(queryCtx, ids.subscriptionIDs)
		return loadErr
	})
	if err := queryGroup.Wait(); err != nil {
		return nil, nil, nil, nil, nil, err
	}
	return users, apiKeys, accounts, groups, subs, nil
}

type usageLogIDs struct {
	userIDs         []int64
	apiKeyIDs       []int64
	accountIDs      []int64
	groupIDs        []int64
	subscriptionIDs []int64
}

func collectUsageLogIDs(logs []service.UsageLog) usageLogIDs {
	idSet := func() map[int64]struct{} { return make(map[int64]struct{}) }

	userIDs := idSet()
	apiKeyIDs := idSet()
	accountIDs := idSet()
	groupIDs := idSet()
	subscriptionIDs := idSet()

	for i := range logs {
		userIDs[logs[i].UserID] = struct{}{}
		apiKeyIDs[logs[i].APIKeyID] = struct{}{}
		accountIDs[logs[i].AccountID] = struct{}{}
		if logs[i].GroupID != nil {
			groupIDs[*logs[i].GroupID] = struct{}{}
		}
		if logs[i].SubscriptionID != nil {
			subscriptionIDs[*logs[i].SubscriptionID] = struct{}{}
		}
	}

	return usageLogIDs{
		userIDs:         setToSlice(userIDs),
		apiKeyIDs:       setToSlice(apiKeyIDs),
		accountIDs:      setToSlice(accountIDs),
		groupIDs:        setToSlice(groupIDs),
		subscriptionIDs: setToSlice(subscriptionIDs),
	}
}

func (r *usageLogRepository) loadUsers(ctx context.Context, ids []int64) (map[int64]*service.User, error) {
	out := make(map[int64]*service.User)
	if len(ids) == 0 {
		return out, nil
	}
	models, err := r.client.User.Query().Where(dbuser.IDIn(ids...)).All(ctx)
	if err != nil {
		return nil, err
	}
	for _, m := range models {
		out[m.ID] = userEntityToService(m)
	}
	return out, nil
}

func (r *usageLogRepository) loadAPIKeys(ctx context.Context, ids []int64) (map[int64]*service.APIKey, error) {
	out := make(map[int64]*service.APIKey)
	if len(ids) == 0 {
		return out, nil
	}
	models, err := r.client.APIKey.Query().Where(dbapikey.IDIn(ids...)).All(ctx)
	if err != nil {
		return nil, err
	}
	for _, m := range models {
		out[m.ID] = apiKeyEntityToService(m)
	}
	return out, nil
}

func (r *usageLogRepository) loadAccounts(ctx context.Context, ids []int64) (map[int64]*service.Account, error) {
	out := make(map[int64]*service.Account)
	if len(ids) == 0 {
		return out, nil
	}
	models, err := r.client.Account.Query().Where(dbaccount.IDIn(ids...)).All(ctx)
	if err != nil {
		return nil, err
	}
	for _, m := range models {
		out[m.ID] = accountEntityToService(m)
	}
	return out, nil
}

func (r *usageLogRepository) loadGroups(ctx context.Context, ids []int64) (map[int64]*service.Group, error) {
	out := make(map[int64]*service.Group)
	if len(ids) == 0 {
		return out, nil
	}
	models, err := r.client.Group.Query().Where(dbgroup.IDIn(ids...)).All(ctx)
	if err != nil {
		return nil, err
	}
	for _, m := range models {
		out[m.ID] = groupEntityToService(m)
	}
	return out, nil
}

func (r *usageLogRepository) loadSubscriptions(ctx context.Context, ids []int64) (map[int64]*service.UserSubscription, error) {
	out := make(map[int64]*service.UserSubscription)
	if len(ids) == 0 {
		return out, nil
	}
	models, err := r.client.UserSubscription.Query().Where(dbusersub.IDIn(ids...)).All(ctx)
	if err != nil {
		return nil, err
	}
	for _, m := range models {
		out[m.ID] = userSubscriptionEntityToService(m)
	}
	return out, nil
}

func scanUsageLog(scanner interface{ Scan(...any) error }) (*service.UsageLog, error) {
	var (
		id                    int64
		userID                int64
		apiKeyID              int64
		accountID             int64
		requestID             sql.NullString
		model                 string
		requestedModel        sql.NullString
		upstreamModel         sql.NullString
		groupID               sql.NullInt64
		subscriptionID        sql.NullInt64
		inputTokens           int
		outputTokens          int
		cacheCreationTokens   int
		cacheReadTokens       int
		cacheCreation5m       int
		cacheCreation1h       int
		inputCost             float64
		outputCost            float64
		cacheCreationCost     float64
		cacheReadCost         float64
		totalCost             float64
		actualCost            float64
		rateMultiplier        float64
		accountRateMultiplier sql.NullFloat64
		billingType           int16
		requestTypeRaw        int16
		stream                bool
		openaiWSMode          bool
		durationMs            sql.NullInt64
		firstTokenMs          sql.NullInt64
		userAgent             sql.NullString
		ipAddress             sql.NullString
		imageCount            int
		imageSize             sql.NullString
		mediaType             sql.NullString
		serviceTier           sql.NullString
		reasoningEffort       sql.NullString
		inboundEndpoint       sql.NullString
		upstreamEndpoint      sql.NullString
		cacheTTLOverridden    bool
		createdAt             time.Time
	)

	if err := scanner.Scan(
		&id,
		&userID,
		&apiKeyID,
		&accountID,
		&requestID,
		&model,
		&requestedModel,
		&upstreamModel,
		&groupID,
		&subscriptionID,
		&inputTokens,
		&outputTokens,
		&cacheCreationTokens,
		&cacheReadTokens,
		&cacheCreation5m,
		&cacheCreation1h,
		&inputCost,
		&outputCost,
		&cacheCreationCost,
		&cacheReadCost,
		&totalCost,
		&actualCost,
		&rateMultiplier,
		&accountRateMultiplier,
		&billingType,
		&requestTypeRaw,
		&stream,
		&openaiWSMode,
		&durationMs,
		&firstTokenMs,
		&userAgent,
		&ipAddress,
		&imageCount,
		&imageSize,
		&mediaType,
		&serviceTier,
		&reasoningEffort,
		&inboundEndpoint,
		&upstreamEndpoint,
		&cacheTTLOverridden,
		&createdAt,
	); err != nil {
		return nil, err
	}

	log := &service.UsageLog{
		ID:                    id,
		UserID:                userID,
		APIKeyID:              apiKeyID,
		AccountID:             accountID,
		Model:                 model,
		RequestedModel:        coalesceTrimmedString(requestedModel, model),
		InputTokens:           inputTokens,
		OutputTokens:          outputTokens,
		CacheCreationTokens:   cacheCreationTokens,
		CacheReadTokens:       cacheReadTokens,
		CacheCreation5mTokens: cacheCreation5m,
		CacheCreation1hTokens: cacheCreation1h,
		InputCost:             inputCost,
		OutputCost:            outputCost,
		CacheCreationCost:     cacheCreationCost,
		CacheReadCost:         cacheReadCost,
		TotalCost:             totalCost,
		ActualCost:            actualCost,
		RateMultiplier:        rateMultiplier,
		AccountRateMultiplier: nullFloat64Ptr(accountRateMultiplier),
		BillingType:           int8(billingType),
		RequestType:           service.RequestTypeFromInt16(requestTypeRaw),
		ImageCount:            imageCount,
		CacheTTLOverridden:    cacheTTLOverridden,
		CreatedAt:             createdAt,
	}
	// 先回填 legacy 字段，再基于 legacy + request_type 计算最终请求类型，保证历史数据兼容。
	log.Stream = stream
	log.OpenAIWSMode = openaiWSMode
	log.RequestType = log.EffectiveRequestType()
	log.Stream, log.OpenAIWSMode = service.ApplyLegacyRequestFields(log.RequestType, stream, openaiWSMode)

	if requestID.Valid {
		log.RequestID = requestID.String
	}
	if groupID.Valid {
		value := groupID.Int64
		log.GroupID = &value
	}
	if subscriptionID.Valid {
		value := subscriptionID.Int64
		log.SubscriptionID = &value
	}
	if durationMs.Valid {
		value := int(durationMs.Int64)
		log.DurationMs = &value
	}
	if firstTokenMs.Valid {
		value := int(firstTokenMs.Int64)
		log.FirstTokenMs = &value
	}
	if userAgent.Valid {
		log.UserAgent = &userAgent.String
	}
	if ipAddress.Valid {
		log.IPAddress = &ipAddress.String
	}
	if imageSize.Valid {
		log.ImageSize = &imageSize.String
	}
	if mediaType.Valid {
		log.MediaType = &mediaType.String
	}
	if serviceTier.Valid {
		log.ServiceTier = &serviceTier.String
	}
	if reasoningEffort.Valid {
		log.ReasoningEffort = &reasoningEffort.String
	}
	if inboundEndpoint.Valid {
		log.InboundEndpoint = &inboundEndpoint.String
	}
	if upstreamEndpoint.Valid {
		log.UpstreamEndpoint = &upstreamEndpoint.String
	}
	if upstreamModel.Valid {
		log.UpstreamModel = &upstreamModel.String
	}

	return log, nil
}

func setToSlice(set map[int64]struct{}) []int64 {
	out := make([]int64, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	return out
}
