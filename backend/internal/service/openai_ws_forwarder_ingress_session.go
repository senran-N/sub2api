package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	coderws "github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/openai"
)

type openAIWSIngressSessionContext struct {
	account               *Account
	wsDecision            OpenAIWSProtocolDecision
	ingressMode           string
	dedicatedMode         bool
	wsHost                string
	wsPath                string
	debugEnabled          bool
	firstPayload          openAIWSClientPayload
	turnState             string
	stateStore            OpenAIWSStateStore
	groupID               int64
	sessionHash           string
	preferredConnID       string
	storeDisabled         bool
	storeDisabledConnMode string
	isCodexCLI            bool
	baseAcquireReq        openAIWSAcquireRequest
	pool                  *openAIWSConnPool
}

func (s *OpenAIGatewayService) buildOpenAIWSIngressSessionContext(
	c *gin.Context,
	account *Account,
	token string,
	firstClientMessage []byte,
	wsDecision OpenAIWSProtocolDecision,
	ingressMode string,
	dedicatedMode bool,
) (*openAIWSIngressSessionContext, error) {
	if s == nil {
		return nil, errors.New("service is nil")
	}
	if c == nil {
		return nil, errors.New("gin context is nil")
	}
	if account == nil {
		return nil, errors.New("account is nil")
	}

	wsURL, err := s.buildOpenAIResponsesWSURL(account)
	if err != nil {
		return nil, fmt.Errorf("build ws url: %w", err)
	}
	wsHost := "-"
	wsPath := "-"
	if parsedURL, parseErr := url.Parse(wsURL); parseErr == nil && parsedURL != nil {
		wsHost = normalizeOpenAIWSLogValue(parsedURL.Host)
		wsPath = normalizeOpenAIWSLogValue(parsedURL.Path)
	}

	firstPayload, err := s.parseOpenAIWSIngressClientPayload(c, account, firstClientMessage)
	if err != nil {
		return nil, err
	}
	firstPayload = s.prepareOpenAIWSClientPayload(account, firstPayload)

	turnState := strings.TrimSpace(c.GetHeader(openAIWSTurnStateHeader))
	stateStore := s.getOpenAIWSStateStore()
	groupID := getOpenAIGroupIDFromContext(c)
	sessionHash := s.GenerateOpenAIWSIngressSessionHash(c, firstPayload.rawForHash)
	if turnState == "" && stateStore != nil && sessionHash != "" {
		if savedTurnState, ok := stateStore.GetSessionTurnState(groupID, sessionHash); ok {
			turnState = savedTurnState
		}
	}

	preferredConnID := ""
	if stateStore != nil && firstPayload.previousResponseID != "" {
		if connID, ok := stateStore.GetResponseConn(firstPayload.previousResponseID); ok {
			preferredConnID = connID
		}
	}

	storeDisabled := firstPayload.storeDisabled
	storeDisabledConnMode := s.openAIWSStoreDisabledConnMode()
	if stateStore != nil && storeDisabled && firstPayload.previousResponseID == "" && sessionHash != "" {
		if connID, ok := stateStore.GetSessionConn(groupID, sessionHash); ok {
			preferredConnID = connID
		}
	}

	isCodexCLI := openai.IsCodexOfficialClientByHeaders(c.GetHeader("User-Agent"), c.GetHeader("originator")) ||
		(s.cfg != nil && s.cfg.Gateway.ForceCodexCLI)
	wsHeaders, _ := s.buildOpenAIWSHeaders(
		c,
		account,
		token,
		wsDecision,
		isCodexCLI,
		turnState,
		strings.TrimSpace(c.GetHeader(openAIWSTurnMetadataHeader)),
		firstPayload.promptCacheKey,
	)
	pool := s.getOpenAIWSConnPool()
	if pool == nil {
		return nil, errors.New("openai ws conn pool is nil")
	}

	session := &openAIWSIngressSessionContext{
		account:               account,
		wsDecision:            wsDecision,
		ingressMode:           ingressMode,
		dedicatedMode:         dedicatedMode,
		wsHost:                wsHost,
		wsPath:                wsPath,
		debugEnabled:          isOpenAIWSModeDebugEnabled(),
		firstPayload:          firstPayload,
		turnState:             turnState,
		stateStore:            stateStore,
		groupID:               groupID,
		sessionHash:           sessionHash,
		preferredConnID:       preferredConnID,
		storeDisabled:         storeDisabled,
		storeDisabledConnMode: storeDisabledConnMode,
		isCodexCLI:            isCodexCLI,
		baseAcquireReq: openAIWSAcquireRequest{
			Account: account,
			WSURL:   wsURL,
			Headers: wsHeaders,
			ProxyURL: func() string {
				if account.ProxyID != nil && account.Proxy != nil {
					return account.Proxy.URL()
				}
				return ""
			}(),
			ForceNewConn: false,
		},
		pool: pool,
	}

	logOpenAIWSModeInfo(
		"ingress_ws_protocol_confirm account_id=%d account_type=%s transport=%s ws_host=%s ws_path=%s ws_mode=%s store_disabled=%v has_session_hash=%v has_previous_response_id=%v",
		account.ID,
		account.Type,
		normalizeOpenAIWSLogValue(string(wsDecision.Transport)),
		wsHost,
		wsPath,
		normalizeOpenAIWSLogValue(ingressMode),
		storeDisabled,
		sessionHash != "",
		firstPayload.previousResponseID != "",
	)

	if session.debugEnabled {
		logOpenAIWSModeDebug(
			"ingress_ws_start account_id=%d account_type=%s transport=%s ws_host=%s preferred_conn_id=%s has_session_hash=%v has_previous_response_id=%v store_disabled=%v",
			account.ID,
			account.Type,
			normalizeOpenAIWSLogValue(string(wsDecision.Transport)),
			wsHost,
			truncateOpenAIWSLogValue(preferredConnID, openAIWSIDValueMaxLen),
			sessionHash != "",
			firstPayload.previousResponseID != "",
			storeDisabled,
		)
	}
	if firstPayload.previousResponseID != "" {
		firstPreviousResponseIDKind := ClassifyOpenAIPreviousResponseIDKind(firstPayload.previousResponseID)
		logOpenAIWSModeInfo(
			"ingress_ws_continuation_probe account_id=%d turn=%d previous_response_id=%s previous_response_id_kind=%s preferred_conn_id=%s session_hash=%s header_session_id=%s header_conversation_id=%s has_turn_state=%v turn_state_len=%d has_prompt_cache_key=%v store_disabled=%v",
			account.ID,
			1,
			truncateOpenAIWSLogValue(firstPayload.previousResponseID, openAIWSIDValueMaxLen),
			normalizeOpenAIWSLogValue(firstPreviousResponseIDKind),
			truncateOpenAIWSLogValue(preferredConnID, openAIWSIDValueMaxLen),
			truncateOpenAIWSLogValue(sessionHash, 12),
			openAIWSHeaderValueForLog(session.baseAcquireReq.Headers, "session_id"),
			openAIWSHeaderValueForLog(session.baseAcquireReq.Headers, "conversation_id"),
			turnState != "",
			len(turnState),
			firstPayload.promptCacheKey != "",
			storeDisabled,
		)
	}

	return session, nil
}

