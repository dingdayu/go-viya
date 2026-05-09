package viya

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const defaultClientID = "go-viya"

// ErrViyaAuthFailed is returned when SAS Viya authentication cannot produce a bearer token.
var ErrViyaAuthFailed = errors.New("viya authentication failed")

// ErrInvalidParameter is returned when required parameters are missing or invalid.
type ErrInvalidParameter struct {
	Parameter string
	Reason    string
}

func (e *ErrInvalidParameter) Error() string {
	return fmt.Sprintf("invalid parameter %s: %s", e.Parameter, e.Reason)
}

func (e *ErrInvalidParameter) Unwrap() error {
	return nil
}

// TokenProvider supplies bearer tokens for authenticated SAS Viya requests.
//
// Implementations may use SAS Logon OAuth2 grants, cached tokens, or an external
// credential source. Implementations should honor ctx cancellation where possible.
// In distributed services, implement this interface with your own shared cache,
// secret storage, refresh-token rotation, and cross-instance locking strategy.
// This package intentionally consumes only bearer access tokens and does not
// expose refresh tokens.
// SAS Logon API reference:
// https://developer.sas.com/rest-apis/SASLogon
type TokenProvider interface {
	Token(ctx context.Context) (string, error)
}

type tokenProviderOptions struct {
	baseURL      *url.URL
	clientID     string
	clientSecret string
}

// TokenProviderOption configures token providers that use SAS Logon OAuth2 clients.
type TokenProviderOption func(*tokenProviderOptions)

// WithOAuthClient configures the OAuth2 client used by password and authorization-code flows.
//
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

// ClientCredentialsTokenProvider provides bearer tokens using the OAuth2 client credentials flow.
//
// It is also used as the shared credential holder by Password/AuthCode providers.
type ClientCredentialsTokenProvider struct {
	baseURL      *url.URL
	clientID     string
	clientSecret string
	httpClient   *http.Client

	mu    sync.Mutex
	token *oauth2.Token
}

func (p *ClientCredentialsTokenProvider) tokenURL() string {
	return p.baseURL.String() + "/SASLogon/oauth/token"
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

// Token returns a bearer token from SAS Logon, refreshing the cached token as needed.
func (p *ClientCredentialsTokenProvider) Token(ctx context.Context) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.token.Valid() {
		return p.token.AccessToken, nil
	}

	oauthCfg := &clientcredentials.Config{
		ClientID:     p.clientID,
		ClientSecret: p.clientSecret,
		TokenURL:     p.tokenURL(),
		AuthStyle:    oauth2.AuthStyleAutoDetect,
	}
	tok, err := oauthCfg.Token(p.tokenContext(ctx))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrViyaAuthFailed, err)
	}
	if tok == nil || tok.AccessToken == "" {
		return "", ErrViyaAuthFailed
	}

	p.token = tok
	return tok.AccessToken, nil
}

// PasswordTokenProvider provides bearer tokens using the OAuth2 password flow.
//
// This flow requires SAS Logon to allow password grants for the configured OAuth client.
type PasswordTokenProvider struct {
	*ClientCredentialsTokenProvider
	username string
	password string
}

// Token returns a bearer token for the configured username and password, refreshing the cached token as needed.
func (p *PasswordTokenProvider) Token(ctx context.Context) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.token.Valid() {
		return p.token.AccessToken, nil
	}

	conf := &oauth2.Config{
		ClientID:     p.clientID,
		ClientSecret: p.clientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL:  p.tokenURL(),
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}

	tokenCtx := p.tokenContext(ctx)
	var (
		tok *oauth2.Token
		err error
	)
	if p.token == nil {
		tok, err = conf.PasswordCredentialsToken(tokenCtx, p.username, p.password)
	} else {
		tok, err = conf.TokenSource(tokenCtx, p.token).Token()
	}
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrViyaAuthFailed, err)
	}
	if tok == nil || tok.AccessToken == "" {
		return "", ErrViyaAuthFailed
	}

	p.token = tok
	return tok.AccessToken, nil
}

