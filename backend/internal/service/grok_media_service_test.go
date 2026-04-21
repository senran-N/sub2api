//go:build unit

package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type grokMediaSettingRepoStub struct {
	values map[string]string
}

func (s *grokMediaSettingRepoStub) Get(context.Context, string) (*Setting, error) {
	panic("unexpected Get call")
}
func (s *grokMediaSettingRepoStub) GetValue(context.Context, string) (string, error) {
	panic("unexpected GetValue call")
}
func (s *grokMediaSettingRepoStub) Set(context.Context, string, string) error {
	panic("unexpected Set call")
}
func (s *grokMediaSettingRepoStub) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	result := make(map[string]string, len(keys))
	for _, key := range keys {
		result[key] = s.values[key]
	}
	return result, nil
}
func (s *grokMediaSettingRepoStub) SetMultiple(context.Context, map[string]string) error {
	panic("unexpected SetMultiple call")
}
func (s *grokMediaSettingRepoStub) GetAll(context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}
func (s *grokMediaSettingRepoStub) Delete(context.Context, string) error {
	panic("unexpected Delete call")
}

type stubGrokVideoJobRepository struct {
	records   map[string]GrokVideoJobRecord
	upserts   []GrokVideoJobRecord
	patches   []GrokVideoJobStatusPatch
	getCalls  []string
	getErr    error
	updateErr error
}

func (s *stubGrokVideoJobRepository) Upsert(_ context.Context, record GrokVideoJobRecord) error {
	s.upserts = append(s.upserts, record)
	if s.records == nil {
		s.records = make(map[string]GrokVideoJobRecord)
	}
	s.records[record.JobID] = record
	return nil
}

func (s *stubGrokVideoJobRepository) GetByJobID(_ context.Context, jobID string) (*GrokVideoJobRecord, error) {
	s.getCalls = append(s.getCalls, jobID)
	if s.getErr != nil {
		return nil, s.getErr
	}
	record, ok := s.records[jobID]
	if !ok {
		return nil, ErrGrokVideoJobNotFound
	}
	recordCopy := record
	return &recordCopy, nil
}

func (s *stubGrokVideoJobRepository) UpdateStatus(_ context.Context, patch GrokVideoJobStatusPatch) error {
	s.patches = append(s.patches, patch)
	if s.updateErr != nil {
		return s.updateErr
	}
	record, ok := s.records[patch.JobID]
	if !ok {
		return ErrGrokVideoJobNotFound
	}
	record.UpstreamStatus = patch.UpstreamStatus
	record.NormalizedStatus = patch.NormalizedStatus
	record.PollAfter = patch.PollAfter
	record.ErrorCode = patch.ErrorCode
	record.ErrorMessage = patch.ErrorMessage
	if patch.OutputAssetID != "" {
		record.OutputAssetID = patch.OutputAssetID
	}
	s.records[patch.JobID] = record
	return nil
}

type stubGrokMediaAssetRepository struct {
	records        map[string]GrokMediaAssetRecord
	upserts        []GrokMediaAssetRecord
	patches        []GrokMediaAssetCachePatch
	accesses       []string
	hashLookups    []string
	deletedBatches [][]string
	localPathRefs  []string
}

func (s *stubGrokMediaAssetRepository) Upsert(_ context.Context, record GrokMediaAssetRecord) error {
	s.upserts = append(s.upserts, record)
	if s.records == nil {
		s.records = make(map[string]GrokMediaAssetRecord)
	}
	s.records[record.AssetID] = record
	return nil
}

func (s *stubGrokMediaAssetRepository) GetByAssetID(_ context.Context, assetID string) (*GrokMediaAssetRecord, error) {
	record, ok := s.records[assetID]
	if !ok {
		return nil, ErrGrokMediaAssetNotFound
	}
	recordCopy := record
	return &recordCopy, nil
}

func (s *stubGrokMediaAssetRepository) FindCachedByHash(_ context.Context, contentHash string) (*GrokMediaAssetRecord, error) {
	s.hashLookups = append(s.hashLookups, contentHash)
	for _, record := range s.records {
		if record.ContentHash == contentHash && record.Status == grokMediaAssetStatusReady && strings.TrimSpace(record.LocalPath) != "" {
			recordCopy := record
			return &recordCopy, nil
		}
	}
	return nil, nil
}

func (s *stubGrokMediaAssetRepository) UpdateCacheState(_ context.Context, patch GrokMediaAssetCachePatch) error {
	s.patches = append(s.patches, patch)
	record, ok := s.records[patch.AssetID]
	if !ok {
		return ErrGrokMediaAssetNotFound
	}
	record.LocalPath = patch.LocalPath
	record.ContentHash = patch.ContentHash
	record.MimeType = patch.MimeType
	record.SizeBytes = patch.SizeBytes
	record.Status = patch.Status
	record.ExpiresAt = patch.ExpiresAt
	record.LastAccessAt = patch.LastAccessAt
	s.records[patch.AssetID] = record
	return nil
}

func (s *stubGrokMediaAssetRepository) MarkAccessed(_ context.Context, assetID string, accessAt time.Time, expiresAt *time.Time) error {
	s.accesses = append(s.accesses, assetID)
	record, ok := s.records[assetID]
	if !ok {
		return ErrGrokMediaAssetNotFound
	}
	record.LastAccessAt = &accessAt
	record.ExpiresAt = expiresAt
	s.records[assetID] = record
	return nil
}

func (s *stubGrokMediaAssetRepository) DeleteExpired(_ context.Context, expiresBefore time.Time, limit int) ([]GrokMediaAssetRecord, error) {
	if limit <= 0 {
		return nil, nil
	}

	expiredIDs := make([]string, 0, limit)
	expired := make([]GrokMediaAssetRecord, 0, limit)
	for assetID, record := range s.records {
		if record.ExpiresAt == nil || record.ExpiresAt.After(expiresBefore) {
			continue
		}
		expiredIDs = append(expiredIDs, assetID)
		expired = append(expired, record)
		if len(expired) >= limit {
			break
		}
	}
	if len(expired) == 0 {
		return nil, nil
	}
	s.deletedBatches = append(s.deletedBatches, expiredIDs)
	for _, assetID := range expiredIDs {
		delete(s.records, assetID)
	}
	return expired, nil
}

func (s *stubGrokMediaAssetRepository) CountByLocalPath(_ context.Context, localPath string) (int, error) {
	s.localPathRefs = append(s.localPathRefs, localPath)
	count := 0
	for _, record := range s.records {
		if record.LocalPath == localPath {
			count++
		}
	}
	return count, nil
}

