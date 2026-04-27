package viya

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"go.opentelemetry.io/otel/codes"
)

// ComputeContext describes a SAS Viya Compute context definition.
type ComputeContext struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	LaunchType  string `json:"launchType"`
	QueueName   string `json:"queueName"`

	LaunchContext map[string]any            `json:"launchContext,omitempty"`
	Attributes    map[string]any            `json:"attributes,omitempty"`
	MediaTypeMap  map[string]string         `json:"mediaTypeMap,omitempty"`
	Resources     []ComputeExternalResource `json:"resources,omitempty"`
	Environment   *ComputeEnvironment       `json:"environment,omitempty"`
	CreatedBy     string                    `json:"createdBy"`
	CreationTime  time.Time                 `json:"creationTimeStamp"`
	ModifiedBy    string                    `json:"modifiedBy"`
	ModifiedTime  time.Time                 `json:"modifiedTimeStamp"`
	Version       int                       `json:"version"`
	Links         []Link                    `json:"links"`
}

// ComputeContextsResponse is a collection of SAS Viya Compute contexts.
type ComputeContextsResponse = ListResponse[ComputeContext]

// GetComputeContexts returns available SAS Viya Compute context definitions.
func (c *Client) GetComputeContexts(ctx context.Context) (resp ComputeContextsResponse, err error) {
	ctx, span := tracer.Start(ctx, "GetComputeContexts")
	defer span.End()

	contextAccept := "application/json, application/vnd.sas.collection+json;version=2, application/vnd.sas.error+json"
	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", contextAccept).
		SetHeader("Accept-Item", "application/vnd.sas.compute.context+json").
		SetResult(&resp).
		Get("/compute/contexts")
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get compute contexts, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// GetComputeContextInfo returns a single SAS Viya Compute context definition.
func (c *Client) GetComputeContextInfo(ctx context.Context, contextId string) (resp ComputeContext, err error) {
	ctx, span := tracer.Start(ctx, "GetComputeContextInfo")
	defer span.End()

	contextAccept := "application/json, application/vnd.sas.compute.context+json, application/vnd.sas.error+json"
	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", contextAccept).
		SetResult(&resp).
		Get(computeContextPath(contextId))
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get compute context info, status code: %d", r.StatusCode())
	}

	return resp, nil
}

func computeContextPath(contextId string) string {
	return fmt.Sprintf("/compute/contexts/%s", url.PathEscape(contextId))
}
