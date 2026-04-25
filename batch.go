package viya

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"go.opentelemetry.io/otel/codes"
)

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

type BatchContextsResponse = ListResponse[BatchContext]

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

type BatchFileSet struct {
	ID string `json:"id"` // JOB_20260423_025718_331_1

	CreatedBy         string    `json:"createdBy"`
	CreationTimeStamp time.Time `json:"creationTimeStamp"`
	ModifiedBy        string    `json:"modifiedBy"`
	ModifiedTimeStamp time.Time `json:"modifiedTimeStamp"`

	ContextID string `json:"contextId"`
	Version   int    `json:"version"`
	Links     []Link `json:"links"`
}

type BatchFileSetsResponse = ListResponse[BatchFileSet]

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

// --------------------------- File ----------------------------

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

type BatchFileResponse = ListResponse[BatchFile]

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

// --------------------------- Job ----------------------------

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

type BatchJobsResponse = ListResponse[BatchJob]

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

type LauncherOptions struct {
	BatchContextID    string `json:"batchContextId"`
	JobName           string `json:"jobName"`
	RequiredResources string `json:"requiredResources"`
	WorkloadQueueName string `json:"workloadQueueName"`
}

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
