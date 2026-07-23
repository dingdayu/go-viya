package viya

import (
	"context"
	"fmt"
	"net/url"

	"go.opentelemetry.io/otel/codes"
)

// SAS Viya CAS Management API reference:
// https://developer.sas.com/rest-apis/casManagement

// LoadCASTableToMemory loads a table from a CAS library into memory.
//
// serverID identifies the CAS server. caslibName and tableName identify the CAS
// library and table. replace controls whether an existing in-memory table can be
// replaced, and scope is passed to the CAS Management API state-change request.
func (c *Client) LoadCASTableToMemory(ctx context.Context, serverID string, caslibName string, tableName string, replace bool, scope string) error {
	if serverID == "" {
		return &ErrInvalidParameter{Parameter: "serverID", Reason: "must not be empty"}
	}
	if caslibName == "" {
		return &ErrInvalidParameter{Parameter: "caslibName", Reason: "must not be empty"}
	}
	if tableName == "" {
		return &ErrInvalidParameter{Parameter: "tableName", Reason: "must not be empty"}
	}

	ctx, span := tracer.Start(ctx, "LoadCASTableToMemory")
	defer span.End()

	body := map[string]any{
		"outputCaslibName": caslibName,
		"outputTableName":  tableName,
		"replace":          replace,
		"scope":            scope,
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", AcceptJSONError).
		SetQueryParam("value", "loaded").
		SetBody(body).
		Put(fmt.Sprintf("/casManagement/servers/%s/caslibs/%s/tables/%s/state", serverID, url.PathEscape(caslibName), url.PathEscape(tableName)))
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	if !resp.IsStatusSuccess() {
		span.SetStatus(codes.Error, resp.String())
		return fmt.Errorf("failed to load CAS table to memory, status code: %d", resp.StatusCode())
	}
	return nil
}

// UnloadCASTableFromMemory unloads a table from CAS memory.
//
// In SAS Visual Analytics workflows, unloading a table can let reports reload
// the latest source data the next time they access the table.
func (c *Client) UnloadCASTableFromMemory(ctx context.Context, serverID string, caslibName string, tableName string) error {
	if serverID == "" {
		return &ErrInvalidParameter{Parameter: "serverID", Reason: "must not be empty"}
	}
	if caslibName == "" {
		return &ErrInvalidParameter{Parameter: "caslibName", Reason: "must not be empty"}
	}
	if tableName == "" {
		return &ErrInvalidParameter{Parameter: "tableName", Reason: "must not be empty"}
	}

	ctx, span := tracer.Start(ctx, "UnloadCASTableFromMemory")
	defer span.End()

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", AcceptJSONError).
		SetQueryParam("value", "unloaded").
		Put(fmt.Sprintf("/casManagement/servers/%s/caslibs/%s/tables/%s/state", serverID, url.PathEscape(caslibName), url.PathEscape(tableName)))
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	if !resp.IsStatusSuccess() {
		span.SetStatus(codes.Error, resp.String())
		return fmt.Errorf("failed to unload CAS table from memory, status code: %d, body: %s", resp.StatusCode(), resp.String())
	}
	return nil
}
