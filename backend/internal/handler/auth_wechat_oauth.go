package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/payment"
	infraerrors "github.com/senran-N/sub2api/internal/pkg/errors"
	"github.com/senran-N/sub2api/internal/pkg/oauth"
	"github.com/senran-N/sub2api/internal/pkg/response"
	"github.com/senran-N/sub2api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/imroc/req/v3"
)

const (
	wechatOAuthCookiePath         = "/api/v1/auth/oauth/wechat"
	wechatOAuthCookieMaxAgeSec    = 10 * 60
	wechatOAuthStateCookieName    = "wechat_oauth_state"
	wechatOAuthRedirectCookieName = "wechat_oauth_redirect"
	wechatOAuthIntentCookieName   = "wechat_oauth_intent"
	wechatOAuthModeCookieName     = "wechat_oauth_mode"
	wechatOAuthBindUserCookieName = "wechat_oauth_bind_user"
	wechatOAuthDefaultRedirectTo  = "/dashboard"
	wechatOAuthDefaultFrontendCB  = "/auth/wechat/callback"
	wechatOAuthProviderKey        = "wechat-main"
	wechatPaymentOAuthCookiePath  = "/api/v1/auth/oauth/wechat/payment"
	wechatPaymentOAuthStateName   = "wechat_payment_oauth_state"
	wechatPaymentOAuthRedirect    = "wechat_payment_oauth_redirect"
	wechatPaymentOAuthContextName = "wechat_payment_oauth_context"
	wechatPaymentOAuthScopeName   = "wechat_payment_oauth_scope"
	wechatPaymentOAuthDefaultTo   = "/purchase"
	wechatPaymentOAuthFrontendCB  = "/auth/wechat/payment/callback"
)

var (
	wechatOAuthAccessTokenURL = "https://api.weixin.qq.com/sns/oauth2/access_token"
	wechatOAuthUserInfoURL    = "https://api.weixin.qq.com/sns/userinfo"
)

type wechatOAuthRuntimeConfig struct {
	mode             string
	appID            string
	appSecret        string
	scope            string
	redirectURI      string
	frontendCallback string
	openEnabled      bool
	mpEnabled        bool
	mobileEnabled    bool
}

type wechatOAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionID      string `json:"unionid"`
	ErrCode      int64  `json:"errcode"`
	ErrMsg       string `json:"errmsg"`
}

