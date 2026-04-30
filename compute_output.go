package viya

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel/codes"
)

// GetComputeJobLog returns log lines for a SAS Viya Compute job.
func (c *Client) GetComputeJobLog(ctx context.Context, sessionId string, jobId string) (resp ComputeLogLinesResponse, err error) {
	return c.getComputeLogLines(ctx, "GetComputeJobLog", fmt.Sprintf("/compute/sessions/%s/jobs/%s/log", sessionId, jobId), "failed to get compute job log")
}

// GetComputeJobListing returns listing lines for a SAS Viya Compute job.
func (c *Client) GetComputeJobListing(ctx context.Context, sessionId string, jobId string) (resp ComputeLogLinesResponse, err error) {
	return c.getComputeLogLines(ctx, "GetComputeJobListing", fmt.Sprintf("/compute/sessions/%s/jobs/%s/listing", sessionId, jobId), "failed to get compute job listing")
}

// GetComputeJobLogText returns the job log as plain text.
func (c *Client) GetComputeJobLogText(ctx context.Context, sessionId string, jobId string) (text string, err error) {
	return c.getComputeText(ctx, "GetComputeJobLogText", fmt.Sprintf("/compute/sessions/%s/jobs/%s/log", sessionId, jobId), "failed to get compute job log text")
}

// GetComputeJobListingText returns the job listing as plain text.
func (c *Client) GetComputeJobListingText(ctx context.Context, sessionId string, jobId string) (text string, err error) {
	return c.getComputeText(ctx, "GetComputeJobListingText", fmt.Sprintf("/compute/sessions/%s/jobs/%s/listing", sessionId, jobId), "failed to get compute job listing text")
}

func (c *Client) getComputeLogLines(ctx context.Context, spanName string, path string, operation string) (resp ComputeLogLinesResponse, err error) {
	ctx, span := tracer.Start(ctx, spanName)
	defer span.End()

	contextAccept := "application/json, application/vnd.sas.collection+json;version=2, application/vnd.sas.error+json"
	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", contextAccept).
		SetResult(&resp).
		Get(path)
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("%s, status code: %d", operation, r.StatusCode())
	}

	return resp, nil
}

func (c *Client) getComputeText(ctx context.Context, spanName string, path string, operation string) (text string, err error) {
	ctx, span := tracer.Start(ctx, spanName)
	defer span.End()

	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", "text/plain, application/vnd.sas.error+json").
		Get(path)
	if err != nil {
		return "", err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return "", fmt.Errorf("%s, status code: %d", operation, r.StatusCode())
	}

	return strings.TrimRight(r.String(), "\n"), nil
}
