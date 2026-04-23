package service

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

type verificationCacheStub struct {
	getVerificationCodeFn           func(ctx context.Context, email string) (*VerificationCodeData, error)
	setVerificationCodeFn           func(ctx context.Context, email string, data *VerificationCodeData, ttl time.Duration) error
	deleteVerificationCodeFn        func(ctx context.Context, email string) error
	getNotifyVerifyCodeFn           func(ctx context.Context, email string) (*VerificationCodeData, error)
	setNotifyVerifyCodeFn           func(ctx context.Context, email string, data *VerificationCodeData, ttl time.Duration) error
	deleteNotifyVerifyCodeFn        func(ctx context.Context, email string) error
	getPasswordResetTokenFn         func(ctx context.Context, email string) (*PasswordResetTokenData, error)
	setPasswordResetTokenFn         func(ctx context.Context, email string, data *PasswordResetTokenData, ttl time.Duration) error
	deletePasswordResetTokenFn      func(ctx context.Context, email string) error
	isPasswordResetCooldownFn       func(ctx context.Context, email string) bool
	setPasswordResetEmailCooldownFn func(ctx context.Context, email string, ttl time.Duration) error
}

func (s *verificationCacheStub) GetVerificationCode(ctx context.Context, email string) (*VerificationCodeData, error) {
	if s.getVerificationCodeFn == nil {
		return nil, errors.New("unexpected GetVerificationCode call")
	}
	return s.getVerificationCodeFn(ctx, email)
}

func (s *verificationCacheStub) SetVerificationCode(ctx context.Context, email string, data *VerificationCodeData, ttl time.Duration) error {
	if s.setVerificationCodeFn == nil {
		return errors.New("unexpected SetVerificationCode call")
	}
	return s.setVerificationCodeFn(ctx, email, data, ttl)
}

func (s *verificationCacheStub) DeleteVerificationCode(ctx context.Context, email string) error {
	if s.deleteVerificationCodeFn == nil {
		return nil
	}
	return s.deleteVerificationCodeFn(ctx, email)
}

func (s *verificationCacheStub) GetNotifyVerifyCode(ctx context.Context, email string) (*VerificationCodeData, error) {
	if s.getNotifyVerifyCodeFn == nil {
		return nil, errors.New("unexpected GetNotifyVerifyCode call")
	}
	return s.getNotifyVerifyCodeFn(ctx, email)
}

func (s *verificationCacheStub) SetNotifyVerifyCode(ctx context.Context, email string, data *VerificationCodeData, ttl time.Duration) error {
	if s.setNotifyVerifyCodeFn == nil {
		return errors.New("unexpected SetNotifyVerifyCode call")
	}
	return s.setNotifyVerifyCodeFn(ctx, email, data, ttl)
}

func (s *verificationCacheStub) DeleteNotifyVerifyCode(ctx context.Context, email string) error {
	if s.deleteNotifyVerifyCodeFn == nil {
		return nil
	}
	return s.deleteNotifyVerifyCodeFn(ctx, email)
}

func (s *verificationCacheStub) GetPasswordResetToken(ctx context.Context, email string) (*PasswordResetTokenData, error) {
	if s.getPasswordResetTokenFn == nil {
		return nil, errors.New("unexpected GetPasswordResetToken call")
	}
	return s.getPasswordResetTokenFn(ctx, email)
}

func (s *verificationCacheStub) SetPasswordResetToken(ctx context.Context, email string, data *PasswordResetTokenData, ttl time.Duration) error {
	if s.setPasswordResetTokenFn == nil {
		return errors.New("unexpected SetPasswordResetToken call")
	}
	return s.setPasswordResetTokenFn(ctx, email, data, ttl)
}

func (s *verificationCacheStub) DeletePasswordResetToken(ctx context.Context, email string) error {
	if s.deletePasswordResetTokenFn == nil {
		return nil
	}
	return s.deletePasswordResetTokenFn(ctx, email)
}

func (s *verificationCacheStub) IsPasswordResetEmailInCooldown(ctx context.Context, email string) bool {
	if s.isPasswordResetCooldownFn == nil {
		return false
	}
	return s.isPasswordResetCooldownFn(ctx, email)
}

func (s *verificationCacheStub) SetPasswordResetEmailCooldown(ctx context.Context, email string, ttl time.Duration) error {
	if s.setPasswordResetEmailCooldownFn == nil {
		return nil
	}
	return s.setPasswordResetEmailCooldownFn(ctx, email, ttl)
}

