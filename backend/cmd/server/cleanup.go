package main

import (
	"context"
	"io"
	"log"
	"reflect"
	"sync"
)

type cleanupStep struct {
	name string
	fn   func() error
}

func runCleanup(ctx context.Context, parallelSteps []cleanupStep, infraSteps []cleanupStep) {
	var wg sync.WaitGroup
	errCh := make(chan error, len(parallelSteps))

	for _, step := range parallelSteps {
		step := step
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := step.fn(); err != nil {
				errCh <- err
				log.Printf("Cleanup error [%s]: %v", step.name, err)
			}
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		log.Printf("Cleanup timeout after %s", ctx.Err())
	}

	close(errCh)

	for _, step := range infraSteps {
		if err := step.fn(); err != nil {
			log.Printf("Cleanup error [%s]: %v", step.name, err)
		}
	}
}

func stopStep(name string, svc interface{ Stop() }) cleanupStep {
	return cleanupStep{
		name: name,
		fn: func() error {
			if isNilValue(svc) {
				return nil
			}
			svc.Stop()
			return nil
		},
	}
}

func callbackStep(name string, fn func()) cleanupStep {
	return cleanupStep{
		name: name,
		fn: func() error {
			if fn != nil {
				fn()
			}
			return nil
		},
	}
}

func closeStep(name string, closer io.Closer) cleanupStep {
	return cleanupStep{
		name: name,
		fn: func() error {
			if isNilValue(closer) {
				return nil
			}
			return closer.Close()
		},
	}
}

func isNilValue(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}
