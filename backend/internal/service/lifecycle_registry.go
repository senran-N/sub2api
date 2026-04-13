package service

import (
	"reflect"
	"sync"
)

// LifecycleCleanupEntry is a named stop callback exported for app cleanup orchestration.
type LifecycleCleanupEntry struct {
	Name string
	Stop func()
}

// LifecycleRegistry records stoppable resources as providers construct/start them.
// It is intentionally stop-focused first so existing providers can migrate gradually
// without forcing every background worker into the same constructor signature.
type LifecycleRegistry struct {
	mu      sync.Mutex
	entries []LifecycleCleanupEntry
}

func NewLifecycleRegistry() *LifecycleRegistry {
	return &LifecycleRegistry{}
}

func (r *LifecycleRegistry) Register(name string, stop func()) {
	if r == nil || stop == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = append(r.entries, LifecycleCleanupEntry{
		Name: name,
		Stop: stop,
	})
}

func (r *LifecycleRegistry) Entries() []LifecycleCleanupEntry {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]LifecycleCleanupEntry, len(r.entries))
	copy(out, r.entries)
	return out
}

func manageStartStopLifecycle[T interface {
	Start()
	Stop()
}](registry *LifecycleRegistry, name string, svc T) T {
	if lifecycleIsNil(svc) {
		return svc
	}
	svc.Start()
	registry.Register(name, svc.Stop)
	return svc
}

func lifecycleIsNil(v any) bool {
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
