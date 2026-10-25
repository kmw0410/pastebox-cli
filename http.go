package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// requestError never formats transport errors, which can contain rejected URLs,
// Location headers, or other server-controlled values.
func requestError(err error) error {
	var redirectErr *getRedirectError
	if errors.As(err, &redirectErr) {
		return fmt.Errorf("request failed: %w", redirectErr)
	}
	if errors.Is(err, context.Canceled) {
		return fmt.Errorf("request failed: %w", context.Canceled)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("request failed: %w", context.DeadlineExceeded)
	}
	var networkErr net.Error
	if errors.As(err, &networkErr) && networkErr.Timeout() {
		return fmt.Errorf("request failed: network timeout")
	}
	return fmt.Errorf("request failed: network or protocol error")
}

func responseError(prefix string, resp *http.Response) error {
	// Error bodies may reflect paste contents, passwords, or tokens. Only use
	// the standard status description, never the server's body or reason phrase.
	message := http.StatusText(resp.StatusCode)
	if message == "" {
		message = "unexpected server response"
	}
	return fmt.Errorf("%s: %s (HTTP %d)", prefix, message, resp.StatusCode)
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
	headers := make(http.Header)
	if password != "" {
		headers.Set("paste-password", password)
	}
	return secureHTTPClient(client, serverURL, headers)
}

func secureHTTPClient(client *http.Client, serverURL string, sensitiveHeaders http.Header) (*http.Client, error) {
	base, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}
	redirectClient := *client
	previousCheckRedirect := client.CheckRedirect
	redirectClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		for name := range sensitiveHeaders {
			req.Header.Del(name)
		}
		if len(via) >= maxGetRedirects {
			return redirectErrorf("redirect refused: stopped after %d redirects", maxGetRedirects)
		}
		if err := validateGetRedirect(base, req.URL); err != nil {
			return err
		}
		for name, values := range sensitiveHeaders {
			for _, value := range values {
				req.Header.Add(name, value)
			}
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
