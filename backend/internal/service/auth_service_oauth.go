package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	dbent "github.com/senran-N/sub2api/ent"
	infraerrors "github.com/senran-N/sub2api/internal/pkg/errors"
	"github.com/senran-N/sub2api/internal/pkg/logger"

	"github.com/golang-jwt/jwt/v5"
)

// LoginOrRegisterOAuth 用于第三方 OAuth/SSO 登录：
// - 如果邮箱已存在：直接登录（不需要本地密码）
// - 如果邮箱不存在：创建新用户并登录
func (s *AuthService) LoginOrRegisterOAuth(ctx context.Context, email, username string) (string, *User, error) {
	user, normalizedUsername, err := s.findOrCreateOAuthUser(ctx, email, username, "", false)
	if err != nil {
		return "", nil, err
	}
	if !user.IsActive() {
		return "", nil, ErrUserNotActive
	}

	s.ensureOAuthUsername(ctx, user, normalizedUsername)

	token, err := s.GenerateToken(user)
	if err != nil {
		return "", nil, fmt.Errorf("generate token: %w", err)
	}
	return token, user, nil
}

// LoginOrRegisterOAuthWithTokenPair 用于第三方 OAuth/SSO 登录，返回完整的 TokenPair。
// invitationCode 仅在邀请码注册模式下新用户注册时使用；已有账号登录时忽略。
func (s *AuthService) LoginOrRegisterOAuthWithTokenPair(ctx context.Context, email, username, invitationCode string) (*TokenPair, *User, error) {
	if s.refreshTokenCache == nil {
		return nil, nil, errors.New("refresh token cache not configured")
	}

	user, normalizedUsername, err := s.findOrCreateOAuthUser(ctx, email, username, invitationCode, true)
	if err != nil {
		return nil, nil, err
	}
	if !user.IsActive() {
		return nil, nil, ErrUserNotActive
	}

	s.ensureOAuthUsername(ctx, user, normalizedUsername)

	tokenPair, err := s.GenerateTokenPair(ctx, user, "")
	if err != nil {
		return nil, nil, fmt.Errorf("generate token pair: %w", err)
	}
	return tokenPair, user, nil
}

// pendingOAuthTokenTTL is the validity period for pending OAuth tokens.
const pendingOAuthTokenTTL = 10 * time.Minute

// pendingOAuthPurpose is the purpose claim value for pending OAuth registration tokens.
const pendingOAuthPurpose = "pending_oauth_registration"

type pendingOAuthClaims struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Purpose  string `json:"purpose"`
	jwt.RegisteredClaims
}

// CreatePendingOAuthToken generates a short-lived JWT that carries the OAuth identity
// while waiting for the user to supply an invitation code.
func (s *AuthService) CreatePendingOAuthToken(email, username string) (string, error) {
	now := time.Now()
	claims := &pendingOAuthClaims{
		Email:    email,
		Username: username,
		Purpose:  pendingOAuthPurpose,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(pendingOAuthTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.Secret))
}

// VerifyPendingOAuthToken validates a pending OAuth token and returns the embedded identity.
func (s *AuthService) VerifyPendingOAuthToken(tokenStr string) (email, username string, err error) {
	if len(tokenStr) > maxTokenLength {
		return "", "", ErrInvalidToken
	}
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
	token, parseErr := parser.ParseWithClaims(tokenStr, &pendingOAuthClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.cfg.JWT.Secret), nil
	})
	if parseErr != nil {
		return "", "", ErrInvalidToken
	}
	claims, ok := token.Claims.(*pendingOAuthClaims)
	if !ok || !token.Valid || claims.Purpose != pendingOAuthPurpose {
		return "", "", ErrInvalidToken
	}
	return claims.Email, claims.Username, nil
}

func (s *AuthService) findOrCreateOAuthUser(ctx context.Context, email, username, invitationCode string, requireInvitation bool) (*User, string, error) {
	normalizedEmail, normalizedUsername, err := normalizeOAuthIdentity(email, username)
	if err != nil {
		return nil, "", err
	}

	user, err := s.userRepo.GetByEmail(ctx, normalizedEmail)
	switch {
	case err == nil:
		return user, normalizedUsername, nil
	case !errors.Is(err, ErrUserNotFound):
		logger.LegacyPrintf("service.auth", "[Auth] Database error during oauth login: %v", err)
		return nil, "", ErrServiceUnavailable
	}

	if s.settingService == nil || !s.settingService.IsRegistrationEnabled(ctx) {
		return nil, "", ErrRegDisabled
	}

	invitationRedeemCode, err := s.resolveOAuthInvitationCode(ctx, invitationCode, requireInvitation)
	if err != nil {
		return nil, "", err
	}

	user, err = s.createOAuthUser(ctx, normalizedEmail, normalizedUsername, invitationRedeemCode)
	if err != nil {
		return nil, "", err
	}
	return user, normalizedUsername, nil
}