func TestGrokMediaServiceHandleImages_ForwardsThroughGrokRuntime(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := []byte(`{"model":"grok-2-image","prompt":"cat"}`)
	upstream := &queuedHTTPUpstream{
		responses: []*http.Response{
			newJSONResponse(http.StatusOK, `{"data":[{"url":"https://media.example/image.png"}]}`),
		},
	}
	repo := &mockAccountRepoForPlatform{
		accounts: []Account{
			newSchedulableGrokMediaAccount(41, map[string]any{
				"grok": map[string]any{
					"tier": map[string]any{
						"normalized": "basic",
					},
				},
			}),
		},
	}
	svc := NewGrokMediaService(&GatewayService{
		accountRepo:  repo,
		cfg:          testConfig(),
		httpUpstream: upstream,
	}, nil, nil)

	c, rec := newGrokMediaTestContext(http.MethodPost, "/v1/images/generations", body)
	handled := svc.HandleImages(c, nil, body)

	require.True(t, handled)
	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, `{"data":[{"url":"https://media.example/image.png"}]}`, rec.Body.String())
	require.Len(t, upstream.requests, 1)
	require.Equal(t, "Bearer sk-grok-media", upstream.requests[0].Header.Get("Authorization"))
	require.Equal(t, "/v1/images/generations", upstream.requests[0].URL.Path)
	require.JSONEq(t, string(body), readTestRequestBody(t, upstream.requests[0]))
}

func TestGrokMediaServiceHandleImages_FailoverSwitchesAccount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := []byte(`{"model":"grok-2-image","prompt":"cat"}`)
	upstream := &queuedHTTPUpstream{
		responses: []*http.Response{
			newJSONResponse(http.StatusTooManyRequests, `{"error":{"message":"rate limit reached"}}`),
			newJSONResponse(http.StatusOK, `{"data":[{"url":"https://media.example/image.png"}]}`),
		},
	}
	repo := &mockAccountRepoForPlatform{
		accounts: []Account{
			{
				ID:          141,
				Name:        "grok-media-first",
				Platform:    PlatformGrok,
				Type:        AccountTypeAPIKey,
				Status:      StatusActive,
				Schedulable: true,
				Credentials: map[string]any{
					"api_key":  "sk-grok-first",
					"base_url": "https://grok.example/v1",
				},
				Extra: map[string]any{
					"grok": map[string]any{
						"tier": map[string]any{
							"normalized": "basic",
						},
					},
				},
			},
			{
				ID:          142,
				Name:        "grok-media-second",
				Platform:    PlatformGrok,
				Type:        AccountTypeAPIKey,
				Status:      StatusActive,
				Schedulable: true,
				Credentials: map[string]any{
					"api_key":  "sk-grok-second",
					"base_url": "https://grok.example/v1",
				},
				Extra: map[string]any{
					"grok": map[string]any{
						"tier": map[string]any{
							"normalized": "basic",
						},
					},
				},
			},
		},
	}
	svc := NewGrokMediaService(&GatewayService{
		accountRepo:  repo,
		cfg:          testConfig(),
		httpUpstream: upstream,
	}, nil, nil)

	c, rec := newGrokMediaTestContext(http.MethodPost, "/v1/images/generations", body)
	handled := svc.HandleImages(c, nil, body)

	require.True(t, handled)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, upstream.requests, 2)
	require.Equal(t, "Bearer sk-grok-first", upstream.requests[0].Header.Get("Authorization"))
	require.Equal(t, "Bearer sk-grok-second", upstream.requests[1].Header.Get("Authorization"))
	require.Len(t, repo.runtimeStates, 2)
	require.Equal(t, "failover", repo.runtimeStates[0]["last_outcome"])
	require.Equal(t, "success", repo.runtimeStates[1]["last_outcome"])
	require.Equal(t, "image", repo.runtimeStates[0]["last_request_capability"])
	require.Equal(t, "image", repo.runtimeStates[1]["last_request_capability"])
}

func TestGrokMediaServiceHandleVideos_CreateBindsJobToSelectedAccount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := []byte(`{"model":"grok-imagine-video","prompt":"launch sequence"}`)
	upstream := &queuedHTTPUpstream{
		responses: []*http.Response{
			newJSONResponse(http.StatusOK, `{"id":"job_123","status":"queued"}`),
		},
	}
	videoJobs := &stubGrokVideoJobRepository{}
	mediaAssets := &stubGrokMediaAssetRepository{}
	repo := &mockAccountRepoForPlatform{
		accounts: []Account{
			newSchedulableGrokMediaAccount(42, map[string]any{
				"grok": map[string]any{
					"tier": map[string]any{
						"normalized": "super",
					},
				},
			}),
		},
	}
	svc := NewGrokMediaService(&GatewayService{
		accountRepo:  repo,
		cfg:          testConfig(),
		httpUpstream: upstream,
	}, videoJobs, mediaAssets)

	c, rec := newGrokMediaTestContext(http.MethodPost, "/v1/videos", body)
	handled := svc.HandleVideos(c, nil, body)

	require.True(t, handled)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, videoJobs.upserts, 1)
	require.Equal(t, "job_123", videoJobs.upserts[0].JobID)
	require.Equal(t, int64(42), videoJobs.upserts[0].AccountID)
	require.Equal(t, "grok-imagine-video", videoJobs.upserts[0].RequestedModel)
	require.Equal(t, "grok-imagine-video", videoJobs.upserts[0].CanonicalModel)
	require.Equal(t, "queued", videoJobs.upserts[0].UpstreamStatus)
	require.Equal(t, "queued", videoJobs.upserts[0].NormalizedStatus)
	require.JSONEq(t, string(body), string(videoJobs.upserts[0].RequestPayloadSnapshot))
	require.Len(t, mediaAssets.upserts, 0)
}

