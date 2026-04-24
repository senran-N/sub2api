//nolint:unused
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

func readJSONSetting[T any](
	ctx context.Context,
	repo SettingRepository,
	key string,
	getOperation string,
	defaultFactory func() *T,
	normalize func(*T),
) (*T, error) {
	value, err := repo.GetValue(ctx, key)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			settings := defaultFactory()
			normalizeJSONSetting(settings, normalize)
			return settings, nil
		}
		return nil, fmt.Errorf("get %s: %w", getOperation, err)
	}
	if strings.TrimSpace(value) == "" {
		settings := defaultFactory()
		normalizeJSONSetting(settings, normalize)
		return settings, nil
	}

	var settings T
	if err := json.Unmarshal([]byte(value), &settings); err != nil {
		fallback := defaultFactory()
		normalizeJSONSetting(fallback, normalize)
		return fallback, nil
	}

	normalizeJSONSetting(&settings, normalize)
	return &settings, nil
}

func normalizeJSONSetting[T any](settings *T, normalize func(*T)) {
	if settings != nil && normalize != nil {
		normalize(settings)
	}
}

func writeJSONSetting(ctx context.Context, repo SettingRepository, key, marshalOperation string, settings any) error {
	data, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", marshalOperation, err)
	}
	return repo.Set(ctx, key, string(data))
}
