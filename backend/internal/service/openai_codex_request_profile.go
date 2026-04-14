package service

import (
	"strings"
	"unicode"

	"github.com/cespare/xxhash/v2"
	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/openai"
	"github.com/tidwall/gjson"
)

const (
	// OpenAIParsedCodexRequestProfileKey caches the structured Codex request profile for the current request.
	OpenAIParsedCodexRequestProfileKey = "openai_parsed_codex_request_profile"
	// OpenAIParsedCodexRequestProfileCacheKey records the body/config binding for the cached request profile.
	OpenAIParsedCodexRequestProfileCacheKey = "openai_parsed_codex_request_profile_cache"
)

const (
	// CodexOfficialClientReasonUnknown means the request did not match any known official-client signal.
	CodexOfficialClientReasonUnknown = "unknown"
	// CodexOfficialClientReasonUserAgent means the official-client match came from User-Agent.
	CodexOfficialClientReasonUserAgent = "user_agent"
	// CodexOfficialClientReasonOriginator means the official-client match came from originator.
	CodexOfficialClientReasonOriginator = "originator"
	// CodexOfficialClientReasonForceCodexCLI means ForceCodexCLI elevated the request into the Codex official path.
	CodexOfficialClientReasonForceCodexCLI = "force_codex_cli"
)

// CodexWireAPI identifies which Responses wire protocol shape the current request is using.
type CodexWireAPI string

const (
	CodexWireAPIUnknown            CodexWireAPI = ""
	CodexWireAPIResponsesHTTP      CodexWireAPI = "responses_http"
	CodexWireAPIResponsesWebSocket CodexWireAPI = "responses_websocket"
)

// CodexRequestHeaderProfile captures the Codex-relevant request headers that influence forwarding and chaining.
type CodexRequestHeaderProfile struct {
	Accept         string
	AcceptLanguage string
	ConversationID string
	OpenAIBeta     string
	Originator     string
	SessionID      string
	TurnMetadata   string
	TurnState      string
	UserAgent      string
	Version        string
}

// CodexRequestBodyProfile captures the request body signals that affect transport, mutation, and continuation.
type CodexRequestBodyProfile struct {
	FunctionCallOutputPresent bool
	InstructionsPresent       bool
	Model                     string
	PreviousResponseID        string
	PromptCacheKey            string
	ReasoningEffort           string
	ReasoningPresent          bool
	RequestType               string
	Store                     bool
	StorePresent              bool
	Stream                    bool
	StreamPresent             bool
}

// CodexContinuationProfile captures chain-related state inferred from headers and body.
type CodexContinuationProfile struct {
	DependsOnPriorResponse bool
	HasPromptCacheKey      bool
	HasTurnMetadata        bool
	HasTurnState           bool
	PreviousResponseIDKind string
}

// CodexRequestProfile is the structured request image consumed by Codex-specific routing decisions.
type CodexRequestProfile struct {
	ClientTransport       OpenAIClientTransport
	CodexVersion          string
	CompactPath           bool
	Continuation          CodexContinuationProfile
	ForceCodexCLI         bool
	Headers               CodexRequestHeaderProfile
	OfficialClient        bool
	OfficialClientReason  string
	TransportFallbackHTTP bool
	Warmup                bool
	WireAPI               CodexWireAPI
	Body                  CodexRequestBodyProfile
}

type codexRequestProfileCache struct {
	BodyBound     bool
	BodyHash      uint64
	ContextHash   uint64
	ForceCodexCLI bool
}

// GetCodexRequestProfile returns the shared Codex request profile and caches it on gin.Context.
func GetCodexRequestProfile(c *gin.Context, body []byte, forceCodexCLI bool) CodexRequestProfile {
	cacheTag := buildCodexRequestProfileCacheTag(c, body, forceCodexCLI)
	if c != nil {
		if cached, ok := c.Get(OpenAIParsedCodexRequestProfileKey); ok {
			if profile, ok := cached.(CodexRequestProfile); ok {
				cacheState, hasCacheState := c.Get(OpenAIParsedCodexRequestProfileCacheKey)
				if !hasCacheState {
					return profile
				}
				if cachedTag, ok := cacheState.(codexRequestProfileCache); ok {
					if codexRequestProfileCacheMatches(cachedTag, cacheTag) {
						return profile
					}
				}
			}
		}
	}

	profile := buildCodexRequestProfile(c, body, forceCodexCLI)
	if c != nil {
		c.Set(OpenAIParsedCodexRequestProfileKey, profile)
		c.Set(OpenAIParsedCodexRequestProfileCacheKey, cacheTag)
	}
	return profile
}

