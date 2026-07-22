package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
)

func getPaste(ctx context.Context, client *http.Client, cfg config, target, password string, output io.Writer) error {
	requestURL, err := resolvePasteURL(cfg.ServerURL, target)
	if err != nil {
		return fmt.Errorf("invalid paste target: %w", err)
	}
	requestURL, err = addRawQuery(requestURL)
	if err != nil {
		return fmt.Errorf("invalid paste target: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return fmt.Errorf("create get request: %w", err)
	}
	if password != "" {
		req.Header.Set("paste-password", password)
	}
	redirectClient, err := getHTTPClient(client, cfg.ServerURL, password)
	if err != nil {
		return fmt.Errorf("configure get redirects: %w", err)
	}
	resp, err := redirectClient.Do(req)
	if err != nil {
		var redirectErr *getRedirectError
		if errors.As(err, &redirectErr) {
			return fmt.Errorf("request failed: %w", redirectErr)
		}
		return fmt.Errorf("request failed for %s: %w", requestURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
		if readErr != nil {
			return fmt.Errorf("read paste response: %w", readErr)
		}
		return responseError("show failed", resp, body)
	}
	if _, err := io.Copy(output, resp.Body); err != nil {
		return fmt.Errorf("write paste content: %w", err)
	}
	return nil
}

const maxGetRedirects = 10

type getRedirectError struct {
	message string
}

func (e *getRedirectError) Error() string {
	return e.message
}

func redirectErrorf(format string, args ...any) error {
	return &getRedirectError{message: fmt.Sprintf(format, args...)}
}

func getHTTPClient(client *http.Client, serverURL, password string) (*http.Client, error) {
	base, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}
	redirectClient := *client
	previousCheckRedirect := client.CheckRedirect
	redirectClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		req.Header.Del("paste-password")
		if len(via) >= maxGetRedirects {
			return redirectErrorf("redirect refused: stopped after %d redirects", maxGetRedirects)
		}
		if err := validateGetRedirect(base, req.URL); err != nil {
			return err
		}
		if password != "" {
			req.Header.Set("paste-password", password)
		}
		if previousCheckRedirect != nil {
			return previousCheckRedirect(req, via)
		}
		return nil
	}
	return &redirectClient, nil
}

func validateGetRedirect(base, destination *url.URL) error {
	if destination.User != nil {
		return redirectErrorf("redirect refused: destination URL contains user credentials")
	}
	if !strings.EqualFold(destination.Scheme, base.Scheme) {
		return redirectErrorf("redirect refused: destination scheme must remain %s", base.Scheme)
	}
	if !strings.EqualFold(destination.Hostname(), base.Hostname()) {
		return redirectErrorf("redirect refused: destination hostname must remain %s", base.Hostname())
	}
	if effectivePort(destination) != effectivePort(base) {
		return redirectErrorf("redirect refused: destination port must remain %s", effectivePort(base))
	}
	basePath := cleanBasePath(base.Path)
	destinationPath := path.Clean(destination.Path)
	if basePath != "" && destinationPath != basePath && !strings.HasPrefix(destinationPath, basePath+"/") {
		return redirectErrorf("redirect refused: destination path must remain below configured server path %s", base.Path)
	}
	return nil
}

func cleanBasePath(basePath string) string {
	if basePath == "" || basePath == "/" {
		return ""
	}
	return strings.TrimRight(path.Clean(basePath), "/")
}

func effectivePort(parsed *url.URL) string {
	if port := parsed.Port(); port != "" {
		return port
	}
	if strings.EqualFold(parsed.Scheme, "https") {
		return "443"
	}
	return "80"
}

func resolvePasteURL(serverURL, target string) (string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return "", fmt.Errorf("code or URL is required")
	}
	if !strings.Contains(target, "://") {
		if strings.ContainsAny(target, "/?#") {
			return "", fmt.Errorf("invalid paste code %q", target)
		}
		return strings.TrimRight(serverURL, "/") + "/" + url.PathEscape(target), nil
	}

	base, err := url.Parse(serverURL)
	if err != nil {
		return "", err
	}
	parsed, err := url.Parse(target)
	if err != nil {
		return "", err
	}
	if parsed.Scheme != base.Scheme || parsed.Host != base.Host {
		return "", fmt.Errorf("paste URL must belong to configured server %s", serverURL)
	}
	if parsed.User != nil {
		return "", fmt.Errorf("paste URL must not contain user credentials")
	}
	basePath := strings.TrimRight(base.EscapedPath(), "/")
	targetPath := parsed.EscapedPath()
	if targetPath == "" || targetPath == "/" || targetPath == basePath || targetPath == basePath+"/" {
		return "", fmt.Errorf("paste URL must include a paste code")
	}
	if basePath != "" && !strings.HasPrefix(targetPath, basePath+"/") {
		return "", fmt.Errorf("paste URL must be below configured server path %s", base.Path)
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("paste URL must not contain a query or fragment")
	}
	return parsed.String(), nil
}
