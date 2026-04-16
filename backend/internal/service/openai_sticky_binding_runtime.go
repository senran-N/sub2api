package service

import (
	"context"
	"strings"
	"sync/atomic"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/logger"
	"go.uber.org/zap"
)

type stickyBindingOutcome string

const (
	stickyBindingOutcomeUsable         stickyBindingOutcome = ""
	stickyBindingOutcomeSoftMiss       stickyBindingOutcome = "soft_miss"
	stickyBindingOutcomeHardInvalidate stickyBindingOutcome = "hard_invalidate"
)

type stickyBindingKind string

const (
	stickyBindingKindSession          stickyBindingKind = "session"
	stickyBindingKindPreviousResponse stickyBindingKind = "previous_response"
)

type stickyBindingDisposition struct {
	Outcome stickyBindingOutcome
	Reason  string
}

type openAIStickyBindingMetricsSnapshot struct {
	StickySoftMissTotal            int64 `json:"sticky_soft_miss_total"`
	StickyHardInvalidateTotal      int64 `json:"sticky_hard_invalidate_total"`
	StickyLookupMissTotal          int64 `json:"sticky_lookup_miss_total"`
	StickyTransportSoftMissTotal   int64 `json:"sticky_transport_soft_miss_total"`
	StickyTemporarySoftMissTotal   int64 `json:"sticky_temporary_soft_miss_total"`
	StickyModelInvalidateTotal     int64 `json:"sticky_model_invalidate_total"`
	PreviousSoftMissTotal          int64 `json:"previous_soft_miss_total"`
	PreviousHardInvalidateTotal    int64 `json:"previous_hard_invalidate_total"`
	PreviousLookupMissTotal        int64 `json:"previous_lookup_miss_total"`
	PreviousTransportSoftMissTotal int64 `json:"previous_transport_soft_miss_total"`
}

type openAIStickyBindingMetrics struct {
	stickySoftMissTotal            atomic.Int64
	stickyHardInvalidateTotal      atomic.Int64
	stickyLookupMissTotal          atomic.Int64
	stickyTransportSoftMissTotal   atomic.Int64
	stickyTemporarySoftMissTotal   atomic.Int64
	stickyModelInvalidateTotal     atomic.Int64
	previousSoftMissTotal          atomic.Int64
	previousHardInvalidateTotal    atomic.Int64
	previousLookupMissTotal        atomic.Int64
	previousTransportSoftMissTotal atomic.Int64
}

var defaultOpenAIStickyBindingMetrics openAIStickyBindingMetrics

func newStickyBindingHardInvalidate(reason string) stickyBindingDisposition {
	return stickyBindingDisposition{
		Outcome: stickyBindingOutcomeHardInvalidate,
		Reason:  reason,
	}
}

func newStickyBindingSoftMiss(reason string) stickyBindingDisposition {
	return stickyBindingDisposition{
		Outcome: stickyBindingOutcomeSoftMiss,
		Reason:  reason,
	}
}

func classifyStickyBindingDisposition(account *Account, requestedModel string) stickyBindingDisposition {
	if account == nil {
		return stickyBindingDisposition{
			Outcome: stickyBindingOutcomeSoftMiss,
			Reason:  "lookup_miss",
		}
	}

	switch account.Status {
	case StatusError:
		return stickyBindingDisposition{Outcome: stickyBindingOutcomeHardInvalidate, Reason: "status_error"}
	case StatusDisabled:
		return stickyBindingDisposition{Outcome: stickyBindingOutcomeHardInvalidate, Reason: "status_disabled"}
	}

	if !account.Schedulable {
		return stickyBindingDisposition{Outcome: stickyBindingOutcomeHardInvalidate, Reason: "schedulable_disabled"}
	}
	if account.AutoPauseOnExpired && account.ExpiresAt != nil && !time.Now().Before(*account.ExpiresAt) {
		return stickyBindingDisposition{Outcome: stickyBindingOutcomeHardInvalidate, Reason: "account_expired"}
	}
	if detail := oauthSelectionCredentialIssue(account); detail != "" {
		return stickyBindingDisposition{Outcome: stickyBindingOutcomeHardInvalidate, Reason: detail}
	}
	if account.TempUnschedulableUntil != nil && time.Now().Before(*account.TempUnschedulableUntil) {
		return stickyBindingDisposition{Outcome: stickyBindingOutcomeSoftMiss, Reason: "temp_unschedulable"}
	}
	if account.IsOverloaded() {
		return stickyBindingDisposition{Outcome: stickyBindingOutcomeSoftMiss, Reason: "account_overloaded"}
	}
	if account.IsRateLimited() {
		return stickyBindingDisposition{Outcome: stickyBindingOutcomeSoftMiss, Reason: "account_rate_limited"}
	}
	if remaining := account.GetRateLimitRemainingTimeWithContext(context.Background(), requestedModel); remaining > 0 {
		return stickyBindingDisposition{Outcome: stickyBindingOutcomeSoftMiss, Reason: "model_rate_limited"}
	}

	return stickyBindingDisposition{}
}

