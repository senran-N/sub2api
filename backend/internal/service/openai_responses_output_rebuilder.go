package service

import (
	"encoding/json"
	"slices"

	"github.com/senran-N/sub2api/internal/pkg/apicompat"
	"github.com/tidwall/gjson"
)

type openAIResponsesOutputCollector struct {
	items                map[int]apicompat.ResponsesOutput
	textParts            map[int]map[int]apicompat.ResponsesContentPart
	textDeltas           map[int]map[int]string
	functionArgsByIndex  map[int]string
	reasoningTextByIndex map[int]map[int]string
}

func newOpenAIResponsesOutputCollector() *openAIResponsesOutputCollector {
	return &openAIResponsesOutputCollector{
		items:                make(map[int]apicompat.ResponsesOutput),
		textParts:            make(map[int]map[int]apicompat.ResponsesContentPart),
		textDeltas:           make(map[int]map[int]string),
		functionArgsByIndex:  make(map[int]string),
		reasoningTextByIndex: make(map[int]map[int]string),
	}
}

func (c *openAIResponsesOutputCollector) ConsumePayload(payload []byte) {
	if len(payload) == 0 || !gjson.ValidBytes(payload) {
		return
	}

	switch gjson.GetBytes(payload, "type").String() {
	case "response.output_item.added", "response.output_item.done":
		c.consumeOutputItem(payload)
	case "response.output_text.delta":
		c.consumeTextDelta(payload)
	case "response.output_text.done":
		c.consumeTextDone(payload)
	case "response.content_part.done":
		c.consumeContentPartDone(payload)
	case "response.function_call_arguments.delta":
		c.consumeFunctionArgumentsDelta(payload)
	case "response.function_call_arguments.done":
		c.consumeFunctionArgumentsDone(payload)
	case "response.reasoning_summary_text.delta":
		c.consumeReasoningSummaryDelta(payload)
	case "response.reasoning_summary_text.done":
		c.consumeReasoningSummaryDone(payload)
	}
}

func (c *openAIResponsesOutputCollector) RepairResponse(final *apicompat.ResponsesResponse) *apicompat.ResponsesResponse {
	if final == nil {
		return nil
	}
	if len(final.Output) > 0 {
		return final
	}

	output := c.BuildOutput()
	if len(output) == 0 {
		return final
	}

	cloned := *final
	cloned.Output = output
	return &cloned
}

func (c *openAIResponsesOutputCollector) BuildOutput() []apicompat.ResponsesOutput {
	if len(c.items) == 0 &&
		len(c.textParts) == 0 &&
		len(c.textDeltas) == 0 &&
		len(c.functionArgsByIndex) == 0 &&
		len(c.reasoningTextByIndex) == 0 {
		return nil
	}

	indexSet := map[int]struct{}{}
	for idx := range c.items {
		indexSet[idx] = struct{}{}
	}
	for idx := range c.textParts {
		indexSet[idx] = struct{}{}
	}
	for idx := range c.textDeltas {
		indexSet[idx] = struct{}{}
	}
	for idx := range c.functionArgsByIndex {
		indexSet[idx] = struct{}{}
	}
	for idx := range c.reasoningTextByIndex {
		indexSet[idx] = struct{}{}
	}

	indices := make([]int, 0, len(indexSet))
	for idx := range indexSet {
		indices = append(indices, idx)
	}
	slices.Sort(indices)

	output := make([]apicompat.ResponsesOutput, 0, len(indices))
	for _, idx := range indices {
		item := c.items[idx]

		if content := c.buildMessageContent(idx); len(content) > 0 {
			if item.Type == "" {
				item.Type = "message"
			}
			if item.Role == "" {
				item.Role = "assistant"
			}
			if item.Status == "" {
				item.Status = "completed"
			}
			if len(item.Content) == 0 {
				item.Content = content
			}
		}

		if summaries := c.buildReasoningSummary(idx); len(summaries) > 0 {
			if item.Type == "" {
				item.Type = "reasoning"
			}
			if len(item.Summary) == 0 {
				item.Summary = summaries
			}
		}

		if args := c.functionArgsByIndex[idx]; args != "" {
			if item.Type == "" {
				item.Type = "function_call"
			}
			if item.Arguments == "" {
				item.Arguments = args
			}
		}

		if item.Type == "" {
			continue
		}
		output = append(output, item)
	}

	return output
}

func (c *openAIResponsesOutputCollector) consumeOutputItem(payload []byte) {
	outputIndex := int(gjson.GetBytes(payload, "output_index").Int())
	itemRaw := gjson.GetBytes(payload, "item")
	if outputIndex < 0 || !itemRaw.Exists() || itemRaw.Raw == "" {
		return
	}

	var item apicompat.ResponsesOutput
	if err := json.Unmarshal([]byte(itemRaw.Raw), &item); err != nil {
		return
	}

	existing := c.items[outputIndex]
	c.items[outputIndex] = mergeResponsesOutput(existing, item)
}