func TestGrokMediaServiceHandleVideos_CreateFailoverSwitchesAccount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := []byte(`{"model":"grok-imagine-video","prompt":"launch sequence"}`)
	upstream := &queuedHTTPUpstream{
		responses: []*http.Response{
			newJSONResponse(http.StatusTooManyRequests, `{"error":{"message":"rate limit reached"}}`),
			newJSONResponse(http.StatusOK, `{"id":"job_456","status":"queued"}`),
		},
	}
	videoJobs := &stubGrokVideoJobRepository{}
	mediaAssets := &stubGrokMediaAssetRepository{}
	repo := &mockAccountRepoForPlatform{
		accounts: []Account{
			{
				ID:          242,
				Name:        "grok-video-first",
				Platform:    PlatformGrok,
				Type:        AccountTypeAPIKey,
				Status:      StatusActive,
				Schedulable: true,
				Credentials: map[string]any{
					"api_key":  "sk-video-first",
					"base_url": "https://grok.example/v1",
				},
				Extra: map[string]any{
					"grok": map[string]any{
						"tier": map[string]any{
							"normalized": "super",
						},
					},
				},
			},
			{
				ID:          243,
				Name:        "grok-video-second",
				Platform:    PlatformGrok,
				Type:        AccountTypeAPIKey,
				Status:      StatusActive,
				Schedulable: true,
				Credentials: map[string]any{
					"api_key":  "sk-video-second",
					"base_url": "https://grok.example/v1",
				},
				Extra: map[string]any{
					"grok": map[string]any{
						"tier": map[string]any{
							"normalized": "super",
						},
					},
				},
			},
		},
	}
	svc := NewGrokMediaService(&GatewayService{
		accountRepo:  repo,
		cfg:          testConfig(),
		httpUpstream: upstream,
	}, videoJobs, mediaAssets)

	c, rec := newGrokMediaTestContext(http.MethodPost, "/v1/videos", body)
	handled := svc.HandleVideos(c, nil, body)

	require.True(t, handled)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, upstream.requests, 2)
	require.Equal(t, "Bearer sk-video-first", upstream.requests[0].Header.Get("Authorization"))
	require.Equal(t, "Bearer sk-video-second", upstream.requests[1].Header.Get("Authorization"))
	require.Len(t, videoJobs.upserts, 1)
	require.Equal(t, int64(243), videoJobs.upserts[0].AccountID)
	require.Equal(t, "job_456", videoJobs.upserts[0].JobID)
	require.Len(t, repo.runtimeStates, 2)
	require.Equal(t, "failover", repo.runtimeStates[0]["last_outcome"])
	require.Equal(t, "success", repo.runtimeStates[1]["last_outcome"])
	require.Equal(t, "video", repo.runtimeStates[0]["last_request_capability"])
	require.Equal(t, "video", repo.runtimeStates[1]["last_request_capability"])
}

func TestGrokMediaServiceHandleVideos_FollowupUsesBoundAccountAndUpdatesStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	selectionCalls := 0
	upstream := &queuedHTTPUpstream{
		responses: []*http.Response{
			newJSONResponse(http.StatusOK, `{"id":"job_123","status":"completed"}`),
		},
	}
	videoJobs := &stubGrokVideoJobRepository{
		records: map[string]GrokVideoJobRecord{
			"job_123": {
				JobID:          "job_123",
				AccountID:      77,
				CanonicalModel: "grok-imagine-video",
			},
		},
	}
	mediaAssets := &stubGrokMediaAssetRepository{}
	repo := &mockAccountRepoForPlatform{
		accountsByID: map[int64]*Account{
			77: accountPtr(newSchedulableGrokMediaAccount(77, map[string]any{
				"grok": map[string]any{
					"tier": map[string]any{
						"normalized": "super",
					},
				},
			})),
		},
		listPlatformFunc: func(context.Context, string) ([]Account, error) {
			selectionCalls++
			return nil, errors.New("unexpected account reselection")
		},
	}
	svc := NewGrokMediaService(&GatewayService{
		accountRepo:  repo,
		cfg:          testConfig(),
		httpUpstream: upstream,
	}, videoJobs, mediaAssets)

	c, rec := newGrokMediaTestContext(http.MethodGet, "/v1/videos/job_123", nil)
	handled := svc.HandleVideos(c, nil, nil)

	require.True(t, handled)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, 0, selectionCalls)
	require.Equal(t, 1, repo.getByIDCalls)
	require.Equal(t, []string{"job_123"}, videoJobs.getCalls)
	require.Len(t, videoJobs.patches, 1)
	require.Equal(t, "job_123", videoJobs.patches[0].JobID)
	require.Equal(t, "completed", videoJobs.patches[0].UpstreamStatus)
	require.Equal(t, "completed", videoJobs.patches[0].NormalizedStatus)
	require.Equal(t, "", videoJobs.patches[0].OutputAssetID)
	require.Len(t, upstream.requests, 1)
	require.Equal(t, "/v1/videos/job_123", upstream.requests[0].URL.Path)
	require.Equal(t, "Bearer sk-grok-media", upstream.requests[0].Header.Get("Authorization"))
}

func TestGrokMediaServiceHandleVideos_MissingJobBindingReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	upstream := &queuedHTTPUpstream{}
	videoJobs := &stubGrokVideoJobRepository{
		getErr: ErrGrokVideoJobNotFound,
	}
	svc := NewGrokMediaService(&GatewayService{
		accountRepo:  &mockAccountRepoForPlatform{},
		cfg:          testConfig(),
		httpUpstream: upstream,
	}, videoJobs, nil)

	c, rec := newGrokMediaTestContext(http.MethodGet, "/v1/videos/job_missing", nil)
	handled := svc.HandleVideos(c, nil, nil)

	require.True(t, handled)
	require.Equal(t, http.StatusNotFound, rec.Code)
	require.JSONEq(t, `{"error":{"code":"not_found_error","message":"Grok video job is not known to this gateway"}}`, rec.Body.String())
	require.Empty(t, upstream.requests)
}

func TestGrokMediaServiceHandleVideos_ContentFollowupDoesNotRewriteStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	upstream := &queuedHTTPUpstream{
		responses: []*http.Response{
			newJSONResponse(http.StatusOK, `{"content_url":"https://media.example/job_123.mp4"}`),
		},
	}
	videoJobs := &stubGrokVideoJobRepository{
		records: map[string]GrokVideoJobRecord{
			"job_123": {
				JobID:          "job_123",
				AccountID:      88,
				CanonicalModel: "grok-imagine-video",
			},
		},
	}
	mediaAssets := &stubGrokMediaAssetRepository{}
	repo := &mockAccountRepoForPlatform{
		accountsByID: map[int64]*Account{
			88: accountPtr(newSchedulableGrokMediaAccount(88, map[string]any{
				"grok": map[string]any{
					"tier": map[string]any{
						"normalized": "super",
					},
				},
			})),
		},
	}
	svc := NewGrokMediaService(&GatewayService{
		accountRepo:  repo,
		cfg:          testConfig(),
		httpUpstream: upstream,
	}, videoJobs, mediaAssets)

	c, rec := newGrokMediaTestContext(http.MethodGet, "/grok/v1/videos/job_123/content", nil)
	handled := svc.HandleVideos(c, nil, nil)

	require.True(t, handled)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Empty(t, videoJobs.patches)
	require.Len(t, upstream.requests, 1)
	require.Equal(t, "/v1/videos/job_123/content", upstream.requests[0].URL.Path)
}

