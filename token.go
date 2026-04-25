package viya

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const defaultClientID = "go-viya"

var ErrViyaAuthFailed = errors.New("viya authentication failed")

// TokenProvider 用于向 batch Client 注入 Bearer Token。
// 你可以在上层对接 OAuth2 client credentials、缓存 token、自动刷新等逻辑。
type TokenProvider interface {
	Token(ctx context.Context) (string, error)
}

type tokenProviderOptions struct {
	baseURL      string
	clientID     string
	clientSecret string
}

// TokenProviderOption configures providers that can optionally use an OAuth2 client.
type TokenProviderOption func(*tokenProviderOptions)

// WithOAuthClient configures the OAuth2 client used by password/auth-code flows.
// clientSecret is optional for public clients.
func WithOAuthClient(clientID, clientSecret string) TokenProviderOption {
	return func(o *tokenProviderOptions) {
		o.clientID = clientID
		o.clientSecret = clientSecret
	}
}

// WithOAuthClientProvider reuses base URL and OAuth2 client settings from an
// existing ClientCredentialsTokenProvider.
func WithOAuthClientProvider(provider *ClientCredentialsTokenProvider) TokenProviderOption {
	return func(o *tokenProviderOptions) {
		if provider == nil {
			return
		}
		o.baseURL = provider.baseURL
		o.clientID = provider.clientID
		o.clientSecret = provider.clientSecret
	}
}

// ClientCredentialsTokenProvider provides bearer tokens using OAuth2 client credentials flow.
// It is also used as the shared credential holder by Password/AuthCode providers.
type ClientCredentialsTokenProvider struct {
	baseURL      string
	clientID     string
	clientSecret string
	httpClient   *http.Client

	mu          sync.Mutex
	tokenSource oauth2.TokenSource
}

func (p *ClientCredentialsTokenProvider) tokenURL() string {
	return p.baseURL + "/SASLogon/oauth/token"
}

func newTokenHTTPClient(capture bool) *http.Client {
	return &http.Client{
		Transport: otelhttp.NewTransport(
			http.DefaultTransport,
			otelhttp.WithSpanOptions(trace.WithAttributes(
				attribute.Bool("http.request.body.capture", capture),
			)),
		),
	}
}

func (p *ClientCredentialsTokenProvider) tokenContext(ctx context.Context) context.Context {
	if p.httpClient == nil {
		return ctx
	}
	return context.WithValue(ctx, oauth2.HTTPClient, p.httpClient)
}

func (p *ClientCredentialsTokenProvider) Token(ctx context.Context) (string, error) {
	p.mu.Lock()
	if p.tokenSource == nil {
		oauthCfg := &clientcredentials.Config{
			ClientID:     p.clientID,
			ClientSecret: p.clientSecret,
			TokenURL:     p.tokenURL(),
			AuthStyle:    oauth2.AuthStyleAutoDetect,
		}
		p.tokenSource = oauthCfg.TokenSource(p.tokenContext(ctx))
	}
	src := p.tokenSource
	p.mu.Unlock()

	tok, err := src.Token()
	if err != nil || tok == nil || tok.AccessToken == "" {
		return "", ErrViyaAuthFailed
	}

	return tok.AccessToken, nil
}

// PasswordTokenProvider provides bearer tokens using OAuth2 password flow.
type PasswordTokenProvider struct {
	*ClientCredentialsTokenProvider
	username string
	password string
}

func (p *PasswordTokenProvider) Token(ctx context.Context) (string, error) {
	p.mu.Lock()
	if p.tokenSource == nil {
		conf := &oauth2.Config{
			ClientID:     p.clientID,
			ClientSecret: p.clientSecret,
			Endpoint: oauth2.Endpoint{
				TokenURL:  p.tokenURL(),
				AuthStyle: oauth2.AuthStyleInParams,
			},
		}

		tokenCtx := p.tokenContext(ctx)
		tok, err := conf.PasswordCredentialsToken(tokenCtx, p.username, p.password)
		if err != nil {
			p.mu.Unlock()
			return "", ErrViyaAuthFailed
		}
		p.tokenSource = conf.TokenSource(tokenCtx, tok)
	}
	src := p.tokenSource
	p.mu.Unlock()

	tok, err := src.Token()
	if err != nil || tok == nil || tok.AccessToken == "" {
		return "", ErrViyaAuthFailed
	}

	return tok.AccessToken, nil
}

