package viya

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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

// BatchFile describes a file stored in a SAS Viya Batch file set.
type BatchFile struct {
	CreationTimeStamp time.Time `json:"creationTimeStamp"`
	CreatedBy         string    `json:"createdBy"`
	ModifiedTimeStamp time.Time `json:"modifiedTimeStamp"`
	ModifiedBy        string    `json:"modifiedBy"`
	FileSetID         string    `json:"fileSetId"`
	Name              string    `json:"name"`
	Size              int64     `json:"size"`
	Links             []Link    `json:"links"`
	Version           int       `json:"version"`
}

// BatchFileResponse is a collection of SAS Viya Batch files.
type BatchFileResponse = ListResponse[BatchFile]

// GetBatchFile returns files in a SAS Viya Batch file set.
func (c *Client) GetBatchFile(ctx context.Context, fileSetId string) (resp BatchFileResponse, err error) {
	ctx, span := tracer.Start(ctx, "GetBatchFileSetFiles")
	defer span.End()

	contextAccept := "application/json, application/vnd.sas.collection+json;version=2, application/vnd.sas.error+json"
	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", contextAccept).
		SetResult(&resp).
		Get(fmt.Sprintf("/batch/fileSets/%s/files", fileSetId))
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get batch file set files, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// GetBatchFileInfo returns metadata for one file in a SAS Viya Batch file set.
func (c *Client) GetBatchFileInfo(ctx context.Context, fileSetId string, fileName string) (resp BatchFile, err error) {
	ctx, span := tracer.Start(ctx, "GetBatchFileInfo")
	defer span.End()

	contextAccept := "application/json, application/vnd.sas.batch.file.set.file+json"
	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", contextAccept).
		SetResult(&resp).
		Get(fmt.Sprintf("/batch/fileSets/%s/files/%s", fileSetId, fileName))
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get batch file info, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// DownloadBatchFile downloads a file from a SAS Viya Batch file set.
func (c *Client) DownloadBatchFile(ctx context.Context, fileSetId string, fileName string) (content []byte, err error) {
	ctx, span := tracer.Start(ctx, "DownloadBatchFile")
	defer span.End()

	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", "application/octet-stream").
		Get(fmt.Sprintf("/batch/fileSets/%s/files/%s", fileSetId, fileName))
	if err != nil {
		return nil, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return nil, fmt.Errorf("failed to download batch file, status code: %d", r.StatusCode())
	}

	defer r.Body.Close()

	content, err = io.ReadAll(r.Body)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to read batch file content: %w", err)
	}

	return content, nil
}

// UploadBatchFile uploads or replaces a file in a SAS Viya Batch file set.
func (c *Client) UploadBatchFile(ctx context.Context, fileSetId string, fileName string, content []byte) (err error) {
	ctx, span := tracer.Start(ctx, "UploadBatchFile")
	defer span.End()

	contextAccept := "application/vnd.sas.error+json"
	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", contextAccept).
		SetContentType("application/octet-stream").
		SetBody(content).
		Put(fmt.Sprintf("/batch/fileSets/%s/files/%s", fileSetId, fileName))
	if err != nil {
		return err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return fmt.Errorf("failed to upload batch file, status code: %d", r.StatusCode())
	}

	return nil
}

// BatchJob describes a SAS Viya Batch job and its execution state.
type BatchJob struct {
	ID string `json:"id"`

	Name               string    `json:"name"`
	State              string    `json:"state"`
	CreatedBy          string    `json:"createdBy"`
	CreationTimeStamp  time.Time `json:"creationTimeStamp"`
	StartedTimeStamp   time.Time `json:"startedTimeStamp"`
	SubmittedTimeStamp time.Time `json:"submittedTimeStamp"`
	EndedTimeStamp     time.Time `json:"endedTimeStamp"`
	ModifiedBy         string    `json:"modifiedBy"`
	ModifiedTimeStamp  time.Time `json:"modifiedTimeStamp"`

	ContextID     string `json:"contextId"`
	FileSetID     string `json:"fileSetId"`
	ExecutionHost string `json:"executionHost"`
	LogFile       string `json:"logFile"`
	ProcessID     string `json:"processId"`
	WorkloadJobID string `json:"workloadJobId"`
	ReturnCode    int    `json:"returnCode"`
	Version       int    `json:"version"`
	Links         []Link `json:"links"`
}

// BatchJobsResponse is a collection of SAS Viya Batch jobs.
type BatchJobsResponse = ListResponse[BatchJob]

// GetBatchJobsList returns SAS Viya Batch jobs visible to the caller.
func (c *Client) GetBatchJobsList(ctx context.Context) (resp BatchJobsResponse, err error) {
	ctx, span := tracer.Start(ctx, "GetBatchJobsList")
	defer span.End()

	contextAccept := "application/json, application/vnd.sas.collection+json;version=2, application/vnd.sas.error+json"
	r, err := c.client.R().SetHeader("Accept", contextAccept).
		SetContext(ctx).
		SetResult(&resp).
		Get("/batch/jobs")
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get batch jobs, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// GetBatchJobInfo returns details for a SAS Viya Batch job.
func (c *Client) GetBatchJobInfo(ctx context.Context, jobId string) (resp BatchJob, err error) {
	ctx, span := tracer.Start(ctx, "GetBatchJobInfo")
	defer span.End()

	contextAccept := "application/json, application/vnd.sas.batch.job.state+json, application/vnd.sas.batch.job.state+json;version=1, application/vnd.sas.batch.job+json, application/vnd.sas.batch.job+json;version=1, application/vnd.sas.error+json"
	r, err := c.client.R().SetHeader("Accept", contextAccept).
		SetContext(ctx).
		SetResult(&resp).
		Get(fmt.Sprintf("/batch/jobs/%s", jobId))
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get batch job info, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// SubmitBatchJobRequest is the request body for creating a SAS Viya Batch job.
type SubmitBatchJobRequest struct {
	FileSetID       string          `json:"fileSetId"`
	LauncherOptions LauncherOptions `json:"launcherOptions"`
	RestartType     string          `json:"restartType"`
	SASProgramName  string          `json:"sasProgramName"`
	SASOptions      string          `json:"sasOptions"`
	WatchOutput     bool            `json:"watchOutput"`
	OutputDir       string          `json:"outputDir"`
	LogFile         string          `json:"logFile"`
	ListFile        string          `json:"listFile"`
	Version         int             `json:"version"`
}

// LauncherOptions configures how SAS Viya Batch launches a job.
type LauncherOptions struct {
	BatchContextID    string `json:"batchContextId"`
	JobName           string `json:"jobName"`
	RequiredResources string `json:"requiredResources"`
	WorkloadQueueName string `json:"workloadQueueName"`
}

// CreateBatchJob submits a SAS Viya Batch job.
//
// req.FileSetID should identify a file set that contains the SAS program named
// by req.SASProgramName and any related input files.
func (c *Client) CreateBatchJob(ctx context.Context, req SubmitBatchJobRequest) (resp BatchJob, err error) {
	ctx, span := tracer.Start(ctx, "CreateBatchJob")
	defer span.End()

	contextAccept := "application/json, application/vnd.sas.batch.job+json, application/vnd.sas.batch.job+json;version=1, application/vnd.sas.error+json"
	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", contextAccept).
		SetBody(req).
		SetContentType("application/json").
		SetResult(&resp).
		Post("/batch/jobs")
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to create batch job, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// DeleteBatchJob deletes a SAS Viya Batch job.
func (c *Client) DeleteBatchJob(ctx context.Context, jobId string) (err error) {
	ctx, span := tracer.Start(ctx, "DeleteBatchJob")
	defer span.End()

	r, err := c.client.R().
		SetContext(ctx).
		Delete(fmt.Sprintf("/batch/jobs/%s", jobId))
	if err != nil {
		return err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return fmt.Errorf("failed to delete batch job, status code: %d, body: %s", r.StatusCode(), r.String())
	}

	return nil
}

// WaitBatchJobCompleted polls a SAS Viya Batch job until it reaches a terminal state.
//
// It returns "completed" or "failed". The wait stops early when ctx is canceled
// or GetBatchJobInfo returns an error.
func (c *Client) WaitBatchJobCompleted(ctx context.Context, jobId string, interval time.Duration) (finalState string, err error) {
	ctx, span := tracer.Start(ctx, "WaitBatchJobCompleted")
	defer span.End()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-ticker.C:
			jobInfo, err := c.GetBatchJobInfo(ctx, jobId)
			if err != nil {
				return "", err
			}
			if jobInfo.State == "completed" || jobInfo.State == "failed" {
				return jobInfo.State, nil
			}
		}
	}
}
