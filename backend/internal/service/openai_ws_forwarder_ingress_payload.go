package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type openAIWSPayloadRewriteOptions struct {
	dropPreviousResponseID bool
	setPreviousResponseID  string
	setInput               bool
	input                  []json.RawMessage
}

func rewriteOpenAIWSPayload(payload []byte, opts openAIWSPayloadRewriteOptions) ([]byte, error) {
	if len(payload) == 0 {
		return payload, nil
	}
	normalizedPrevID := strings.TrimSpace(opts.setPreviousResponseID)
	if !opts.dropPreviousResponseID && normalizedPrevID == "" && !opts.setInput {
		return payload, nil
	}
	if updated, ok, err := tryRewriteOpenAIWSPayloadFastPath(payload, opts, normalizedPrevID); ok {
		return updated, err
	}

	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return nil, err
	}
	if decoded == nil {
		decoded = make(map[string]json.RawMessage)
	}

	if opts.dropPreviousResponseID {
		delete(decoded, "previous_response_id")
	}
	if normalizedPrevID != "" {
		prevRaw, err := json.Marshal(normalizedPrevID)
		if err != nil {
			return nil, err
		}
		decoded["previous_response_id"] = prevRaw
	}
	if opts.setInput {
		inputForMarshal := opts.input
		if inputForMarshal == nil {
			inputForMarshal = []json.RawMessage{}
		}
		inputRaw, err := json.Marshal(inputForMarshal)
		if err != nil {
			return nil, err
		}
		decoded["input"] = inputRaw
	}
	return json.Marshal(decoded)
}

func tryRewriteOpenAIWSPayloadFastPath(
	payload []byte,
	opts openAIWSPayloadRewriteOptions,
	normalizedPrevID string,
) ([]byte, bool, error) {
	updated := payload
	if opts.dropPreviousResponseID {
		next, _, err := dropPreviousResponseIDFromRawPayload(updated)
		if err != nil {
			return nil, false, nil
		}
		updated = next
	}
	if normalizedPrevID != "" {
		next, err := setPreviousResponseIDToRawPayload(updated, normalizedPrevID)
		if err != nil {
			return nil, false, nil
		}
		updated = next
	}
	if opts.setInput {
		next, err := setOpenAIWSPayloadInputSequence(updated, opts.input, true)
		if err != nil {
			return nil, false, nil
		}
		updated = next
	}
	return updated, true, nil
}

func dropPreviousResponseIDFromRawPayload(payload []byte) ([]byte, bool, error) {
	return dropPreviousResponseIDFromRawPayloadWithDeleteFn(payload, sjson.DeleteBytes)
}

func dropPreviousResponseIDFromRawPayloadWithDeleteFn(
	payload []byte,
	deleteFn func([]byte, string) ([]byte, error),
) ([]byte, bool, error) {
	if len(payload) == 0 {
		return payload, false, nil
	}
	if !gjson.GetBytes(payload, "previous_response_id").Exists() {
		return payload, false, nil
	}
	if deleteFn == nil {
		deleteFn = sjson.DeleteBytes
	}

	updated := payload
	for i := 0; i < openAIWSMaxPrevResponseIDDeletePasses &&
		gjson.GetBytes(updated, "previous_response_id").Exists(); i++ {
		next, err := deleteFn(updated, "previous_response_id")
		if err != nil {
			return payload, false, err
		}
		updated = next
	}
	return updated, !gjson.GetBytes(updated, "previous_response_id").Exists(), nil
}

func setPreviousResponseIDToRawPayload(payload []byte, previousResponseID string) ([]byte, error) {
	normalizedPrevID := strings.TrimSpace(previousResponseID)
	if len(payload) == 0 || normalizedPrevID == "" {
		return payload, nil
	}
	if bytes.Contains(payload, openAIWSIngressPayloadPreviousResponseIDKey) {
		currentPrevID := strings.TrimSpace(gjson.GetBytes(payload, "previous_response_id").String())
		if currentPrevID == normalizedPrevID {
			return payload, nil
		}
	}
	updated, err := sjson.SetBytes(payload, "previous_response_id", normalizedPrevID)
	if err == nil {
		return updated, nil
	}

	var reqBody map[string]any
	if unmarshalErr := json.Unmarshal(payload, &reqBody); unmarshalErr != nil {
		return nil, err
	}
	reqBody["previous_response_id"] = normalizedPrevID
	rebuilt, marshalErr := json.Marshal(reqBody)
	if marshalErr != nil {
		return nil, marshalErr
	}
	return rebuilt, nil
}

