package viya

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewPasswordTokenProviderDefaultsOAuthClient(t *testing.T) {
	baseURL, _ := url.Parse("https://viya.example.com")
	provider, err := NewPasswordTokenProvider("user", "password", baseURL)
	if err != nil {
		t.Fatalf("NewPasswordTokenProvider() error = %v", err)
	}

	passwordProvider, ok := provider.(*PasswordTokenProvider)
	if !ok {
		t.Fatalf("provider type = %T, want *PasswordTokenProvider", provider)
	}
	if passwordProvider.clientID != defaultClientID {
		t.Fatalf("clientID = %q, want %q", passwordProvider.clientID, defaultClientID)
	}
	if passwordProvider.clientSecret != "" {
		t.Fatalf("clientSecret = %q, want empty", passwordProvider.clientSecret)
	}
}

func TestNewAuthCodeTokenProviderUsesOAuthClientProvider(t *testing.T) {
	baseURL, _ := url.Parse("https://viya.example.com")
	base, err := NewClientCredentialsTokenProvider("client-id", "client-secret", baseURL)
	if err != nil {
		t.Fatalf("NewClientCredentialsTokenProvider() error = %v", err)
	}

	provider, err := NewAuthCodeTokenProvider("auth-code", baseURL, WithOAuthClientProvider(base))
	if err != nil {
		t.Fatalf("NewAuthCodeTokenProvider() error = %v", err)
	}

	authCodeProvider, ok := provider.(*AuthCodeTokenProvider)
	if !ok {
		t.Fatalf("provider type = %T, want *AuthCodeTokenProvider", provider)
	}
	if authCodeProvider.baseURL != base.baseURL {
		t.Fatalf("baseURL = %q, want %q", authCodeProvider.baseURL, base.baseURL)
	}
	if authCodeProvider.clientID != base.clientID {
		t.Fatalf("clientID = %q, want %q", authCodeProvider.clientID, base.clientID)
	}
	if authCodeProvider.clientSecret != base.clientSecret {
		t.Fatalf("clientSecret = %q, want %q", authCodeProvider.clientSecret, base.clientSecret)
	}
}

func TestNewClientCredentialsTokenProviderRequiresSecret(t *testing.T) {
	baseURL, _ := url.Parse("https://viya.example.com")
	if _, err := NewClientCredentialsTokenProvider("client-id", "", baseURL); err == nil {
		t.Fatal("NewClientCredentialsTokenProvider() error = nil, want error")
	}
}

func TestClientCredentialsTokenProviderRefreshUsesCurrentContext(t *testing.T) {
	var requests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/SASLogon/oauth/token"; got != want {
			t.Errorf("path = %q, want %q", got, want)
		}
		if got, want := r.FormValue("grant_type"), "client_credentials"; got != want {
			t.Errorf("grant_type = %q, want %q", got, want)
		}

		tokenNumber := atomic.AddInt32(&requests, 1)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{
			"access_token": "token-" + strconv.FormatInt(int64(tokenNumber), 10),
			"token_type":   "bearer",
			"expires_in":   3600,
		}); err != nil {
			t.Errorf("encode token response: %v", err)
		}
	}))
	defer server.Close()

	serverURL, _ := url.Parse(server.URL)
	provider, err := NewClientCredentialsTokenProvider("client-id", "client-secret", serverURL)
	if err != nil {
		t.Fatalf("NewClientCredentialsTokenProvider() error = %v", err)
	}

	firstCtx, cancel := context.WithCancel(t.Context())
	token, err := provider.Token(firstCtx)
	if err != nil {
		t.Fatalf("Token(firstCtx) error = %v", err)
	}
	if token != "token-1" {
		t.Fatalf("Token(firstCtx) = %q, want token-1", token)
	}
	cancel()

	provider.mu.Lock()
	provider.token.Expiry = time.Now().Add(-time.Hour)
	provider.mu.Unlock()

	secondCtx, secondCancel := context.WithTimeout(t.Context(), time.Second)
	defer secondCancel()
	token, err = provider.Token(secondCtx)
	if err != nil {
		t.Fatalf("Token(secondCtx) error = %v", err)
	}
	if token != "token-2" {
		t.Fatalf("Token(secondCtx) = %q, want token-2", token)
	}
	if got := atomic.LoadInt32(&requests); got != 2 {
		t.Fatalf("token requests = %d, want 2", got)
	}
}
