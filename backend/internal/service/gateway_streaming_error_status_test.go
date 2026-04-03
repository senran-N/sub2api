package service

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestInferStreamingErrorStatusCode(t *testing.T) {
	tests := []struct {
		name       string
		httpStatus int
		body       string
		want       int
	}{
		{
			name:       "explicit numeric status in payload wins",
			httpStatus: http.StatusOK,
			body:       `{"error":{"status_code":429}}`,
			want:       http.StatusTooManyRequests,
		},
		{
			name:       "nested JSON string message is parsed",
			httpStatus: http.StatusOK,
			body:       `{"error":{"message":"{\"statusCode\":\"404\"}"}}`,
			want:       http.StatusNotFound,
		},
		{
			name:       "error type maps to status",
			httpStatus: http.StatusOK,
			body:       `{"error":{"type":"permission_error"}}`,
			want:       http.StatusForbidden,
		},
		{
			name:       "http error status used when payload has none",
			httpStatus: http.StatusBadGateway,
			body:       `{"error":{"message":"upstream failed"}}`,
			want:       http.StatusBadGateway,
		},
		{
			name:       "sse error defaults to forbidden",
			httpStatus: http.StatusOK,
			body:       `{"error":{"message":"stream failed"}}`,
			want:       http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inferStreamingErrorStatusCode(tt.httpStatus, []byte(tt.body))
			require.Equal(t, tt.want, got)
		})
	}
}

func TestParseStreamingErrorStatusValue(t *testing.T) {
	require.Equal(t, 429, parseStreamingErrorStatusValue(mustGJSON(`429`)))
	require.Equal(t, 403, parseStreamingErrorStatusValue(mustGJSON(`"403"`)))
	require.Equal(t, 0, parseStreamingErrorStatusValue(mustGJSON(`"bad"`)))
}

func mustGJSON(raw string) gjson.Result {
	return gjson.Parse(raw)
}
