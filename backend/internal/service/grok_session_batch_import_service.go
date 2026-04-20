package service

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	infraerrors "github.com/senran-N/sub2api/internal/pkg/errors"
)

const (
	defaultGrokSessionBatchImportNamePrefix = "grok-sso"
	grokSessionBatchDedupeSkipExisting      = "skip_existing"
)

type GrokSessionBatchImportInput struct {
	RawInput        string
	NamePrefix      string
	GroupIDs        []int64
	ProxyID         *int64
	Priority        int
	Concurrency     int
	RateMultiplier  *float64
	LoadFactor      *int
	Notes           *string
	DedupeStrategy  string
	DryRun          bool
	TestAfterCreate bool
}

type GrokSessionBatchImportResult struct {
	Total             int                                `json:"total"`
	Created           int                                `json:"created"`
	Skipped           int                                `json:"skipped"`
	Invalid           int                                `json:"invalid"`
	DryRun            bool                               `json:"dry_run,omitempty"`
	Results           []GrokSessionBatchImportLineResult `json:"results"`
	CreatedAccountIDs []int64                            `json:"-"`
}

type GrokSessionBatchImportLineResult struct {
	Line        int    `json:"line"`
	Name        string `json:"name,omitempty"`
	Success     bool   `json:"success"`
	AccountID   int64  `json:"account_id,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
	Reason      string `json:"reason,omitempty"`
}

type grokSessionBatchImportLine struct {
	line int
	raw  string
}

func (s *adminServiceImpl) BatchImportGrokSessionAccounts(ctx context.Context, input *GrokSessionBatchImportInput) (*GrokSessionBatchImportResult, error) {
	if input == nil {
		return nil, infraerrors.BadRequest("GROK_SESSION_BATCH_IMPORT_INVALID", "import input is required")
	}
	if err := validateAccountRateMultiplier(input.RateMultiplier); err != nil {
		return nil, infraerrors.BadRequest("GROK_SESSION_BATCH_IMPORT_INVALID", err.Error())
	}

	dedupeStrategy := normalizeGrokSessionBatchDedupeStrategy(input.DedupeStrategy)
	if dedupeStrategy != grokSessionBatchDedupeSkipExisting {
		return nil, infraerrors.BadRequest("GROK_SESSION_BATCH_IMPORT_INVALID", "unsupported dedupe_strategy")
	}

	lines := parseGrokSessionBatchImportLines(input.RawInput)
	result := &GrokSessionBatchImportResult{
		Total:             len(lines),
		DryRun:            input.DryRun,
		Results:           make([]GrokSessionBatchImportLineResult, 0, len(lines)),
		CreatedAccountIDs: make([]int64, 0, len(lines)),
	}
	if len(lines) == 0 {
		return result, nil
	}

	existingFingerprints, err := s.listExistingGrokSessionFingerprints(ctx)
	if err != nil {
		return nil, err
	}
	seenFingerprints := make(map[string]struct{}, len(lines))
	namePrefix := normalizeGrokSessionBatchNamePrefix(input.NamePrefix)
	nextNameIndex := 1

	for _, line := range lines {
		normalizedCookie, err := ValidateGrokSessionImportToken(line.raw)
		if err != nil {
			result.Invalid++
			result.Results = append(result.Results, GrokSessionBatchImportLineResult{
				Line:    line.line,
				Success: false,
				Reason:  err.Error(),
			})
			continue
		}

		fingerprint := FingerprintGrokSessionToken(normalizedCookie)
		maskedFingerprint := MaskGrokSessionFingerprint(fingerprint)
		if _, exists := existingFingerprints[fingerprint]; exists {
			result.Skipped++
			result.Results = append(result.Results, GrokSessionBatchImportLineResult{
				Line:        line.line,
				Success:     false,
				Fingerprint: maskedFingerprint,
				Reason:      "existing session token fingerprint",
			})
			continue
		}
		if _, exists := seenFingerprints[fingerprint]; exists {
			result.Skipped++
			result.Results = append(result.Results, GrokSessionBatchImportLineResult{
				Line:        line.line,
				Success:     false,
				Fingerprint: maskedFingerprint,
				Reason:      "duplicate session token fingerprint in batch",
			})
			continue
		}

		accountName := fmt.Sprintf("%s-%03d", namePrefix, nextNameIndex)
		nextNameIndex++
		seenFingerprints[fingerprint] = struct{}{}
		existingFingerprints[fingerprint] = struct{}{}

		if input.DryRun {
			result.Created++
			result.Results = append(result.Results, GrokSessionBatchImportLineResult{
				Line:        line.line,
				Name:        accountName,
				Success:     true,
				Fingerprint: maskedFingerprint,
			})
			continue
		}

		account, err := s.CreateAccount(ctx, &CreateAccountInput{
			Name:           accountName,
			Notes:          input.Notes,
			Platform:       PlatformGrok,
			Type:           AccountTypeSession,
			Credentials:    map[string]any{"session_token": normalizedCookie},
			Extra:          buildGrokSessionBatchImportExtra(fingerprint),
			ProxyID:        input.ProxyID,
			Concurrency:    input.Concurrency,
			Priority:       input.Priority,
			RateMultiplier: input.RateMultiplier,
			LoadFactor:     input.LoadFactor,
			GroupIDs:       input.GroupIDs,
		})
		if err != nil {
			delete(existingFingerprints, fingerprint)
			result.Invalid++
			result.Results = append(result.Results, GrokSessionBatchImportLineResult{
				Line:        line.line,
				Name:        accountName,
				Success:     false,
				Fingerprint: maskedFingerprint,
				Reason:      strings.TrimSpace(err.Error()),
			})
			continue
		}

		result.Created++
		result.CreatedAccountIDs = append(result.CreatedAccountIDs, account.ID)
		result.Results = append(result.Results, GrokSessionBatchImportLineResult{
			Line:        line.line,
			Name:        accountName,
			Success:     true,
			AccountID:   account.ID,
			Fingerprint: maskedFingerprint,
		})
	}

	return result, nil
}

func (s *adminServiceImpl) listExistingGrokSessionFingerprints(ctx context.Context) (map[string]struct{}, error) {
	if s == nil || s.accountRepo == nil {
		return map[string]struct{}{}, nil
	}

	accounts, err := s.accountRepo.ListByPlatform(ctx, PlatformGrok)
	if err != nil {
		return nil, err
	}

	fingerprints := make(map[string]struct{}, len(accounts))
	for i := range accounts {
		account := &accounts[i]
		if account == nil || account.Type != AccountTypeSession {
			continue
		}

		if fingerprint := strings.TrimSpace(getStringFromMaps(account.grokExtraMap(), nil, "auth_fingerprint")); fingerprint != "" {
			fingerprints[fingerprint] = struct{}{}
		}

		normalizedCookie, err := ValidateGrokSessionImportToken(account.GetGrokSessionToken())
		if err != nil {
			continue
		}
		if fingerprint := FingerprintGrokSessionToken(normalizedCookie); fingerprint != "" {
			fingerprints[fingerprint] = struct{}{}
		}
	}
	return fingerprints, nil
}

func parseGrokSessionBatchImportLines(rawInput string) []grokSessionBatchImportLine {
	scanner := bufio.NewScanner(strings.NewReader(rawInput))
	scanner.Buffer(make([]byte, 0, 1024), 1024*1024)

	lines := make([]grokSessionBatchImportLine, 0)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" {
			continue
		}
		lines = append(lines, grokSessionBatchImportLine{
			line: lineNumber,
			raw:  raw,
		})
	}
	return lines
}

func normalizeGrokSessionBatchNamePrefix(raw string) string {
	if value := strings.TrimSpace(raw); value != "" {
		return value
	}
	return defaultGrokSessionBatchImportNamePrefix
}

func normalizeGrokSessionBatchDedupeStrategy(raw string) string {
	if value := strings.TrimSpace(raw); value != "" {
		return value
	}
	return grokSessionBatchDedupeSkipExisting
}

func buildGrokSessionBatchImportExtra(fingerprint string) map[string]any {
	return map[string]any{
		"grok": map[string]any{
			"auth_fingerprint": fingerprint,
		},
	}
}