func (c *openAIResponsesOutputCollector) consumeTextDelta(payload []byte) {
	outputIndex := int(gjson.GetBytes(payload, "output_index").Int())
	if outputIndex < 0 {
		outputIndex = 0
	}
	contentIndex := int(gjson.GetBytes(payload, "content_index").Int())
	c.ensureMessageItem(outputIndex, gjson.GetBytes(payload, "item_id").String())
	if _, ok := c.textDeltas[outputIndex]; !ok {
		c.textDeltas[outputIndex] = make(map[int]string)
	}
	c.textDeltas[outputIndex][contentIndex] += gjson.GetBytes(payload, "delta").String()
}

func (c *openAIResponsesOutputCollector) consumeTextDone(payload []byte) {
	outputIndex := int(gjson.GetBytes(payload, "output_index").Int())
	if outputIndex < 0 {
		outputIndex = 0
	}
	contentIndex := int(gjson.GetBytes(payload, "content_index").Int())
	c.ensureMessageItem(outputIndex, gjson.GetBytes(payload, "item_id").String())
	text := gjson.GetBytes(payload, "text").String()
	if text == "" {
		text = c.textDeltas[outputIndex][contentIndex]
	}
	if text == "" {
		return
	}
	c.setContentPart(outputIndex, contentIndex, apicompat.ResponsesContentPart{
		Type: "output_text",
		Text: text,
	})
}

func (c *openAIResponsesOutputCollector) consumeContentPartDone(payload []byte) {
	partRaw := gjson.GetBytes(payload, "part")
	if !partRaw.Exists() || partRaw.Raw == "" {
		return
	}

	var part apicompat.ResponsesContentPart
	if err := json.Unmarshal([]byte(partRaw.Raw), &part); err != nil {
		return
	}

	outputIndex := int(gjson.GetBytes(payload, "output_index").Int())
	if outputIndex < 0 {
		outputIndex = 0
	}
	contentIndex := int(gjson.GetBytes(payload, "content_index").Int())
	c.ensureMessageItem(outputIndex, gjson.GetBytes(payload, "item_id").String())
	if part.Type == "output_text" && part.Text == "" {
		part.Text = c.textDeltas[outputIndex][contentIndex]
	}
	c.setContentPart(outputIndex, contentIndex, part)
}

func (c *openAIResponsesOutputCollector) consumeFunctionArgumentsDelta(payload []byte) {
	outputIndex := int(gjson.GetBytes(payload, "output_index").Int())
	if outputIndex < 0 {
		outputIndex = 0
	}
	item := c.items[outputIndex]
	if item.Type == "" {
		item.Type = "function_call"
	}
	if item.CallID == "" {
		item.CallID = gjson.GetBytes(payload, "call_id").String()
	}
	if item.Name == "" {
		item.Name = gjson.GetBytes(payload, "name").String()
	}
	c.items[outputIndex] = item
	c.functionArgsByIndex[outputIndex] += gjson.GetBytes(payload, "delta").String()
}

func (c *openAIResponsesOutputCollector) consumeFunctionArgumentsDone(payload []byte) {
	outputIndex := int(gjson.GetBytes(payload, "output_index").Int())
	if outputIndex < 0 {
		outputIndex = 0
	}
	item := c.items[outputIndex]
	if item.Type == "" {
		item.Type = "function_call"
	}
	if item.CallID == "" {
		item.CallID = gjson.GetBytes(payload, "call_id").String()
	}
	if item.Name == "" {
		item.Name = gjson.GetBytes(payload, "name").String()
	}
	if item.Arguments == "" {
		item.Arguments = gjson.GetBytes(payload, "arguments").String()
		if item.Arguments == "" {
			item.Arguments = c.functionArgsByIndex[outputIndex]
		}
	}
	c.items[outputIndex] = item
}

func (c *openAIResponsesOutputCollector) consumeReasoningSummaryDelta(payload []byte) {
	outputIndex := int(gjson.GetBytes(payload, "output_index").Int())
	if outputIndex < 0 {
		outputIndex = 0
	}
	summaryIndex := int(gjson.GetBytes(payload, "summary_index").Int())
	item := c.items[outputIndex]
	if item.Type == "" {
		item.Type = "reasoning"
	}
	c.items[outputIndex] = item
	if _, ok := c.reasoningTextByIndex[outputIndex]; !ok {
		c.reasoningTextByIndex[outputIndex] = make(map[int]string)
	}
	c.reasoningTextByIndex[outputIndex][summaryIndex] += gjson.GetBytes(payload, "delta").String()
}