func normalizeOAuthIdentity(email, username string) (string, string, error) {
	normalizedEmail := strings.TrimSpace(email)
	if normalizedEmail == "" || len(normalizedEmail) > 255 {
		return "", "", infraerrors.BadRequest("INVALID_EMAIL", "invalid email")
	}
	if _, err := mail.ParseAddress(normalizedEmail); err != nil {
		return "", "", infraerrors.BadRequest("INVALID_EMAIL", "invalid email")
	}

	normalizedUsername := strings.TrimSpace(username)
	if len([]rune(normalizedUsername)) > 100 {
		normalizedUsername = string([]rune(normalizedUsername)[:100])
	}
	return normalizedEmail, normalizedUsername, nil
}

func (s *AuthService) resolveOAuthInvitationCode(ctx context.Context, invitationCode string, requireInvitation bool) (*RedeemCode, error) {
	if !requireInvitation || s.settingService == nil || !s.settingService.IsInvitationCodeEnabled(ctx) {
		return nil, nil
	}
	if invitationCode == "" {
		return nil, ErrOAuthInvitationRequired
	}

	redeemCode, err := s.redeemRepo.GetByCode(ctx, invitationCode)
	if err != nil {
		return nil, ErrInvitationCodeInvalid
	}
	if redeemCode.Type != RedeemTypeInvitation || redeemCode.Status != StatusUnused {
		return nil, ErrInvitationCodeInvalid
	}
	return redeemCode, nil
}

func (s *AuthService) createOAuthUser(ctx context.Context, email, username string, invitationRedeemCode *RedeemCode) (*User, error) {
	hashedPassword, err := s.generateOAuthPasswordHash()
	if err != nil {
		return nil, err
	}

	newUser := s.newOAuthUser(ctx, email, username, hashedPassword)
	if invitationRedeemCode != nil && s.entClient != nil {
		return s.createOAuthUserWithInvitationTx(ctx, email, newUser, invitationRedeemCode)
	}
	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return s.resolveOAuthCreateConflict(ctx, email, err)
	}

	s.assignDefaultSubscriptions(ctx, newUser.ID)
	if invitationRedeemCode != nil {
		if err := s.redeemRepo.Use(ctx, invitationRedeemCode.ID, newUser.ID); err != nil {
			return nil, ErrInvitationCodeInvalid
		}
	}
	return newUser, nil
}

func (s *AuthService) generateOAuthPasswordHash() (string, error) {
	randomPassword, err := randomHexString(32)
	if err != nil {
		logger.LegacyPrintf("service.auth", "[Auth] Failed to generate random password for oauth signup: %v", err)
		return "", ErrServiceUnavailable
	}
	hashedPassword, err := s.HashPassword(randomPassword)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return hashedPassword, nil
}

func (s *AuthService) newOAuthUser(ctx context.Context, email, username, hashedPassword string) *User {
	defaultBalance := s.cfg.Default.UserBalance
	defaultConcurrency := s.cfg.Default.UserConcurrency
	if s.settingService != nil {
		defaultBalance = s.settingService.GetDefaultBalance(ctx)
		defaultConcurrency = s.settingService.GetDefaultConcurrency(ctx)
	}

	return &User{
		Email:        email,
		Username:     username,
		PasswordHash: hashedPassword,
		Role:         RoleUser,
		Balance:      defaultBalance,
		Concurrency:  defaultConcurrency,
		Status:       StatusActive,
	}
}

func (s *AuthService) createOAuthUserWithInvitationTx(ctx context.Context, email string, newUser *User, invitationRedeemCode *RedeemCode) (*User, error) {
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		logger.LegacyPrintf("service.auth", "[Auth] Failed to begin transaction for oauth registration: %v", err)
		return nil, ErrServiceUnavailable
	}
	defer func() { _ = tx.Rollback() }()

	txCtx := dbent.NewTxContext(ctx, tx)
	if err := s.userRepo.Create(txCtx, newUser); err != nil {
		return s.resolveOAuthCreateConflict(ctx, email, err)
	}
	if err := s.redeemRepo.Use(txCtx, invitationRedeemCode.ID, newUser.ID); err != nil {
		return nil, ErrInvitationCodeInvalid
	}
	if err := tx.Commit(); err != nil {
		logger.LegacyPrintf("service.auth", "[Auth] Failed to commit oauth registration transaction: %v", err)
		return nil, ErrServiceUnavailable
	}

	s.assignDefaultSubscriptions(ctx, newUser.ID)
	return newUser, nil
}

func (s *AuthService) resolveOAuthCreateConflict(ctx context.Context, email string, createErr error) (*User, error) {
	if !errors.Is(createErr, ErrEmailExists) {
		logger.LegacyPrintf("service.auth", "[Auth] Database error creating oauth user: %v", createErr)
		return nil, ErrServiceUnavailable
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		logger.LegacyPrintf("service.auth", "[Auth] Database error getting user after conflict: %v", err)
		return nil, ErrServiceUnavailable
	}
	return user, nil
}

func (s *AuthService) ensureOAuthUsername(ctx context.Context, user *User, username string) {
	if user == nil || user.Username != "" || username == "" {
		return
	}

	user.Username = username
	if err := s.userRepo.Update(ctx, user); err != nil {
		logger.LegacyPrintf("service.auth", "[Auth] Failed to update username after oauth login: %v", err)
	}
}
