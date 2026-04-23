package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	infraerrors "github.com/senran-N/sub2api/internal/pkg/errors"
	"github.com/senran-N/sub2api/internal/pkg/response"
	middleware2 "github.com/senran-N/sub2api/internal/server/middleware"

	"github.com/gin-gonic/gin"
)

const (
	oauthBindUserCookieName = "oauth_bind_user"
	oauthBindUserCookiePath = "/api/v1/auth/oauth"
	oauthBindUserCookieTTL  = 10 * time.Minute
)

func normalizeOAuthIntent(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "login":
		return "login"
	case "bind", "bind_current_user":
		return "bind_current_user"
	default:
		return "login"
	}
}

func (h *AuthHandler) buildOAuthBindUserCookieFromContext(c *gin.Context) (string, error) {
	if value, err := readOAuthBindUserCookie(c); err == nil {
		if _, parseErr := parseOAuthBindUserCookieValue(value, h.oauthBindCookieSecret(), time.Now()); parseErr == nil {
			return value, nil
		}
	}

	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		return "", infraerrors.Unauthorized("UNAUTHORIZED", "authentication required")
	}
	return buildOAuthBindUserCookieValue(subject.UserID, h.oauthBindCookieSecret(), time.Now()), nil
}

func (h *AuthHandler) readOAuthBindUserIDFromCookie(c *gin.Context, cookieName string) (int64, error) {
	value, err := readCookieDecoded(c, cookieName)
	if err != nil {
		return 0, err
	}
	return parseOAuthBindUserCookieValue(value, h.oauthBindCookieSecret(), time.Now())
}

func (h *AuthHandler) oauthBindCookieSecret() string {
	if h != nil && h.cfg != nil && strings.TrimSpace(h.cfg.JWT.Secret) != "" {
		return h.cfg.JWT.Secret
	}
	return "sub2api-oauth-bind"
}

func buildOAuthBindUserCookieValue(userID int64, secret string, now time.Time) string {
	expiresAt := now.UTC().Add(oauthBindUserCookieTTL).Unix()
	payload := fmt.Sprintf("%d:%d", userID, expiresAt)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return base64.RawURLEncoding.EncodeToString([]byte(payload + ":" + signature))
}

func parseOAuthBindUserCookieValue(raw, secret string, now time.Time) (int64, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return 0, infraerrors.Unauthorized("INVALID_BIND_COOKIE", "invalid oauth bind cookie")
	}
	parts := strings.Split(string(decoded), ":")
	if len(parts) != 3 {
		return 0, infraerrors.Unauthorized("INVALID_BIND_COOKIE", "invalid oauth bind cookie")
	}
	payload := strings.Join(parts[:2], ":")
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return 0, infraerrors.Unauthorized("INVALID_BIND_COOKIE", "invalid oauth bind cookie")
	}
	userID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || userID <= 0 {
		return 0, infraerrors.Unauthorized("INVALID_BIND_COOKIE", "invalid oauth bind cookie")
	}
	expiresAt, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || now.UTC().Unix() > expiresAt {
		return 0, infraerrors.Unauthorized("BIND_COOKIE_EXPIRED", "oauth bind cookie expired")
	}
	return userID, nil
}

func setOAuthBindUserCookie(c *gin.Context, value string, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     oauthBindUserCookieName,
		Value:    encodeCookieValue(value),
		Path:     oauthBindUserCookiePath,
		MaxAge:   int(oauthBindUserCookieTTL.Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearOAuthBindUserCookie(c *gin.Context, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     oauthBindUserCookieName,
		Value:    "",
		Path:     oauthBindUserCookiePath,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func readOAuthBindUserCookie(c *gin.Context) (string, error) {
	return readCookieDecoded(c, oauthBindUserCookieName)
}

func (h *AuthHandler) PrepareOAuthBindAccessTokenCookie(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	setOAuthBindUserCookie(c, buildOAuthBindUserCookieValue(subject.UserID, h.oauthBindCookieSecret(), time.Now()), isRequestHTTPS(c))
	response.Success(c, gin.H{"message": "ok"})
}
