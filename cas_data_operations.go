package viya

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"go.opentelemetry.io/otel/codes"
)

// SAS Viya CAS Management API reference:
// https://developer.sas.com/rest-apis/casManagement

// UploadCSVToCASTable uploads CSV data into a CAS table.
//
// The CSV data should include a header row. The target table is created in
// caslibName on serverID using the supplied tableName.
func (c *Client) UploadCSVToCASTable(ctx context.Context, serverID string, caslibName string, tableName string, csv []byte) (resp CASTable, err error) {
	return c.UploadCSVToCASTableFromReader(ctx, serverID, caslibName, tableName, bytes.NewReader(csv))
}

// UploadCSVToCASTableFromReader uploads CSV data from r into a CAS table.
//
// The CSV data should include a header row. The target table is created in
// caslibName on serverID using the supplied tableName.
func (c *Client) UploadCSVToCASTableFromReader(ctx context.Context, serverID string, caslibName string, tableName string, r io.Reader) (resp CASTable, err error) {
	ctx, span := tracer.Start(ctx, "UploadCSVToCASTableFromReader")
	defer span.End()

	httpResp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json, application/vnd.sas.error+json").
		SetMultipartFormData(map[string]string{
			"tableName":         tableName,
			"format":            "csv",
			"containsHeaderRow": "true",
		}).
		SetMultipartField("file", "data.csv", "text/csv", r).
		SetResult(&resp).
		Post(fmt.Sprintf("%s/tables", caslibPath(serverID, caslibName)))
	if err != nil {
		return resp, err
	}
	if !httpResp.IsSuccess() {
		span.SetStatus(codes.Error, httpResp.String())
		return resp, fmt.Errorf("failed to upload CSV to CAS table, status code: %d", httpResp.StatusCode())
	}

	return resp, nil
}

// PromoteCASTable promotes a CAS table to global scope.
//
// A promoted table is visible outside the session that created it.
func (c *Client) PromoteCASTable(ctx context.Context, serverID string, caslibName string, tableName string) (resp CASTable, err error) {
	ctx, span := tracer.Start(ctx, "PromoteCASTable")
	defer span.End()

	httpResp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json, application/vnd.sas.error+json").
		SetContentType("application/json").
		SetBody(map[string]any{"scope": "global"}).
		SetResult(&resp).
		Post(casTablePath(serverID, caslibName, tableName))
	if err != nil {
		return resp, err
	}
	if !httpResp.IsSuccess() {
		span.SetStatus(codes.Error, httpResp.String())
		return resp, fmt.Errorf("failed to promote CAS table, status code: %d", httpResp.StatusCode())
	}

	return resp, nil
}
