package viya

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"go.opentelemetry.io/otel/codes"
)

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
		Get(fmt.Sprintf("/batch/fileSets/%s/files/%s", fileSetId, url.PathEscape(fileName)))
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
		Get(fmt.Sprintf("/batch/fileSets/%s/files/%s", fileSetId, url.PathEscape(fileName)))
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

	err = c.uploadBatchFileFromReader(ctx, fileSetId, fileName, bytes.NewReader(content))
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

// UploadBatchFileFromReader uploads or replaces a file in a SAS Viya Batch file set
// from content read from r.
func (c *Client) UploadBatchFileFromReader(ctx context.Context, fileSetId string, fileName string, r io.Reader) (err error) {
	ctx, span := tracer.Start(ctx, "UploadBatchFileFromReader")
	defer span.End()

	err = c.uploadBatchFileFromReader(ctx, fileSetId, fileName, r)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

func (c *Client) uploadBatchFileFromReader(ctx context.Context, fileSetId string, fileName string, r io.Reader) (err error) {
	contextAccept := "application/vnd.sas.error+json"
	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", contextAccept).
		SetContentType("application/octet-stream").
		SetBody(r).
		Put(fmt.Sprintf("/batch/fileSets/%s/files/%s", fileSetId, url.PathEscape(fileName)))
	if err != nil {
		return err
	}
	if !resp.IsSuccess() {
		return fmt.Errorf("failed to upload batch file, status code: %d", resp.StatusCode())
	}

	return nil
}
