package service

import (
	"sync"
	"time"
)

type concurrencySlotCleanupLifecycle struct {
	svc         *ConcurrencyService
	accountRepo AccountRepository
	interval    time.Duration
	stopCh      chan struct{}
	stopOnce    sync.Once
	wg          sync.WaitGroup
}

func newConcurrencySlotCleanupLifecycle(
	svc *ConcurrencyService,
	accountRepo AccountRepository,
	interval time.Duration,
) *concurrencySlotCleanupLifecycle {
	if svc == nil || svc.cache == nil || interval <= 0 {
		return nil
	}
	if accountRepo == nil {
		if _, ok := svc.cache.(accountSlotCleanupSweeper); !ok {
			return nil
		}
	}
	return &concurrencySlotCleanupLifecycle{
		svc:         svc,
		accountRepo: accountRepo,
		interval:    interval,
		stopCh:      make(chan struct{}),
	}
}

func (l *concurrencySlotCleanupLifecycle) Start() {
	if l == nil {
		return
	}
	l.wg.Add(1)
	go func() {
		defer l.wg.Done()
		ticker := time.NewTicker(l.interval)
		defer ticker.Stop()

		l.svc.runSlotCleanupWorkerOnce(l.accountRepo)
		for {
			select {
			case <-l.stopCh:
				return
			case <-ticker.C:
				l.svc.runSlotCleanupWorkerOnce(l.accountRepo)
			}
		}
	}()
}

func (l *concurrencySlotCleanupLifecycle) Stop() {
	if l == nil {
		return
	}
	l.stopOnce.Do(func() {
		close(l.stopCh)
	})
	l.wg.Wait()
}
