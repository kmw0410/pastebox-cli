package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func TestRunGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/pastebox/abc123" || r.URL.Query().Get("raw") != "1" {
			t.Errorf("URL = %s", r.URL.String())
		}
		if r.Header.Get("paste-password") != "top-secret" {
			t.Errorf("paste-password header missing")
		}
		io.WriteString(w, "paste content")
	}))
	defer server.Close()
	app, stdout, stderr := testApplication(serverConfig(t, server.URL+"/pastebox"), strings.NewReader(""))
	if code := app.run([]string{"get", "--password", "top-secret", "abc123"}); code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr.String())
	}
	if stdout.String() != "paste content" {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunGetAllowsSafeRedirects(t *testing.T) {
	tests := []struct {
		name     string
		location func(string) string
	}{
		{name: "same origin", location: func(serverURL string) string {
			return serverURL + "/pastebox/final?raw=1"
		}},
		{name: "different path below base", location: func(serverURL string) string {
			return serverURL + "/pastebox/nested/final?raw=1"
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/pastebox/start":
					http.Redirect(w, r, tt.location(server.URL), http.StatusFound)
				case "/pastebox/final", "/pastebox/nested/final":
					if r.Header.Get("paste-password") != "top-secret" {
						t.Errorf("paste-password header = %q", r.Header.Get("paste-password"))
					}
					io.WriteString(w, "redirected paste")
				default:
					http.NotFound(w, r)
				}
			}))
			defer server.Close()

			app, stdout, stderr := testApplication(serverConfig(t, server.URL+"/pastebox"), strings.NewReader(""))
			if code := app.run([]string{"get", "--password", "top-secret", "start"}); code != 0 {
				t.Fatalf("exit = %d, stderr = %s", code, stderr.String())
			}
			if stdout.String() != "redirected paste" {
				t.Fatalf("stdout = %q", stdout.String())
			}
		})
	}
}

func TestRunGetRejectsUnsafeRedirectsWithoutSendingPassword(t *testing.T) {
	tests := []struct {
		name         string
		location     func(sourceURL, targetURL string) string
		want         string
		usesExternal bool
	}{
		{
			name: "different hostname",
			location: func(sourceURL, _ string) string {
				parsed, _ := url.Parse(sourceURL + "/pastebox/blocked")
				parsed.Host = "localhost:" + parsed.Port()
				return parsed.String()
			},
			want: "hostname must remain",
		},
		{
			name: "different port",
			location: func(_, targetURL string) string {
				return targetURL + "/pastebox/blocked"
			},
			want:         "port must remain",
			usesExternal: true,
		},
		{
			name: "different scheme",
			location: func(sourceURL, _ string) string {
				return "https" + strings.TrimPrefix(sourceURL, "http") + "/pastebox/blocked"
			},
			want: "scheme must remain",
		},
		{
			name: "outside base path",
			location: func(sourceURL, _ string) string {
				return sourceURL + "/outside/blocked"
			},
			want: "below configured server path",
		},
		{
			name: "user information",
			location: func(_, targetURL string) string {
				parsed, _ := url.Parse(targetURL + "/pastebox/blocked")
				parsed.User = url.UserPassword("redirect-user", "redirect-secret")
				return parsed.String()
			},
			want:         "contains user credentials",
			usesExternal: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var externalReceivedPassword atomic.Bool
			external := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("paste-password") != "" {
					externalReceivedPassword.Store(true)
				}
				io.WriteString(w, "unsafe destination")
			}))
			defer external.Close()

			var source *httptest.Server
			var blockedReceivedPassword atomic.Bool
			source = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/pastebox/start" {
					http.Redirect(w, r, tt.location(source.URL, external.URL), http.StatusFound)
					return
				}
				if r.Header.Get("paste-password") != "" {
					blockedReceivedPassword.Store(true)
				}
				io.WriteString(w, "unsafe destination")
			}))
			defer source.Close()

			app, stdout, stderr := testApplication(serverConfig(t, source.URL+"/pastebox"), strings.NewReader(""))
			if code := app.run([]string{"get", "--password", "top-secret", "start"}); code != 1 {
				t.Fatalf("exit = %d, stdout = %q, stderr = %q", code, stdout.String(), stderr.String())
			}
			if !strings.Contains(stderr.String(), tt.want) {
				t.Fatalf("stderr = %q, want %q", stderr.String(), tt.want)
			}
			if strings.Contains(stderr.String(), "top-secret") || strings.Contains(stderr.String(), "redirect-secret") {
				t.Fatalf("stderr exposed a secret: %q", stderr.String())
			}
			if blockedReceivedPassword.Load() {
				t.Fatal("blocked source destination received paste-password")
			}
			if tt.usesExternal && externalReceivedPassword.Load() {
				t.Fatal("blocked external destination received paste-password")
			}
		})
	}
}

