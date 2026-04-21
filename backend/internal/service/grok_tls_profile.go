package service

import "github.com/senran-N/sub2api/internal/pkg/tlsfingerprint"

func resolveGrokTLSProfile(account *Account, svc *TLSFingerprintProfileService) *tlsfingerprint.Profile {
	if account == nil || !account.IsTLSFingerprintEnabled() {
		return nil
	}
	if svc != nil {
		if profile := svc.ResolveTLSProfile(account); profile != nil {
			return profile
		}
	}
	if account.GetTLSFingerprintProfileID() > 0 {
		return nil
	}
	return &tlsfingerprint.Profile{Name: "Built-in Default (Node.js 24.x)"}
}
