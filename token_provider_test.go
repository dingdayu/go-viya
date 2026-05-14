package viya

import (
	"context"
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"resty.dev/v3"
)

func TestErrInvalidParameterError(t *testing.T) {
	t.Parallel()

	err := &ErrInvalidParameter{Parameter: "jobID", Reason: "must not be empty"}
	assert.Equal(t, "invalid parameter jobID: must not be empty", err.Error())
	assert.False(t, errors.Is(err, ErrViyaAuthFailed))
}

func TestErrViyaAuthFailedSentinel(t *testing.T) {
	t.Parallel()
	assert.ErrorIs(t, ErrViyaAuthFailed, ErrViyaAuthFailed)
}

func TestNewClientWithAllOptions(t *testing.T) {
	t.Parallel()

	baseURL, err := url.Parse("https://viya.example.com")
	require.NoError(t, err)

	client := NewClient(context.Background(),
		WithBaseURL(baseURL),
		WithRoundTripper(nil),
		WithTokenProvider(staticTokenProvider("test-token")),
	)
	assert.NotNil(t, client)
}

func TestWithAuthMiddlewareAndTokenProviderConflict(t *testing.T) {
	t.Parallel()

	baseURL, err := url.Parse("https://viya.example.com")
	require.NoError(t, err)

	client := NewClient(context.Background(),
		WithBaseURL(baseURL),
		WithTokenProvider(staticTokenProvider("token-from-provider")),
		WithAuthMiddleware(func(_ *resty.Client, _ *resty.Request) error {
			return nil
		}),
	)
	assert.NotNil(t, client)
}
