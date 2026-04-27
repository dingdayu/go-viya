package viya

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/codes"
)

// SAS Viya Batch API reference:
// https://developer.sas.com/rest-apis/batch

// BatchContext describes a SAS Viya Batch execution context.
//
// A context defines how batch jobs are launched, including queue, resource,
// server, and SAS option settings.
type BatchContext struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	State             string    `json:"state"`
	CreatedBy         string    `json:"createdBy"`
	CreationTimeStamp time.Time `json:"creationTimeStamp"`
	IsMultiServer     bool      `json:"isMultiServer"`
	LauncherContextID string    `json:"launcherContextId"`
	MaxServers        int       `json:"maxServers"`
	MinServers        int       `json:"minServers"`
	ModifiedBy        string    `json:"modifiedBy"`
	ModifiedTimeStamp time.Time `json:"modifiedTimeStamp"`
	ProviderOptions   string    `json:"providerOptions"`
	QueueName         string    `json:"queueName"`
	RequiredResources string    `json:"requiredResources"`
	RunServerAs       string    `json:"runServerAs"`
	SASOptions        string    `json:"sasOptions"`
	ServerTimeout     int       `json:"serverTimeout"`
	Version           int       `json:"version"`
	Links             []Link    `json:"links"`
}

// BatchContextsResponse is a collection of SAS Viya Batch contexts.
type BatchContextsResponse = ListResponse[BatchContext]

// GetBatchContexts returns the available SAS Viya Batch execution contexts.
func (c *Client) GetBatchContexts(ctx context.Context) (resp BatchContextsResponse, err error) {
	ctx, span := tracer.Start(ctx, "GetBatchContexts")
	defer span.End()

	contextAccept := "application/json, application/vnd.sas.collection+json;version=2, application/vnd.sas.error+json"
	r, err := c.client.R().
		SetHeader("Accept", contextAccept).
		SetContext(ctx).
		SetResult(&resp).
		Get("/batch/contexts")
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get batch contexts, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// GetBatchContextByName returns a SAS Viya Batch execution context by name.
func (c *Client) GetBatchContextByName(ctx context.Context, name string) (resp BatchContext, err error) {
	ctx, span := tracer.Start(ctx, "GetBatchContextByName")
	defer span.End()

	contextAccept := "application/json, application/vnd.sas.batch.context+json, application/vnd.sas.error+json"
	r, err := c.client.R().
		SetHeader("Accept", contextAccept).
		SetContext(ctx).
		SetQueryParam("name", name).
		SetResult(&resp).
		Get("/batch/contexts/@item")
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get batch context by name, status code: %d", r.StatusCode())
	}

	return resp, nil
}
