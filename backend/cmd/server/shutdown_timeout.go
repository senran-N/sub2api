package main

import (
	"time"

	"github.com/senran-N/sub2api/internal/config"
)

const defaultShutdownTimeout = 45 * time.Second

func resolveShutdownTimeout(cfg *config.Config) time.Duration {
	if cfg != nil && cfg.Server.ShutdownTimeout > 0 {
		return time.Duration(cfg.Server.ShutdownTimeout) * time.Second
	}

	return defaultShutdownTimeout
}
