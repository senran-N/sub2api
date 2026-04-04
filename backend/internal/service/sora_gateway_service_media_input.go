package service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const soraImageInputMaxBytes = 20 << 20
const soraImageInputMaxRedirects = 3
const soraImageInputTimeout = 20 * time.Second
const soraVideoInputMaxBytes = 200 << 20
const soraVideoInputMaxRedirects = 3
const soraVideoInputTimeout = 60 * time.Second

var soraBlockedHostnames = map[string]struct{}{
	"localhost":                 {},
	"localhost.localdomain":     {},
	"metadata.google.internal":  {},
	"metadata.google.internal.": {},
}

var soraBlockedCIDRs = mustParseCIDRs([]string{
	"0.0.0.0/8",
	"10.0.0.0/8",
	"100.64.0.0/10",
	"127.0.0.0/8",
	"169.254.0.0/16",
	"172.16.0.0/12",
	"192.168.0.0/16",
	"224.0.0.0/4",
	"240.0.0.0/4",
	"::/128",
	"::1/128",
	"fc00::/7",
	"fe80::/10",
})

func decodeSoraImageInput(ctx context.Context, input string) ([]byte, string, error) {
	raw := strings.TrimSpace(input)
	if raw == "" {
		return nil, "", errors.New("empty image input")
	}
	if strings.HasPrefix(raw, "data:") {
		parts := strings.SplitN(raw, ",", 2)
		if len(parts) != 2 {
			return nil, "", errors.New("invalid data url")
		}
		meta := parts[0]
		payload := parts[1]
		decoded, err := decodeBase64WithLimit(payload, soraImageInputMaxBytes)
		if err != nil {
			return nil, "", err
		}
		ext := ""
		if strings.HasPrefix(meta, "data:") {
			metaParts := strings.SplitN(meta[5:], ";", 2)
			if len(metaParts) > 0 {
				if exts, err := mime.ExtensionsByType(metaParts[0]); err == nil && len(exts) > 0 {
					ext = exts[0]
				}
			}
		}
		filename := "image" + ext
		return decoded, filename, nil
	}
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return downloadSoraImageInput(ctx, raw)
	}
	decoded, err := decodeBase64WithLimit(raw, soraImageInputMaxBytes)
	if err != nil {
		return nil, "", errors.New("invalid base64 image")
	}
	return decoded, "image.png", nil
}

func decodeSoraVideoInput(ctx context.Context, input string) ([]byte, error) {
	raw := strings.TrimSpace(input)
	if raw == "" {
		return nil, errors.New("empty video input")
	}
	if strings.HasPrefix(raw, "data:") {
		parts := strings.SplitN(raw, ",", 2)
		if len(parts) != 2 {
			return nil, errors.New("invalid video data url")
		}
		decoded, err := decodeBase64WithLimit(parts[1], soraVideoInputMaxBytes)
		if err != nil {
			return nil, errors.New("invalid base64 video")
		}
		if len(decoded) == 0 {
			return nil, errors.New("empty video data")
		}
		return decoded, nil
	}
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return downloadSoraVideoInput(ctx, raw)
	}
	decoded, err := decodeBase64WithLimit(raw, soraVideoInputMaxBytes)
	if err != nil {
		return nil, errors.New("invalid base64 video")
	}
	if len(decoded) == 0 {
		return nil, errors.New("empty video data")
	}
	return decoded, nil
}

func downloadSoraImageInput(ctx context.Context, rawURL string) ([]byte, string, error) {
	parsed, err := validateSoraRemoteURL(rawURL)
	if err != nil {
		return nil, "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, "", err
	}
	client := &http.Client{
		Timeout: soraImageInputTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= soraImageInputMaxRedirects {
				return errors.New("too many redirects")
			}
			return validateSoraRemoteURLValue(req.URL)
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("download image failed: %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, soraImageInputMaxBytes))
	if err != nil {
		return nil, "", err
	}
	ext := fileExtFromURL(parsed.String())
	if ext == "" {
		ext = fileExtFromContentType(resp.Header.Get("Content-Type"))
	}
	filename := "image" + ext
	return data, filename, nil
}

func downloadSoraVideoInput(ctx context.Context, rawURL string) ([]byte, error) {
	parsed, err := validateSoraRemoteURL(rawURL)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Timeout: soraVideoInputTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= soraVideoInputMaxRedirects {
				return errors.New("too many redirects")
			}
			return validateSoraRemoteURLValue(req.URL)
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download video failed: %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, soraVideoInputMaxBytes))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New("empty video content")
	}
	return data, nil
}

func decodeBase64WithLimit(encoded string, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		return nil, errors.New("invalid max bytes limit")
	}
	decoder := base64.NewDecoder(base64.StdEncoding, strings.NewReader(encoded))
	limited := io.LimitReader(decoder, maxBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("input exceeds %d bytes limit", maxBytes)
	}
	return data, nil
}

func validateSoraRemoteURL(raw string) (*url.URL, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, errors.New("empty remote url")
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid remote url: %w", err)
	}
	if err := validateSoraRemoteURLValue(parsed); err != nil {
		return nil, err
	}
	return parsed, nil
}

func validateSoraRemoteURLValue(parsed *url.URL) error {
	if parsed == nil {
		return errors.New("invalid remote url")
	}
	scheme := strings.ToLower(strings.TrimSpace(parsed.Scheme))
	if scheme != "http" && scheme != "https" {
		return errors.New("only http/https remote url is allowed")
	}
	if parsed.User != nil {
		return errors.New("remote url cannot contain userinfo")
	}
	host := strings.ToLower(strings.TrimSpace(parsed.Hostname()))
	if host == "" {
		return errors.New("remote url missing host")
	}
	if _, blocked := soraBlockedHostnames[host]; blocked {
		return errors.New("remote url is not allowed")
	}
	if ip := net.ParseIP(host); ip != nil {
		if isSoraBlockedIP(ip) {
			return errors.New("remote url is not allowed")
		}
		return nil
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("resolve remote url failed: %w", err)
	}
	for _, ip := range ips {
		if isSoraBlockedIP(ip) {
			return errors.New("remote url is not allowed")
		}
	}
	return nil
}

func isSoraBlockedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	for _, cidr := range soraBlockedCIDRs {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

func mustParseCIDRs(values []string) []*net.IPNet {
	out := make([]*net.IPNet, 0, len(values))
	for _, val := range values {
		_, cidr, err := net.ParseCIDR(val)
		if err != nil {
			continue
		}
		out = append(out, cidr)
	}
	return out
}
