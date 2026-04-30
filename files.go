package viya

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"go.opentelemetry.io/otel/codes"
)

// FileListOptions configures SAS Viya Files Service collection paging and filtering.
type FileListOptions struct {
	Start      int
	Limit      int
	FilterName string
}

// ViyaFile describes a file stored in the SAS Viya Files Service.
type ViyaFile struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	ContentType       string    `json:"contentType,omitempty"`
	Size              int64     `json:"size,omitempty"`
	CreatedBy         string    `json:"createdBy,omitempty"`
	CreationTimeStamp time.Time `json:"creationTimeStamp,omitempty"`
	ModifiedBy        string    `json:"modifiedBy,omitempty"`
	ModifiedTimeStamp time.Time `json:"modifiedTimeStamp,omitempty"`
	Links             []Link    `json:"links,omitempty"`
	Version           int       `json:"version,omitempty"`
}

// ViyaFilesResponse is a collection of SAS Viya Files Service files.
type ViyaFilesResponse = ListResponse[ViyaFile]

// GetFiles returns files visible to the caller in the SAS Viya Files Service.
func (c *Client) GetFiles(ctx context.Context, opts FileListOptions) (resp ViyaFilesResponse, err error) {
	ctx, span := tracer.Start(ctx, "GetFiles")
	defer span.End()

	req := c.collectionRequest(ctx, ListOptions{Start: opts.Start, Limit: opts.Limit}).
		SetResult(&resp)
	if opts.FilterName != "" {
		req.SetQueryParam("filter", fmt.Sprintf("contains(name,'%s')", strings.ReplaceAll(opts.FilterName, "'", "''")))
	}

	r, err := req.Get("/files/files")
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get files, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// UploadFile uploads content to the SAS Viya Files Service.
func (c *Client) UploadFile(ctx context.Context, fileName string, contentType string, content []byte) (resp ViyaFile, err error) {
	return c.UploadFileFromReader(ctx, fileName, contentType, bytes.NewReader(content))
}

// UploadFileFromReader uploads content from r to the SAS Viya Files Service.
func (c *Client) UploadFileFromReader(ctx context.Context, fileName string, contentType string, r io.Reader) (resp ViyaFile, err error) {
	ctx, span := tracer.Start(ctx, "UploadFileFromReader")
	defer span.End()

	if contentType == "" {
		contentType = "application/octet-stream"
	}

	httpResp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json, application/vnd.sas.error+json").
		SetHeader("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, strings.ReplaceAll(fileName, `"`, `\"`))).
		SetContentType(contentType).
		SetBody(r).
		SetResult(&resp).
		Post("/files/files")
	if err != nil {
		return resp, err
	}
	if !httpResp.IsSuccess() {
		span.SetStatus(codes.Error, httpResp.String())
		return resp, fmt.Errorf("failed to upload file, status code: %d", httpResp.StatusCode())
	}

	return resp, nil
}

// DownloadFile downloads content from the SAS Viya Files Service.
func (c *Client) DownloadFile(ctx context.Context, fileID string) (content []byte, err error) {
	ctx, span := tracer.Start(ctx, "DownloadFile")
	defer span.End()

	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", "application/octet-stream, application/vnd.sas.error+json").
		Get(fmt.Sprintf("/files/files/%s/content", fileID))
	if err != nil {
		return nil, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return nil, fmt.Errorf("failed to download file, status code: %d", r.StatusCode())
	}

	defer r.Body.Close()
	content, err = io.ReadAll(r.Body)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	return content, nil
}
