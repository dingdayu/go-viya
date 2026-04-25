package viya

import (
	"context"
	"fmt"
)

// SAS Viya CAS Management API reference:
// https://developer.sas.com/rest-apis/casManagement

// CAS_SERVER_NAME is the default shared CAS server name used by SAS Viya deployments.
const CAS_SERVER_NAME = "cas-shared-default"

// LoadCasLibTableToMemory loads a table from a CAS library into memory.
//
// casLibName and table identify the CAS library and table. replace controls whether
// an existing in-memory table can be replaced, and scope is passed to the CAS
// Management API state-change request.
func (c *Client) LoadCasLibTableToMemory(ctx context.Context, casLibName, table string, replace bool, scope string) error {
	body := map[string]any{
		"outputCaslibName": casLibName,
		"outputTableName":  table,
		"replace":          replace,
		"scope":            scope,
	}

	resp, err := c.client.R().SetContext(ctx).
		SetQueryParam("value", "loaded").
		SetBody(body).
		Put(fmt.Sprintf("/casManagement/servers/%s/caslibs/%s/tables/%s/state", CAS_SERVER_NAME, casLibName, table))
	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("failed to load CAS library table: %s", resp.Status())
	}
	return nil
}

// UnLoadCasLibTableInMemory unloads a table from CAS memory.
//
// In SAS Visual Analytics workflows, unloading a table can let reports reload
// the latest source data the next time they access the table.
func (c *Client) UnLoadCasLibTableInMemory(ctx context.Context, casLibName, table string) error {
	resp, err := c.client.R().SetContext(ctx).
		SetQueryParam("value", "unloaded").
		Put(fmt.Sprintf("/casManagement/servers/%s/caslibs/%s/tables/%s/state", CAS_SERVER_NAME, casLibName, table))
	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("failed to load CAS library table: %s, %s", resp.Status(), resp.String())
	}
	return nil
}
