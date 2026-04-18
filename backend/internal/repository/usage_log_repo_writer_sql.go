package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/senran-N/sub2api/internal/service"
)

var usageLogInsertColumnNames = []string{
	"user_id",
	"api_key_id",
	"account_id",
	"request_id",
	"model",
	"requested_model",
	"upstream_model",
	"group_id",
	"subscription_id",
	"input_tokens",
	"output_tokens",
	"cache_creation_tokens",
	"cache_read_tokens",
	"cache_creation_5m_tokens",
	"cache_creation_1h_tokens",
	"image_output_tokens",
	"image_output_cost",
	"input_cost",
	"output_cost",
	"cache_creation_cost",
	"cache_read_cost",
	"total_cost",
	"actual_cost",
	"rate_multiplier",
	"account_rate_multiplier",
	"billing_type",
	"request_type",
	"stream",
	"openai_ws_mode",
	"duration_ms",
	"first_token_ms",
	"user_agent",
	"ip_address",
	"image_count",
	"image_size",
	"media_type",
	"service_tier",
	"reasoning_effort",
	"inbound_endpoint",
	"upstream_endpoint",
	"cache_ttl_overridden",
	"channel_id",
	"model_mapping_chain",
	"billing_tier",
	"billing_mode",
	"created_at",
}

var usageLogInsertColumnsSQL = strings.Join(usageLogInsertColumnNames, ",\n\t\t\t")

var usageLogSingleInsertReturningQuery = fmt.Sprintf(`
	INSERT INTO usage_logs (
		%s
	) VALUES (
		%s
	)
	ON CONFLICT (request_id, api_key_id) DO NOTHING
	RETURNING id, created_at
`, usageLogInsertColumnsSQL, buildSequentialPlaceholdersSQL(1, len(usageLogInsertColumnNames)))

var usageLogSingleInsertNoResultQuery = fmt.Sprintf(`
	INSERT INTO usage_logs (
		%s
	) VALUES (
		%s
	)
	ON CONFLICT (request_id, api_key_id) DO NOTHING
`, usageLogInsertColumnsSQL, buildSequentialPlaceholdersSQL(1, len(usageLogInsertColumnNames)))

type usageLogInsertPrepared struct {
	createdAt      time.Time
	requestID      string
	rateMultiplier float64
	requestType    int16
	args           []any
}

type usageLogBatchState struct {
	ID        int64
	CreatedAt time.Time
}

type usageLogBatchRow struct {
	RequestID string    `json:"request_id"`
	APIKeyID  int64     `json:"api_key_id"`
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Inserted  bool      `json:"inserted"`
}

func init() {
	if len(usageLogInsertColumnNames) != len(usageLogInsertArgTypes) {
		panic(fmt.Sprintf("usage log insert metadata mismatch: columns=%d arg_types=%d", len(usageLogInsertColumnNames), len(usageLogInsertArgTypes)))
	}
}

func buildSequentialPlaceholdersSQL(start, count int) string {
	placeholders := make([]string, 0, count)
	for i := 0; i < count; i++ {
		placeholders = append(placeholders, "$"+strconv.Itoa(start+i))
	}
	return strings.Join(placeholders, ", ")
}

func appendUsageLogPreparedCTEValueRow(builder *strings.Builder, args []any, argPos *int, prepared usageLogInsertPrepared, includeInputIndex bool, inputIndex int) []any {
	_, _ = builder.WriteString("(")
	if includeInputIndex {
		_, _ = builder.WriteString("$")
		_, _ = builder.WriteString(strconv.Itoa(*argPos))
		args = append(args, inputIndex)
		*argPos = *argPos + 1
		if len(prepared.args) > 0 {
			_, _ = builder.WriteString(",")
		}
	}
	for i, arg := range prepared.args {
		if i > 0 {
			_, _ = builder.WriteString(",")
		}
		_, _ = builder.WriteString("$")
		_, _ = builder.WriteString(strconv.Itoa(*argPos))
		if i < len(usageLogInsertArgTypes) {
			_, _ = builder.WriteString("::")
			_, _ = builder.WriteString(usageLogInsertArgTypes[i])
		}
		args = append(args, arg)
		*argPos = *argPos + 1
	}
	_, _ = builder.WriteString(")")
	return args
}

func buildUsageLogBatchInsertQuery(keys []string, preparedByKey map[string]usageLogInsertPrepared) (string, []any) {
	var query strings.Builder
	_, _ = query.WriteString(`
		WITH input (
			input_idx,
			`)
	_, _ = query.WriteString(usageLogInsertColumnsSQL)
	_, _ = query.WriteString(`
		) AS (VALUES `)

	args := make([]any, 0, len(keys)*(len(usageLogInsertColumnNames)+1))
	argPos := 1
	for idx, key := range keys {
		if idx > 0 {
			_, _ = query.WriteString(",")
		}
		args = appendUsageLogPreparedCTEValueRow(&query, args, &argPos, preparedByKey[key], true, idx)
	}
	_, _ = query.WriteString(`
		),
		inserted AS (
			INSERT INTO usage_logs (
				`)
	_, _ = query.WriteString(usageLogInsertColumnsSQL)
	_, _ = query.WriteString(`
			)
			SELECT
				`)
	_, _ = query.WriteString(usageLogInsertColumnsSQL)
	_, _ = query.WriteString(`
			FROM input
			ON CONFLICT (request_id, api_key_id) DO NOTHING
			RETURNING request_id, api_key_id, id, created_at
		),
		resolved AS (
			SELECT
				input.input_idx,
				input.request_id,
				input.api_key_id,
				COALESCE(inserted.id, existing.id) AS id,
				COALESCE(inserted.created_at, existing.created_at) AS created_at,
				(inserted.id IS NOT NULL) AS inserted
			FROM input
			LEFT JOIN inserted
				ON inserted.request_id = input.request_id
				AND inserted.api_key_id = input.api_key_id
			LEFT JOIN usage_logs existing
				ON existing.request_id = input.request_id
				AND existing.api_key_id = input.api_key_id
		)
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'request_id', resolved.request_id,
					'api_key_id', resolved.api_key_id,
					'id', resolved.id,
					'created_at', resolved.created_at,
					'inserted', resolved.inserted
				)
				ORDER BY resolved.input_idx
			),
			'[]'::json
		)
		FROM resolved
	`)
	return query.String(), args
}

