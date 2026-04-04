package service

import (
	"context"
	"time"
)

func (s *OpsAlertEvaluatorService) Start() {
	if s == nil {
		return
	}
	s.startOnce.Do(func() {
		if s.stopCh == nil {
			s.stopCh = make(chan struct{})
		}
		s.wg.Add(1)
		go s.run()
	})
}

func (s *OpsAlertEvaluatorService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		if s.stopCh != nil {
			close(s.stopCh)
		}
	})
	s.wg.Wait()
}

func (s *OpsAlertEvaluatorService) run() {
	defer s.wg.Done()

	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			interval := s.getInterval()
			s.evaluateOnce(interval)
			timer.Reset(interval)
		case <-s.stopCh:
			return
		}
	}
}

func (s *OpsAlertEvaluatorService) getInterval() time.Duration {
	interval := 60 * time.Second
	if s == nil || s.opsService == nil {
		return interval
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cfg, err := s.opsService.GetOpsAlertRuntimeSettings(ctx)
	if err != nil || cfg == nil {
		return interval
	}
	if cfg.EvaluationIntervalSeconds <= 0 || cfg.EvaluationIntervalSeconds < 1 {
		return interval
	}
	if cfg.EvaluationIntervalSeconds > int((24 * time.Hour).Seconds()) {
		return interval
	}
	return time.Duration(cfg.EvaluationIntervalSeconds) * time.Second
}