func (s *verificationCacheStub) IncrNotifyCodeUserRate(ctx context.Context, userID int64, window time.Duration) (int64, error) {
	return 0, nil
}

func (s *verificationCacheStub) GetNotifyCodeUserRate(ctx context.Context, userID int64) (int64, error) {
	return 0, nil
}

func TestEmailServiceVerifyCodePreservesRemainingTTL(t *testing.T) {
	now := time.Now()
	cached := &VerificationCodeData{
		Code:      "123456",
		Attempts:  0,
		CreatedAt: now.Add(-14 * time.Minute),
		ExpiresAt: now.Add(45 * time.Second),
	}

	var savedTTL time.Duration
	cache := &verificationCacheStub{
		getVerificationCodeFn: func(ctx context.Context, email string) (*VerificationCodeData, error) {
			return cached, nil
		},
		setVerificationCodeFn: func(ctx context.Context, email string, data *VerificationCodeData, ttl time.Duration) error {
			savedTTL = ttl
			return nil
		},
	}

	svc := &EmailService{cache: cache}
	err := svc.VerifyCode(context.Background(), "user@example.com", "000000")
	if !errors.Is(err, ErrInvalidVerifyCode) {
		t.Fatalf("expected ErrInvalidVerifyCode, got %v", err)
	}
	if savedTTL <= 0 || savedTTL > time.Minute {
		t.Fatalf("expected remaining ttl to be preserved, got %v", savedTTL)
	}
}

func TestEmailServiceVerifyCodeSupportsLegacyExpirylessCache(t *testing.T) {
	now := time.Now()
	cached := &VerificationCodeData{
		Code:      "123456",
		Attempts:  0,
		CreatedAt: now.Add(-14 * time.Minute),
	}

	var savedTTL time.Duration
	cache := &verificationCacheStub{
		getVerificationCodeFn: func(ctx context.Context, email string) (*VerificationCodeData, error) {
			return cached, nil
		},
		setVerificationCodeFn: func(ctx context.Context, email string, data *VerificationCodeData, ttl time.Duration) error {
			savedTTL = ttl
			return nil
		},
	}

	svc := &EmailService{cache: cache}
	err := svc.VerifyCode(context.Background(), "user@example.com", "000000")
	if !errors.Is(err, ErrInvalidVerifyCode) {
		t.Fatalf("expected ErrInvalidVerifyCode, got %v", err)
	}
	if savedTTL <= 0 || savedTTL > time.Minute+5*time.Second {
		t.Fatalf("expected ttl to fall back to legacy created_at window, got %v", savedTTL)
	}
}

func TestEmailServiceSendEmailWithConfigSanitizesHeaders(t *testing.T) {
	originalSMTPDial := smtpDialFunc
	originalSMTPTLSDial := smtpTLSDialFunc
	originalSMTPNewClient := smtpNewClientFunc
	defer func() {
		smtpDialFunc = originalSMTPDial
		smtpTLSDialFunc = originalSMTPTLSDial
		smtpNewClientFunc = originalSMTPNewClient
	}()

	client := &smtpClientStub{startTLSSupported: true}
	smtpDialFunc = func(addr string) (smtpClient, error) {
		return client, nil
	}
	smtpTLSDialFunc = func(network, addr string, config *tls.Config) (net.Conn, error) {
		return nil, errors.New("unexpected implicit TLS path")
	}
	smtpNewClientFunc = func(conn net.Conn, host string) (smtpClient, error) {
		return nil, errors.New("unexpected smtp.NewClient call")
	}

	svc := &EmailService{}
	err := svc.SendEmailWithConfig(&SMTPConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "demo",
		Password: "secret",
		From:     "from@example.com\r\nBcc:bad@example.com",
		FromName: "Sender\r\nCc:bad@example.com",
		UseTLS:   true,
	}, "to@example.com\r\nBcc:bad@example.com", "subject\r\nX-Injected: yes", "<p>body</p>")
	if err != nil {
		t.Fatalf("SendEmailWithConfig returned error: %v", err)
	}

	message := client.dataBuffer.String()
	if strings.Contains(message, "\r\nBcc:") || strings.Contains(message, "\r\nCc:") || strings.Contains(message, "\r\nX-Injected:") {
		t.Fatalf("expected header injection to be sanitized, got message: %q", message)
	}
	if !strings.Contains(message, "Subject: subjectX-Injected: yes") {
		t.Fatalf("expected sanitized subject in message, got: %q", message)
	}
}

var _ smtpClient = (*smtpClientStub)(nil)
var _ io.WriteCloser = nopWriteCloser{}