func shouldInferIngressFunctionCallOutputPreviousResponseID(
	storeDisabled bool,
	turn int,
	hasFunctionCallOutput bool,
	currentPreviousResponseID string,
	expectedPreviousResponseID string,
) bool {
	if !storeDisabled || turn <= 1 || !hasFunctionCallOutput {
		return false
	}
	if strings.TrimSpace(currentPreviousResponseID) != "" {
		return false
	}
	return strings.TrimSpace(expectedPreviousResponseID) != ""
}

func alignStoreDisabledPreviousResponseID(
	payload []byte,
	expectedPreviousResponseID string,
) ([]byte, bool, error) {
	if len(payload) == 0 {
		return payload, false, nil
	}
	expected := strings.TrimSpace(expectedPreviousResponseID)
	if expected == "" {
		return payload, false, nil
	}
	current := openAIWSPayloadStringFromRaw(payload, "previous_response_id")
	if current == "" || current == expected {
		return payload, false, nil
	}
	updated, err := rewriteOpenAIWSPayload(payload, openAIWSPayloadRewriteOptions{
		dropPreviousResponseID: true,
		setPreviousResponseID:  expected,
	})
	if err != nil {
		return payload, false, err
	}
	return updated, true, nil
}

func cloneOpenAIWSPayloadBytes(payload []byte) []byte {
	if len(payload) == 0 {
		return nil
	}
	cloned := make([]byte, len(payload))
	copy(cloned, payload)
	return cloned
}

func cloneOpenAIWSRawMessages(items []json.RawMessage) []json.RawMessage {
	if items == nil {
		return nil
	}
	cloned := make([]json.RawMessage, 0, len(items))
	for idx := range items {
		cloned = append(cloned, json.RawMessage(cloneOpenAIWSPayloadBytes(items[idx])))
	}
	return cloned
}

func normalizeOpenAIWSJSONForCompare(raw []byte) ([]byte, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return nil, errors.New("json is empty")
	}
	var decoded any
	if err := json.Unmarshal(trimmed, &decoded); err != nil {
		return nil, err
	}
	return json.Marshal(decoded)
}

func normalizeOpenAIWSJSONForCompareOrRaw(raw []byte) []byte {
	normalized, err := normalizeOpenAIWSJSONForCompare(raw)
	if err != nil {
		return bytes.TrimSpace(raw)
	}
	return normalized
}

func openAIWSRawJSONMessagesEqual(a, b json.RawMessage) bool {
	aTrimmed := bytes.TrimSpace(a)
	bTrimmed := bytes.TrimSpace(b)
	if bytes.Equal(aTrimmed, bTrimmed) {
		return true
	}
	return bytes.Equal(
		normalizeOpenAIWSJSONForCompareOrRaw(aTrimmed),
		normalizeOpenAIWSJSONForCompareOrRaw(bTrimmed),
	)
}

func normalizeOpenAIWSPayloadWithoutInputAndPreviousResponseID(payload []byte) ([]byte, error) {
	if len(payload) == 0 {
		return nil, errors.New("payload is empty")
	}
	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return nil, err
	}
	delete(decoded, "input")
	delete(decoded, "previous_response_id")
	return json.Marshal(decoded)
}

func openAIWSExtractNormalizedInputSequence(payload []byte) ([]json.RawMessage, bool, error) {
	if len(payload) == 0 {
		return nil, false, nil
	}
	if !gjson.ValidBytes(payload) {
		hasInputKey := bytes.Contains(payload, []byte(`"input"`))
		return nil, hasInputKey, errors.New("invalid websocket payload json")
	}
	inputValue := gjson.GetBytes(payload, "input")
	if !inputValue.Exists() {
		return nil, false, nil
	}
	if inputValue.Type == gjson.JSON {
		raw := strings.TrimSpace(inputValue.Raw)
		if strings.HasPrefix(raw, "[") {
			return openAIWSExtractJSONArrayItems(inputValue, raw)
		}
		return []json.RawMessage{cloneOpenAIWSRawString(raw)}, true, nil
	}
	if inputValue.Type == gjson.String {
		encoded, _ := json.Marshal(inputValue.String())
		return []json.RawMessage{encoded}, true, nil
	}
	return []json.RawMessage{cloneOpenAIWSRawString(inputValue.Raw)}, true, nil
}

