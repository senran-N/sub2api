package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type openAIWSClientPayload struct {
	payloadRaw         []byte
	rawForHash         []byte
	promptCacheKey     string
	previousResponseID string
	originalModel      string
	payloadBytes       int
	storeDisabled      bool
	payloadMeta        openAIWSIngressPayloadMeta
}

func applyOpenAIWSIngressPayloadMutation(current []byte, path string, value any) ([]byte, error) {
	next, err := sjson.SetBytes(current, path, value)
	if err == nil {
		return next, nil
	}

	// 仅在确实需要修改 payload 且 sjson 失败时，退回 map 路径确保兼容性。
	payload := make(map[string]any)
	if unmarshalErr := json.Unmarshal(current, &payload); unmarshalErr != nil {
		return nil, err
	}
	switch path {
	case "type", "model":
		payload[path] = value
	case "client_metadata." + openAIWSTurnMetadataHeader:
		setOpenAIWSTurnMetadata(payload, fmt.Sprintf("%v", value))
	default:
		return nil, err
	}
	rebuilt, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return nil, marshalErr
	}
	return rebuilt, nil
}

func parseOpenAIWSIngressClientPayload(c *gin.Context, account *Account, raw []byte) (openAIWSClientPayload, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return openAIWSClientPayload{}, NewOpenAIWSClientCloseError(websocket.StatusPolicyViolation, "empty websocket request payload", nil)
	}
	if !gjson.ValidBytes(trimmed) {
		return openAIWSClientPayload{}, NewOpenAIWSClientCloseError(websocket.StatusPolicyViolation, "invalid websocket request payload", errors.New("invalid json"))
	}

	values := gjson.GetManyBytes(trimmed, "type", "model", "prompt_cache_key", "previous_response_id")
	eventType := strings.TrimSpace(values[0].String())
	normalized := trimmed
	switch eventType {
	case "":
		eventType = "response.create"
		next, setErr := applyOpenAIWSIngressPayloadMutation(normalized, "type", eventType)
		if setErr != nil {
			return openAIWSClientPayload{}, NewOpenAIWSClientCloseError(websocket.StatusPolicyViolation, "invalid websocket request payload", setErr)
		}
		normalized = next
	case "response.create":
	case "response.append":
		return openAIWSClientPayload{}, NewOpenAIWSClientCloseError(
			websocket.StatusPolicyViolation,
			"response.append is not supported in ws v2; use response.create with previous_response_id",
			nil,
		)
	default:
		return openAIWSClientPayload{}, NewOpenAIWSClientCloseError(
			websocket.StatusPolicyViolation,
			fmt.Sprintf("unsupported websocket request type: %s", eventType),
			nil,
		)
	}

	originalModel := strings.TrimSpace(values[1].String())
	if originalModel == "" {
		return openAIWSClientPayload{}, NewOpenAIWSClientCloseError(
			websocket.StatusPolicyViolation,
			"model is required in response.create payload",
			nil,
		)
	}
	promptCacheKey := strings.TrimSpace(values[2].String())
	previousResponseID := strings.TrimSpace(values[3].String())
	previousResponseIDKind := ClassifyOpenAIPreviousResponseIDKind(previousResponseID)
	if previousResponseID != "" && previousResponseIDKind == OpenAIPreviousResponseIDKindMessageID {
		return openAIWSClientPayload{}, NewOpenAIWSClientCloseError(
			websocket.StatusPolicyViolation,
			"previous_response_id must be a response.id (resp_*), not a message id",
			nil,
		)
	}
	if turnMetadata := strings.TrimSpace(c.GetHeader(openAIWSTurnMetadataHeader)); turnMetadata != "" {
		next, setErr := applyOpenAIWSIngressPayloadMutation(normalized, "client_metadata."+openAIWSTurnMetadataHeader, turnMetadata)
		if setErr != nil {
			return openAIWSClientPayload{}, NewOpenAIWSClientCloseError(websocket.StatusPolicyViolation, "invalid websocket request payload", setErr)
		}
		normalized = next
	}
	mappedModel := resolveOpenAIForwardModel(account, originalModel, "")
	if normalizedModel := normalizeOpenAIModelForUpstream(account, mappedModel); normalizedModel != "" {
		mappedModel = normalizedModel
	}
	if mappedModel != originalModel {
		next, setErr := applyOpenAIWSIngressPayloadMutation(normalized, "model", mappedModel)
		if setErr != nil {
			return openAIWSClientPayload{}, NewOpenAIWSClientCloseError(websocket.StatusPolicyViolation, "invalid websocket request payload", setErr)
		}
		normalized = next
	}

	return openAIWSClientPayload{
		payloadRaw:         normalized,
		rawForHash:         trimmed,
		promptCacheKey:     promptCacheKey,
		previousResponseID: previousResponseID,
		originalModel:      originalModel,
		payloadBytes:       len(normalized),
	}, nil
}

func (s *OpenAIGatewayService) prepareOpenAIWSClientPayload(
	account *Account,
	payload openAIWSClientPayload,
) openAIWSClientPayload {
	if s == nil || len(payload.payloadRaw) == 0 {
		return payload
	}
	payload.storeDisabled = s.isOpenAIWSStoreDisabledInRequestRaw(payload.payloadRaw, account)
	payload.payloadMeta = s.buildOpenAIWSIngressPayloadMeta(payload.payloadRaw, account, payload.storeDisabled)
	return payload
}
