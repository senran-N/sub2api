package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func (s *SoraGatewayService) pollImageTask(ctx context.Context, c *gin.Context, account *Account, taskID string, stream bool) ([]string, error) {
	interval := s.pollInterval()
	maxAttempts := s.pollMaxAttempts()
	lastPing := time.Now()
	for attempt := 0; attempt < maxAttempts; attempt++ {
		status, err := s.soraClient.GetImageTask(ctx, account, taskID)
		if err != nil {
			return nil, err
		}
		switch strings.ToLower(status.Status) {
		case "succeeded", "completed":
			return status.URLs, nil
		case "failed":
			if status.ErrorMsg != "" {
				return nil, errors.New(status.ErrorMsg)
			}
			return nil, errors.New("sora image generation failed")
		}
		if stream {
			s.maybeSendPing(c, &lastPing)
		}
		if err := sleepWithContext(ctx, interval); err != nil {
			return nil, err
		}
	}
	return nil, errors.New("sora image generation timeout")
}

func (s *SoraGatewayService) pollVideoTaskDetailed(ctx context.Context, c *gin.Context, account *Account, taskID string, stream bool) (*SoraVideoTaskStatus, error) {
	interval := s.pollInterval()
	maxAttempts := s.pollMaxAttempts()
	lastPing := time.Now()
	for attempt := 0; attempt < maxAttempts; attempt++ {
		status, err := s.soraClient.GetVideoTask(ctx, account, taskID)
		if err != nil {
			return nil, err
		}
		switch strings.ToLower(status.Status) {
		case "completed", "succeeded":
			return status, nil
		case "failed":
			if status.ErrorMsg != "" {
				return nil, errors.New(status.ErrorMsg)
			}
			return nil, errors.New("sora video generation failed")
		}
		if stream {
			s.maybeSendPing(c, &lastPing)
		}
		if err := sleepWithContext(ctx, interval); err != nil {
			return nil, err
		}
	}
	return nil, errors.New("sora video generation timeout")
}

func (s *SoraGatewayService) pollInterval() time.Duration {
	if s == nil || s.cfg == nil {
		return 2 * time.Second
	}
	interval := s.cfg.Sora.Client.PollIntervalSeconds
	if interval <= 0 {
		interval = 2
	}
	return time.Duration(interval) * time.Second
}

func (s *SoraGatewayService) pollMaxAttempts() int {
	if s == nil || s.cfg == nil {
		return 600
	}
	maxAttempts := s.cfg.Sora.Client.MaxPollAttempts
	if maxAttempts <= 0 {
		maxAttempts = 600
	}
	return maxAttempts
}

func (s *SoraGatewayService) maybeSendPing(c *gin.Context, lastPing *time.Time) {
	if c == nil {
		return
	}
	interval := 10 * time.Second
	if s != nil && s.cfg != nil && s.cfg.Concurrency.PingInterval > 0 {
		interval = time.Duration(s.cfg.Concurrency.PingInterval) * time.Second
	}
	if time.Since(*lastPing) < interval {
		return
	}
	if _, err := fmt.Fprint(c.Writer, ":\n\n"); err == nil {
		if flusher, ok := c.Writer.(http.Flusher); ok {
			flusher.Flush()
		}
		*lastPing = time.Now()
	}
}
