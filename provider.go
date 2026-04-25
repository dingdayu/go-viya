package viya

import "errors"

// ErrDefaultClientNotSet is returned when the process-wide default client has
// not been configured.
var ErrDefaultClientNotSet = errors.New("viya default client is not set")

var defaultClient *Client

// SetDefaultClient stores the process-wide default client.
//
// Prefer passing *Client explicitly in new code. The default client exists for
// applications that want package-level wiring around a single SAS Viya deployment.
// Passing nil clears the default client.
func SetDefaultClient(client *Client) {
	defaultClient = client
}

// GetDefaultClient returns the process-wide default client.
//
// It returns ErrDefaultClientNotSet when SetDefaultClient has not been called
// or when the default client has been cleared.
func GetDefaultClient() (*Client, error) {
	if defaultClient == nil {
		return nil, ErrDefaultClientNotSet
	}
	return defaultClient, nil
}

// MustGetDefaultClient returns the process-wide default client.
//
// It panics with ErrDefaultClientNotSet when SetDefaultClient has not been
// called or when the default client has been cleared.
func MustGetDefaultClient() *Client {
	client, err := GetDefaultClient()
	if err != nil {
		panic(err)
	}
	return client
}
