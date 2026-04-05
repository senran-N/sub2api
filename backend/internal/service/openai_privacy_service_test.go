package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/imroc/req/v3"
	"github.com/stretchr/testify/require"
)

func TestDisableOpenAITraining_UsesAPIRequestHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method)
		require.Equal(t, "/backend-api/settings/account_user_setting", r.URL.Path)
		require.Equal(t, "application/json", r.Header.Get("Accept"))
		require.Equal(t, "cors", r.Header.Get("sec-fetch-mode"))
		require.Equal(t, "same-origin", r.Header.Get("sec-fetch-site"))
		require.Equal(t, "empty", r.Header.Get("sec-fetch-dest"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	target, err := url.Parse(server.URL)
	require.NoError(t, err)

	mode := disableOpenAITraining(context.Background(), func(proxyURL string) (*req.Client, error) {
		client := req.C()
		client.GetTransport().WrapRoundTripFunc(func(rt http.RoundTripper) req.HttpRoundTripFunc {
			return func(r *http.Request) (*http.Response, error) {
				if r.URL.String() == openAISettingsURL+"?feature=training_allowed&value=false" {
					r.URL.Scheme = target.Scheme
					r.URL.Host = target.Host
					r.Host = target.Host
				}
				return rt.RoundTrip(r)
			}
		})
		return client, nil
	}, "token-1", "")

	require.Equal(t, PrivacyModeTrainingOff, mode)
}
