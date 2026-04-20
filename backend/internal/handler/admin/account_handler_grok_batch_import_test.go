package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestAccountHandlerBatchImportGrokSession_DoesNotEchoRawInput(t *testing.T) {
	gin.SetMode(gin.TestMode)

	adminSvc := newStubAdminService()
	adminSvc.batchImportResult = &service.GrokSessionBatchImportResult{
		Total:   1,
		Created: 1,
		Results: []service.GrokSessionBatchImportLineResult{
			{
				Line:        1,
				Name:        "grok-sso-001",
				Success:     true,
				AccountID:   101,
				Fingerprint: "sha256:ab12cd34...ef56",
			},
		},
	}
	handler := NewAccountHandler(adminSvc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	body := map[string]any{
		"raw_input":         "super-secret-token",
		"name_prefix":       "grok-sso",
		"dry_run":           false,
		"test_after_create": false,
		"dedupe_strategy":   "skip_existing",
	}
	payload, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts/grok/session/batch-import", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	handler.BatchImportGrokSession(ctx)

	require.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, adminSvc.lastBatchImportInput)
	require.Equal(t, "super-secret-token", adminSvc.lastBatchImportInput.RawInput)
	require.NotContains(t, w.Body.String(), "super-secret-token")
	require.Contains(t, w.Body.String(), "sha256:ab12cd34...ef56")
}
