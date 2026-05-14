package viya

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

// testServer creates an httptest.Server with the given handler and
// returns a *Client connected to it. The server is automatically
// cleaned up when the test completes.
func testServer(t *testing.T, h http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()

	server := httptest.NewServer(h)
	t.Cleanup(server.Close)

	u, err := url.Parse(server.URL)
	require.NoError(t, err)

	client := NewClient(t.Context(), WithBaseURL(u), WithTokenProvider(staticTokenProvider("test-token")))
	return client, server
}

// testTokenServer creates an httptest server that responds to
// /SASLogon/oauth/token with a JSON token response.
func testTokenServer(t *testing.T, h http.HandlerFunc) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(h)
	t.Cleanup(server.Close)
	return server
}
