package viya

import (
	"context"
	"fmt"
)

// SAS Viya CAS Management API reference:
// https://developer.sas.com/rest-apis/casManagement

// LoadCASTableToMemory loads a table from a CAS library into memory.
//
// serverID identifies the CAS server. caslibName and tableName identify the CAS
// library and table. replace controls whether an existing in-memory table can be
// replaced, and scope is passed to the CAS Management API state-change request.
func (c *Client) LoadCASTableToMemory(ctx context.Context, serverID string, caslibName string, tableName string, replace bool, scope string) error {
	body := map[string]any{
		"outputCaslibName": caslibName,
		"outputTableName":  tableName,
		"replace":          replace,
		"scope":            scope,
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetQueryParam("value", "loaded").
		SetBody(body).
		Put(fmt.Sprintf("%s/state", casTablePath(serverID, caslibName, tableName)))
	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("failed to load CAS table to memory, status code: %d", resp.StatusCode())
	}
	return nil
}

// UnloadCASTableFromMemory unloads a table from CAS memory.
//
// In SAS Visual Analytics workflows, unloading a table can let reports reload
// the latest source data the next time they access the table.
func (c *Client) UnloadCASTableFromMemory(ctx context.Context, serverID string, caslibName string, tableName string) error {
	resp, err := c.client.R().
		SetContext(ctx).
		SetQueryParam("value", "unloaded").
		Put(fmt.Sprintf("%s/state", casTablePath(serverID, caslibName, tableName)))
	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("failed to unload CAS table from memory, status code: %d, body: %s", resp.StatusCode(), resp.String())
	}
	return nil
}
