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
		{name: "show", args: []string{"show", "--password", "top-secret", "abc123"}, input: strings.NewReader("")},
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
