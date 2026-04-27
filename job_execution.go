package viya

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"go.opentelemetry.io/otel/codes"
)

// JobExecutionError describes an error returned on a SAS Viya Job Execution job.
type JobExecutionError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// JobExecutionJob describes a SAS Viya Job Execution service job.
type JobExecutionJob struct {
	ID                string             `json:"id"`
	Name              string             `json:"name,omitempty"`
	State             string             `json:"state,omitempty"`
	CreatedBy         string             `json:"createdBy,omitempty"`
	CreationTimeStamp time.Time          `json:"creationTimeStamp,omitempty"`
	ModifiedBy        string             `json:"modifiedBy,omitempty"`
	ModifiedTimeStamp time.Time          `json:"modifiedTimeStamp,omitempty"`
	Results           map[string]string  `json:"results,omitempty"`
	Error             *JobExecutionError `json:"error,omitempty"`
	Links             []Link             `json:"links,omitempty"`
	Version           int                `json:"version,omitempty"`
}

// JobExecutionJobsResponse is a collection of SAS Viya Job Execution jobs.
type JobExecutionJobsResponse = ListResponse[JobExecutionJob]

// SubmitJobExecutionCodeRequest is the request for submitting SAS code through Job Execution.
type SubmitJobExecutionCodeRequest struct {
	Name        string
	Code        string
	ContextName string
}

type jobExecutionSubmitBody struct {
	Name          string                 `json:"name,omitempty"`
	JobDefinition jobExecutionDefinition `json:"jobDefinition"`
	Arguments     map[string]string      `json:"arguments,omitempty"`
}

type jobExecutionDefinition struct {
	Type string `json:"type"`
	Code string `json:"code"`
}

// SubmitJobExecutionCode submits SAS code as an asynchronous Job Execution job.
func (c *Client) SubmitJobExecutionCode(ctx context.Context, req SubmitJobExecutionCodeRequest) (resp JobExecutionJob, err error) {
	ctx, span := tracer.Start(ctx, "SubmitJobExecutionCode")
	defer span.End()

	body := jobExecutionSubmitBody{
		Name: req.Name,
		JobDefinition: jobExecutionDefinition{
			Type: "Compute",
			Code: req.Code,
		},
	}
	if req.ContextName != "" {
		body.Arguments = map[string]string{"_contextName": req.ContextName}
	}

	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json, application/vnd.sas.error+json").
		SetContentType("application/json").
		SetBody(body).
		SetResult(&resp).
		Post("/jobExecution/jobs")
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to submit job execution code, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// GetJobExecutionJob returns details for a SAS Viya Job Execution job.
func (c *Client) GetJobExecutionJob(ctx context.Context, jobID string) (resp JobExecutionJob, err error) {
	ctx, span := tracer.Start(ctx, "GetJobExecutionJob")
	defer span.End()

	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json, application/vnd.sas.error+json").
		SetResult(&resp).
		Get(jobExecutionJobPath(jobID))
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get job execution job, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// GetJobExecutionJobs returns Job Execution jobs visible to the caller.
func (c *Client) GetJobExecutionJobs(ctx context.Context, opts ListOptions) (resp JobExecutionJobsResponse, err error) {
	ctx, span := tracer.Start(ctx, "GetJobExecutionJobs")
	defer span.End()

	r, err := c.collectionRequest(ctx, opts).
		SetResult(&resp).
		Get("/jobExecution/jobs")
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get job execution jobs, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// CancelJobExecutionJob cancels a Job Execution job.
func (c *Client) CancelJobExecutionJob(ctx context.Context, jobID string) (err error) {
	ctx, span := tracer.Start(ctx, "CancelJobExecutionJob")
	defer span.End()

	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", "application/vnd.sas.error+json").
		Delete(jobExecutionJobPath(jobID))
	if err != nil {
		return err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return fmt.Errorf("failed to cancel job execution job, status code: %d", r.StatusCode())
	}

	return nil
}

// GetJobExecutionJobLog retrieves the log text for a Job Execution job when available.
func (c *Client) GetJobExecutionJobLog(ctx context.Context, jobID string) (string, error) {
	ctx, span := tracer.Start(ctx, "GetJobExecutionJobLog")
	defer span.End()

	job, err := c.GetJobExecutionJob(ctx, jobID)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}

	logURI := jobExecutionLogURI(job.Results)
	if logURI == "" {
		if job.Error != nil && job.Error.Message != "" {
			return fmt.Sprintf("Job %s: %s", job.State, job.Error.Message), nil
		}
		return fmt.Sprintf("No log available. Job state: %s", job.State), nil
	}

	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", "text/plain, application/octet-stream, application/vnd.sas.error+json").
		Get(logURI + "/content")
	if err != nil {
		return "", err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return "", fmt.Errorf("failed to get job execution job log, status code: %d", r.StatusCode())
	}

	defer r.Body.Close()
	content, err := io.ReadAll(r.Body)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", fmt.Errorf("failed to read job execution job log: %w", err)
	}

	return string(content), nil
}

func jobExecutionLogURI(results map[string]string) string {
	for key, value := range results {
		if strings.HasSuffix(key, ".log.txt") {
			return value
		}
	}
	for key, value := range results {
		if strings.HasSuffix(key, ".log") {
			return value
		}
	}
	return ""
}

func jobExecutionJobPath(jobID string) string {
	return fmt.Sprintf("/jobExecution/jobs/%s", url.PathEscape(jobID))
}
