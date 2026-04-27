package viya

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"go.opentelemetry.io/otel/codes"
)

// ComputeSession describes a SAS Viya Compute session.
type ComputeSession struct {
	ID                   string                    `json:"id"`
	Name                 string                    `json:"name"`
	Description          string                    `json:"description"`
	State                string                    `json:"state"`
	Owner                string                    `json:"owner"`
	ServerID             string                    `json:"serverId"`
	ServiceAPIVersion    int                       `json:"serviceAPIVersion"`
	SessionConditionCode int                       `json:"sessionConditionCode"`
	StateElapsedTime     int                       `json:"stateElapsedTime"`
	Environment          *ComputeEnvironment       `json:"environment,omitempty"`
	Attributes           map[string]any            `json:"attributes,omitempty"`
	Resources            []ComputeExternalResource `json:"resources,omitempty"`
	LogStatistics        *ComputeOutputStatistics  `json:"logStatistics,omitempty"`
	ListingStatistics    *ComputeOutputStatistics  `json:"listingStatistics,omitempty"`
	CreationTimeStamp    time.Time                 `json:"creationTimeStamp"`
	Version              int                       `json:"version"`
	Links                []Link                    `json:"links"`
}

// ComputeSessionsResponse is a collection of SAS Viya Compute sessions.
type ComputeSessionsResponse = ListResponse[ComputeSession]

// CreateComputeSessionRequest is the request body for creating a Compute session.
type CreateComputeSessionRequest struct {
	Version     int                       `json:"version,omitempty"`
	Name        string                    `json:"name,omitempty"`
	Description string                    `json:"description,omitempty"`
	Attributes  map[string]any            `json:"attributes,omitempty"`
	Environment *ComputeEnvironment       `json:"environment,omitempty"`
	Resources   []ComputeExternalResource `json:"resources,omitempty"`
}

// CreateComputeSession creates a Compute session from the specified context definition.
func (c *Client) CreateComputeSession(ctx context.Context, contextId string, req CreateComputeSessionRequest) (resp ComputeSession, err error) {
	ctx, span := tracer.Start(ctx, "CreateComputeSession")
	defer span.End()

	contextAccept := "application/json, application/vnd.sas.compute.session+json, application/vnd.sas.error+json"
	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", contextAccept).
		SetContentType("application/json").
		SetBody(req).
		SetResult(&resp).
		Post(fmt.Sprintf("%s/sessions", computeContextPath(contextId)))
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to create compute session, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// GetComputeSessionsList returns active SAS Viya Compute sessions visible to the caller.
func (c *Client) GetComputeSessionsList(ctx context.Context) (resp ComputeSessionsResponse, err error) {
	ctx, span := tracer.Start(ctx, "GetComputeSessionsList")
	defer span.End()

	contextAccept := "application/json, application/vnd.sas.collection+json;version=2, application/vnd.sas.error+json"
	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", contextAccept).
		SetHeader("Accept-Item", "application/vnd.sas.compute.session+json").
		SetResult(&resp).
		Get("/compute/sessions")
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get compute sessions, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// GetComputeSessionInfo returns information about a SAS Viya Compute session.
func (c *Client) GetComputeSessionInfo(ctx context.Context, sessionId string) (resp ComputeSession, err error) {
	ctx, span := tracer.Start(ctx, "GetComputeSessionInfo")
	defer span.End()

	contextAccept := "application/json, application/vnd.sas.compute.session+json, application/vnd.sas.error+json"
	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", contextAccept).
		SetResult(&resp).
		Get(computeSessionPath(sessionId))
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get compute session info, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// DeleteComputeSession deletes a SAS Viya Compute session.
func (c *Client) DeleteComputeSession(ctx context.Context, sessionId string) (err error) {
	ctx, span := tracer.Start(ctx, "DeleteComputeSession")
	defer span.End()

	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", "application/vnd.sas.error+json").
		Delete(computeSessionPath(sessionId))
	if err != nil {
		return err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return fmt.Errorf("failed to delete compute session, status code: %d", r.StatusCode())
	}

	return nil
}

// GetComputeSessionState returns the plain-text state for a SAS Viya Compute session.
func (c *Client) GetComputeSessionState(ctx context.Context, sessionId string) (state string, err error) {
	ctx, span := tracer.Start(ctx, "GetComputeSessionState")
	defer span.End()

	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", "text/plain, application/vnd.sas.error+json").
		Get(fmt.Sprintf("%s/state", computeSessionPath(sessionId)))
	if err != nil {
		return "", err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return "", fmt.Errorf("failed to get compute session state, status code: %d", r.StatusCode())
	}

	return strings.TrimSpace(r.String()), nil
}

// SetComputeSessionState updates the state of a SAS Viya Compute session.
//
// SAS Viya requires the current ETag in ifMatch for this operation.
func (c *Client) SetComputeSessionState(ctx context.Context, sessionId string, state string, ifMatch string) (newState string, err error) {
	ctx, span := tracer.Start(ctx, "SetComputeSessionState")
	defer span.End()

	req := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", "text/plain, application/vnd.sas.error+json").
		SetQueryParam("value", state)
	if ifMatch != "" {
		req.SetHeader("If-Match", ifMatch)
	}

	r, err := req.Put(fmt.Sprintf("%s/state", computeSessionPath(sessionId)))
	if err != nil {
		return "", err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return "", fmt.Errorf("failed to set compute session state, status code: %d", r.StatusCode())
	}

	return strings.TrimSpace(r.String()), nil
}

// CancelComputeSession requests cancellation of running work in a SAS Viya Compute session.
func (c *Client) CancelComputeSession(ctx context.Context, sessionId string, ifMatch string) (state string, err error) {
	return c.SetComputeSessionState(ctx, sessionId, "canceled", ifMatch)
}

func computeSessionPath(sessionId string) string {
	return fmt.Sprintf("/compute/sessions/%s", url.PathEscape(sessionId))
}