func buildUsageLogBestEffortInsertQuery(preparedList []usageLogInsertPrepared) (string, []any) {
	var query strings.Builder
	_, _ = query.WriteString(`
		WITH input (
			`)
	_, _ = query.WriteString(usageLogInsertColumnsSQL)
	_, _ = query.WriteString(`
		) AS (VALUES `)

	args := make([]any, 0, len(preparedList)*len(usageLogInsertColumnNames))
	argPos := 1
	for idx, prepared := range preparedList {
		if idx > 0 {
			_, _ = query.WriteString(",")
		}
		args = appendUsageLogPreparedCTEValueRow(&query, args, &argPos, prepared, false, 0)
	}

	_, _ = query.WriteString(`
		)
		INSERT INTO usage_logs (
			`)
	_, _ = query.WriteString(usageLogInsertColumnsSQL)
	_, _ = query.WriteString(`
		)
		SELECT
			`)
	_, _ = query.WriteString(usageLogInsertColumnsSQL)
	_, _ = query.WriteString(`
		FROM input
		ON CONFLICT (request_id, api_key_id) DO NOTHING
	`)

	return query.String(), args
}

func execUsageLogInsertNoResult(ctx context.Context, sqlq sqlExecutor, prepared usageLogInsertPrepared) error {
	_, err := sqlq.ExecContext(ctx, usageLogSingleInsertNoResultQuery, prepared.args...)
	return err
}

func prepareUsageLogInsert(log *service.UsageLog) usageLogInsertPrepared {
	createdAt := log.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	requestID := strings.TrimSpace(log.RequestID)
	log.RequestID = requestID

	rateMultiplier := log.RateMultiplier
	log.SyncRequestTypeAndLegacyFields()
	requestType := int16(log.RequestType)

	groupID := nullInt64(log.GroupID)
	subscriptionID := nullInt64(log.SubscriptionID)
	duration := nullInt(log.DurationMs)
	firstToken := nullInt(log.FirstTokenMs)
	userAgent := nullString(log.UserAgent)
	ipAddress := nullString(log.IPAddress)
	imageSize := nullString(log.ImageSize)
	mediaType := nullString(log.MediaType)
	serviceTier := nullString(log.ServiceTier)
	reasoningEffort := nullString(log.ReasoningEffort)
	inboundEndpoint := nullString(log.InboundEndpoint)
	upstreamEndpoint := nullString(log.UpstreamEndpoint)
	channelID := nullInt64(log.ChannelID)
	modelMappingChain := nullString(log.ModelMappingChain)
	billingTier := nullString(log.BillingTier)
	billingMode := nullString(log.BillingMode)
	requestedModel := strings.TrimSpace(log.RequestedModel)
	if requestedModel == "" {
		requestedModel = strings.TrimSpace(log.Model)
	}
	upstreamModel := nullString(log.UpstreamModel)

	var requestIDArg any
	if requestID != "" {
		requestIDArg = requestID
	}

	return usageLogInsertPrepared{
		createdAt:      createdAt,
		requestID:      requestID,
		rateMultiplier: rateMultiplier,
		requestType:    requestType,
		args: []any{
			log.UserID,
			log.APIKeyID,
			log.AccountID,
			requestIDArg,
			log.Model,
			nullString(&requestedModel),
			upstreamModel,
			groupID,
			subscriptionID,
			log.InputTokens,
			log.OutputTokens,
			log.CacheCreationTokens,
			log.CacheReadTokens,
			log.CacheCreation5mTokens,
			log.CacheCreation1hTokens,
			log.ImageOutputTokens,
			log.ImageOutputCost,
			log.InputCost,
			log.OutputCost,
			log.CacheCreationCost,
			log.CacheReadCost,
			log.TotalCost,
			log.ActualCost,
			rateMultiplier,
			log.AccountRateMultiplier,
			log.BillingType,
			requestType,
			log.Stream,
			log.OpenAIWSMode,
			duration,
			firstToken,
			userAgent,
			ipAddress,
			log.ImageCount,
			imageSize,
			mediaType,
			serviceTier,
			reasoningEffort,
			inboundEndpoint,
			upstreamEndpoint,
			log.CacheTTLOverridden,
			channelID,
			modelMappingChain,
			billingTier,
			billingMode,
			createdAt,
		},
	}
}

func usageLogBatchKey(requestID string, apiKeyID int64) string {
	return requestID + "\x1f" + strconv.FormatInt(apiKeyID, 10)
}