type wechatOAuthUserInfoResponse struct {
	OpenID     string `json:"openid"`
	Nickname   string `json:"nickname"`
	HeadImgURL string `json:"headimgurl"`
	UnionID    string `json:"unionid"`
	ErrCode    int64  `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

type wechatPaymentOAuthContext struct {
	PaymentType string `json:"payment_type"`
	Amount      string `json:"amount,omitempty"`
	OrderType   string `json:"order_type,omitempty"`
	PlanID      int64  `json:"plan_id,omitempty"`
}

func (h *AuthHandler) WeChatOAuthStart(c *gin.Context) {
	cfg, err := h.getWeChatOAuthConfig(c.Request.Context(), c.Query("mode"))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if cfg.mode == "mobile" {
		response.BadRequest(c, "wechat mobile app oauth cannot be started from browser")
		return
	}

	state, err := oauth.GenerateState()
	if err != nil {
		response.ErrorFrom(c, infraerrors.InternalServer("OAUTH_STATE_GEN_FAILED", "failed to generate oauth state").WithCause(err))
		return
	}
	redirectTo := sanitizeFrontendRedirectPath(c.Query("redirect"))
	if redirectTo == "" {
		redirectTo = wechatOAuthDefaultRedirectTo
	}
	browserSessionKey, err := generateOAuthPendingBrowserSession()
	if err != nil {
		response.ErrorFrom(c, infraerrors.InternalServer("OAUTH_BROWSER_SESSION_GEN_FAILED", "failed to generate oauth browser session").WithCause(err))
		return
	}

	intent := normalizeOAuthIntent(c.Query("intent"))
	secureCookie := isRequestHTTPS(c)
	wechatSetCookie(c, wechatOAuthStateCookieName, encodeCookieValue(state), wechatOAuthCookieMaxAgeSec, secureCookie)
	wechatSetCookie(c, wechatOAuthRedirectCookieName, encodeCookieValue(redirectTo), wechatOAuthCookieMaxAgeSec, secureCookie)
	wechatSetCookie(c, wechatOAuthIntentCookieName, encodeCookieValue(intent), wechatOAuthCookieMaxAgeSec, secureCookie)
	wechatSetCookie(c, wechatOAuthModeCookieName, encodeCookieValue(cfg.mode), wechatOAuthCookieMaxAgeSec, secureCookie)
	setOAuthPendingBrowserCookie(c, browserSessionKey, secureCookie)
	clearOAuthPendingSessionCookie(c, secureCookie)
	if intent == oauthIntentBindCurrentUser {
		bindCookieValue, bindErr := h.buildOAuthBindUserCookieFromContext(c)
		if bindErr != nil {
			response.ErrorFrom(c, bindErr)
			return
		}
		wechatSetCookie(c, wechatOAuthBindUserCookieName, encodeCookieValue(bindCookieValue), wechatOAuthCookieMaxAgeSec, secureCookie)
	} else {
		wechatClearCookie(c, wechatOAuthBindUserCookieName, secureCookie)
	}

	authURL, err := buildWeChatAuthorizeURL(cfg, state)
	if err != nil {
		response.ErrorFrom(c, infraerrors.InternalServer("OAUTH_BUILD_URL_FAILED", "failed to build oauth authorization url").WithCause(err))
		return
	}
	c.Redirect(http.StatusFound, authURL)
}

func (h *AuthHandler) WeChatOAuthCallback(c *gin.Context) {
	mode, _ := readCookieDecoded(c, wechatOAuthModeCookieName)
	cfg, err := h.getWeChatOAuthConfig(c.Request.Context(), mode)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	frontendCallback := strings.TrimSpace(cfg.frontendCallback)
	if frontendCallback == "" {
		frontendCallback = wechatOAuthDefaultFrontendCB
	}

	if providerErr := strings.TrimSpace(c.Query("error")); providerErr != "" {
		redirectOAuthError(c, frontendCallback, "provider_error", providerErr, c.Query("error_description"))
		return
	}

	code := strings.TrimSpace(c.Query("code"))
	state := strings.TrimSpace(c.Query("state"))
	if code == "" || state == "" {
		redirectOAuthError(c, frontendCallback, "missing_params", "missing code/state", "")
		return
	}

	expectedState, err := readCookieDecoded(c, wechatOAuthStateCookieName)
	if err != nil || !strings.EqualFold(strings.TrimSpace(expectedState), state) {
		redirectOAuthError(c, frontendCallback, "invalid_state", "invalid oauth state", "")
		return
	}
	redirectTo, _ := readCookieDecoded(c, wechatOAuthRedirectCookieName)
	redirectTo = sanitizeFrontendRedirectPath(redirectTo)
	if redirectTo == "" {
		redirectTo = wechatOAuthDefaultRedirectTo
	}
	browserSessionKey, err := readOAuthPendingBrowserCookie(c)
	if err != nil || strings.TrimSpace(browserSessionKey) == "" {
		redirectOAuthError(c, frontendCallback, "missing_browser_session", "missing oauth browser session", "")
		return
	}
	intent, _ := readCookieDecoded(c, wechatOAuthIntentCookieName)
	intent = normalizeOAuthIntent(intent)

	secureCookie := isRequestHTTPS(c)
	wechatClearCookie(c, wechatOAuthStateCookieName, secureCookie)
	wechatClearCookie(c, wechatOAuthRedirectCookieName, secureCookie)
	wechatClearCookie(c, wechatOAuthIntentCookieName, secureCookie)
	wechatClearCookie(c, wechatOAuthModeCookieName, secureCookie)
	wechatClearCookie(c, wechatOAuthBindUserCookieName, secureCookie)

	tokenResp, err := exchangeWeChatOAuthCode(c.Request.Context(), cfg, code)
	if err != nil {
		redirectOAuthError(c, frontendCallback, "token_exchange_failed", "failed to exchange oauth code", singleLine(err.Error()))
		return
	}

	userInfo, _ := fetchWeChatOAuthUserInfo(c.Request.Context(), tokenResp)
	unionID := strings.TrimSpace(firstNonEmpty(userInfo.UnionID, tokenResp.UnionID))
	openID := strings.TrimSpace(firstNonEmpty(userInfo.OpenID, tokenResp.OpenID))
	providerSubject := strings.TrimSpace(firstNonEmpty(unionID, openID))
	if providerSubject == "" {
		redirectOAuthError(c, frontendCallback, "missing_subject", "missing wechat subject", "")
		return
	}

	providerKey := wechatOAuthProviderKey
	if unionID == "" {
		providerKey = "wechat-" + cfg.mode
	}
	username := strings.TrimSpace(firstNonEmpty(userInfo.Nickname, wechatFallbackUsername(providerSubject)))
	identityRef := service.PendingAuthIdentityKey{
		ProviderType:    "wechat",
		ProviderKey:     providerKey,
		ProviderSubject: providerSubject,
	}
	upstreamClaims := map[string]any{
		"email":                  wechatSyntheticEmail(providerSubject),
		"username":               username,
		"subject":                providerSubject,
		"openid":                 openID,
		"unionid":                unionID,
		"mode":                   cfg.mode,
		"suggested_display_name": strings.TrimSpace(userInfo.Nickname),
		"suggested_avatar_url":   strings.TrimSpace(userInfo.HeadImgURL),
	}

	if intent == oauthIntentBindCurrentUser {
		targetUserID, bindErr := h.readOAuthBindUserIDFromCookie(c, wechatOAuthBindUserCookieName)
		if bindErr != nil {
			redirectOAuthError(c, frontendCallback, "invalid_state", "invalid oauth bind target", "")
			return
		}
		if err := h.createOAuthPendingSession(c, oauthPendingSessionPayload{
			Intent:                 oauthIntentBindCurrentUser,
			Identity:               identityRef,
			TargetUserID:           &targetUserID,
			ResolvedEmail:          "",
			RedirectTo:             redirectTo,
			BrowserSessionKey:      browserSessionKey,
			UpstreamIdentityClaims: upstreamClaims,
			CompletionResponse: map[string]any{
				"redirect": redirectTo,
			},
		}); err != nil {
			redirectOAuthError(c, frontendCallback, "session_error", "failed to continue oauth bind", "")
			return
		}
		redirectToFrontendCallback(c, frontendCallback)
		return
	}

	existingIdentityUser, err := h.findOAuthIdentityUser(c.Request.Context(), identityRef)
	if err != nil {
		redirectOAuthError(c, frontendCallback, "session_error", infraerrors.Reason(err), infraerrors.Message(err))
		return
	}
	if existingIdentityUser != nil {
		serviceUser, getErr := h.userService.GetByID(c.Request.Context(), existingIdentityUser.ID)
		if getErr != nil {
			redirectOAuthError(c, frontendCallback, "login_failed", infraerrors.Reason(getErr), infraerrors.Message(getErr))
			return
		}
		if backendModeBlocksLogin(h.settingSvc.IsBackendModeEnabled(c.Request.Context()), serviceUser) {
			redirectOAuthError(c, frontendCallback, "login_failed", "backend_mode_forbidden", "")
			return
		}
		tokenPair, tokenErr := h.authService.GenerateTokenPair(c.Request.Context(), serviceUser, "")
		if tokenErr != nil {
			redirectOAuthError(c, frontendCallback, "login_failed", "token_generation_failed", "")
			return
		}
		_ = h.authService.BindOAuthIdentity(c.Request.Context(), serviceUser.ID, "wechat", providerKey, providerSubject, map[string]any{
			"openid":   openID,
			"unionid":  unionID,
			"mode":     cfg.mode,
			"username": username,
			"source":   "wechat_oauth_identity_login",
		})
		fragment := url.Values{}
		fragment.Set("access_token", tokenPair.AccessToken)
		fragment.Set("refresh_token", tokenPair.RefreshToken)
		fragment.Set("expires_in", fmt.Sprintf("%d", tokenPair.ExpiresIn))
		fragment.Set("token_type", "Bearer")
		fragment.Set("redirect", redirectTo)
		redirectWithFragment(c, frontendCallback, fragment)
		return
	}

	if err := h.createOAuthPendingSession(c, oauthPendingSessionPayload{
		Intent:                 oauthIntentLogin,
		Identity:               identityRef,
		ResolvedEmail:          "",
		RedirectTo:             redirectTo,
		BrowserSessionKey:      browserSessionKey,
		UpstreamIdentityClaims: upstreamClaims,
		CompletionResponse: map[string]any{
			"step":                      oauthPendingChoiceStep,
			"adoption_required":         true,
			"redirect":                  redirectTo,
			"email":                     "",
			"resolved_email":            "",
			"existing_account_email":    "",
			"existing_account_bindable": true,
			"create_account_allowed":    true,
			"choice_reason":             "third_party_signup",
		},
	}); err != nil {
		redirectOAuthError(c, frontendCallback, "session_error", "failed to continue oauth login", "")
		return
	}
	redirectToFrontendCallback(c, frontendCallback)
}

// WeChatPaymentOAuthStart starts the WeChat payment OAuth flow.
// GET /api/v1/auth/oauth/wechat/payment/start
func (h *AuthHandler) WeChatPaymentOAuthStart(c *gin.Context) {
	cfg, err := h.getWeChatOAuthConfig(c.Request.Context(), "mp")
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	paymentType := normalizeWeChatPaymentType(c.Query("payment_type"))
	if paymentType == "" {
		response.BadRequest(c, "Invalid payment type")
		return
	}

	state, err := oauth.GenerateState()
	if err != nil {
		response.ErrorFrom(c, infraerrors.InternalServer("OAUTH_STATE_GEN_FAILED", "failed to generate oauth state").WithCause(err))
		return
	}

	redirectTo := normalizeWeChatPaymentRedirectPath(sanitizeFrontendRedirectPath(c.Query("redirect")))
	if redirectTo == "" {
		redirectTo = wechatPaymentOAuthDefaultTo
	}
	rawContext, err := encodeWeChatPaymentOAuthContext(wechatPaymentOAuthContext{
		PaymentType: paymentType,
		Amount:      strings.TrimSpace(c.Query("amount")),
		OrderType:   strings.TrimSpace(c.Query("order_type")),
		PlanID:      parseWeChatPaymentPlanID(c.Query("plan_id")),
	})
	if err != nil {
		response.ErrorFrom(c, infraerrors.InternalServer("OAUTH_CONTEXT_ENCODE_FAILED", "failed to encode oauth context").WithCause(err))
		return
	}

	scope := normalizeWeChatPaymentScope(c.Query("scope"))
	secureCookie := isRequestHTTPS(c)
	wechatPaymentSetCookie(c, wechatPaymentOAuthStateName, encodeCookieValue(state), wechatOAuthCookieMaxAgeSec, secureCookie)
	wechatPaymentSetCookie(c, wechatPaymentOAuthRedirect, encodeCookieValue(redirectTo), wechatOAuthCookieMaxAgeSec, secureCookie)
	wechatPaymentSetCookie(c, wechatPaymentOAuthContextName, encodeCookieValue(rawContext), wechatOAuthCookieMaxAgeSec, secureCookie)
	wechatPaymentSetCookie(c, wechatPaymentOAuthScopeName, encodeCookieValue(scope), wechatOAuthCookieMaxAgeSec, secureCookie)

	cfg.redirectURI = h.resolveWeChatPaymentOAuthCallbackURL(c)
	cfg.scope = scope
	authURL, err := buildWeChatAuthorizeURL(cfg, state)
	if err != nil {
		response.ErrorFrom(c, infraerrors.InternalServer("OAUTH_BUILD_URL_FAILED", "failed to build oauth authorization url").WithCause(err))
		return
	}

	c.Redirect(http.StatusFound, authURL)
}

// WeChatPaymentOAuthCallback exchanges the code for an OpenID and forwards the
// browser back to the frontend callback route.
func (h *AuthHandler) WeChatPaymentOAuthCallback(c *gin.Context) {
	frontendCallback := wechatPaymentOAuthFrontendCB

	if providerErr := strings.TrimSpace(c.Query("error")); providerErr != "" {
		redirectOAuthError(c, frontendCallback, "provider_error", providerErr, c.Query("error_description"))
		return
	}

	code := strings.TrimSpace(c.Query("code"))
	state := strings.TrimSpace(c.Query("state"))
	if code == "" || state == "" {
		redirectOAuthError(c, frontendCallback, "missing_params", "missing code/state", "")
		return
	}

	secureCookie := isRequestHTTPS(c)
	defer func() {
		wechatPaymentClearCookie(c, wechatPaymentOAuthStateName, secureCookie)
		wechatPaymentClearCookie(c, wechatPaymentOAuthRedirect, secureCookie)
		wechatPaymentClearCookie(c, wechatPaymentOAuthContextName, secureCookie)
		wechatPaymentClearCookie(c, wechatPaymentOAuthScopeName, secureCookie)
	}()

	expectedState, err := readCookieDecoded(c, wechatPaymentOAuthStateName)
	if err != nil || expectedState == "" || state != expectedState {
		redirectOAuthError(c, frontendCallback, "invalid_state", "invalid oauth state", "")
		return
	}

	redirectTo, _ := readCookieDecoded(c, wechatPaymentOAuthRedirect)
	redirectTo = normalizeWeChatPaymentRedirectPath(sanitizeFrontendRedirectPath(redirectTo))
	if redirectTo == "" {
		redirectTo = wechatPaymentOAuthDefaultTo
	}

	rawContext, _ := readCookieDecoded(c, wechatPaymentOAuthContextName)
	paymentContext, err := decodeWeChatPaymentOAuthContext(rawContext)
	if err != nil {
		redirectOAuthError(c, frontendCallback, "invalid_context", "invalid oauth context", "")
		return
	}
	if paymentContext.PaymentType == "" {
		paymentContext.PaymentType = payment.TypeWxpay
	}

	scope, _ := readCookieDecoded(c, wechatPaymentOAuthScopeName)
	scope = normalizeWeChatPaymentScope(scope)

	cfg, err := h.getWeChatOAuthConfig(c.Request.Context(), "mp")
	if err != nil {
		redirectOAuthError(c, frontendCallback, "provider_error", infraerrors.Reason(err), infraerrors.Message(err))
		return
	}
	cfg.redirectURI = h.resolveWeChatPaymentOAuthCallbackURL(c)

	tokenResp, err := exchangeWeChatOAuthCode(c.Request.Context(), cfg, code)
	if err != nil {
		redirectOAuthError(c, frontendCallback, "token_exchange_failed", "failed to exchange oauth code", singleLine(err.Error()))
		return
	}

	openid := strings.TrimSpace(tokenResp.OpenID)
	if openid == "" {
		redirectOAuthError(c, frontendCallback, "missing_openid", "missing openid", "")
		return
	}
	if strings.TrimSpace(tokenResp.Scope) != "" {
		scope = strings.TrimSpace(tokenResp.Scope)
	}

	resumeToken, err := h.wechatPaymentResumeService().CreateWeChatPaymentResumeToken(service.WeChatPaymentResumeClaims{
		OpenID:      openid,
		PaymentType: paymentContext.PaymentType,
		Amount:      paymentContext.Amount,
		OrderType:   paymentContext.OrderType,
		PlanID:      paymentContext.PlanID,
		RedirectTo:  redirectTo,
		Scope:       scope,
	})
	if err != nil {
		redirectOAuthError(c, frontendCallback, "invalid_context", "failed to encode payment resume context", "")
		return
	}

	fragment := url.Values{}
	fragment.Set("wechat_resume_token", resumeToken)
	fragment.Set("openid", openid)
	fragment.Set("payment_type", paymentContext.PaymentType)
	fragment.Set("redirect", redirectTo)
	if paymentContext.Amount != "" {
		fragment.Set("amount", paymentContext.Amount)
	}
	if paymentContext.OrderType != "" {
		fragment.Set("order_type", paymentContext.OrderType)
	}
	if paymentContext.PlanID > 0 {
		fragment.Set("plan_id", strconv.FormatInt(paymentContext.PlanID, 10))
	}
	if scope != "" {
		fragment.Set("scope", scope)
	}
	redirectWithFragment(c, frontendCallback, fragment)
}

func (h *AuthHandler) getWeChatOAuthConfig(ctx context.Context, requestedMode string) (*wechatOAuthRuntimeConfig, error) {
	var base config.WeChatConnectConfig
	var err error
	if h != nil && h.settingSvc != nil {
		base, err = h.settingSvc.GetWeChatConnectOAuthConfig(ctx)
	} else if h != nil && h.cfg != nil {
		base = h.cfg.WeChat
	} else {
		return nil, infraerrors.ServiceUnavailable("CONFIG_NOT_READY", "config not loaded")
	}
	if err != nil {
		return nil, err
	}

	openEnabled := base.OpenEnabled
	mpEnabled := base.MPEnabled
	mobileEnabled := base.MobileEnabled
	if !openEnabled && !mpEnabled && !mobileEnabled {
		switch normalizeWeChatOAuthMode(base.Mode) {
		case "mp":
			mpEnabled = true
		case "mobile":
			mobileEnabled = true
		default:
			openEnabled = true
		}
	}

	mode := normalizeWeChatOAuthMode(firstNonEmpty(requestedMode, base.Mode))
	switch mode {
	case "mp":
		if !mpEnabled {
			if openEnabled {
				mode = "open"
			} else if mobileEnabled {
				mode = "mobile"
			}
		}
	case "mobile":
		if !mobileEnabled {
			if openEnabled {
				mode = "open"
			} else if mpEnabled {
				mode = "mp"
			}
		}
	default:
		if !openEnabled {
			if mpEnabled {
				mode = "mp"
			} else if mobileEnabled {
				mode = "mobile"
			}
		}
	}
	appID, appSecret := wechatConfigCredentialsForMode(base, mode)
	if strings.TrimSpace(appID) == "" {
		return nil, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "wechat app id not configured")
	}
	if strings.TrimSpace(appSecret) == "" {
		return nil, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "wechat app secret not configured")
	}
	frontendCallback := strings.TrimSpace(base.FrontendRedirectURL)
	if frontendCallback == "" {
		frontendCallback = wechatOAuthDefaultFrontendCB
	}
	return &wechatOAuthRuntimeConfig{
		mode:             mode,
		appID:            appID,
		appSecret:        appSecret,
		scope:            normalizeWeChatOAuthScope(base.Scopes, mode),
		redirectURI:      strings.TrimSpace(base.RedirectURL),
		frontendCallback: frontendCallback,
		openEnabled:      openEnabled,
		mpEnabled:        mpEnabled,
		mobileEnabled:    mobileEnabled,
	}, nil
}

func normalizeWeChatOAuthMode(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "mp":
		return "mp"
	case "mobile":
		return "mobile"
	default:
		return "open"
	}
}

func normalizeWeChatOAuthScope(raw, mode string) string {
	switch normalizeWeChatOAuthMode(mode) {
	case "mp":
		scope := strings.TrimSpace(raw)
		switch scope {
		case "snsapi_base", "snsapi_userinfo":
			return scope
		default:
			return "snsapi_userinfo"
		}
	case "mobile":
		return ""
	default:
		scope := strings.TrimSpace(raw)
		if scope == "" {
			return "snsapi_login"
		}
		return scope
	}
}

func wechatConfigCredentialsForMode(cfg config.WeChatConnectConfig, mode string) (string, string) {
	switch normalizeWeChatOAuthMode(mode) {
	case "mp":
		return strings.TrimSpace(firstNonEmpty(cfg.MPAppID, cfg.AppID)), strings.TrimSpace(firstNonEmpty(cfg.MPAppSecret, cfg.AppSecret))
	case "mobile":
		return strings.TrimSpace(firstNonEmpty(cfg.MobileAppID, cfg.AppID)), strings.TrimSpace(firstNonEmpty(cfg.MobileAppSecret, cfg.AppSecret))
	default:
		return strings.TrimSpace(firstNonEmpty(cfg.OpenAppID, cfg.AppID)), strings.TrimSpace(firstNonEmpty(cfg.OpenAppSecret, cfg.AppSecret))
	}
}

func buildWeChatAuthorizeURL(cfg *wechatOAuthRuntimeConfig, state string) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("missing wechat oauth config")
	}
	redirectURI := strings.TrimSpace(cfg.redirectURI)
	if redirectURI == "" {
		return "", fmt.Errorf("missing redirect uri")
	}
	query := url.Values{}
	query.Set("appid", cfg.appID)
	query.Set("redirect_uri", redirectURI)
	query.Set("response_type", "code")
	query.Set("scope", normalizeWeChatOAuthScope(cfg.scope, cfg.mode))
	query.Set("state", state)

	baseURL := "https://open.weixin.qq.com/connect/qrconnect"
	if cfg.mode == "mp" {
		baseURL = "https://open.weixin.qq.com/connect/oauth2/authorize"
	}
	return baseURL + "?" + query.Encode() + "#wechat_redirect", nil
}

func exchangeWeChatOAuthCode(ctx context.Context, cfg *wechatOAuthRuntimeConfig, code string) (*wechatOAuthTokenResponse, error) {
	resp, err := req.C().SetTimeout(30 * time.Second).R().
		SetContext(ctx).
		SetSuccessResult(&wechatOAuthTokenResponse{}).
		SetQueryParams(map[string]string{
			"appid":      cfg.appID,
			"secret":     cfg.appSecret,
			"code":       strings.TrimSpace(code),
			"grant_type": "authorization_code",
		}).
		Get(wechatOAuthAccessTokenURL)
	if err != nil {
		return nil, err
	}
	result, _ := resp.SuccessResult().(*wechatOAuthTokenResponse)
	if result == nil {
		return nil, fmt.Errorf("invalid wechat token response")
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat errcode=%d errmsg=%s", result.ErrCode, strings.TrimSpace(result.ErrMsg))
	}
	if strings.TrimSpace(result.OpenID) == "" {
		return nil, fmt.Errorf("missing openid")
	}
	return result, nil
}

func fetchWeChatOAuthUserInfo(ctx context.Context, tokenResp *wechatOAuthTokenResponse) (*wechatOAuthUserInfoResponse, error) {
	if tokenResp == nil || strings.TrimSpace(tokenResp.AccessToken) == "" || strings.TrimSpace(tokenResp.OpenID) == "" {
		return &wechatOAuthUserInfoResponse{}, nil
	}
	resp, err := req.C().SetTimeout(30 * time.Second).R().
		SetContext(ctx).
		SetSuccessResult(&wechatOAuthUserInfoResponse{}).
		SetQueryParams(map[string]string{
			"access_token": strings.TrimSpace(tokenResp.AccessToken),
			"openid":       strings.TrimSpace(tokenResp.OpenID),
			"lang":         "zh_CN",
		}).
		Get(wechatOAuthUserInfoURL)
	if err != nil {
		return nil, err
	}
	result, _ := resp.SuccessResult().(*wechatOAuthUserInfoResponse)
	if result == nil {
		return nil, fmt.Errorf("invalid wechat userinfo response")
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat errcode=%d errmsg=%s", result.ErrCode, strings.TrimSpace(result.ErrMsg))
	}
	return result, nil
}

func wechatSyntheticEmail(subject string) string {
	subject = strings.TrimSpace(subject)
	if subject == "" {
		return ""
	}
	return "wechat-" + subject + service.WeChatConnectSyntheticEmailDomain
}

func wechatFallbackUsername(subject string) string {
	subject = strings.TrimSpace(subject)
	if subject == "" {
		return "wechat_user"
	}
	if len(subject) > 8 {
		subject = subject[:8]
	}
	return "wechat_" + subject
}

func wechatSetCookie(c *gin.Context, name, value string, maxAge int, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     wechatOAuthCookiePath,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func wechatClearCookie(c *gin.Context, name string, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     wechatOAuthCookiePath,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *AuthHandler) wechatPaymentResumeService() *service.PaymentResumeService {
	var legacyKey []byte
	if key, err := payment.ProvideEncryptionKey(h.cfg); err == nil {
		legacyKey = []byte(key)
	}
	return service.NewLegacyAwarePaymentResumeService(legacyKey)
}

func (h *AuthHandler) resolveWeChatPaymentOAuthCallbackURL(c *gin.Context) string {
	scheme := "http"
	if isRequestHTTPS(c) {
		scheme = "https"
	}
	host := strings.TrimSpace(c.Request.Host)
	if host == "" {
		return ""
	}
	return scheme + "://" + host + "/api/v1/auth/oauth/wechat/payment/callback"
}

func normalizeWeChatPaymentType(raw string) string {
	normalized := service.NormalizeVisibleMethod(raw)
	if normalized == "" {
		return payment.TypeWxpay
	}
	if normalized != payment.TypeWxpay {
		return ""
	}
	return normalized
}

func normalizeWeChatPaymentRedirectPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return wechatPaymentOAuthDefaultTo
	}
	if path == "/payment" {
		return wechatPaymentOAuthDefaultTo
	}
	if strings.HasPrefix(path, "/payment?") {
		return wechatPaymentOAuthDefaultTo + path[len("/payment"):]
	}
	if !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") || strings.Contains(path, "://") {
		return wechatPaymentOAuthDefaultTo
	}
	return path
}

func normalizeWeChatPaymentScope(raw string) string {
	switch strings.TrimSpace(raw) {
	case "snsapi_userinfo":
		return "snsapi_userinfo"
	default:
		return "snsapi_base"
	}
}

func encodeWeChatPaymentOAuthContext(ctx wechatPaymentOAuthContext) (string, error) {
	payload, err := json.Marshal(ctx)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

func decodeWeChatPaymentOAuthContext(raw string) (*wechatPaymentOAuthContext, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return &wechatPaymentOAuthContext{}, nil
	}
	var out wechatPaymentOAuthContext
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	out.PaymentType = normalizeWeChatPaymentType(out.PaymentType)
	return &out, nil
}

func parseWeChatPaymentPlanID(raw string) int64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	planID, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || planID <= 0 {
		return 0
	}
	return planID
}

func wechatPaymentSetCookie(c *gin.Context, name, value string, maxAge int, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     wechatPaymentOAuthCookiePath,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func wechatPaymentClearCookie(c *gin.Context, name string, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     wechatPaymentOAuthCookiePath,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}
