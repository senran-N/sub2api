package service

import (
	"errors"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/util/urlvalidator"
)

func validateCompatibleUpstreamBaseURL(cfg *config.Config, raw string) (string, error) {
	if cfg == nil {
		return "", errors.New("config is not available")
	}
	if !cfg.Security.URLAllowlist.Enabled {
		return urlvalidator.ValidateURLFormat(raw, cfg.Security.URLAllowlist.AllowInsecureHTTP)
	}
	return urlvalidator.ValidateHTTPSURL(raw, urlvalidator.ValidationOptions{
		AllowedHosts:     cfg.Security.URLAllowlist.UpstreamHosts,
		RequireAllowlist: true,
		AllowPrivate:     cfg.Security.URLAllowlist.AllowPrivateHosts,
	})
}