func TestGrokMediaServiceHandleImages_SessionAccountUsesProviderTransport(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := []byte(`{"model":"grok-2-image","prompt":"cat"}`)
	upstream := &queuedHTTPUpstream{
		responses: []*http.Response{
			newGrokSessionImageGenerationResponse("https://media.example/image.png"),
		},
	}
	repo := &mockAccountRepoForPlatform{
		accounts: []Account{
			newSchedulableGrokSessionMediaAccount(91, map[string]any{
				"grok": map[string]any{
					"tier": map[string]any{
						"normalized": "basic",
					},
				},
			}),
		},
	}
	svc := NewGrokMediaService(&GatewayService{
		accountRepo:  repo,
		cfg:          testConfig(),
		httpUpstream: upstream,
	}, nil, nil)

	c, rec := newGrokMediaTestContext(http.MethodPost, "/v1/images/generations", body)
	handled := svc.HandleImages(c, nil, body)

	require.True(t, handled)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, upstream.requests, 1)
	require.Equal(t, "https://grok.com/rest/app-chat/conversations/new", upstream.requests[0].URL.String())
	require.Equal(t, "sso=session-token-91; sso-rw=session-token-91-rw", upstream.requests[0].Header.Get("Cookie"))
	require.Equal(t, grokWebBaseURL, upstream.requests[0].Header.Get("Origin"))
	require.Equal(t, grokWebBaseURL+"/", upstream.requests[0].Header.Get("Referer"))
	require.Equal(t, grokSessionProbeUserAgent, upstream.requests[0].Header.Get("User-Agent"))

	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(readTestRequestBody(t, upstream.requests[0])), &payload))
	require.Equal(t, "Drawing: cat", payload["message"])
	require.Equal(t, true, payload["disableTextFollowUps"])
	require.Equal(t, true, payload["enableImageGeneration"])
	require.Equal(t, true, payload["enableImageStreaming"])
	require.Equal(t, false, payload["returnImageBytes"])
	require.Equal(t, false, payload["returnRawGrokInXaiRequest"])
	require.Equal(t, float64(1), payload["imageGenerationCount"])
	require.Equal(t, grokSessionModeFast, payload["modeId"])
	require.Equal(t, "https://media.example/image.png", gjson.Get(rec.Body.String(), "data.0.url").String())
}

func TestGrokMediaServiceHandleImages_SessionAccountUsesConfiguredRuntimeSessionBaseURL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := []byte(`{"model":"grok-2-image","prompt":"cat"}`)
	upstream := &queuedHTTPUpstream{
		responses: []*http.Response{
			newGrokSessionImageGenerationResponse("https://media.example/image.png"),
		},
	}
	repo := &mockAccountRepoForPlatform{
		accounts: []Account{
			newSchedulableGrokSessionMediaAccount(191, map[string]any{
				"grok": map[string]any{
					"tier": map[string]any{
						"normalized": "basic",
					},
				},
			}),
		},
	}
	settingService := NewSettingService(&grokMediaSettingRepoStub{
		values: map[string]string{
			SettingKeyGrokSessionBaseURL: "https://session.grok.example/root",
		},
	}, testConfig())
	svc := NewGrokMediaService(&GatewayService{
		accountRepo:    repo,
		cfg:            testConfig(),
		httpUpstream:   upstream,
		settingService: settingService,
	}, nil, nil)

	c, rec := newGrokMediaTestContext(http.MethodPost, "/v1/images/generations", body)
	handled := svc.HandleImages(c, nil, body)

	require.True(t, handled)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, upstream.requests, 1)
	require.Equal(t, "https://session.grok.example/root/rest/app-chat/conversations/new", upstream.requests[0].URL.String())
	require.Equal(t, "https://session.grok.example/root", upstream.requests[0].Header.Get("Origin"))
	require.Equal(t, "https://session.grok.example/root/", upstream.requests[0].Header.Get("Referer"))
	require.Equal(t, "https://media.example/image.png", gjson.Get(rec.Body.String(), "data.0.url").String())
}

func TestGrokMediaServiceHandleImages_SessionImageEditForwardsAllImageReferences(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := []byte(`{
		"model":"grok-imagine-image-edit",
		"prompt":"replace the sky",
		"image":[
			"data:image/png;base64,QUFB",
			"data:image/png;base64,QkJC"
		]
	}`)
	upstream := &queuedHTTPUpstream{
		responses: []*http.Response{
			newJSONResponse(http.StatusOK, `{"fileMetadataId":"file_1","fileUri":"https://assets.grok.com/users/u_1/file_1/content"}`),
			newJSONResponse(http.StatusOK, `{"fileMetadataId":"file_2","fileUri":"https://assets.grok.com/users/u_1/file_2/content"}`),
			newJSONResponse(http.StatusOK, `{"post":{"id":"post_123"}}`),
			newJSONResponse(http.StatusOK, strings.Join([]string{
				`{"result":{"response":{"streamingImageGenerationResponse":{"progress":100,"imageUrl":"https://media.example/edited.png"}}}}`,
			}, "\n")),
		},
	}
	repo := &mockAccountRepoForPlatform{
		accounts: []Account{
			newSchedulableGrokSessionMediaAccount(291, map[string]any{
				"grok": map[string]any{
					"tier": map[string]any{
						"normalized": "super",
					},
				},
			}),
		},
	}
	svc := NewGrokMediaService(&GatewayService{
		accountRepo:  repo,
		cfg:          testConfig(),
		httpUpstream: upstream,
	}, nil, nil)

	c, rec := newGrokMediaTestContext(http.MethodPost, "/v1/images/edits", body)
	handled := svc.HandleImages(c, nil, body)

	require.True(t, handled)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, upstream.requests, 4)
	require.Equal(t, "https://grok.com/rest/app-chat/upload-file", upstream.requests[0].URL.String())
	require.Equal(t, "https://grok.com/rest/app-chat/upload-file", upstream.requests[1].URL.String())
	require.Equal(t, "https://grok.com/rest/media/post/create", upstream.requests[2].URL.String())
	require.Equal(t, "https://grok.com/rest/app-chat/conversations/new", upstream.requests[3].URL.String())

	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(readTestRequestBody(t, upstream.requests[3])), &payload))
	require.Equal(t, "imagine-image-edit", payload["modelName"])
	require.Equal(t, float64(2), payload["imageGenerationCount"])

	references := gjson.GetBytes([]byte(readTestRequestBody(t, upstream.requests[3])), "responseMetadata.modelConfigOverride.modelMap.imageEditModelConfig.imageReferences").Array()
	require.Len(t, references, 2)
	require.Equal(t, "https://assets.grok.com/users/u_1/file_1/content", references[0].String())
	require.Equal(t, "https://assets.grok.com/users/u_1/file_2/content", references[1].String())
	require.Equal(t, "post_123", gjson.GetBytes([]byte(readTestRequestBody(t, upstream.requests[3])), "responseMetadata.modelConfigOverride.modelMap.imageEditModelConfig.parentPostId").String())
	require.Equal(t, "https://media.example/edited.png", gjson.Get(rec.Body.String(), "data.0.url").String())
}