func buildCodexRequestProfileCacheTag(c *gin.Context, body []byte, forceCodexCLI bool) codexRequestProfileCache {
	cacheTag := codexRequestProfileCache{
		ContextHash:   hashCodexRequestProfileContext(c),
		ForceCodexCLI: forceCodexCLI,
	}
	if len(body) > 0 {
		cacheTag.BodyBound = true
		cacheTag.BodyHash = xxhash.Sum64(body)
	}
	return cacheTag
}

func codexRequestProfileCacheMatches(cached, current codexRequestProfileCache) bool {
	return cached.BodyBound == current.BodyBound &&
		cached.BodyHash == current.BodyHash &&
		cached.ContextHash == current.ContextHash &&
		cached.ForceCodexCLI == current.ForceCodexCLI
}

func hashCodexRequestProfileContext(c *gin.Context) uint64 {
	if c == nil {
		return 0
	}

	hasher := xxhash.New()
	writeValue := func(value string) {
		_, _ = hasher.WriteString(strings.TrimSpace(value))
		_, _ = hasher.Write([]byte{0})
	}

	if c.Request != nil && c.Request.URL != nil {
		writeValue(c.Request.URL.Path)
	}
	writeValue(string(GetOpenAIClientTransport(c)))
	for _, key := range []string{
		"Accept",
		"Accept-Language",
		"conversation_id",
		"OpenAI-Beta",
		"originator",
		"session_id",
		openAIWSTurnMetadataHeader,
		openAIWSTurnStateHeader,
		"User-Agent",
		"version",
	} {
		writeValue(getTrimmedCodexRequestHeader(c, key))
	}

	return hasher.Sum64()
}

func buildCodexRequestProfile(c *gin.Context, body []byte, forceCodexCLI bool) CodexRequestProfile {
	headers := CodexRequestHeaderProfile{
		Accept:         getTrimmedCodexRequestHeader(c, "Accept"),
		AcceptLanguage: getTrimmedCodexRequestHeader(c, "Accept-Language"),
		ConversationID: getTrimmedCodexRequestHeader(c, "conversation_id"),
		OpenAIBeta:     getTrimmedCodexRequestHeader(c, "OpenAI-Beta"),
		Originator:     getTrimmedCodexRequestHeader(c, "originator"),
		SessionID:      getTrimmedCodexRequestHeader(c, "session_id"),
		TurnMetadata:   getTrimmedCodexRequestHeader(c, openAIWSTurnMetadataHeader),
		TurnState:      getTrimmedCodexRequestHeader(c, openAIWSTurnStateHeader),
		UserAgent:      getTrimmedCodexRequestHeader(c, "User-Agent"),
		Version:        getTrimmedCodexRequestHeader(c, "version"),
	}
	officialClient, officialReason := detectCodexOfficialClient(headers.UserAgent, headers.Originator, forceCodexCLI)
	meta := getOpenAIRequestMeta(c, body)
	storeValue := gjson.GetBytes(body, "store")
	requestType := strings.TrimSpace(gjson.GetBytes(body, "type").String())

	profile := CodexRequestProfile{
		ClientTransport:       GetOpenAIClientTransport(c),
		CodexVersion:          resolveCodexVersion(headers.UserAgent, headers.Version),
		CompactPath:           isOpenAIResponsesCompactPath(c),
		ForceCodexCLI:         forceCodexCLI,
		Headers:               headers,
		OfficialClient:        officialClient,
		OfficialClientReason:  officialReason,
		TransportFallbackHTTP: false,
		WireAPI:               resolveCodexWireAPI(c),
		Body: CodexRequestBodyProfile{
			FunctionCallOutputPresent: gjson.GetBytes(body, `input.#(type=="function_call_output")`).Exists(),
			InstructionsPresent:       hasNonEmptyOpenAIRequestField(body, "instructions"),
			Model:                     meta.Model,
			PreviousResponseID:        meta.PreviousResponseID,
			PromptCacheKey:            meta.PromptCacheKey,
			ReasoningEffort:           meta.ReasoningEffort,
			ReasoningPresent:          meta.ReasoningPresent,
			RequestType:               requestType,
			Store:                     storeValue.Type == gjson.True,
			StorePresent:              storeValue.Exists(),
			Stream:                    meta.Stream,
			StreamPresent:             meta.StreamExists,
		},
	}
	profile.Continuation = CodexContinuationProfile{
		DependsOnPriorResponse: profile.Body.PreviousResponseID != "" || headers.TurnState != "" || profile.Body.FunctionCallOutputPresent,
		HasPromptCacheKey:      profile.Body.PromptCacheKey != "",
		HasTurnMetadata:        headers.TurnMetadata != "",
		HasTurnState:           headers.TurnState != "",
		PreviousResponseIDKind: ClassifyOpenAIPreviousResponseIDKind(profile.Body.PreviousResponseID),
	}
	profile.Warmup = isCodexWarmupRequest(profile, body)
	profile.TransportFallbackHTTP = profile.ClientTransport == OpenAIClientTransportHTTP && profile.WireAPI == CodexWireAPIResponsesHTTP
	return profile
}

