package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunDeleteWithPrivateURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/pastebox/api/v1/pastes/AbC123" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
		}
		if got := r.Header.Get("paste-delete-token"); got != "delete-secret" {
			t.Errorf("paste-delete-token = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"deleted":true,"id":"AbC123"}`)
	}))
	defer server.Close()

	app, stdout, stderr := testApplication(serverConfig(t, server.URL+"/pastebox"), strings.NewReader(""))
	target := server.URL + "/pastebox/AbC123?delete=delete-secret"
	if code := app.run([]string{"delete", target}); code != 0 {
		t.Fatalf("exit = %d, stderr = %q", code, stderr.String())
	}
	if stdout.String() != "deleted: AbC123\n" {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunDeletePromptsForTokenWithCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("paste-delete-token"); got != "delete-secret" {
			t.Errorf("paste-delete-token = %q", got)
		}
		io.WriteString(w, `{"deleted":true}`)
	}))
	defer server.Close()

	app, _, stderr := testApplication(serverConfig(t, server.URL), strings.NewReader(""))
	app.readPassword = testPasswordReader("delete-secret")
	if code := app.run([]string{"delete", "AbC123"}); code != 0 {
		t.Fatalf("exit = %d, stderr = %q", code, stderr.String())
	}
}

func TestRunDeleteRejectsForeignPrivateURLWithoutExposingToken(t *testing.T) {
	app, _, stderr := testApplication(serverConfig(t, "https://paste.example.com"), strings.NewReader(""))
	if code := app.run([]string{"delete", "https://other.example/AbC123?delete=delete-secret"}); code != 2 {
		t.Fatalf("exit = %d", code)
	}
	if strings.Contains(stderr.String(), "delete-secret") {
		t.Fatalf("stderr exposed token: %q", stderr.String())
	}
}

func TestRunDeleteRejectsUnsafeRedirectWithoutSendingToken(t *testing.T) {
	receivedToken := false
	destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedToken = r.Header.Get("paste-delete-token") != ""
		io.WriteString(w, `{"deleted":true}`)
	}))
	defer destination.Close()

	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, destination.URL+"/delete", http.StatusTemporaryRedirect)
	}))
	defer source.Close()

	app, _, stderr := testApplication(serverConfig(t, source.URL), strings.NewReader(""))
	app.readPassword = testPasswordReader("delete-secret")
	if code := app.run([]string{"delete", "AbC123"}); code != 1 {
		t.Fatalf("exit = %d, stderr = %q", code, stderr.String())
	}
	if receivedToken {
		t.Fatal("unsafe redirect received delete token")
	}
	if strings.Contains(stderr.String(), "delete-secret") {
		t.Fatalf("stderr exposed token: %q", stderr.String())
	}
}
