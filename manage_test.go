package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const managedPasteJSON = `{"id":"AbC123","filename":"server.log","label":"production","url":"https://paste.example/AbC123","created_at":"2026-07-22T10:00:00+09:00","expires":"2026-07-23T10:00:00+09:00","data_policy":"temporary","size":42,"content_type":"text/plain","password_protected":true}`

func TestRunManageShow(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/pastebox/api/v1/pastes/AbC123" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
		}
		if got := r.Header.Get("paste-manage-token"); got != "manage-secret" {
			t.Errorf("paste-manage-token = %q", got)
		}
		io.WriteString(w, managedPasteJSON)
	}))
	defer server.Close()

	app, stdout, stderr := testApplication(serverConfig(t, server.URL+"/pastebox"), strings.NewReader(""))
	target := server.URL + "/pastebox/AbC123?manage=manage-secret"
	if code := app.run([]string{"manage", "show", target}); code != 0 {
		t.Fatalf("exit = %d, stderr = %q", code, stderr.String())
	}
	for _, want := range []string{"code: AbC123", "filename: server.log", "label: production", "policy: temporary", "password-protected: yes"} {
		if !strings.Contains(stdout.String(), want) {
			t.Errorf("stdout = %q, missing %q", stdout.String(), want)
		}
	}
}

func TestRunManagePatchActions(t *testing.T) {
	tests := []struct {
		name       string
		args       func(string) []string
		passwords  []string
		wantAction string
		wantField  string
		wantValue  string
	}{
		{name: "label", args: func(target string) []string { return []string{"manage", "label", target, "deploy log"} }, wantAction: "set_label", wantField: "label", wantValue: "deploy log"},
		{name: "policy", args: func(target string) []string { return []string{"manage", "policy", target, "12h"} }, wantAction: "set_policy", wantField: "data_policy", wantValue: "12h"},
		{name: "enable password", args: func(target string) []string { return []string{"manage", "password", "enable", target} }, passwords: []string{"custom-secret", "custom-secret"}, wantAction: "enable_password", wantField: "new_password", wantValue: "custom-secret"},
		{name: "disable password", args: func(target string) []string { return []string{"manage", "password", "disable", target} }, passwords: []string{"custom-secret"}, wantAction: "disable_password", wantField: "password", wantValue: "custom-secret"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch || r.Header.Get("paste-manage-token") != "manage-secret" {
					t.Errorf("unexpected request: %s %s headers=%v", r.Method, r.URL.String(), r.Header)
				}
				var body map[string]string
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				if body["action"] != tt.wantAction || body[tt.wantField] != tt.wantValue {
					t.Errorf("body = %#v", body)
				}
				io.WriteString(w, managedPasteJSON)
			}))
			defer server.Close()

			app, _, stderr := testApplication(serverConfig(t, server.URL), strings.NewReader(""))
			if len(tt.passwords) > 0 {
				app.readPassword = testPasswordReader(tt.passwords...)
			}
			target := server.URL + "/AbC123?manage=manage-secret"
			if code := app.run(tt.args(target)); code != 0 {
				t.Fatalf("exit = %d, stderr = %q", code, stderr.String())
			}
		})
	}
}

func TestRunManageDeleteDirectsToDeleteCommand(t *testing.T) {
	app, stdout, stderr := testApplication("/missing/config.json", strings.NewReader(""))
	if code := app.run([]string{"manage", "delete", "AbC123"}); code != 2 {
		t.Fatalf("exit = %d, stderr = %q", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if got := stderr.String(); got != "pb manage delete has been removed; use pb delete to delete a paste\n" {
		t.Fatalf("stderr = %q", got)
	}
}

func TestRunManageRejectsUnsafeRedirectWithoutSendingToken(t *testing.T) {
	receivedToken := false
	destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedToken = r.Header.Get("paste-manage-token") != ""
		io.WriteString(w, managedPasteJSON)
	}))
	defer destination.Close()

	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, destination.URL+"/manage", http.StatusTemporaryRedirect)
	}))
	defer source.Close()

	app, _, stderr := testApplication(serverConfig(t, source.URL), strings.NewReader(""))
	app.readPassword = testPasswordReader("manage-secret")
	if code := app.run([]string{"manage", "show", "AbC123"}); code != 1 {
		t.Fatalf("exit = %d, stderr = %q", code, stderr.String())
	}
	if receivedToken {
		t.Fatal("unsafe redirect received manage token")
	}
	if strings.Contains(stderr.String(), "manage-secret") {
		t.Fatalf("stderr exposed token: %q", stderr.String())
	}
}
