package viya

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel/codes"
)

// ComputeJob describes a SAS Viya Compute job submitted to a session.
type ComputeJob struct {
	ID                string                   `json:"id"`
	SessionID         string                   `json:"sessionId"`
	Name              string                   `json:"name"`
	Description       string                   `json:"description"`
	State             string                   `json:"state"`
	JobConditionCode  int                      `json:"jobConditionCode"`
	CreationTime      time.Time                `json:"creationTimeStamp"`
	CompletedTime     time.Time                `json:"completedTimeStamp"`
	LogStatistics     *ComputeOutputStatistics `json:"logStatistics,omitempty"`
	ListingStatistics *ComputeOutputStatistics `json:"listingStatistics,omitempty"`
	Version           int                      `json:"version"`
	Links             []Link                   `json:"links"`
}

// ComputeJobsResponse is a collection of SAS Viya Compute jobs.
type ComputeJobsResponse = ListResponse[ComputeJob]

// CreateComputeJobRequest is the request body for submitting SAS code to a Compute session.
type CreateComputeJobRequest struct {
	Version     int                       `json:"version,omitempty"`
	Name        string                    `json:"name,omitempty"`
	Description string                    `json:"description,omitempty"`
	Environment *ComputeEnvironment       `json:"environment,omitempty"`
	Variables   map[string]any            `json:"variables,omitempty"`
	Code        []string                  `json:"code,omitempty"`
	CodeURI     string                    `json:"codeUri,omitempty"`
	Resources   []ComputeExternalResource `json:"resources,omitempty"`
	Attributes  map[string]any            `json:"attributes,omitempty"`
}

// CreateComputeJob executes SAS code asynchronously in a SAS Viya Compute session.
func (c *Client) CreateComputeJob(ctx context.Context, sessionId string, req CreateComputeJobRequest) (resp ComputeJob, err error) {
	ctx, span := tracer.Start(ctx, "CreateComputeJob")
	defer span.End()

	contextAccept := "application/json, application/vnd.sas.compute.job+json, application/vnd.sas.error+json"
	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", contextAccept).
		SetContentType("application/json").
		SetBody(req).
		SetResult(&resp).
		Post(fmt.Sprintf("/compute/sessions/%s/jobs", sessionId))
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to create compute job, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// GetComputeJobsList returns current jobs for a SAS Viya Compute session.
func (c *Client) GetComputeJobsList(ctx context.Context, sessionId string) (resp ComputeJobsResponse, err error) {
	ctx, span := tracer.Start(ctx, "GetComputeJobsList")
	defer span.End()

	contextAccept := "application/json, application/vnd.sas.collection+json;version=2, application/vnd.sas.error+json"
	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", contextAccept).
		SetResult(&resp).
		Get(fmt.Sprintf("/compute/sessions/%s/jobs", sessionId))
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get compute jobs, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// GetComputeJobInfo returns information about a SAS Viya Compute job.
func (c *Client) GetComputeJobInfo(ctx context.Context, sessionId string, jobId string) (resp ComputeJob, err error) {
	ctx, span := tracer.Start(ctx, "GetComputeJobInfo")
	defer span.End()

	contextAccept := "application/json, application/vnd.sas.compute.job+json, application/vnd.sas.error+json"
	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", contextAccept).
		SetResult(&resp).
		Get(fmt.Sprintf("/compute/sessions/%s/jobs/%s", sessionId, jobId))
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get compute job info, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// DeleteComputeJob deletes a SAS Viya Compute job and its session access points.
func (c *Client) DeleteComputeJob(ctx context.Context, sessionId string, jobId string) (err error) {
	ctx, span := tracer.Start(ctx, "DeleteComputeJob")
	defer span.End()

	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", "application/vnd.sas.error+json").
		Delete(fmt.Sprintf("/compute/sessions/%s/jobs/%s", sessionId, jobId))
	if err != nil {
		return err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return fmt.Errorf("failed to delete compute job, status code: %d", r.StatusCode())
	}

	return nil
}

// GetComputeJobState returns the plain-text state for a SAS Viya Compute job.
func (c *Client) GetComputeJobState(ctx context.Context, sessionId string, jobId string) (state string, err error) {
	ctx, span := tracer.Start(ctx, "GetComputeJobState")
	defer span.End()

	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", "text/plain, application/vnd.sas.error+json").
		Get(fmt.Sprintf("/compute/sessions/%s/jobs/%s/state", sessionId, jobId))
	if err != nil {
		return "", err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return "", fmt.Errorf("failed to get compute job state, status code: %d", r.StatusCode())
	}

	return strings.TrimSpace(r.String()), nil
}

// SetComputeJobState updates the state of a SAS Viya Compute job.
//
// SAS Viya requires the current ETag in ifMatch for this operation.
func (c *Client) SetComputeJobState(ctx context.Context, sessionId string, jobId string, state string, ifMatch string) (newState string, err error) {
	ctx, span := tracer.Start(ctx, "SetComputeJobState")
	defer span.End()

	req := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", "text/plain, application/vnd.sas.error+json").
		SetQueryParam("value", state)
	if ifMatch != "" {
		req.SetHeader("If-Match", ifMatch)
	}

	r, err := req.Put(fmt.Sprintf("/compute/sessions/%s/jobs/%s/state", sessionId, jobId))
	if err != nil {
		return "", err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return "", fmt.Errorf("failed to set compute job state, status code: %d", r.StatusCode())
	}

	return strings.TrimSpace(r.String()), nil
}

// CancelComputeJob requests cancellation of a SAS Viya Compute job.
func (c *Client) CancelComputeJob(ctx context.Context, sessionId string, jobId string, ifMatch string) (state string, err error) {
	return c.SetComputeJobState(ctx, sessionId, jobId, "canceled", ifMatch)
}
