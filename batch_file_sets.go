package viya

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/codes"
)

// BatchFileSet describes a SAS Viya Batch file set.
//
// File sets group input, output, list, and log files associated with batch jobs.
type BatchFileSet struct {
	ID string `json:"id"`

	CreatedBy         string    `json:"createdBy"`
	CreationTimeStamp time.Time `json:"creationTimeStamp"`
	ModifiedBy        string    `json:"modifiedBy"`
	ModifiedTimeStamp time.Time `json:"modifiedTimeStamp"`

	ContextID string `json:"contextId"`
	Version   int    `json:"version"`
	Links     []Link `json:"links"`
}

// BatchFileSetsResponse is a collection of SAS Viya Batch file sets.
type BatchFileSetsResponse = ListResponse[BatchFileSet]

// GetBatchFileSetsList returns the available SAS Viya Batch file sets.
func (c *Client) GetBatchFileSetsList(ctx context.Context) (resp BatchFileSetsResponse, err error) {
	ctx, span := tracer.Start(ctx, "GetBatchFileSets")
	defer span.End()

	contextAccept := "application/json, application/vnd.sas.collection+json;version=2, application/vnd.sas.error+json"
	r, err := c.client.R().
		SetHeader("Accept", contextAccept).
		SetContext(ctx).
		SetResult(&resp).
		Get("/batch/fileSets")
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get batch file sets, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// GetBatchFileSetsInfo returns metadata for a single SAS Viya Batch file set.
func (c *Client) GetBatchFileSetsInfo(ctx context.Context, id string) (resp BatchFileSet, err error) {
	ctx, span := tracer.Start(ctx, "GetBatchFileSetsInfo")
	defer span.End()

	contextAccept := "application/json, application/vnd.sas.batch.file.set+json, application/vnd.sas.batch.file.set+json;version=1, application/vnd.sas.error+json"
	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", contextAccept).
		SetResult(&resp).
		Get(fmt.Sprintf("/batch/fileSets/%s", id))
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get batch file sets, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// CreateBatchFileSet creates a SAS Viya Batch file set for the supplied context ID.
func (c *Client) CreateBatchFileSet(ctx context.Context, contextId string) (resp BatchFileSet, err error) {
	ctx, span := tracer.Start(ctx, "CreateBatchFileSet")
	defer span.End()

	contextAccept := "application/json, application/vnd.sas.batch.file.set+json, application/vnd.sas.batch.file.set+json;version=1, application/vnd.sas.error+json"
	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", contextAccept).
		SetBody(map[string]any{
			"contextId": contextId,
		}).
		SetContentType("application/json").
		SetResult(&resp).
		Post("/batch/fileSets")
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to create batch file set, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// DeleteBatchFileSet deletes a SAS Viya Batch file set.
//
// SAS Viya can reject deletion with HTTP 409 when the file set is still referenced
// by another resource, such as a job.
func (c *Client) DeleteBatchFileSet(ctx context.Context, fileSetId string) (err error) {
	ctx, span := tracer.Start(ctx, "DeleteBatchFileSet")
	defer span.End()

	contextAccept := "application/vnd.sas.error+json"
	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", contextAccept).
		Delete(fmt.Sprintf("/batch/fileSets/%s", fileSetId))
	if err != nil {
		return err
	}

	if !r.IsSuccess() {
		statusCode := r.StatusCode()
		rawBody := r.String()

		span.SetStatus(codes.Error, rawBody)

		if statusCode == 409 {
			var errorResp ErrorResponse
			if json.Unmarshal([]byte(rawBody), &errorResp) == nil {
				span.SetStatus(codes.Error, errorResp.Error())
				return fmt.Errorf("failed to delete batch file set due to conflict: %s", errorResp.Error())
			}
		}

		return fmt.Errorf("failed to delete batch file set, status code: %d, body: %s", statusCode, rawBody)
	}

	return nil
}
