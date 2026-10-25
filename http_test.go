package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestCommandsKeepHTTPFailuresSecretSafe(t *testing.T) {
	commands := []struct {
		name      string
		args      []string
		passwords []string
	}{
		{"upload", []string{"--password"}, []string{"custom-secret", "custom-secret"}},
		{"show", []string{"show", "--password", "abc"}, []string{"source-secret"}},
		{"clone", []string{"clone", "--source-password", "--password", "abc"}, []string{"source-secret", "custom-secret", "custom-secret"}},
		{"delete", []string{"delete", "abc"}, []string{"delete-secret"}},
		{"manage", []string{"manage", "show", "abc"}, []string{"manage-secret"}},
	}
	for _, command := range commands {
		for _, failure := range []string{"credentials redirect", "query redirect", "invalid redirect", "json error", "text error"} {
			t.Run(command.name+"/"+failure, func(t *testing.T) {
				var destinationRequests atomic.Int32
				destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					destinationRequests.Add(1)
				}))
				defer destination.Close()
				source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_, _ = io.Copy(io.Discard, r.Body)
					switch failure {
					case "credentials redirect":
						location := strings.Replace(destination.URL, "http://", "http://redirect-user:redirect-secret@", 1)
						http.Redirect(w, r, location+"/paste", http.StatusTemporaryRedirect)
					case "query redirect":
						http.Redirect(w, r, destination.URL+"/paste?token=query-secret", http.StatusTemporaryRedirect)
					case "invalid redirect":
						w.Header().Set("Location", "http://redirect-user:redirect-secret@invalid%host/paste")
						w.WriteHeader(http.StatusTemporaryRedirect)
					case "json error":
						w.WriteHeader(http.StatusForbidden)
						fmt.Fprint(w, `{"error":"custom-secret source-secret delete-secret manage-secret paste-content-secret"}`)
					case "text error":
						w.WriteHeader(http.StatusForbidden)
						fmt.Fprint(w, "custom-secret source-secret delete-secret manage-secret paste-content-secret")
					}
				}))
				defer source.Close()
				app, stdout, stderr := testApplication(serverConfig(t, source.URL), strings.NewReader("paste-content-secret"))
				app.readPassword = testPasswordReader(command.passwords...)
				if code := app.run(command.args); code != 1 {
					t.Fatalf("exit = %d, want 1", code)
				}
				if stdout.Len() != 0 {
					t.Fatal("failed request produced stdout")
				}
				if destinationRequests.Load() != 0 {
					t.Fatal("rejected redirect contacted destination")
				}
				for _, secret := range []string{"redirect-user", "redirect-secret", "query-secret", "custom-secret", "source-secret", "delete-secret", "manage-secret", "paste-content-secret"} {
					if strings.Contains(stderr.String(), secret) {
						t.Errorf("diagnostic exposed %s", secret)
					}
				}
				if strings.HasSuffix(failure, "error") && !strings.Contains(stderr.String(), "Forbidden (HTTP 403)") {
					t.Fatal("diagnostic omitted HTTP status")
				}
			})
		}
	}
}