// AuthCodeTokenProvider provides bearer tokens using OAuth2 authorization code flow.
type AuthCodeTokenProvider struct {
	*ClientCredentialsTokenProvider
	code string
}

func (p *AuthCodeTokenProvider) Token(ctx context.Context) (string, error) {
	p.mu.Lock()
	if p.tokenSource == nil {
		conf := &oauth2.Config{
			ClientID:     p.clientID,
			ClientSecret: p.clientSecret,
			Endpoint: oauth2.Endpoint{
				TokenURL:  p.tokenURL(),
				AuthStyle: oauth2.AuthStyleInParams,
			},
		}

		tokenCtx := p.tokenContext(ctx)
		tok, err := conf.Exchange(tokenCtx, p.code)
		if err != nil {
			p.mu.Unlock()
			return "", ErrViyaAuthFailed
		}
		p.tokenSource = conf.TokenSource(tokenCtx, tok)
	}
	src := p.tokenSource
	p.mu.Unlock()

	tok, err := src.Token()
	if err != nil || tok == nil || tok.AccessToken == "" {
		return "", ErrViyaAuthFailed
	}

	return tok.AccessToken, nil
}

func newCredentialBase(options tokenProviderOptions, requireSecret bool) (*ClientCredentialsTokenProvider, error) {
	baseURL := options.baseURL
	clientID := options.clientID
	clientSecret := options.clientSecret

	if baseURL == "" {
		return nil, ErrViyaAuthFailed
	}
	if clientID == "" {
		clientID = defaultClientID
	}
	if requireSecret && clientSecret == "" {
		return nil, ErrViyaAuthFailed
	}

	return &ClientCredentialsTokenProvider{
		baseURL:      baseURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   newTokenHTTPClient(false),
		tokenSource:  nil,
	}, nil
}

func providerOptions(baseURL string, opts ...TokenProviderOption) tokenProviderOptions {
	options := tokenProviderOptions{
		baseURL:  baseURL,
		clientID: defaultClientID,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	if options.baseURL == "" {
		options.baseURL = baseURL
	}
	if options.clientID == "" {
		options.clientID = defaultClientID
	}
	return options
}

func NewPasswordTokenProvider(baseURL, username, password string, opts ...TokenProviderOption) (TokenProvider, error) {
	if username == "" || password == "" {
		return nil, ErrViyaAuthFailed
	}

	options := providerOptions(baseURL, opts...)
	base, err := newCredentialBase(options, false)
	if err != nil {
		return nil, ErrViyaAuthFailed
	}

	return &PasswordTokenProvider{
		ClientCredentialsTokenProvider: base,
		username:                       username,
		password:                       password,
	}, nil
}

func NewAuthCodeTokenProvider(baseURL, code string, opts ...TokenProviderOption) (TokenProvider, error) {
	if code == "" {
		return nil, ErrViyaAuthFailed
	}

	options := providerOptions(baseURL, opts...)
	base, err := newCredentialBase(options, false)
	if err != nil {
		return nil, ErrViyaAuthFailed
	}

	return &AuthCodeTokenProvider{
		ClientCredentialsTokenProvider: base,
		code:                           code,
	}, nil
}

func NewClientCredentialsTokenProvider(baseURL, clientID, clientSecret string) (*ClientCredentialsTokenProvider, error) {
	base, err := newCredentialBase(tokenProviderOptions{
		baseURL:      baseURL,
		clientID:     clientID,
		clientSecret: clientSecret,
	}, true)
	if err != nil {
		return nil, ErrViyaAuthFailed
	}

	return base, nil
}
