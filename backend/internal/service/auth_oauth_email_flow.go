package service

import (
	"context"
	"net/mail"
	"strings"
	"time"
)

func (s *AuthService) SendPendingOAuthVerifyCode(ctx context.Context, email string) (*SendVerifyCodeResult, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return nil, ErrEmailVerifyRequired
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, ErrEmailVerifyRequired
	}
	if isReservedEmail(email) {
		return nil, ErrEmailReserved
	}
	if s == nil || s.emailService == nil {
		return nil, ErrServiceUnavailable
	}

	siteName := "Sub2API"
	if s.settingService != nil {
		siteName = s.settingService.GetSiteName(ctx)
	}
	if err := s.emailService.SendVerifyCode(ctx, email, siteName); err != nil {
		return nil, err
	}
	return &SendVerifyCodeResult{Countdown: int(verifyCodeCooldown / time.Second)}, nil
}