func (s *OpenAIGatewayService) acquireOpenAIWSIngressTurnLease(
	ctx context.Context,
	session *openAIWSIngressSessionContext,
	turn int,
	preferredConnID string,
	forcePreferredConn bool,
) (*openAIWSConnLease, error) {
	if s == nil {
		return nil, errors.New("service is nil")
	}
	if session == nil {
		return nil, errors.New("ingress session context is nil")
	}
	if session.pool == nil {
		return nil, errors.New("openai ws conn pool is nil")
	}
	if session.account == nil {
		return nil, errors.New("account is nil")
	}

	acquireTimeout := s.openAIWSAcquireTimeout()
	if acquireTimeout <= 0 {
		acquireTimeout = 30 * time.Second
	}

	req := normalizeOpenAIWSAcquireRequest(session.baseAcquireReq)
	req.PreferredConnID = strings.TrimSpace(preferredConnID)
	req.ForcePreferredConn = forcePreferredConn
	req.ForceNewConn = session.dedicatedMode

	acquireCtx, acquireCancel := context.WithTimeout(ctx, acquireTimeout)
	lease, acquireErr := session.pool.Acquire(acquireCtx, req)
	acquireCancel()
	if acquireErr != nil {
		dialStatus, dialClass, dialCloseStatus, dialCloseReason, dialRespServer, dialRespVia, dialRespCFRay, dialRespReqID := summarizeOpenAIWSDialError(acquireErr)
		logOpenAIWSModeInfo(
			"ingress_ws_upstream_acquire_fail account_id=%d turn=%d reason=%s dial_status=%d dial_class=%s dial_close_status=%s dial_close_reason=%s dial_resp_server=%s dial_resp_via=%s dial_resp_cf_ray=%s dial_resp_x_request_id=%s cause=%s preferred_conn_id=%s force_preferred_conn=%v ws_host=%s ws_path=%s proxy_enabled=%v",
			session.account.ID,
			turn,
			normalizeOpenAIWSLogValue(classifyOpenAIWSAcquireError(acquireErr)),
			dialStatus,
			dialClass,
			dialCloseStatus,
			truncateOpenAIWSLogValue(dialCloseReason, openAIWSHeaderValueMaxLen),
			dialRespServer,
			dialRespVia,
			dialRespCFRay,
			dialRespReqID,
			truncateOpenAIWSLogValue(acquireErr.Error(), openAIWSLogValueMaxLen),
			truncateOpenAIWSLogValue(preferredConnID, openAIWSIDValueMaxLen),
			forcePreferredConn,
			session.wsHost,
			session.wsPath,
			session.account.ProxyID != nil && session.account.Proxy != nil,
		)
		var dialErr *openAIWSDialError
		if errors.As(acquireErr, &dialErr) && dialErr != nil && dialErr.StatusCode == http.StatusTooManyRequests {
			s.persistOpenAIWSRateLimitSignal(ctx, session.account, dialErr.ResponseHeaders, nil, "rate_limit_exceeded", "rate_limit_error", strings.TrimSpace(acquireErr.Error()))
		}
		if errors.Is(acquireErr, errOpenAIWSPreferredConnUnavailable) {
			return nil, NewOpenAIWSClientCloseError(
				coderws.StatusPolicyViolation,
				"upstream continuation connection is unavailable; please restart the conversation",
				acquireErr,
			)
		}
		if errors.Is(acquireErr, context.DeadlineExceeded) || errors.Is(acquireErr, errOpenAIWSConnQueueFull) {
			return nil, NewOpenAIWSClientCloseError(
				coderws.StatusTryAgainLater,
				"upstream websocket is busy, please retry later",
				acquireErr,
			)
		}
		return nil, acquireErr
	}

	connID := strings.TrimSpace(lease.ConnID())
	if handshakeTurnState := strings.TrimSpace(lease.HandshakeHeader(openAIWSTurnStateHeader)); handshakeTurnState != "" {
		session.turnState = handshakeTurnState
		if session.stateStore != nil && session.sessionHash != "" {
			session.stateStore.BindSessionTurnState(session.groupID, session.sessionHash, handshakeTurnState, s.openAIWSSessionStickyTTL())
		}
		updatedHeaders := cloneHeader(session.baseAcquireReq.Headers)
		if updatedHeaders == nil {
			updatedHeaders = make(http.Header)
		}
		updatedHeaders.Set(openAIWSTurnStateHeader, handshakeTurnState)
		session.baseAcquireReq.Headers = updatedHeaders
	}
	logOpenAIWSModeInfo(
		"ingress_ws_upstream_connected account_id=%d turn=%d conn_id=%s conn_reused=%v conn_pick_ms=%d queue_wait_ms=%d preferred_conn_id=%s",
		session.account.ID,
		turn,
		truncateOpenAIWSLogValue(connID, openAIWSIDValueMaxLen),
		lease.Reused(),
		lease.ConnPickDuration().Milliseconds(),
		lease.QueueWaitDuration().Milliseconds(),
		truncateOpenAIWSLogValue(preferredConnID, openAIWSIDValueMaxLen),
	)
	return lease, nil
}