func openAIWSExtractJSONArrayItems(inputValue gjson.Result, raw string) ([]json.RawMessage, bool, error) {
	items := make([]json.RawMessage, 0, 4)
	itemCount := 0
	inputValue.ForEach(func(_, item gjson.Result) bool {
		itemCount++
		itemRaw := strings.TrimSpace(item.Raw)
		if itemRaw == "" {
			items = nil
			itemCount = -1
			return false
		}
		items = append(items, cloneOpenAIWSRawString(itemRaw))
		return true
	})
	if itemCount == 0 {
		return []json.RawMessage{}, true, nil
	}
	if itemCount < 0 {
		var fallback []json.RawMessage
		if err := json.Unmarshal([]byte(raw), &fallback); err != nil {
			return nil, true, err
		}
		return fallback, true, nil
	}
	return items, true, nil
}

func cloneOpenAIWSRawString(raw string) json.RawMessage {
	if raw == "" {
		return json.RawMessage{}
	}
	cloned := make([]byte, len(raw))
	copy(cloned, raw)
	return json.RawMessage(cloned)
}

func openAIWSInputIsPrefixExtended(previousPayload, currentPayload []byte) (bool, error) {
	previousItems, previousExists, prevErr := openAIWSExtractNormalizedInputSequence(previousPayload)
	if prevErr != nil {
		return false, prevErr
	}
	currentItems, currentExists, currentErr := openAIWSExtractNormalizedInputSequence(currentPayload)
	if currentErr != nil {
		return false, currentErr
	}
	if !previousExists && !currentExists {
		return true, nil
	}
	if !previousExists {
		return len(currentItems) == 0, nil
	}
	if !currentExists {
		return len(previousItems) == 0, nil
	}
	if len(currentItems) < len(previousItems) {
		return false, nil
	}

	for idx := range previousItems {
		if !openAIWSRawJSONMessagesEqual(previousItems[idx], currentItems[idx]) {
			return false, nil
		}
	}
	return true, nil
}

func openAIWSRawItemsHasPrefix(items []json.RawMessage, prefix []json.RawMessage) bool {
	if len(prefix) == 0 {
		return true
	}
	if len(items) < len(prefix) {
		return false
	}
	for idx := range prefix {
		if !openAIWSRawJSONMessagesEqual(prefix[idx], items[idx]) {
			return false
		}
	}
	return true
}

func buildOpenAIWSReplayInputSequence(
	previousFullInput []json.RawMessage,
	previousFullInputExists bool,
	currentPayload []byte,
	hasPreviousResponseID bool,
) ([]json.RawMessage, bool, error) {
	currentItems, currentExists, currentErr := openAIWSExtractNormalizedInputSequence(currentPayload)
	if currentErr != nil {
		return nil, false, currentErr
	}
	if !hasPreviousResponseID {
		return currentItems, currentExists, nil
	}
	if !previousFullInputExists {
		return currentItems, currentExists, nil
	}
	if !currentExists || len(currentItems) == 0 {
		return previousFullInput, true, nil
	}
	if openAIWSRawItemsHasPrefix(currentItems, previousFullInput) {
		return currentItems, true, nil
	}
	// json.RawMessage values are treated as immutable after extraction, so merged
	// sequences can safely reference existing entries without deep-copying bytes.
	merged := make([]json.RawMessage, 0, len(previousFullInput)+len(currentItems))
	merged = append(merged, previousFullInput...)
	merged = append(merged, currentItems...)
	return merged, true, nil
}

func setOpenAIWSPayloadInputSequence(
	payload []byte,
	fullInput []json.RawMessage,
	fullInputExists bool,
) ([]byte, error) {
	if !fullInputExists {
		return payload, nil
	}
	// Preserve [] vs null semantics when input exists but is empty.
	inputForMarshal := fullInput
	if inputForMarshal == nil {
		inputForMarshal = []json.RawMessage{}
	}
	inputRaw := buildOpenAIWSRawJSONArray(inputForMarshal)
	return sjson.SetRawBytes(payload, "input", inputRaw)
}

