package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type privatePasteReference struct {
	code  string
	token string
}

func (a application) runDelete(args []string) int {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Fprint(a.stdout, deleteUsageText)
		return 0
	}
	if len(args) != 1 || strings.TrimSpace(args[0]) == "" {
		fmt.Fprint(a.stderr, deleteUsageText)
		return 2
	}

	cfg, err := loadConfig(a.configPath)
	if err != nil {
		fmt.Fprintln(a.stderr, err)
		return 2
	}
	reference, err := resolvePrivatePasteReference(cfg.ServerURL, args[0], "delete")
	if err != nil {
		fmt.Fprintf(a.stderr, "invalid delete target: %v\n", err)
		return 2
	}
	if reference.token == "" {
		reference.token, err = a.promptSecret("Delete token: ", "delete token must not be empty")
		if err != nil {
			fmt.Fprintf(a.stderr, "cannot read delete token: %v\n", err)
			return 2
		}
	}
	if err := deletePaste(a.requestContext(), a.httpClient, cfg, reference.code, reference.token); err != nil {
		fmt.Fprintln(a.stderr, err)
		return 1
	}
	if _, err := fmt.Fprintf(a.stdout, "deleted: %s\n", reference.code); err != nil {
		fmt.Fprintf(a.stderr, "write output: %v\n", err)
		return 1
	}
	return 0
}

func deletePaste(ctx context.Context, client *http.Client, cfg config, code, token string) error {
	endpoint := strings.TrimRight(cfg.ServerURL, "/") + "/api/v1/pastes/" + url.PathEscape(code)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create delete request: %w", err)
	}
	req.Header.Set("paste-delete-token", token)

	sensitiveHeaders := make(http.Header)
	sensitiveHeaders.Set("paste-delete-token", token)
	redirectClient, err := secureHTTPClient(client, cfg.ServerURL, sensitiveHeaders)
	if err != nil {
		return fmt.Errorf("configure delete redirects: %w", err)
	}
	resp, err := redirectClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete request failed for %s: %w", cfg.ServerURL, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if err != nil {
		return fmt.Errorf("read delete response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return responseError("delete failed", resp, body)
	}
	var result struct {
		Deleted bool `json:"deleted"`
	}
	if err := json.Unmarshal(body, &result); err != nil || !result.Deleted {
		return fmt.Errorf("invalid delete response from server")
	}
	return nil
}

func resolvePrivatePasteReference(serverURL, target, tokenName string) (privatePasteReference, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return privatePasteReference{}, fmt.Errorf("code or private URL is required")
	}
	if !strings.Contains(target, "://") {
		if _, err := resolvePasteURL(serverURL, target); err != nil {
			return privatePasteReference{}, err
		}
		return privatePasteReference{code: target}, nil
	}

	base, err := url.Parse(serverURL)
	if err != nil {
		return privatePasteReference{}, err
	}
	parsed, err := url.Parse(target)
	if err != nil {
		return privatePasteReference{}, fmt.Errorf("private URL is invalid")
	}
	if parsed.User != nil {
		return privatePasteReference{}, fmt.Errorf("private URL must not contain user credentials")
	}
	if !strings.EqualFold(parsed.Scheme, base.Scheme) || !strings.EqualFold(parsed.Host, base.Host) {
		return privatePasteReference{}, fmt.Errorf("private URL must belong to the configured server")
	}
	if parsed.Fragment != "" {
		return privatePasteReference{}, fmt.Errorf("private URL must not contain a fragment")
	}

	basePath := strings.TrimRight(base.EscapedPath(), "/")
	prefix := basePath + "/"
	if !strings.HasPrefix(parsed.EscapedPath(), prefix) {
		return privatePasteReference{}, fmt.Errorf("private URL must be below the configured server path")
	}
	escapedCode := strings.TrimPrefix(parsed.EscapedPath(), prefix)
	if escapedCode == "" || strings.Contains(escapedCode, "/") {
		return privatePasteReference{}, fmt.Errorf("private URL must include one paste code")
	}
	code, err := url.PathUnescape(escapedCode)
	if err != nil {
		return privatePasteReference{}, fmt.Errorf("private URL contains an invalid paste code")
	}
	if _, err := resolvePasteURL(serverURL, code); err != nil {
		return privatePasteReference{}, err
	}

	query := parsed.Query()
	values, ok := query[tokenName]
	if len(query) != 1 || !ok || len(values) != 1 || strings.TrimSpace(values[0]) == "" {
		return privatePasteReference{}, fmt.Errorf("private URL must contain exactly one %s token", tokenName)
	}
	return privatePasteReference{code: code, token: values[0]}, nil
}
