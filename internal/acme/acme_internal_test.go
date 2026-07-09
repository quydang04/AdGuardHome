package acme

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/challenge/http01"
)

func TestHTTPProvider(t *testing.T) {
	p := newHTTPProvider()

	const (
		token   = "test-token"
		keyAuth = "test-token.key-auth"
	)

	req := httptest.NewRequest(http.MethodGet, http01.ChallengePath(token), nil)
	w := httptest.NewRecorder()
	p.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("before Present: got status %d, want %d", w.Code, http.StatusNotFound)
	}

	err := p.Present("example.com", token, keyAuth)
	if err != nil {
		t.Fatalf("Present: %s", err)
	}

	w = httptest.NewRecorder()
	p.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("after Present: got status %d, want %d", w.Code, http.StatusOK)
	}

	if got := w.Body.String(); got != keyAuth {
		t.Errorf("after Present: got body %q, want %q", got, keyAuth)
	}

	err = p.CleanUp("example.com", token, keyAuth)
	if err != nil {
		t.Fatalf("CleanUp: %s", err)
	}

	w = httptest.NewRecorder()
	p.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("after CleanUp: got status %d, want %d", w.Code, http.StatusNotFound)
	}
}