func TestGrokMediaServiceHandleImages_SessionImageEditRetriesToFillMissingFinals(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := []byte(`{
		"model":"grok-imagine-image-edit",
		"prompt":"replace background",
		"n":2,
		"image":[
			"data:image/png;base64,QUFB",
			"data:image/png;base64,QkJC"
		]
	}`)
	upstream := &queuedHTTPUpstream{
		responses: []*http.Response{
			newJSONResponse(http.StatusOK, `{"fileMetadataId":"file_1","fileUri":"https://assets.grok.com/users/u_1/file_1/content"}`),
			newJSONResponse(http.StatusOK, `{"fileMetadataId":"file_2","fileUri":"https://assets.grok.com/users/u_1/file_2/content"}`),
			newJSONResponse(http.StatusOK, `{"post":{"id":"post_123"}}`),
			newJSONResponse(http.StatusOK, strings.Join([]string{
				`{"result":{"response":{"streamingImageGenerationResponse":{"imageIndex":0,"progress":100,"imageUrl":"https://media.example/edited-a.png"}}}}`,
			}, "\n")),
			newJSONResponse(http.StatusOK, strings.Join([]string{
				`{"result":{"response":{"streamingImageGenerationResponse":{"imageIndex":0,"progress":100,"imageUrl":"https://media.example/edited-a.png"}}}}`,
				`{"result":{"response":{"streamingImageGenerationResponse":{"imageIndex":1,"progress":100,"imageUrl":"https://media.example/edited-b.png"}}}}`,
			}, "\n")),
		},
	}
	repo := &mockAccountRepoForPlatform{
		accounts: []Account{
			newSchedulableGrokSessionMediaAccount(291, map[string]any{
				"grok": map[string]any{
					"tier": map[string]any{
						"normalized": "super",
					},
				},
			}),
		},
	}
	svc := NewGrokMediaService(&GatewayService{
		accountRepo:  repo,
		cfg:          testConfig(),
		httpUpstream: upstream,
	}, nil, nil)

	c, rec := newGrokMediaTestContext(http.MethodPost, "/v1/images/edits", body)
	handled := svc.HandleImages(c, nil, body)

	require.True(t, handled)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, upstream.requests, 5)
	require.Equal(t, "https://grok.com/rest/app-chat/upload-file", upstream.requests[0].URL.String())
	require.Equal(t, "https://grok.com/rest/app-chat/upload-file", upstream.requests[1].URL.String())
	require.Equal(t, "https://grok.com/rest/media/post/create", upstream.requests[2].URL.String())
	require.Equal(t, "https://grok.com/rest/app-chat/conversations/new", upstream.requests[3].URL.String())
	require.Equal(t, "https://grok.com/rest/app-chat/conversations/new", upstream.requests[4].URL.String())

	require.Equal(t, float64(2), gjson.GetBytes([]byte(readTestRequestBody(t, upstream.requests[3])), "imageGenerationCount").Num)
	require.Equal(t, float64(2), gjson.GetBytes([]byte(readTestRequestBody(t, upstream.requests[4])), "imageGenerationCount").Num)
	require.Equal(t, "https://media.example/edited-a.png", gjson.Get(rec.Body.String(), "data.0.url").String())
	require.Equal(t, "https://media.example/edited-b.png", gjson.Get(rec.Body.String(), "data.1.url").String())
}

func TestGrokMediaServiceHandleVideos_FollowupUsesBoundSessionAccount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	videoJobs := &stubGrokVideoJobRepository{
		records: map[string]GrokVideoJobRecord{
			"job_123": {
				JobID:            "job_123",
				AccountID:        177,
				CanonicalModel:   "grok-imagine-video",
				NormalizedStatus: "completed",
				UpstreamStatus:   "completed",
				CreatedAt:        time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC),
				UpdatedAt:        time.Date(2026, 4, 20, 12, 6, 0, 0, time.UTC),
			},
		},
	}
	repo := &mockAccountRepoForPlatform{
		accountsByID: map[int64]*Account{
			177: accountPtr(newSchedulableGrokSessionMediaAccount(177, map[string]any{
				"grok": map[string]any{
					"tier": map[string]any{
						"normalized": "super",
					},
				},
			})),
		},
	}
	svc := NewGrokMediaService(&GatewayService{
		accountRepo:  repo,
		cfg:          testConfig(),
		httpUpstream: &queuedHTTPUpstream{},
	}, videoJobs, nil)

	c, rec := newGrokMediaTestContext(http.MethodGet, "/v1/videos/job_123", nil)
	handled := svc.HandleVideos(c, nil, nil)

	require.True(t, handled)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, 1, repo.getByIDCalls)
	require.Empty(t, videoJobs.patches)
	require.JSONEq(t, `{"id":"job_123","object":"video","created_at":1776686400,"status":"completed","model":"grok-imagine-video","progress":100,"prompt":"","seconds":"6","size":"720x1280","quality":"standard","completed_at":1776686760}`, rec.Body.String())
}

func TestGrokMediaServiceHandleImages_RewritesToLocalProxyAssetURL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := []byte(`{"model":"grok-2-image","prompt":"cat"}`)
	upstream := &queuedHTTPUpstream{
		responses: []*http.Response{
			newJSONResponse(http.StatusOK, `{"data":[{"url":"https://media.example/image.png"}]}`),
		},
	}
	repo := &mockAccountRepoForPlatform{
		accounts: []Account{
			newSchedulableGrokMediaAccount(41, map[string]any{
				"grok": map[string]any{
					"tier": map[string]any{
						"normalized": "basic",
					},
				},
			}),
		},
	}
	mediaAssets := &stubGrokMediaAssetRepository{}
	svc := NewGrokMediaService(&GatewayService{
		accountRepo:  repo,
		cfg:          testConfig(),
		httpUpstream: upstream,
	}, nil, mediaAssets)

	c, rec := newGrokMediaTestContext(http.MethodPost, "/v1/images/generations", body)
	c.Request.Host = "gateway.example"
	c.Request.Header.Set("X-Forwarded-Proto", "https")
	handled := svc.HandleImages(c, nil, body)

	require.True(t, handled)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, mediaAssets.upserts, 1)
	require.Contains(t, rec.Body.String(), `"url":"https://gateway.example/grok/media/assets/`)
	require.Contains(t, rec.Body.String(), `"upstream_url":"https://media.example/image.png"`)
}

