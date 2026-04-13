package main

import (
	"testing"
	"time"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestResolveShutdownTimeout(t *testing.T) {
	t.Run("defaults when config is nil", func(t *testing.T) {
		require.Equal(t, defaultShutdownTimeout, resolveShutdownTimeout(nil))
	})

	t.Run("defaults when configured timeout is non-positive", func(t *testing.T) {
		cfg := &config.Config{}
		require.Equal(t, defaultShutdownTimeout, resolveShutdownTimeout(cfg))
	})

	t.Run("uses configured timeout seconds", func(t *testing.T) {
		cfg := &config.Config{
			Server: config.ServerConfig{
				ShutdownTimeout: 60,
			},
		}

		require.Equal(t, 60*time.Second, resolveShutdownTimeout(cfg))
	})
}
