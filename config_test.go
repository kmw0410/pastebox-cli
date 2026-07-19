package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTestConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadConfig(t *testing.T) {
	path := writeTestConfig(t, `{"server_url":"https://example.com/pastebox/"}`)
	cfg, err := loadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ServerURL != "https://example.com/pastebox" {
		t.Fatalf("ServerURL = %q", cfg.ServerURL)
	}
}

func TestLoadConfigDiagnostics(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{name: "syntax", content: "{\n  \"server_url\": ]\n}", want: ":2:17:"},
		{name: "unknown field", content: `{"server_ur1":"https://example.com"}`, want: `unknown field "server_ur1"`},
		{name: "missing field", content: `{}`, want: `required field "server_url" is missing`},
		{name: "wrong type", content: `{"server_url":42}`, want: "server_url must be a string"},
		{name: "empty", content: `{"server_url":" "}`, want: "server_url must not be empty"},
		{name: "missing scheme", content: `{"server_url":"example.com"}`, want: "is missing http:// or https://"},
		{name: "unsupported scheme", content: `{"server_url":"ftp://example.com"}`, want: "must use http:// or https://"},
		{name: "missing host", content: `{"server_url":"https:///pastebox"}`, want: "must include a host"},
		{name: "credentials", content: `{"server_url":"https://user:pass@example.com"}`, want: "must not contain user credentials"},
		{name: "query", content: `{"server_url":"https://example.com?x=1"}`, want: "must not contain a query or fragment"},
		{name: "fragment", content: `{"server_url":"https://example.com/#x"}`, want: "must not contain a query or fragment"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeTestConfig(t, tt.content)
			_, err := loadConfig(path)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want substring %q", err, tt.want)
			}
			if !strings.Contains(err.Error(), path) {
				t.Fatalf("error does not include config path: %v", err)
			}
		})
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")
	_, err := loadConfig(path)
	if err == nil {
		t.Fatal("expected error")
	}
	for _, want := range []string{path, "Run pb without arguments", "pb config set server <URL>"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error = %q, want %q", err, want)
		}
	}
}

func TestInitializeConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "pastebox", "config.json")
	created, err := initializeConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if !created {
		t.Fatal("config was not created")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != initialConfig {
		t.Fatalf("config = %q", data)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("config mode = %o", info.Mode().Perm())
	}

	created, err = initializeConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if created {
		t.Fatal("existing config was overwritten")
	}
}

func TestRunConfigSetServerAndShow(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "pastebox", "config.json")
	app, stdout, stderr := testApplication(path, bytes.NewReader(nil))
	if code := app.run([]string{"config", "set", "server", " https://example.com/pastebox/ "}); code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr.String())
	}
	if got := stdout.String(); !strings.Contains(got, "updated config: "+path) || !strings.Contains(got, "server_url: https://example.com/pastebox") {
		t.Fatalf("stdout = %q", got)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "{\n  \"server_url\": \"https://example.com/pastebox\"\n}\n" {
		t.Fatalf("config = %q", data)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("config mode = %o", info.Mode().Perm())
	}

	stdout.Reset()
	stderr.Reset()
	if code := app.run([]string{"config", "show"}); code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr.String())
	}
	if got := stdout.String(); !strings.Contains(got, "config: "+path) || !strings.Contains(got, "server_url: https://example.com/pastebox") {
		t.Fatalf("stdout = %q", got)
	}
}

func TestRunConfigSetServerUpdatesInitializedConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pastebox", "config.json")
	created, err := initializeConfig(path)
	if err != nil || !created {
		t.Fatalf("created = %v, err = %v", created, err)
	}

	app, _, stderr := testApplication(path, bytes.NewReader(nil))
	if code := app.run([]string{"config", "set", "server", "https://paste.example.com"}); code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr.String())
	}
	cfg, err := loadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ServerURL != "https://paste.example.com" {
		t.Fatalf("ServerURL = %q", cfg.ServerURL)
	}
}

func TestRunConfigSetServerRejectsInvalidURLWithoutChangingFile(t *testing.T) {
	path := writeTestConfig(t, "{\n  \"server_url\": \"https://original.example\"\n}\n")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	app, _, stderr := testApplication(path, bytes.NewReader(nil))
	if code := app.run([]string{"config", "set", "server", "ftp://invalid.example"}); code != 2 {
		t.Fatalf("exit = %d", code)
	}
	if !strings.Contains(stderr.String(), "must use http:// or https://") {
		t.Fatalf("stderr = %q", stderr.String())
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, before) {
		t.Fatalf("config changed: before %q, after %q", before, after)
	}
}

func TestRunConfigUsage(t *testing.T) {
	app, _, stderr := testApplication("unused", bytes.NewReader(nil))
	for _, args := range [][]string{
		{"config"},
		{"config", "set", "server"},
		{"config", "set", "unknown", "https://example.com"},
	} {
		stderr.Reset()
		if code := app.run(args); code != 2 {
			t.Fatalf("args = %v, exit = %d", args, code)
		}
		if !strings.Contains(stderr.String(), "pb config set server <URL>") {
			t.Fatalf("args = %v, stderr = %q", args, stderr.String())
		}
	}
}
