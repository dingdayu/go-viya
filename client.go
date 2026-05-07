package viya

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	resty "resty.dev/v3"
)

// ParseURL creates a Client option from the given SAS Viya base URL.
// The base URL should be the root of a SAS Viya deployment, without a trailing service path.
func ParseURL(baseURL string) (Option, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	return WithBaseURL(u), nil
}

// Option configures a Client.
type Option func(*clientOptions)

type clientOptions struct {
	baseURL        *url.URL
	rt             http.RoundTripper
	tokenProvider  TokenProvider
	authMiddleware resty.RequestMiddleware
}

// WithRoundTripper configures the HTTP transport used by the underlying Resty client.
// It is primarily useful for tests, tracing, proxies, or custom TLS settings.
func WithRoundTripper(rt http.RoundTripper) Option {
	return func(o *clientOptions) {
		o.rt = rt
	}
}

// WithAuthMiddleware injects a custom authentication middleware for outgoing requests.
// If set, this middleware is used instead of the default token-provider middleware.
// The middleware should set the Authorization header or other auth details on the request.
func WithAuthMiddleware(mw resty.RequestMiddleware) Option {
	return func(o *clientOptions) {
		o.authMiddleware = mw
	}
}

// WithTokenProvider configures the client to use the given TokenProvider for authentication.
// It is mutually exclusive with WithAuthMiddleware; if both are set, WithAuthMiddleware takes precedence.
func WithTokenProvider(tp TokenProvider) Option {
	return func(o *clientOptions) {
		o.tokenProvider = tp
	}
}

// WithBaseURL configures the client's base URL for API requests.
func WithBaseURL(u *url.URL) Option {
	return func(o *clientOptions) {
		o.baseURL = u
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
//
// The ctx parameter is reserved for future context-aware initialization.
// Callers may pass context.Background() if no specific context is needed.
func NewClient(ctx context.Context, opts ...Option) *Client {
	cfg := &clientOptions{
		baseURL:        nil,
		rt:             nil,
		tokenProvider:  nil,
		authMiddleware: nil,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}

	client := resty.New().SetBaseURL(cfg.baseURL.String())
	if cfg.rt != nil {
		client.SetTransport(cfg.rt)
	}

	result := &Client{
		baseURL:       cfg.baseURL.String(),
		client:        client,
		tokenProvider: cfg.tokenProvider,
	}

	if cfg.authMiddleware != nil {
		client.AddRequestMiddleware(cfg.authMiddleware)
	} else if cfg.tokenProvider != nil {
		provider := cfg.tokenProvider
		client.AddRequestMiddleware(func(_ *resty.Client, r *resty.Request) error {
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