func shouldClearStickySession(account *Account, requestedModel string) bool {
	return classifyStickyBindingDisposition(account, requestedModel).Outcome == stickyBindingOutcomeHardInvalidate
}

func recordOpenAIStickyBindingDisposition(
	ctx context.Context,
	kind stickyBindingKind,
	disposition stickyBindingDisposition,
	accountID int64,
	sessionHash string,
	previousResponseID string,
) {
	if disposition.Outcome == stickyBindingOutcomeUsable {
		return
	}

	switch kind {
	case stickyBindingKindPreviousResponse:
		if disposition.Outcome == stickyBindingOutcomeHardInvalidate {
			defaultOpenAIStickyBindingMetrics.previousHardInvalidateTotal.Add(1)
		} else {
			defaultOpenAIStickyBindingMetrics.previousSoftMissTotal.Add(1)
		}
		switch disposition.Reason {
		case "lookup_miss", "lookup_error", "db_recheck_miss":
			defaultOpenAIStickyBindingMetrics.previousLookupMissTotal.Add(1)
		case "transport_cooling", "transport_incompatible":
			defaultOpenAIStickyBindingMetrics.previousTransportSoftMissTotal.Add(1)
		}
	default:
		if disposition.Outcome == stickyBindingOutcomeHardInvalidate {
			defaultOpenAIStickyBindingMetrics.stickyHardInvalidateTotal.Add(1)
		} else {
			defaultOpenAIStickyBindingMetrics.stickySoftMissTotal.Add(1)
		}
		switch disposition.Reason {
		case "lookup_miss", "lookup_error", "db_recheck_miss":
			defaultOpenAIStickyBindingMetrics.stickyLookupMissTotal.Add(1)
		case "transport_cooling", "transport_incompatible":
			defaultOpenAIStickyBindingMetrics.stickyTransportSoftMissTotal.Add(1)
		case "temp_unschedulable", "account_overloaded", "account_rate_limited", "model_rate_limited":
			defaultOpenAIStickyBindingMetrics.stickyTemporarySoftMissTotal.Add(1)
		case "model_unsupported", "platform_mismatch":
			defaultOpenAIStickyBindingMetrics.stickyModelInvalidateTotal.Add(1)
		}
	}

	fields := []zap.Field{
		zap.String("binding_kind", string(kind)),
		zap.String("outcome", string(disposition.Outcome)),
		zap.String("reason", disposition.Reason),
	}
	if accountID > 0 {
		fields = append(fields, zap.Int64("account_id", accountID))
	}
	if trimmed := shortSessionHash(sessionHash); trimmed != "" {
		fields = append(fields, zap.String("session_hash", trimmed))
	}
	if trimmed := shortOpenAIPreviousResponseID(previousResponseID); trimmed != "" {
		fields = append(fields, zap.String("previous_response_id", trimmed))
	}

	log := logger.FromContext(ctx).With(fields...)
	if disposition.Outcome == stickyBindingOutcomeHardInvalidate {
		log.Warn("openai.sticky_binding_hard_invalidate")
		return
	}
	log.Info("openai.sticky_binding_soft_miss")
}

func snapshotOpenAIStickyBindingMetrics() openAIStickyBindingMetricsSnapshot {
	return openAIStickyBindingMetricsSnapshot{
		StickySoftMissTotal:            defaultOpenAIStickyBindingMetrics.stickySoftMissTotal.Load(),
		StickyHardInvalidateTotal:      defaultOpenAIStickyBindingMetrics.stickyHardInvalidateTotal.Load(),
		StickyLookupMissTotal:          defaultOpenAIStickyBindingMetrics.stickyLookupMissTotal.Load(),
		StickyTransportSoftMissTotal:   defaultOpenAIStickyBindingMetrics.stickyTransportSoftMissTotal.Load(),
		StickyTemporarySoftMissTotal:   defaultOpenAIStickyBindingMetrics.stickyTemporarySoftMissTotal.Load(),
		StickyModelInvalidateTotal:     defaultOpenAIStickyBindingMetrics.stickyModelInvalidateTotal.Load(),
		PreviousSoftMissTotal:          defaultOpenAIStickyBindingMetrics.previousSoftMissTotal.Load(),
		PreviousHardInvalidateTotal:    defaultOpenAIStickyBindingMetrics.previousHardInvalidateTotal.Load(),
		PreviousLookupMissTotal:        defaultOpenAIStickyBindingMetrics.previousLookupMissTotal.Load(),
		PreviousTransportSoftMissTotal: defaultOpenAIStickyBindingMetrics.previousTransportSoftMissTotal.Load(),
	}
}

func shortOpenAIPreviousResponseID(value string) string {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) <= 16 {
		return trimmed
	}
	return trimmed[:16]
}