// AuthCodeTokenProvider provides bearer tokens using the OAuth2 authorization code flow.
type AuthCodeTokenProvider struct {
	*ClientCredentialsTokenProvider
	code string
}

// Token exchanges the configured authorization code for a bearer token, then refreshes the cached token as needed.
func (p *AuthCodeTokenProvider) Token(ctx context.Context) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.token.Valid() {
		return p.token.AccessToken, nil
	}

	conf := &oauth2.Config{
		ClientID:     p.clientID,
		ClientSecret: p.clientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL:  p.tokenURL(),
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}

	tokenCtx := p.tokenContext(ctx)
	var (
		tok *oauth2.Token
		err error
	)
	if p.token == nil {
		tok, err = conf.Exchange(tokenCtx, p.code)
	} else {
		tok, err = conf.TokenSource(tokenCtx, p.token).Token()
	}
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrViyaAuthFailed, err)
	}
	if tok == nil || tok.AccessToken == "" {
		return "", ErrViyaAuthFailed
	}

	p.token = tok
	return tok.AccessToken, nil
}

func newCredentialBase(options tokenProviderOptions, requireSecret bool) (*ClientCredentialsTokenProvider, error) {
	baseURL := options.baseURL
	clientID := options.clientID
	clientSecret := options.clientSecret

	if baseURL == nil {
		return nil, &ErrInvalidParameter{Parameter: "baseURL", Reason: "must not be nil"}
	}
	if clientID == "" {
		clientID = defaultClientID
	}
	if requireSecret && clientSecret == "" {
		return nil, &ErrInvalidParameter{Parameter: "clientSecret", Reason: "must not be empty for client credentials flow"}
	}

	return &ClientCredentialsTokenProvider{
		baseURL:      baseURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   newTokenHTTPClient(false),
	}, nil
}

// NewPasswordTokenProvider creates a token provider that authenticates with username and password.
func NewPasswordTokenProvider(username, password string, baseURL *url.URL, opts ...TokenProviderOption) (TokenProvider, error) {
	if username == "" {
		return nil, &ErrInvalidParameter{Parameter: "username", Reason: "must not be empty"}
	}
	if password == "" {
		return nil, &ErrInvalidParameter{Parameter: "password", Reason: "must not be empty"}
	}

	options := tokenProviderOptions{
		baseURL:  baseURL,
		clientID: defaultClientID,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	base, err := newCredentialBase(options, false)
	if err != nil {
		return nil, err
	}

	return &PasswordTokenProvider{
		ClientCredentialsTokenProvider: base,
		username:                       username,
		password:                       password,
	}, nil
}

// NewAuthCodeTokenProvider creates a token provider that exchanges an OAuth2 authorization code.
func NewAuthCodeTokenProvider(code string, baseURL *url.URL, opts ...TokenProviderOption) (TokenProvider, error) {
	if code == "" {
		return nil, &ErrInvalidParameter{Parameter: "code", Reason: "must not be empty"}
	}
	if baseURL == nil {
		return nil, &ErrInvalidParameter{Parameter: "baseURL", Reason: "must not be nil"}
	}

	options := tokenProviderOptions{
		baseURL:  baseURL,
		clientID: defaultClientID,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	// Placeholder allows auth-code flow to proceed without required secret validation
	if options.clientSecret == "" {
		options.clientSecret = "placeholder-for-auth-code"
	}

	base, err := newCredentialBase(options, false)
	if err != nil {
		return nil, err
	}

	return &AuthCodeTokenProvider{
		ClientCredentialsTokenProvider: base,
		code:                           code,
	}, nil
}

// NewClientCredentialsTokenProvider creates a token provider for service-to-service authentication.
func NewClientCredentialsTokenProvider(clientID, clientSecret string, baseURL *url.URL) (*ClientCredentialsTokenProvider, error) {
	options := tokenProviderOptions{
		baseURL:      baseURL,
		clientID:     clientID,
		clientSecret: clientSecret,
	}
	base, err := newCredentialBase(options, true)
	if err != nil {
		return nil, err
	}

	return base, nil
}