func buildOpenAIWSRawJSONArray(items []json.RawMessage) []byte {
	if len(items) == 0 {
		return []byte("[]")
	}
	total := 2 + len(items) - 1
	for idx := range items {
		if len(items[idx]) == 0 {
			total += len("null")
			continue
		}
		total += len(items[idx])
	}
	out := make([]byte, 0, total)
	out = append(out, '[')
	for idx := range items {
		if idx > 0 {
			out = append(out, ',')
		}
		if len(items[idx]) == 0 {
			out = append(out, "null"...)
			continue
		}
		out = append(out, items[idx]...)
	}
	out = append(out, ']')
	return out
}

func shouldKeepIngressPreviousResponseID(
	previousPayload []byte,
	currentPayload []byte,
	lastTurnResponseID string,
	hasFunctionCallOutput bool,
) (bool, string, error) {
	if hasFunctionCallOutput {
		return true, "has_function_call_output", nil
	}
	currentPreviousResponseID := strings.TrimSpace(openAIWSPayloadStringFromRaw(currentPayload, "previous_response_id"))
	if currentPreviousResponseID == "" {
		return false, "missing_previous_response_id", nil
	}
	expectedPreviousResponseID := strings.TrimSpace(lastTurnResponseID)
	if expectedPreviousResponseID == "" {
		return false, "missing_last_turn_response_id", nil
	}
	if currentPreviousResponseID != expectedPreviousResponseID {
		return false, "previous_response_id_mismatch", nil
	}
	if len(previousPayload) == 0 {
		return false, "missing_previous_turn_payload", nil
	}

	previousComparable, previousComparableErr := normalizeOpenAIWSPayloadWithoutInputAndPreviousResponseID(previousPayload)
	if previousComparableErr != nil {
		return false, "non_input_compare_error", previousComparableErr
	}
	currentComparable, currentComparableErr := normalizeOpenAIWSPayloadWithoutInputAndPreviousResponseID(currentPayload)
	if currentComparableErr != nil {
		return false, "non_input_compare_error", currentComparableErr
	}
	if !bytes.Equal(previousComparable, currentComparable) {
		return false, "non_input_changed", nil
	}
	return true, "strict_incremental_ok", nil
}

type openAIWSIngressPreviousTurnStrictState struct {
	nonInputComparable []byte
}

func buildOpenAIWSIngressPreviousTurnStrictStateFromComparable(
	nonInputComparable []byte,
	nonInputErr error,
) (*openAIWSIngressPreviousTurnStrictState, error) {
	if nonInputErr != nil {
		return nil, nonInputErr
	}
	if len(nonInputComparable) == 0 {
		return nil, nil
	}
	return &openAIWSIngressPreviousTurnStrictState{
		nonInputComparable: nonInputComparable,
	}, nil
}

func buildOpenAIWSIngressPreviousTurnStrictState(payload []byte) (*openAIWSIngressPreviousTurnStrictState, error) {
	if len(payload) == 0 {
		return nil, nil
	}
	nonInputComparable, nonInputErr := normalizeOpenAIWSPayloadWithoutInputAndPreviousResponseID(payload)
	return buildOpenAIWSIngressPreviousTurnStrictStateFromComparable(nonInputComparable, nonInputErr)
}

func shouldKeepIngressPreviousResponseIDWithStrictState(
	previousState *openAIWSIngressPreviousTurnStrictState,
	currentPayload []byte,
	currentComparable []byte,
	currentComparableErr error,
	lastTurnResponseID string,
	hasFunctionCallOutput bool,
) (bool, string, error) {
	if hasFunctionCallOutput {
		return true, "has_function_call_output", nil
	}
	currentPreviousResponseID := strings.TrimSpace(openAIWSPayloadStringFromRaw(currentPayload, "previous_response_id"))
	if currentPreviousResponseID == "" {
		return false, "missing_previous_response_id", nil
	}
	expectedPreviousResponseID := strings.TrimSpace(lastTurnResponseID)
	if expectedPreviousResponseID == "" {
		return false, "missing_last_turn_response_id", nil
	}
	if currentPreviousResponseID != expectedPreviousResponseID {
		return false, "previous_response_id_mismatch", nil
	}
	if previousState == nil {
		return false, "missing_previous_turn_payload", nil
	}

	if currentComparableErr != nil {
		return false, "non_input_compare_error", currentComparableErr
	}
	if !bytes.Equal(previousState.nonInputComparable, currentComparable) {
		return false, "non_input_changed", nil
	}
	return true, "strict_incremental_ok", nil
}
