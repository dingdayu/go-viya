package viya

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"go.opentelemetry.io/otel/codes"
)

// BatchServer describes a reusable SAS Viya Batch server.
type BatchServer struct {
	ID string `json:"id"`

	CreatedBy         string    `json:"createdBy"`
	CreationTimeStamp time.Time `json:"creationTimeStamp"`
	ModifiedBy        string    `json:"modifiedBy"`
	ModifiedTimeStamp time.Time `json:"modifiedTimeStamp"`

	ContextID   string `json:"contextId"`
	HostName    string `json:"hostName"`
	IsPermanent bool   `json:"isPermanent"`
	ProcessID   string `json:"processId"`
	QueueName   string `json:"queueName"`
	RuleID      string `json:"ruleId"`
	RunServerAs string `json:"runServerAs"`
	State       string `json:"state"`
	Version     int    `json:"version"`
	Links       []Link `json:"links"`
}

// BatchServersResponse is a collection of SAS Viya Batch servers.
type BatchServersResponse = ListResponse[BatchServer]

// GetBatchServersList returns reusable SAS Viya Batch servers visible to the caller.
func (c *Client) GetBatchServersList(ctx context.Context) (resp BatchServersResponse, err error) {
	ctx, span := tracer.Start(ctx, "GetBatchServersList")
	defer span.End()

	contextAccept := "application/json, application/vnd.sas.collection+json;version=2, application/vnd.sas.error+json"
	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", contextAccept).
		SetResult(&resp).
		Get("/batch/servers")
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get batch servers, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// GetBatchServerInfo returns details for a reusable SAS Viya Batch server.
func (c *Client) GetBatchServerInfo(ctx context.Context, serverId string) (resp BatchServer, err error) {
	ctx, span := tracer.Start(ctx, "GetBatchServerInfo")
	defer span.End()

	contextAccept := "application/json, application/vnd.sas.batch.server+json, application/vnd.sas.error+json"
	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", contextAccept).
		SetResult(&resp).
		Get(batchServerPath(serverId))
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get batch server info, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// DeleteBatchServer deletes a reusable SAS Viya Batch server.
func (c *Client) DeleteBatchServer(ctx context.Context, serverId string) (err error) {
	ctx, span := tracer.Start(ctx, "DeleteBatchServer")
	defer span.End()

	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", "application/vnd.sas.error+json").
		Delete(batchServerPath(serverId))
	if err != nil {
		return err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return fmt.Errorf("failed to delete batch server, status code: %d", r.StatusCode())
	}

	return nil
}

func batchServerPath(serverId string) string {
	return fmt.Sprintf("/batch/servers/%s", url.PathEscape(serverId))
}
