package service

import (
	"strings"
	"testing"
)

func TestComputeAttributionFingerprintMatchesObservedAlgorithm(t *testing.T) {
	message := "abcdefghijklmnopqrstuvwxyz"

	if got := computeAttributionFingerprint(message, "2.1.87"); got != "449" {
		t.Fatalf("computeAttributionFingerprint(..., 2.1.87)=%q, want %q", got, "449")
	}
	if got := computeAttributionFingerprint(message, "2.1.88"); got != "13b" {
		t.Fatalf("computeAttributionFingerprint(..., 2.1.88)=%q, want %q", got, "13b")
	}
}

func TestBuildAttributionHeaderTextTracksRuntimeVersion(t *testing.T) {
	body := []byte(`{"messages":[{"role":"user","content":"abcdefghijklmnopqrstuvwxyz"}]}`)

	h1 := buildAttributionHeaderText(body, "claude-cli/2.1.87 (external, cli)")
	h2 := buildAttributionHeaderText(body, "claude-cli/2.1.88 (external, cli)")

	if h1 == "" || h2 == "" {
		t.Fatalf("expected non-empty attribution headers, got %q and %q", h1, h2)
	}
	if h1 == h2 {
		t.Fatalf("expected version drift to change attribution header, got %q", h1)
	}
	if !strings.Contains(h1, "cc_version=2.1.87.449") {
		t.Fatalf("expected runtime version in header, got %q", h1)
	}
	if !strings.Contains(h2, "cc_version=2.1.88.13b") {
		t.Fatalf("expected runtime version in header, got %q", h2)
	}
}
