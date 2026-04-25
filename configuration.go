package viya

import "context"

// Metadata contains metadata fields commonly returned by SAS Viya configuration resources.
type Metadata struct {
	CreationTimeStamp string   `json:"creationTimeStamp"`
	CreatedBy         string   `json:"createdBy"`
	ModifiedTimeStamp string   `json:"modifiedTimeStamp"`
	ModifiedBy        string   `json:"modifiedBy"`
	IsDefault         bool     `json:"isDefault"`
	MediaType         string   `json:"mediaType"`
	Services          []string `json:"services"`
}

// ConfigurationsResp is a SAS Viya configuration collection response.
//
// Configuration payloads vary by definition, so items are represented as map[string]any.
type ConfigurationsResp = ListResponse[map[string]any]

// GetConfiguration returns configuration instances for a SAS Viya definition name.
//
// definitionName is passed to /configuration/configurations as the definitionName query parameter.
// The response body is returned as a string because configuration shapes are dynamic.
func (c *Client) GetConfiguration(ctx context.Context, definitionName string) (string, error) {
	request := c.client.R().SetContext(ctx)

	response, err := request.SetQueryParam("definitionName", definitionName).Get("/configuration/configurations")
	if err != nil {
		return "", err
	}

	return response.String(), nil
}