func TestGrokMediaServiceHandleImages_UsesConfiguredUpstreamOutputFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := []byte(`{"model":"grok-2-image","prompt":"cat"}`)
	upstream := &queuedHTTPUpstream{
		responses: []*http.Response{
			newJSONResponse(http.StatusOK, `{"data":[{"url":"https://media.example/image.png"}]}`),
		},
	}
	repo := &mockAccountRepoForPlatform{
		accounts: []Account{
			newSchedulableGrokMediaAccount(41, map[string]any{
				"grok": map[string]any{
					"tier": map[string]any{
						"normalized": "basic",
					},
				},
			}),
		},
	}
	mediaAssets := &stubGrokMediaAssetRepository{}
	settingService := NewSettingService(&grokMediaSettingRepoStub{
		values: map[string]string{
			SettingKeyGrokImageOutputFormat: GrokMediaOutputFormatUpstreamURL,
			SettingKeyGrokMediaProxyEnabled: "true",
		},
	}, testConfig())
	svc := NewGrokMediaService(&GatewayService{
		accountRepo:    repo,
		cfg:            testConfig(),
		httpUpstream:   upstream,
		settingService: settingService,
	}, nil, mediaAssets)

	c, rec := newGrokMediaTestContext(http.MethodPost, "/v1/images/generations", body)
	handled := svc.HandleImages(c, nil, body)

	require.True(t, handled)
	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, `{"data":[{"url":"https://media.example/image.png"}]}`, rec.Body.String())
	require.Len(t, mediaAssets.upserts, 0)
}

func TestGrokMediaServiceHandleImages_UsesConfiguredBase64OutputAndRetention(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := []byte(`{"model":"grok-2-image","prompt":"cat"}`)
	upstream := &queuedHTTPUpstream{
		responses: []*http.Response{
			newJSONResponse(http.StatusOK, `{"data":[{"url":"https://media.example/image.png"}]}`),
			{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"image/png"}},
				Body:       io.NopCloser(bytes.NewReader([]byte("PNGDATA"))),
			},
		},
	}
	repo := &mockAccountRepoForPlatform{
		accounts: []Account{
			newSchedulableGrokMediaAccount(41, map[string]any{
				"grok": map[string]any{
					"tier": map[string]any{
						"normalized": "basic",
					},
				},
			}),
		},
	}
	mediaAssets := &stubGrokMediaAssetRepository{}
	settingService := NewSettingService(&grokMediaSettingRepoStub{
		values: map[string]string{
			SettingKeyGrokImageOutputFormat:        GrokMediaOutputFormatBase64,
			SettingKeyGrokMediaProxyEnabled:        "true",
			SettingKeyGrokMediaCacheRetentionHours: "2",
		},
	}, testConfig())
	svc := NewGrokMediaService(&GatewayService{
		accountRepo:    repo,
		cfg:            testConfig(),
		httpUpstream:   upstream,
		settingService: settingService,
	}, nil, mediaAssets)
	now := time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)
	svc.mediaAssets.cacheRoot = t.TempDir()
	svc.mediaAssets.now = func() time.Time { return now }

	c, rec := newGrokMediaTestContext(http.MethodPost, "/v1/images/generations", body)
	handled := svc.HandleImages(c, nil, body)

	require.True(t, handled)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"url":"data:image/png;base64,UE5HREFUQQ=="`)
	require.Contains(t, rec.Body.String(), `"upstream_url":"https://media.example/image.png"`)
	require.Len(t, mediaAssets.upserts, 1)
	require.NotNil(t, mediaAssets.upserts[0].ExpiresAt)
	require.Equal(t, now.Add(2*time.Hour), mediaAssets.upserts[0].ExpiresAt.UTC())
}

func TestGrokMediaServiceHandleImages_ResponseFormatOverrideReturnsB64JSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := []byte(`{"model":"grok-2-image","prompt":"cat","response_format":"b64_json"}`)
	upstream := &queuedHTTPUpstream{
		responses: []*http.Response{
			newJSONResponse(http.StatusOK, `{"data":[{"url":"https://media.example/image.png"}]}`),
			{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"image/png"}},
				Body:       io.NopCloser(bytes.NewReader([]byte("PNGDATA"))),
			},
		},
	}
	repo := &mockAccountRepoForPlatform{
		accounts: []Account{
			newSchedulableGrokMediaAccount(141, map[string]any{
				"grok": map[string]any{
					"tier": map[string]any{
						"normalized": "basic",
					},
				},
			}),
		},
	}
	mediaAssets := &stubGrokMediaAssetRepository{}
	settingService := NewSettingService(&grokMediaSettingRepoStub{
		values: map[string]string{
			SettingKeyGrokImageOutputFormat: GrokMediaOutputFormatUpstreamURL,
			SettingKeyGrokMediaProxyEnabled: "true",
		},
	}, testConfig())
	svc := NewGrokMediaService(&GatewayService{
		accountRepo:    repo,
		cfg:            testConfig(),
		httpUpstream:   upstream,
		settingService: settingService,
	}, nil, mediaAssets)
	svc.mediaAssets.cacheRoot = t.TempDir()

	c, rec := newGrokMediaTestContext(http.MethodPost, "/v1/images/generations", body)
	handled := svc.HandleImages(c, nil, body)

	require.True(t, handled)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "UE5HREFUQQ==", gjson.Get(rec.Body.String(), "data.0.b64_json").String())
	require.False(t, gjson.Get(rec.Body.String(), "data.0.url").Exists())
	require.Len(t, mediaAssets.upserts, 1)
}

func TestGrokMediaServiceHandleImages_ResponseFormatOverrideReturnsURL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := []byte(`{"model":"grok-2-image","prompt":"cat","response_format":"url"}`)
	upstream := &queuedHTTPUpstream{
		responses: []*http.Response{
			newJSONResponse(http.StatusOK, `{"data":[{"url":"https://media.example/image.png"}]}`),
		},
	}
	repo := &mockAccountRepoForPlatform{
		accounts: []Account{
			newSchedulableGrokMediaAccount(142, map[string]any{
				"grok": map[string]any{
					"tier": map[string]any{
						"normalized": "basic",
					},
				},
			}),
		},
	}
	mediaAssets := &stubGrokMediaAssetRepository{}
	settingService := NewSettingService(&grokMediaSettingRepoStub{
		values: map[string]string{
			SettingKeyGrokImageOutputFormat: GrokMediaOutputFormatBase64,
			SettingKeyGrokMediaProxyEnabled: "true",
		},
	}, testConfig())
	svc := NewGrokMediaService(&GatewayService{
		accountRepo:    repo,
		cfg:            testConfig(),
		httpUpstream:   upstream,
		settingService: settingService,
	}, nil, mediaAssets)

	c, rec := newGrokMediaTestContext(http.MethodPost, "/v1/images/generations", body)
	c.Request.Host = "gateway.example"
	c.Request.Header.Set("X-Forwarded-Proto", "https")
	handled := svc.HandleImages(c, nil, body)

	require.True(t, handled)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"url":"https://gateway.example/grok/media/assets/`)
	require.False(t, gjson.Get(rec.Body.String(), "data.0.b64_json").Exists())
	require.Len(t, mediaAssets.upserts, 1)
}