func (c *openAIResponsesOutputCollector) consumeReasoningSummaryDone(payload []byte) {
	outputIndex := int(gjson.GetBytes(payload, "output_index").Int())
	if outputIndex < 0 {
		outputIndex = 0
	}
	summaryIndex := int(gjson.GetBytes(payload, "summary_index").Int())
	item := c.items[outputIndex]
	if item.Type == "" {
		item.Type = "reasoning"
	}
	c.items[outputIndex] = item
	if _, ok := c.reasoningTextByIndex[outputIndex]; !ok {
		c.reasoningTextByIndex[outputIndex] = make(map[int]string)
	}
	text := gjson.GetBytes(payload, "text").String()
	if text == "" {
		text = c.reasoningTextByIndex[outputIndex][summaryIndex]
	}
	if text != "" {
		c.reasoningTextByIndex[outputIndex][summaryIndex] = text
	}
}

func (c *openAIResponsesOutputCollector) ensureMessageItem(outputIndex int, itemID string) {
	item := c.items[outputIndex]
	if item.Type == "" {
		item.Type = "message"
	}
	if item.Role == "" {
		item.Role = "assistant"
	}
	if item.Status == "" {
		item.Status = "completed"
	}
	if item.ID == "" && itemID != "" {
		item.ID = itemID
	}
	c.items[outputIndex] = item
}

func (c *openAIResponsesOutputCollector) setContentPart(outputIndex, contentIndex int, part apicompat.ResponsesContentPart) {
	if _, ok := c.textParts[outputIndex]; !ok {
		c.textParts[outputIndex] = make(map[int]apicompat.ResponsesContentPart)
	}
	c.textParts[outputIndex][contentIndex] = part
}

func (c *openAIResponsesOutputCollector) buildMessageContent(outputIndex int) []apicompat.ResponsesContentPart {
	partsByIndex, hasParts := c.textParts[outputIndex]
	deltasByIndex, hasDeltas := c.textDeltas[outputIndex]
	if !hasParts && !hasDeltas {
		return nil
	}

	indexSet := map[int]struct{}{}
	for idx := range partsByIndex {
		indexSet[idx] = struct{}{}
	}
	for idx := range deltasByIndex {
		indexSet[idx] = struct{}{}
	}

	indices := make([]int, 0, len(indexSet))
	for idx := range indexSet {
		indices = append(indices, idx)
	}
	slices.Sort(indices)

	parts := make([]apicompat.ResponsesContentPart, 0, len(indices))
	for _, idx := range indices {
		part := partsByIndex[idx]
		if part.Type == "" {
			part.Type = "output_text"
		}
		if part.Type == "output_text" && part.Text == "" {
			part.Text = deltasByIndex[idx]
		}
		if part.Type == "" {
			continue
		}
		parts = append(parts, part)
	}

	return parts
}

func (c *openAIResponsesOutputCollector) buildReasoningSummary(outputIndex int) []apicompat.ResponsesSummary {
	partsByIndex := c.reasoningTextByIndex[outputIndex]
	if len(partsByIndex) == 0 {
		return nil
	}

	indices := make([]int, 0, len(partsByIndex))
	for idx := range partsByIndex {
		indices = append(indices, idx)
	}
	slices.Sort(indices)

	summary := make([]apicompat.ResponsesSummary, 0, len(indices))
	for _, idx := range indices {
		text := partsByIndex[idx]
		if text == "" {
			continue
		}
		summary = append(summary, apicompat.ResponsesSummary{
			Type: "summary_text",
			Text: text,
		})
	}
	return summary
}

func mergeResponsesOutput(base, next apicompat.ResponsesOutput) apicompat.ResponsesOutput {
	if base.Type == "" {
		base.Type = next.Type
	}
	if base.ID == "" {
		base.ID = next.ID
	}
	if base.Role == "" {
		base.Role = next.Role
	}
	if len(base.Content) == 0 && len(next.Content) > 0 {
		base.Content = next.Content
	}
	if base.Status == "" {
		base.Status = next.Status
	}
	if base.EncryptedContent == "" {
		base.EncryptedContent = next.EncryptedContent
	}
	if len(base.Summary) == 0 && len(next.Summary) > 0 {
		base.Summary = next.Summary
	}
	if base.CallID == "" {
		base.CallID = next.CallID
	}
	if base.Name == "" {
		base.Name = next.Name
	}
	if base.Arguments == "" {
		base.Arguments = next.Arguments
	}
	if base.Action == nil && next.Action != nil {
		base.Action = next.Action
	}
	return base
}
