package viya

import "testing"

func TestNewPasswordTokenProviderDefaultsOAuthClient(t *testing.T) {
	provider, err := NewPasswordTokenProvider("https://viya.example.com", "user", "password")
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
	if passwordProvider.httpClient == nil {
		t.Fatal("httpClient = nil, want token HTTP client with capture disabled")
	}
}

func TestNewAuthCodeTokenProviderUsesOAuthClientProvider(t *testing.T) {
	base, err := NewClientCredentialsTokenProvider("https://viya.example.com", "client-id", "client-secret")
	if err != nil {
		t.Fatalf("NewClientCredentialsTokenProvider() error = %v", err)
	}

	provider, err := NewAuthCodeTokenProvider("", "auth-code", WithOAuthClientProvider(base))
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
	if authCodeProvider.httpClient == nil {
		t.Fatal("httpClient = nil, want token HTTP client with capture disabled")
	}
}

func TestNewClientCredentialsTokenProviderRequiresSecret(t *testing.T) {
	if _, err := NewClientCredentialsTokenProvider("https://viya.example.com", "client-id", ""); err == nil {
		t.Fatal("NewClientCredentialsTokenProvider() error = nil, want error")
	}
}