func getTrimmedCodexRequestHeader(c *gin.Context, key string) string {
	if c == nil || c.Request == nil {
		return ""
	}
	return strings.TrimSpace(c.GetHeader(key))
}

func detectCodexOfficialClient(userAgent, originator string, forceCodexCLI bool) (bool, string) {
	switch {
	case openai.IsCodexOfficialClientRequest(userAgent):
		return true, CodexOfficialClientReasonUserAgent
	case openai.IsCodexOfficialClientOriginator(originator):
		return true, CodexOfficialClientReasonOriginator
	case forceCodexCLI:
		return true, CodexOfficialClientReasonForceCodexCLI
	default:
		return false, CodexOfficialClientReasonUnknown
	}
}

func resolveCodexWireAPI(c *gin.Context) CodexWireAPI {
	switch GetOpenAIClientTransport(c) {
	case OpenAIClientTransportWS:
		return CodexWireAPIResponsesWebSocket
	case OpenAIClientTransportHTTP:
		if c != nil && c.Request != nil && strings.Contains(c.Request.URL.Path, "/responses") {
			return CodexWireAPIResponsesHTTP
		}
	}
	return CodexWireAPIUnknown
}

func resolveCodexVersion(userAgent, versionHeader string) string {
	if version := strings.TrimSpace(versionHeader); version != "" {
		return version
	}
	ua := strings.TrimSpace(userAgent)
	if ua == "" {
		return ""
	}

	normalizedUA := strings.ToLower(ua)
	for _, marker := range []string{
		"codex_cli_rs/",
		"codex_vscode/",
		"codex_app/",
		"codex_chatgpt_desktop/",
		"codex_atlas/",
		"codex_exec/",
		"codex_sdk_ts/",
		"codex desktop/",
		"codex/",
	} {
		index := strings.Index(normalizedUA, marker)
		if index < 0 || index+len(marker) >= len(ua) {
			continue
		}
		if version := trimCodexVersionToken(ua[index+len(marker):]); version != "" {
			return version
		}
	}
	return ""
}

func trimCodexVersionToken(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	for i, r := range value {
		if unicode.IsSpace(r) || r == '(' || r == ')' || r == ';' {
			return strings.TrimSpace(value[:i])
		}
	}
	return value
}

func hasNonEmptyOpenAIRequestField(body []byte, path string) bool {
	value := gjson.GetBytes(body, path)
	if !value.Exists() {
		return false
	}
	if value.Type != gjson.String {
		return true
	}
	return strings.TrimSpace(value.String()) != ""
}

func isCodexWarmupRequest(profile CodexRequestProfile, body []byte) bool {
	if profile.WireAPI != CodexWireAPIResponsesWebSocket {
		return false
	}
	if profile.Body.RequestType != "" && profile.Body.RequestType != "response.create" {
		return false
	}
	generateValue := gjson.GetBytes(body, "generate")
	return generateValue.Exists() && generateValue.Type == gjson.False
}
