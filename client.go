package viya

import (
	"context"
	"fmt"
	"net/http"

	"resty.dev/v3"
)

// Option configures a Client.
type Option func(*clientOptions)

type clientOptions struct {
	rt            http.RoundTripper
	tokenProvider TokenProvider
}

// WithRoundTripper configures the HTTP transport used by the underlying Resty client.
// It is primarily useful for tests, tracing, proxies, or custom TLS settings.
func WithRoundTripper(rt http.RoundTripper) Option {
	return func(o *clientOptions) {
		o.rt = rt
	}
}

// WithTokenProvider configures a provider that supplies bearer tokens for each request.
// Token lookup happens lazily in request middleware so callers' contexts can cancel token fetches.
func WithTokenProvider(provider TokenProvider) Option {
	return func(o *clientOptions) {
		o.tokenProvider = provider
	}
}

// Client is a small SAS Viya REST API client.
//
// New clients should be constructed with NewClient and a SAS Viya base URL, for example
// "https://example.viya.sas.com". Authentication can be supplied with WithTokenProvider.
type Client struct {
	baseURL string

	client        *resty.Client
	tokenProvider TokenProvider
}

// NewClient creates a SAS Viya REST client bound to baseURL.
//
// The baseURL should be the root of a SAS Viya deployment, without a trailing service path.
// See the SAS Viya REST API usage notes:
// https://developer.sas.com/docs/rest-apis/getting-started/authentication
func NewClient(ctx context.Context, baseURL string, opts ...Option) *Client {
	_ = ctx

	cfg := &clientOptions{
		rt:            nil,
		tokenProvider: nil,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}

	client := resty.New().SetBaseURL(baseURL)
	if cfg.rt != nil {
		client.SetTransport(cfg.rt)
	}

	result := &Client{
		baseURL:       baseURL,
		client:        client,
		tokenProvider: cfg.tokenProvider,
	}

	if cfg.tokenProvider != nil {
		provider := cfg.tokenProvider

		// Fetch token lazily per request; resty propagates provider errors to callers.
		result.client.AddRequestMiddleware(func(_ *resty.Client, r *resty.Request) error {
			token, err := provider.Token(r.Context())
			if err != nil {
				return err
			}
			if token == "" {
				return ErrViyaAuthFailed
			}
			r.SetAuthScheme("Bearer").SetAuthToken(token)
			return nil
		})
	}

	return result
}

// TokenURL returns the SAS Logon OAuth2 token endpoint for the client's base URL.
//
// SAS Logon API reference:
// https://developer.sas.com/rest-apis/SASLogon
func (c *Client) TokenURL() string {
	return fmt.Sprintf("%s/SASLogon/oauth/token", c.baseURL)
}
