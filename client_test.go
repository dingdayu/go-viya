package viya

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

type staticTokenProvider string

func (p staticTokenProvider) Token(context.Context) (string, error) {
	return string(p), nil
}

func TestParseURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{name: "valid URL", url: "https://viya.example.com", wantErr: false},
		{name: "valid URL with port", url: "https://viya.example.com:8443", wantErr: false},
		{name: "invalid URL", url: "://bad", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			opt, err := ParseURL(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, opt)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, opt)
			}
		})
	}
}

func TestTokenURL(t *testing.T) {
	t.Parallel()

	baseURL, _ := url.Parse("https://viya.example.com")
	client := NewClient(context.Background(), WithBaseURL(baseURL))
	assert.Equal(t, "https://viya.example.com/SASLogon/oauth/token", client.TokenURL())
}

func TestTokenURLWithTrailingSlash(t *testing.T) {
	t.Parallel()

	baseURL, _ := url.Parse("https://viya.example.com/")
	client := NewClient(context.Background(), WithBaseURL(baseURL))
	// TokenURL uses fmt.Sprintf and does not remove trailing slash from baseURL;
	// the double-slash is a known cosmetic quirk.
	assert.Equal(t, "https://viya.example.com//SASLogon/oauth/token", client.TokenURL())
}

func TestWithRoundTripper(t *testing.T) {
	t.Parallel()

	baseURL, _ := url.Parse("https://viya.example.com")
	client := NewClient(context.Background(), WithBaseURL(baseURL), WithRoundTripper(nil))
	assert.NotNil(t, client)
}

func TestNilOptions(t *testing.T) {
	t.Parallel()

	baseURL, _ := url.Parse("https://viya.example.com")
	client := NewClient(context.Background(), WithBaseURL(baseURL))
	assert.NotNil(t, client)
}

func TestNewClientSetsBearerAuthorizationHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Header.Get("Authorization"), "Bearer token-value"; got != want {
			t.Fatalf("Authorization = %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	client := NewClient(t.Context(), WithBaseURL(u), WithTokenProvider(staticTokenProvider("token-value")))

	resp, err := client.client.R().SetContext(t.Context()).Get("/")
	if err != nil {
		t.Fatalf("GET error = %v", err)
	}
	if resp.StatusCode() != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.StatusCode(), http.StatusNoContent)
	}
}
