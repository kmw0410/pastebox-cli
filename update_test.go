package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeOSRelease(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "os-release")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRunUpdateUsesParuForArch(t *testing.T) {
	previousVersion := version
	version = "v26.07.20-1"
	t.Cleanup(func() { version = previousVersion })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(latestRelease{TagName: "v26.07.21-1"})
	}))
	defer server.Close()

	app, stdout, stderr := testApplication("", bytes.NewReader(nil))
	app.osReleasePath = writeOSRelease(t, "ID=endeavouros\nID_LIKE=arch\n")
	app.releaseAPIURL = server.URL
	app.httpClient = server.Client()
	app.lookPath = func(name string) (string, error) {
		return "/usr/bin/" + name, nil
	}
	var commandName string
	var commandArgs []string
	app.runCommand = func(_ context.Context, name string, args ...string) error {
		commandName = name
		commandArgs = append([]string(nil), args...)
		return nil
	}

	if code := app.run([]string{"update"}); code != 0 {
		t.Fatalf("exit = %d, stderr = %q", code, stderr.String())
	}
	if commandName != "paru" || len(commandArgs) != 2 || commandArgs[0] != "-S" || commandArgs[1] != "pastebox-cli" {
		t.Fatalf("command = %q %q", commandName, commandArgs)
	}
	if !strings.Contains(stdout.String(), "Updating pastebox-cli to v26.07.21-1 with paru") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunUpdateFallsBackToYayForArch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(latestRelease{TagName: "v26.07.21-1"})
	}))
	defer server.Close()

	app, _, stderr := testApplication("", bytes.NewReader(nil))
	app.osReleasePath = writeOSRelease(t, "ID=arch\n")
	app.releaseAPIURL = server.URL
	app.httpClient = server.Client()
	app.lookPath = func(name string) (string, error) {
		if name == "yay" {
			return "/usr/bin/yay", nil
		}
		return "", os.ErrNotExist
	}
	var commandName string
	app.runCommand = func(_ context.Context, name string, _ ...string) error {
		commandName = name
		return nil
	}

	if code := app.run([]string{"update"}); code != 0 {
		t.Fatalf("exit = %d, stderr = %q", code, stderr.String())
	}
	if commandName != "yay" {
		t.Fatalf("command = %q, want yay", commandName)
	}
}

