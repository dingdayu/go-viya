package viya

var defaultClient *Client

func SetDefaultClient(client *Client) {
	defaultClient = client
}

func GetDefaultClient() *Client {
	return defaultClient
}
