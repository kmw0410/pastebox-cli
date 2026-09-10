package main

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestNewHTTPClientTimeouts(t *testing.T) {
	client := newHTTPClient()
	if client.Timeout != 0 {
		t.Fatalf("client timeout = %s, want no whole-request timeout", client.Timeout)
	}
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("transport type = %T", client.Transport)
	}
	if transport.DialContext == nil {
		t.Fatal("DialContext is nil")
	}
	if transport.TLSHandshakeTimeout != tlsHandshakeTimeout {
		t.Fatalf("TLS handshake timeout = %s", transport.TLSHandshakeTimeout)
	}
	if transport.ResponseHeaderTimeout != responseHeaderTimeout {
		t.Fatalf("response header timeout = %s", transport.ResponseHeaderTimeout)
	}
}

func TestRunRequestCancellation(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		input io.Reader
	}{
		{name: "upload", input: strings.NewReader("paste body")},
		{name: "show", args: []string{"show", "--password", "abc123"}, input: strings.NewReader("")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			started := make(chan struct{})
			client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				close(started)
				<-req.Context().Done()
				return nil, req.Context().Err()
			})}

			app, _, stderr := testApplication(serverConfig(t, "https://paste.example.com"), tt.input)
			app.httpClient = client
			app.ctx = ctx
			app.readPassword = testPasswordReader("top-secret")
			done := make(chan int, 1)
			go func() {
				done <- app.run(tt.args)
			}()

			select {
			case <-started:
			case <-time.After(time.Second):
				t.Fatal("request did not start")
			}
			cancel()

			select {
			case code := <-done:
				if code != 1 {
					t.Fatalf("exit = %d, stderr = %q", code, stderr.String())
				}
			case <-time.After(time.Second):
				t.Fatal("request did not stop after cancellation")
			}
			if got := stderr.String(); !strings.Contains(got, "context canceled") {
				t.Fatalf("stderr = %q", got)
			}
			if strings.Contains(stderr.String(), "top-secret") {
				t.Fatalf("stderr exposed password: %q", stderr.String())
			}
		})
	}
}

func TestRunCompletion(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "zsh", args: []string{"completion", "zsh"}, want: "#compdef pb"},
		{name: "bash", args: []string{"completion", "bash"}, want: "complete -F _pb pb"},
		{name: "fish", args: []string{"completion", "fish"}, want: "complete -c pb -f"},
		{name: "help", args: []string{"completion", "--help"}, want: "pb completion <zsh|bash|fish>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, stdout, stderr := testApplication("/missing/config.json", strings.NewReader(""))
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

func TestRunCompletionRejectsUnsupportedShell(t *testing.T) {
	app, stdout, stderr := testApplication("/missing/config.json", strings.NewReader(""))
	if code := app.run([]string{"completion", "powershell"}); code != 2 {
		t.Fatalf("exit = %d, stderr = %q", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "unsupported shell") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}
