package viya

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel/codes"
)

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

// BatchJobInputRequest is the request body for sending STDIN to a running Batch job.
type BatchJobInputRequest struct {
	Input   []string `json:"input"`
	Version int      `json:"version"`
}

// SendBatchJobInput sends STDIN text lines to a running SAS Viya Batch job.
func (c *Client) SendBatchJobInput(ctx context.Context, jobId string, input []string) (err error) {
	ctx, span := tracer.Start(ctx, "SendBatchJobInput")
	defer span.End()

	contextAccept := "application/vnd.sas.error+json"
	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", contextAccept).
		SetContentType("application/json").
		SetBody(BatchJobInputRequest{
			Input:   input,
			Version: 1,
		}).
		Post(fmt.Sprintf("/batch/jobs/%s/input", jobId))
	if err != nil {
		return err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return fmt.Errorf("failed to send batch job input, status code: %d", r.StatusCode())
	}

	return nil
}

// BatchJobOutput describes STDOUT and STDERR text retrieved from a running Batch job.
type BatchJobOutput struct {
	Errors  []string `json:"errors"`
	Output  []string `json:"output"`
	Version int      `json:"version"`
}

// GetBatchJobOutput retrieves STDOUT and STDERR text from a running SAS Viya Batch job.
func (c *Client) GetBatchJobOutput(ctx context.Context, jobId string) (resp BatchJobOutput, err error) {
	ctx, span := tracer.Start(ctx, "GetBatchJobOutput")
	defer span.End()

	contextAccept := "application/json, application/vnd.sas.error+json"
	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", contextAccept).
		SetResult(&resp).
		Post(fmt.Sprintf("/batch/jobs/%s/output", jobId))
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get batch job output, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// GetBatchJobState returns the plain-text state for a SAS Viya Batch job.
func (c *Client) GetBatchJobState(ctx context.Context, jobId string) (state string, err error) {
	ctx, span := tracer.Start(ctx, "GetBatchJobState")
	defer span.End()

	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", "text/plain, application/vnd.sas.error+json").
		Get(fmt.Sprintf("/batch/jobs/%s/state", jobId))
	if err != nil {
		return "", err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return "", fmt.Errorf("failed to get batch job state, status code: %d", r.StatusCode())
	}

	return strings.TrimSpace(r.String()), nil
}

// CancelBatchJob requests cancellation of a SAS Viya Batch job.
func (c *Client) CancelBatchJob(ctx context.Context, jobId string) (err error) {
	ctx, span := tracer.Start(ctx, "CancelBatchJob")
	defer span.End()

	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", "application/vnd.sas.error+json").
		SetQueryParam("value", "canceled").
		Put(fmt.Sprintf("/batch/jobs/%s/state", jobId))
	if err != nil {
		return err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return fmt.Errorf("failed to cancel batch job, status code: %d", r.StatusCode())
	}

	return nil
}

// WaitBatchJobCompleted polls a SAS Viya Batch job until it reaches a terminal state.
//
// It returns the final job details when the job state is "completed" or "failed".
// The wait stops early when ctx is canceled or GetBatchJobInfo returns an error.
// In that case, the returned BatchJob contains the most recent job details, if any.
func (c *Client) WaitBatchJobCompleted(ctx context.Context, jobId string, interval time.Duration) (jobInfo BatchJob, err error) {
	ctx, span := tracer.Start(ctx, "WaitBatchJobCompleted")
	defer span.End()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return jobInfo, ctx.Err()
		case <-ticker.C:
			nextJobInfo, err := c.GetBatchJobInfo(ctx, jobId)
			if err != nil {
				return jobInfo, err
			}
			jobInfo = nextJobInfo
			if jobInfo.State == "completed" || jobInfo.State == "failed" {
				return jobInfo, nil
			}
		}
	}
}
