package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func clonePaste(ctx context.Context, client *http.Client, cfg config, target, sourcePassword string, opts uploadOptions) (uploadResponse, []byte, error) {
	requestURL, err := resolvePasteURL(cfg.ServerURL, target)
	if err != nil {
		return uploadResponse{}, nil, fmt.Errorf("invalid paste target: %w", err)
	}
	requestURL, err = addFormatQuery(requestURL)
	if err != nil {
		return uploadResponse{}, nil, fmt.Errorf("invalid paste target: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, nil)
	if err != nil {
		return uploadResponse{}, nil, fmt.Errorf("create clone request: %w", err)
	}
	if sourcePassword != "" {
		req.Header.Set("paste-password", sourcePassword)
	}
	if policy := opts.policy(); policy != "" {
		req.Header.Set("data-policy", policy)
	}
	if opts.usePassword {
		req.Header.Set("usepassword", "true")
	}
	if opts.newPassword != "" {
		req.Header.Set("new-paste-password", opts.newPassword)
	}
	if opts.code != "" {
		req.Header.Set("code", opts.code)
	}

	sensitiveHeaders := make(http.Header)
	if sourcePassword != "" {
		sensitiveHeaders.Set("paste-password", sourcePassword)
	}
	if opts.newPassword != "" {
		sensitiveHeaders.Set("new-paste-password", opts.newPassword)
	}
	redirectClient, err := secureHTTPClient(client, cfg.ServerURL, sensitiveHeaders)
	if err != nil {
		return uploadResponse{}, nil, fmt.Errorf("configure clone redirects: %w", err)
	}
	resp, err := redirectClient.Do(req)
	if err != nil {
		return uploadResponse{}, nil, fmt.Errorf("clone request failed for %s: %w", cfg.ServerURL, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return uploadResponse{}, nil, fmt.Errorf("read clone response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return uploadResponse{}, nil, responseError("clone failed", resp, raw)
	}

	var result uploadResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return uploadResponse{}, nil, fmt.Errorf("invalid clone response from server: %w", err)
	}
	if result.URL == "" {
		return uploadResponse{}, nil, fmt.Errorf("invalid clone response from server: missing url")
	}
	if opts.newPassword != "" && (result.PasswordProtected == nil || !*result.PasswordProtected) {
		return uploadResponse{}, nil, fmt.Errorf("server did not confirm password protection; update the Pastebox server before using --password")
	}
	return result, raw, nil
}

func addFormatQuery(target string) (string, error) {
	parsed, err := url.Parse(target)
	if err != nil {
		return "", err
	}
	query := parsed.Query()
	query.Set("format", "json")
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}