func TestGrokMediaServiceHandleImages_SessionResponseFormatOverrideReturnsB64JSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := []byte(`{"model":"grok-2-image","prompt":"cat","response_format":"b64_json"}`)
	upstream := &queuedHTTPUpstream{
		responses: []*http.Response{
			newGrokSessionImageGenerationResponse("https://media.example/image.png"),
			{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"image/png"}},
				Body:       io.NopCloser(bytes.NewReader([]byte("PNGDATA"))),
			},
		},
	}
	repo := &mockAccountRepoForPlatform{
		accounts: []Account{
			newSchedulableGrokSessionMediaAccount(391, map[string]any{
				"grok": map[string]any{
					"tier": map[string]any{
						"normalized": "basic",
					},
				},
			}),
		},
	}
	mediaAssets := &stubGrokMediaAssetRepository{}
	settingService := NewSettingService(&grokMediaSettingRepoStub{
		values: map[string]string{
			SettingKeyGrokImageOutputFormat: GrokMediaOutputFormatUpstreamURL,
			SettingKeyGrokMediaProxyEnabled: "true",
		},
	}, testConfig())
	svc := NewGrokMediaService(&GatewayService{
		accountRepo:    repo,
		cfg:            testConfig(),
		httpUpstream:   upstream,
		settingService: settingService,
	}, nil, mediaAssets)
	svc.mediaAssets.cacheRoot = t.TempDir()

	c, rec := newGrokMediaTestContext(http.MethodPost, "/v1/images/generations", body)
	handled := svc.HandleImages(c, nil, body)

	require.True(t, handled)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "UE5HREFUQQ==", gjson.Get(rec.Body.String(), "data.0.b64_json").String())
	require.False(t, gjson.Get(rec.Body.String(), "data.0.url").Exists())
	require.Len(t, mediaAssets.upserts, 1)
	require.Len(t, upstream.requests, 2)
	require.Equal(t, "https://media.example/image.png", upstream.requests[1].URL.String())
}

func TestGrokMediaServiceHandleVideos_ContentFollowup_UsesConfiguredHTMLFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	upstream := &queuedHTTPUpstream{
		responses: []*http.Response{
			newJSONResponse(http.StatusOK, `{"content_url":"https://media.example/job_123.mp4"}`),
		},
	}
	videoJobs := &stubGrokVideoJobRepository{
		records: map[string]GrokVideoJobRecord{
			"job_123": {
				JobID:          "job_123",
				AccountID:      88,
				CanonicalModel: "grok-imagine-video",
			},
		},
	}
	mediaAssets := &stubGrokMediaAssetRepository{}
	repo := &mockAccountRepoForPlatform{
		accountsByID: map[int64]*Account{
			88: accountPtr(newSchedulableGrokMediaAccount(88, nil)),
		},
	}
	settingService := NewSettingService(&grokMediaSettingRepoStub{
		values: map[string]string{
			SettingKeyGrokVideoOutputFormat: GrokMediaOutputFormatHTML,
			SettingKeyGrokMediaProxyEnabled: "false",
		},
	}, testConfig())
	svc := NewGrokMediaService(&GatewayService{
		accountRepo:    repo,
		cfg:            testConfig(),
		httpUpstream:   upstream,
		settingService: settingService,
	}, videoJobs, mediaAssets)

	c, rec := newGrokMediaTestContext(http.MethodGet, "/grok/v1/videos/job_123/content", nil)
	handled := svc.HandleVideos(c, nil, nil)

	require.True(t, handled)
	require.Equal(t, http.StatusOK, rec.Code)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	require.Equal(t, `<video controls src="https://media.example/job_123.mp4"></video>`, payload["content_url"])
	require.Equal(t, "https://media.example/job_123.mp4", payload["upstream_content_url"])
	require.Empty(t, mediaAssets.upserts)
}

func TestGrokMediaServiceHandleMediaAssetContent_DownloadsAndCachesProxyAsset(t *testing.T) {
	gin.SetMode(gin.TestMode)

	upstream := &queuedHTTPUpstream{
		responses: []*http.Response{
			{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"image/png"}},
				Body:       io.NopCloser(bytes.NewReader([]byte("PNGDATA"))),
			},
		},
	}
	mediaAssets := &stubGrokMediaAssetRepository{
		records: map[string]GrokMediaAssetRecord{
			"asset_123": {
				AssetID:     "asset_123",
				AccountID:   88,
				AssetType:   "image",
				UpstreamURL: "https://media.example/image.png",
			},
		},
	}
	repo := &mockAccountRepoForPlatform{
		accountsByID: map[int64]*Account{
			88: accountPtr(newSchedulableGrokMediaAccount(88, nil)),
		},
	}
	cacheRoot := t.TempDir()
	svc := NewGrokMediaService(&GatewayService{
		accountRepo:  repo,
		cfg:          testConfig(),
		httpUpstream: upstream,
	}, nil, mediaAssets)
	svc.mediaAssets.cacheRoot = cacheRoot

	c, rec := newGrokMediaTestContext(http.MethodGet, "/grok/media/assets/asset_123", nil)
	handled := svc.HandleAssetContent(c, "asset_123")

	require.True(t, handled)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "image/png", rec.Header().Get("Content-Type"))
	require.Equal(t, "PNGDATA", rec.Body.String())
	require.Len(t, mediaAssets.patches, 1)
	require.NotEmpty(t, mediaAssets.records["asset_123"].LocalPath)
	require.FileExists(t, mediaAssets.records["asset_123"].LocalPath)
	require.Len(t, mediaAssets.accesses, 1)
}

