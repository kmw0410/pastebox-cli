package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunClone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/pastebox/source" || r.URL.Query().Get("format") != "json" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
		}
		if r.Header.Get("paste-password") != "source-secret" {
			t.Errorf("paste-password = %q", r.Header.Get("paste-password"))
		}
		if r.Header.Get("data-policy") != "12h" || r.Header.Get("new-paste-password") != "clone-secret" || r.Header.Get("usepassword") != "" || r.Header.Get("code") != "cloned" {
			t.Errorf("unexpected clone headers: %v", r.Header)
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"url":"https://public.example/cloned","expires":"soon","manage":"manage-url","delete":"delete-url","password_protected":true}`)
	}))
	defer server.Close()

	app, stdout, stderr := testApplication(serverConfig(t, server.URL+"/pastebox"), strings.NewReader(""))
	app.readPassword = testPasswordReader("source-secret", "clone-secret", "clone-secret")
	code := app.run([]string{"clone", "--source-password", "--expires", "12h", "--password", "--code", "cloned", "source"})
	if code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr.String())
	}
	for _, want := range []string{"url: https://public.example/cloned", "expires: soon", "manage: manage-url", "delete: delete-url"} {
		if !strings.Contains(stdout.String(), want) {
			t.Errorf("stdout = %q, missing %q", stdout.String(), want)
		}
	}
}

func TestRunCloneQuiet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"url":"https://public.example/cloned"}`)
	}))
	defer server.Close()

	app, stdout, stderr := testApplication(serverConfig(t, server.URL), strings.NewReader(""))
	if code := app.run([]string{"clone", "--quiet", "source"}); code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr.String())
	}
	if stdout.String() != "https://public.example/cloned\n" {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunCloneEmptyPasswordPromptRequestsRandomPassword(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("usepassword") != "true" || r.Header.Get("new-paste-password") != "" {
			t.Errorf("unexpected password headers: %v", r.Header)
		}
		io.WriteString(w, `{"url":"https://public.example/random-clone","password":"generated-secret","password_protected":true}`)
	}))
	defer server.Close()

	app, stdout, stderr := testApplication(serverConfig(t, server.URL), strings.NewReader(""))
	app.readPassword = testPasswordReader("")
	if code := app.run([]string{"clone", "--password", "source"}); code != 0 {
		t.Fatalf("exit = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "password: generated-secret") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunCloneArgumentErrors(t *testing.T) {
	app, _, stderr := testApplication("unused", strings.NewReader(""))
	if code := app.run([]string{"clone", "--permanent", "--once", "source"}); code != 2 {
		t.Fatalf("exit = %d", code)
	}
	if !strings.Contains(stderr.String(), "mutually exclusive") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestRunCloneServerErrorDoesNotExposeSourcePassword(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, `{"error":"password required or invalid"}`)
	}))
	defer server.Close()

	app, _, stderr := testApplication(serverConfig(t, server.URL), strings.NewReader(""))
	app.readPassword = testPasswordReader("source-secret")
	if code := app.run([]string{"clone", "--source-password", "source"}); code != 1 {
		t.Fatalf("exit = %d", code)
	}
	if strings.Contains(stderr.String(), "source-secret") {
		t.Fatalf("stderr exposed password: %q", stderr.String())
	}
}

func TestRunCloneRejectsUnsafeRedirectWithoutSendingPassword(t *testing.T) {
	receivedPassword := false
	destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPassword = r.Header.Get("paste-password") != "" || r.Header.Get("new-paste-password") != ""
		io.WriteString(w, `{"url":"https://public.example/cloned"}`)
	}))
	defer destination.Close()

	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, destination.URL+"/cloned", http.StatusTemporaryRedirect)
	}))
	defer source.Close()

	app, _, stderr := testApplication(serverConfig(t, source.URL), strings.NewReader(""))
	app.readPassword = testPasswordReader("source-secret", "clone-secret", "clone-secret")
	if code := app.run([]string{"clone", "--source-password", "--password", "source"}); code != 1 {
		t.Fatalf("exit = %d, stderr = %q", code, stderr.String())
	}
	if receivedPassword {
		t.Fatal("unsafe redirect received paste password")
	}
	if strings.Contains(stderr.String(), "source-secret") {
		t.Fatalf("stderr exposed password: %q", stderr.String())
	}
}

func TestRunCloneRejectsServerWithoutPasswordConfirmation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"url":"https://public.example/unprotected"}`)
	}))
	defer server.Close()

	app, _, stderr := testApplication(serverConfig(t, server.URL), strings.NewReader(""))
	app.readPassword = testPasswordReader("clone-secret", "clone-secret")
	if code := app.run([]string{"clone", "--password", "source"}); code != 1 {
		t.Fatalf("exit = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "did not confirm password protection") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}
