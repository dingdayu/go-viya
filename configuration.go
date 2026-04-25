package viya

import "context"

type Metadata struct {
	CreationTimeStamp string   `json:"creationTimeStamp"`
	CreatedBy         string   `json:"createdBy"`
	ModifiedTimeStamp string   `json:"modifiedTimeStamp"`
	ModifiedBy        string   `json:"modifiedBy"`
	IsDefault         bool     `json:"isDefault"`
	MediaType         string   `json:"mediaType"`
	Services          []string `json:"services"`
}

type ConfigurationsResp = ListResponse[map[string]any]

func (c *Client) GetConfiguration(ctx context.Context, definitionName string) (string, error) {
	request := c.client.R().SetContext(ctx)

	response, err := request.SetQueryParam("definitionName", definitionName).Get("/configuration/configurations")
	if err != nil {
		return "", err
	}

	return response.String(), nil
}