func TestRunUpdateGuidesArchUserWithoutAURHelper(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(latestRelease{TagName: "v26.07.21-1"})
	}))
	defer server.Close()

	app, stdout, stderr := testApplication("", bytes.NewReader(nil))
	app.osReleasePath = writeOSRelease(t, "ID=arch\n")
	app.releaseAPIURL = server.URL
	app.httpClient = server.Client()
	app.lookPath = func(string) (string, error) { return "", os.ErrNotExist }
	app.runCommand = func(context.Context, string, ...string) error {
		t.Fatal("AUR helper ran when none was installed")
		return nil
	}

	if code := app.run([]string{"update"}); code != 0 {
		t.Fatalf("exit = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Install paru or yay, then run pb update again") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunUpdateDownloadsVerifiesAndInstallsDeb(t *testing.T) {
	packageBody := []byte("synthetic deb package")
	digest := sha256.Sum256(packageBody)
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/latest":
			if got := r.Header.Get("Accept"); got != "application/vnd.github+json" {
				t.Errorf("Accept = %q", got)
			}
			json.NewEncoder(w).Encode(latestRelease{
				TagName: "v26.07.20-1",
				Assets: []releaseAsset{
					{Name: "pastebox-cli_26.07.20.1-1_arm64.deb", BrowserDownloadURL: server.URL + "/arm64.deb", Digest: "sha256:" + strings.Repeat("0", 64)},
					{Name: "pastebox-cli_26.07.20.1-1_amd64.deb", BrowserDownloadURL: server.URL + "/amd64.deb", Digest: "sha256:" + hex.EncodeToString(digest[:])},
				},
			})
		case "/amd64.deb":
			w.Write(packageBody)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	app, stdout, stderr := testApplication("", bytes.NewReader(nil))
	app.osReleasePath = writeOSRelease(t, "ID=ubuntu\nID_LIKE=debian\n")
	app.goarch = "amd64"
	app.releaseAPIURL = server.URL + "/latest"
	app.httpClient = server.Client()
	app.effectiveUID = func() int { return 1000 }
	var commandName string
	var commandArgs []string
	app.runCommand = func(_ context.Context, name string, args ...string) error {
		commandName = name
		commandArgs = append([]string(nil), args...)
		data, err := os.ReadFile(args[len(args)-1])
		if err != nil {
			return err
		}
		if !bytes.Equal(data, packageBody) {
			return fmt.Errorf("package body = %q", data)
		}
		return nil
	}

	if code := app.run([]string{"update"}); code != 0 {
		t.Fatalf("exit = %d, stderr = %q", code, stderr.String())
	}
	if commandName != "sudo" || len(commandArgs) != 4 || commandArgs[0] != "apt-get" || commandArgs[1] != "install" || commandArgs[2] != "-y" {
		t.Fatalf("command = %q %q", commandName, commandArgs)
	}
	if !strings.Contains(stdout.String(), "Updated pastebox-cli to v26.07.20-1") {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if _, err := os.Stat(commandArgs[3]); !os.IsNotExist(err) {
		t.Fatalf("temporary package was not removed: %v", err)
	}
}

func TestRunUpdateRejectsDigestMismatch(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/latest" {
			json.NewEncoder(w).Encode(latestRelease{
				TagName: "v26.07.20-1",
				Assets: []releaseAsset{{
					Name:               "pastebox-cli_26.07.20.1-1_amd64.deb",
					BrowserDownloadURL: server.URL + "/package.deb",
					Digest:             "sha256:" + strings.Repeat("0", 64),
				}},
			})
			return
		}
		io.WriteString(w, "tampered package")
	}))
	defer server.Close()

	app, _, stderr := testApplication("", bytes.NewReader(nil))
	app.osReleasePath = writeOSRelease(t, "ID=debian\n")
	app.goarch = "amd64"
	app.releaseAPIURL = server.URL + "/latest"
	app.httpClient = server.Client()
	app.runCommand = func(context.Context, string, ...string) error {
		t.Fatal("installer ran for a package with a mismatched digest")
		return nil
	}

	if code := app.run([]string{"update"}); code != 1 {
		t.Fatalf("exit = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "SHA-256 mismatch") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestRunUpdateSkipsInstalledLatestVersion(t *testing.T) {
	previousVersion := version
	version = "v26.07.20-1"
	t.Cleanup(func() { version = previousVersion })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(latestRelease{TagName: version})
	}))
	defer server.Close()

	app, stdout, stderr := testApplication("", bytes.NewReader(nil))
	app.osReleasePath = writeOSRelease(t, "ID=debian\n")
	app.goarch = "amd64"
	app.releaseAPIURL = server.URL
	app.httpClient = server.Client()
	app.runCommand = func(context.Context, string, ...string) error {
		t.Fatal("installer ran for the current version")
		return nil
	}

	if code := app.run([]string{"update"}); code != 0 {
		t.Fatalf("exit = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "already up to date") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestSelectUpdateTarget(t *testing.T) {
	tests := []struct {
		name      string
		osRelease map[string]string
		goarch    string
		installer string
		suffix    string
		wantErr   string
	}{
		{name: "Debian arm64", osRelease: map[string]string{"ID": "debian"}, goarch: "arm64", installer: "apt-get", suffix: "_arm64.deb"},
		{name: "Rocky amd64", osRelease: map[string]string{"ID": "rocky", "ID_LIKE": "rhel centos fedora"}, goarch: "amd64", installer: "dnf", suffix: ".x86_64.rpm"},
		{name: "RPM arm64", osRelease: map[string]string{"ID": "fedora"}, goarch: "arm64", wantErr: "do not support architecture arm64"},
		{name: "unsupported", osRelease: map[string]string{"ID": "alpine"}, goarch: "amd64", wantErr: "supported only"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target, err := selectUpdateTarget(tt.osRelease, tt.goarch)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if target.installer != tt.installer || target.assetSuffix != tt.suffix {
				t.Fatalf("target = %+v", target)
			}
		})
	}
}

func TestRunUpdateHelpAndInvalidArguments(t *testing.T) {
	app, stdout, stderr := testApplication("", bytes.NewReader(nil))
	if code := app.run([]string{"update", "--help"}); code != 0 {
		t.Fatalf("help exit = %d", code)
	}
	if !strings.Contains(stdout.String(), "pb update") {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if code := app.run([]string{"update", "unexpected"}); code != 2 {
		t.Fatalf("invalid exit = %d", code)
	}
	if !strings.Contains(stderr.String(), "pb update") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}