func TestGrokMediaServiceHandleMediaAssetContent_ReusesCachedFileForDuplicateContent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	upstream := &queuedHTTPUpstream{
		responses: []*http.Response{
			{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"image/png"}},
				Body:       io.NopCloser(bytes.NewReader([]byte("SAMEPNG"))),
			},
			{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"image/png"}},
				Body:       io.NopCloser(bytes.NewReader([]byte("SAMEPNG"))),
			},
		},
	}
	mediaAssets := &stubGrokMediaAssetRepository{
		records: map[string]GrokMediaAssetRecord{
			"asset_a": {
				AssetID:     "asset_a",
				AccountID:   88,
				AssetType:   "image",
				UpstreamURL: "https://media.example/image-a.png",
			},
			"asset_b": {
				AssetID:     "asset_b",
				AccountID:   88,
				AssetType:   "image",
				UpstreamURL: "https://media.example/image-b.png",
			},
		},
	}
	repo := &mockAccountRepoForPlatform{
		accountsByID: map[int64]*Account{
			88: accountPtr(newSchedulableGrokMediaAccount(88, nil)),
		},
	}
	svc := NewGrokMediaService(&GatewayService{
		accountRepo:  repo,
		cfg:          testConfig(),
		httpUpstream: upstream,
	}, nil, mediaAssets)
	svc.mediaAssets.cacheRoot = t.TempDir()

	firstCtx, _ := newGrokMediaTestContext(http.MethodGet, "/grok/media/assets/asset_a", nil)
	require.True(t, svc.HandleAssetContent(firstCtx, "asset_a"))

	secondCtx, _ := newGrokMediaTestContext(http.MethodGet, "/grok/media/assets/asset_b", nil)
	require.True(t, svc.HandleAssetContent(secondCtx, "asset_b"))

	require.Len(t, mediaAssets.patches, 2)
	require.NotEmpty(t, mediaAssets.records["asset_a"].LocalPath)
	require.Equal(t, mediaAssets.records["asset_a"].LocalPath, mediaAssets.records["asset_b"].LocalPath)
	require.FileExists(t, mediaAssets.records["asset_a"].LocalPath)
	require.Len(t, upstream.requests, 2)
	require.NotEmpty(t, mediaAssets.hashLookups)
}

func TestGrokMediaAssetServiceCleanupExpiredNow_RemovesOnlyUnreferencedFiles(t *testing.T) {
	now := time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)
	cacheRoot := t.TempDir()
	sharedPath := cacheRoot + "/shared.png"
	orphanPath := cacheRoot + "/orphan.png"
	require.NoError(t, os.WriteFile(sharedPath, []byte("shared"), 0o644))
	require.NoError(t, os.WriteFile(orphanPath, []byte("orphan"), 0o644))

	repo := &stubGrokMediaAssetRepository{
		records: map[string]GrokMediaAssetRecord{
			"expired_shared": {
				AssetID:     "expired_shared",
				LocalPath:   sharedPath,
				Status:      grokMediaAssetStatusReady,
				ExpiresAt:   grokTestTimePtr(now.Add(-time.Hour)),
				ContentHash: "hash-shared",
			},
			"active_shared": {
				AssetID:     "active_shared",
				LocalPath:   sharedPath,
				Status:      grokMediaAssetStatusReady,
				ExpiresAt:   grokTestTimePtr(now.Add(time.Hour)),
				ContentHash: "hash-shared",
			},
			"expired_orphan": {
				AssetID:     "expired_orphan",
				LocalPath:   orphanPath,
				Status:      grokMediaAssetStatusReady,
				ExpiresAt:   grokTestTimePtr(now.Add(-2 * time.Hour)),
				ContentHash: "hash-orphan",
			},
		},
	}
	svc := NewGrokMediaAssetService(&GatewayService{cfg: testConfig()}, repo)
	svc.cacheRoot = cacheRoot
	svc.now = func() time.Time { return now }

	require.NoError(t, svc.CleanupExpiredNow(context.Background()))

	_, sharedErr := os.Stat(sharedPath)
	require.NoError(t, sharedErr)
	_, orphanErr := os.Stat(orphanPath)
	require.True(t, errors.Is(orphanErr, os.ErrNotExist))
	_, expiredSharedStillPresent := repo.records["expired_shared"]
	require.False(t, expiredSharedStillPresent)
	_, expiredOrphanStillPresent := repo.records["expired_orphan"]
	require.False(t, expiredOrphanStillPresent)
	_, activeSharedStillPresent := repo.records["active_shared"]
	require.True(t, activeSharedStillPresent)
}

func newGrokMediaTestContext(method, path string, body []byte) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	var reader io.Reader
	if len(body) > 0 {
		reader = bytes.NewReader(body)
	}
	c.Request = httptest.NewRequest(method, path, reader)
	if len(body) > 0 {
		c.Request.Header.Set("Content-Type", "application/json")
	}
	return c, rec
}

func newGrokSessionImageGenerationResponse(url string) *http.Response {
	payload := strings.Join([]string{
		`{"result":{"response":{"token":"partial"}}}`,
		`{"result":{"response":{"modelResponse":{"generatedImageUrls":["` + url + `"]}}}}`,
	}, "\n")
	return newJSONResponse(http.StatusOK, payload)
}

func newSchedulableGrokMediaAccount(id int64, extra map[string]any) Account {
	return Account{
		ID:          id,
		Name:        "grok-media",
		Platform:    PlatformGrok,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: true,
		Credentials: map[string]any{
			"api_key":  "sk-grok-media",
			"base_url": "https://grok.example/v1",
		},
		Extra: extra,
	}
}

func newSchedulableGrokSessionMediaAccount(id int64, extra map[string]any) Account {
	return Account{
		ID:          id,
		Name:        "grok-media-session",
		Platform:    PlatformGrok,
		Type:        AccountTypeSession,
		Status:      StatusActive,
		Schedulable: true,
		Credentials: map[string]any{
			"session_token": "sso=session-token-" + strconv.FormatInt(id, 10) + "; sso-rw=session-token-" + strconv.FormatInt(id, 10) + "-rw",
		},
		Extra: extra,
	}
}

func accountPtr(account Account) *Account {
	accountCopy := account
	return &accountCopy
}

func grokTestTimePtr(ts time.Time) *time.Time {
	value := ts
	return &value
}

func readTestRequestBody(t *testing.T, req *http.Request) string {
	t.Helper()
	if req == nil || req.Body == nil {
		return ""
	}
	raw, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	require.NoError(t, req.Body.Close())
	req.Body = io.NopCloser(bytes.NewReader(raw))
	return string(raw)
}

func TestExtractGrokVideoPollAfter_NumberParsesIntoFutureTime(t *testing.T) {
	before := time.Now().UTC()
	pollAfter := extractGrokVideoPollAfter([]byte(`{"poll_after":5}`))
	after := time.Now().UTC()

	require.NotNil(t, pollAfter)
	require.False(t, pollAfter.Before(before.Add(4*time.Second)))
	require.False(t, pollAfter.After(after.Add(6*time.Second)))
}
