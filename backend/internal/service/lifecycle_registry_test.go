package service

import "testing"

type lifecycleProbe struct {
	started int
	stopped int
}

func (p *lifecycleProbe) Start() { p.started++ }
func (p *lifecycleProbe) Stop()  { p.stopped++ }

func TestManageStartStopLifecycle_RegistersStop(t *testing.T) {
	registry := NewLifecycleRegistry()
	probe := &lifecycleProbe{}

	got := manageStartStopLifecycle(registry, "probe", probe)
	if got != probe {
		t.Fatalf("manageStartStopLifecycle returned unexpected probe")
	}
	if probe.started != 1 {
		t.Fatalf("started = %d, want 1", probe.started)
	}

	entries := registry.Entries()
	if len(entries) != 1 {
		t.Fatalf("entries len = %d, want 1", len(entries))
	}
	if entries[0].Name != "probe" {
		t.Fatalf("entry name = %q, want probe", entries[0].Name)
	}

	entries[0].Stop()
	if probe.stopped != 1 {
		t.Fatalf("stopped = %d, want 1", probe.stopped)
	}
}

func TestManageStartStopLifecycle_IgnoresNilService(t *testing.T) {
	registry := NewLifecycleRegistry()
	var probe *lifecycleProbe

	got := manageStartStopLifecycle(registry, "probe", probe)
	if got != nil {
		t.Fatalf("got = %v, want nil", got)
	}
	if len(registry.Entries()) != 0 {
		t.Fatalf("entries len = %d, want 0", len(registry.Entries()))
	}
}
