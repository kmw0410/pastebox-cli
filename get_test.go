package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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