func TestRunGetRejectsTooManyRedirects(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestNumber := requests.Add(1)
		if r.Header.Get("paste-password") != "top-secret" {
			t.Errorf("request %d paste-password header = %q", requestNumber, r.Header.Get("paste-password"))
		}
		http.Redirect(w, r, fmt.Sprintf("/pastebox/redirect-%d", requestNumber), http.StatusFound)
	}))
	defer server.Close()

	app, _, stderr := testApplication(serverConfig(t, server.URL+"/pastebox"), strings.NewReader(""))
	if code := app.run([]string{"get", "--password", "top-secret", "start"}); code != 1 {
		t.Fatalf("exit = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "stopped after 10 redirects") {
		t.Fatalf("stderr = %q", stderr.String())
	}
	if got := requests.Load(); got != maxGetRedirects {
		t.Fatalf("requests = %d, want %d", got, maxGetRedirects)
	}
}

func TestValidateGetRedirectRejectsHTTPSDowngrade(t *testing.T) {
	base, err := url.Parse("https://paste.example.com/base")
	if err != nil {
		t.Fatal(err)
	}
	destination, err := url.Parse("http://paste.example.com:443/base/next")
	if err != nil {
		t.Fatal(err)
	}
	if err := validateGetRedirect(base, destination); err == nil || !strings.Contains(err.Error(), "scheme must remain https") {
		t.Fatalf("error = %v", err)
	}
}

func TestRunCommandHelp(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "get long", args: []string{"get", "--help"}, want: "pb get [--password PASSWORD] <code|url>"},
		{name: "get short", args: []string{"get", "-h"}, want: "pb get [--password PASSWORD] <code|url>"},
		{name: "config long", args: []string{"config", "--help"}, want: "pb config set server <URL>"},
		{name: "config short", args: []string{"config", "-h"}, want: "pb config set server <URL>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			missingConfig := filepath.Join(t.TempDir(), "missing.json")
			app, stdout, stderr := testApplication(missingConfig, strings.NewReader(""))
			if code := app.run(tt.args); code != 0 {
				t.Fatalf("exit = %d, stderr = %q", code, stderr.String())
			}
			if !strings.Contains(stdout.String(), tt.want) {
				t.Fatalf("stdout = %q, want %q", stdout.String(), tt.want)
			}
			if stderr.Len() != 0 {
				t.Fatalf("stderr = %q", stderr.String())
			}
		})
	}
}

func TestResolvePasteURLRejectsDifferentServer(t *testing.T) {
	_, err := resolvePasteURL("https://paste.example.com/base", "https://evil.example/abc")
	if err == nil || !strings.Contains(err.Error(), "configured server") {
		t.Fatalf("error = %v", err)
	}
}

func TestResolvePasteURLRejectsUnsafeTargets(t *testing.T) {
	tests := []struct {
		target string
		want   string
	}{
		{target: "https://paste.example.com/base", want: "include a paste code"},
		{target: "https://paste.example.com/other/abc", want: "below configured server path"},
		{target: "https://user:pass@paste.example.com/base/abc", want: "user credentials"},
		{target: "https://paste.example.com/base/abc?delete=secret", want: "query or fragment"},
	}
	for _, tt := range tests {
		_, err := resolvePasteURL("https://paste.example.com/base", tt.target)
		if err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("resolvePasteURL(%q) error = %v, want %q", tt.target, err, tt.want)
		}
	}
}

func TestRunConfigValidate(t *testing.T) {
	path := writeTestConfig(t, `{"server_url":"https://paste.example.com/"}`)
	app, stdout, stderr := testApplication(path, bytes.NewReader(nil))
	if code := app.run([]string{"config", "validate"}); code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "valid config: "+path) || !strings.Contains(stdout.String(), "server_url: https://paste.example.com") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunWithoutArgumentsCreatesConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".config", "pastebox", "config.json")
	app, stdout, stderr := testApplication(path, bytes.NewReader(nil))
	app.stdinTTY = true
	if code := app.run(nil); code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "created config: "+path) || !strings.Contains(stdout.String(), "pb config set server <URL>") {
		t.Fatalf("stdout = %q", stdout.String())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != initialConfig {
		t.Fatalf("config = %q", data)
	}
}

func TestRunWithoutArgumentsDoesNotOverwriteConfig(t *testing.T) {
	path := writeTestConfig(t, `{"server_url":"https://paste.example.com"}`)
	app, _, stderr := testApplication(path, bytes.NewReader(nil))
	app.stdinTTY = true
	if code := app.run(nil); code != 2 {
		t.Fatalf("exit = %d", code)
	}
	if !strings.Contains(stderr.String(), "no input") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}
