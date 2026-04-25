package viya

import (
	"context"
	"fmt"
	"net/http"

	"resty.dev/v3"
)

type Option func(*clientOptions)

type clientOptions struct {
	rt            http.RoundTripper
	tokenProvider TokenProvider
}

func WithRoundTripper(rt http.RoundTripper) Option {
	return func(o *clientOptions) {
		o.rt = rt
	}
}

func WithTokenProvider(provider TokenProvider) Option {
	return func(o *clientOptions) {
		o.tokenProvider = provider
	}
}

type Client struct {
	baseURL string

	client        *resty.Client
	tokenProvider TokenProvider
}

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

func (c *Client) TokenURL() string {
	return fmt.Sprintf("%s/SASLogon/oauth/token", c.baseURL)
}
