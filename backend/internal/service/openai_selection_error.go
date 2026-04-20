package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var ErrOpenAIRequestedModelUnavailable = errors.New("no available OpenAI accounts supporting model")

func newOpenAIRequestedModelUnavailableError(requestedModel string) error {
	model := strings.TrimSpace(requestedModel)
	if model == "" {
		return ErrOpenAIRequestedModelUnavailable
	}
	return fmt.Errorf("%w: %s", ErrOpenAIRequestedModelUnavailable, model)
}

func isOpenAIRequestedModelUnavailableError(err error) bool {
	return errors.Is(err, ErrOpenAIRequestedModelUnavailable)
}

func extractOpenAIRequestedModelUnavailable(err error) string {
	if err == nil {
		return ""
	}
	if !isOpenAIRequestedModelUnavailableError(err) {
		return ""
	}

	message := strings.TrimSpace(err.Error())
	if message == "" {
		return ""
	}

	const marker = "supporting model:"
	index := strings.LastIndex(strings.ToLower(message), marker)
	if index < 0 {
		return ""
	}
	model := strings.TrimSpace(message[index+len(marker):])
	model = strings.TrimSuffix(model, ")")
	return strings.TrimSpace(model)
}

func openAIRequestedModelAvailable(accounts []Account, requestedModel string) bool {
	return openAIRequestedModelAvailableForPlatformWithContext(context.TODO(), accounts, requestedModel, PlatformOpenAI)
}

func openAIRequestedModelAvailableForPlatform(accounts []Account, requestedModel string, platform string) bool {
	return openAIRequestedModelAvailableForPlatformWithContext(context.TODO(), accounts, requestedModel, platform)
}

func openAIRequestedModelAvailableForPlatformWithContext(
	ctx context.Context,
	accounts []Account,
	requestedModel string,
	platform string,
) bool {
	return compatibleRequestedModelAvailableForPlatformWithContext(ctx, accounts, requestedModel, platform)
}

func IsOpenAIRequestedModelUnavailableError(err error) bool {
	return isOpenAIRequestedModelUnavailableError(err)
}

func ExtractOpenAIRequestedModelUnavailable(err error) string {
	return extractOpenAIRequestedModelUnavailable(err)
}
