package viya

var defaultClient *Client

// SetDefaultClient stores the process-wide default client.
//
// Prefer passing *Client explicitly in new code. The default client exists for
// applications that want package-level wiring around a single SAS Viya deployment.
func SetDefaultClient(client *Client) {
	defaultClient = client
}

// GetDefaultClient returns the process-wide default client, if one has been set.
func GetDefaultClient() *Client {
	return defaultClient
}
