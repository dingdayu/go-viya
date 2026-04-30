package viya

import (
	"context"
	"fmt"
	"net/url"

	"go.opentelemetry.io/otel/codes"
	"resty.dev/v3"
)

// ListOptions configures SAS Viya collection paging.
type ListOptions struct {
	Start int
	Limit int
}

// CASServer describes a CAS server returned by the CAS Management API.
type CASServer struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Version     int    `json:"version,omitempty"`
	Links       []Link `json:"links,omitempty"`
}

// CASServersResponse is a collection of SAS Viya CAS servers.
type CASServersResponse = ListResponse[CASServer]

// CASLib describes a CAS library returned by the CAS Management API.
type CASLib struct {
	Name        string `json:"name"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
	Version     int    `json:"version,omitempty"`
	Links       []Link `json:"links,omitempty"`
}

// CASLibsResponse is a collection of SAS Viya CAS libraries.
type CASLibsResponse = ListResponse[CASLib]

// CASTable describes a CAS table returned by the CAS Management API.
type CASTable struct {
	Name        string `json:"name"`
	CaslibName  string `json:"caslibName,omitempty"`
	RowCount    int64  `json:"rowCount,omitempty"`
	ColumnCount int    `json:"columnCount,omitempty"`
	Scope       string `json:"scope,omitempty"`
	SourceName  string `json:"sourceName,omitempty"`
	Version     int    `json:"version,omitempty"`
	Links       []Link `json:"links,omitempty"`
}

// CASTablesResponse is a collection of SAS Viya CAS tables.
type CASTablesResponse = ListResponse[CASTable]

// CASTableColumn describes a CAS table column.
type CASTableColumn struct {
	Name      string `json:"name"`
	Type      string `json:"type,omitempty"`
	RawLength int    `json:"rawLength,omitempty"`
	Length    int    `json:"length,omitempty"`
	Label     string `json:"label,omitempty"`
	Format    string `json:"format,omitempty"`
	Index     int    `json:"index,omitempty"`
}

// CASTableColumnsResponse is a collection of SAS Viya CAS table columns.
type CASTableColumnsResponse = ListResponse[CASTableColumn]

// CASTableRowsResponse contains sample rows from a CAS table.
type CASTableRowsResponse struct {
	Columns []string         `json:"columns"`
	Rows    []map[string]any `json:"rows"`
	Count   int              `json:"count"`
	Start   int              `json:"start"`
	Limit   int              `json:"limit"`
}

// GetCASServers returns CAS servers visible to the caller.
func (c *Client) GetCASServers(ctx context.Context, opts ListOptions) (resp CASServersResponse, err error) {
	ctx, span := tracer.Start(ctx, "GetCASServers")
	defer span.End()

	r, err := c.collectionRequest(ctx, opts).
		SetResult(&resp).
		Get("/casManagement/servers")
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get CAS servers, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// GetCASLibs returns CAS libraries for a CAS server.
func (c *Client) GetCASLibs(ctx context.Context, serverID string, opts ListOptions) (resp CASLibsResponse, err error) {
	ctx, span := tracer.Start(ctx, "GetCASLibs")
	defer span.End()

	r, err := c.collectionRequest(ctx, opts).
		SetResult(&resp).
		Get(fmt.Sprintf("/casManagement/servers/%s/caslibs", serverID))
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get CAS libraries, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// GetCASTables returns tables in a CAS library.
func (c *Client) GetCASTables(ctx context.Context, serverID string, caslibName string, opts ListOptions) (resp CASTablesResponse, err error) {
	ctx, span := tracer.Start(ctx, "GetCASTables")
	defer span.End()

	r, err := c.collectionRequest(ctx, opts).
		SetResult(&resp).
		Get(fmt.Sprintf("/casManagement/servers/%s/caslibs/%s/tables", serverID, url.PathEscape(caslibName)))
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get CAS tables, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// GetCASTableInfo returns metadata for a CAS table.
func (c *Client) GetCASTableInfo(ctx context.Context, serverID string, caslibName string, tableName string) (resp CASTable, err error) {
	ctx, span := tracer.Start(ctx, "GetCASTableInfo")
	defer span.End()

	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json, application/vnd.sas.error+json").
		SetResult(&resp).
		Get(fmt.Sprintf("/casManagement/servers/%s/caslibs/%s/tables/%s", serverID, url.PathEscape(caslibName), url.PathEscape(tableName)))
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get CAS table info, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// GetCASTableColumns returns column metadata for a CAS table.
func (c *Client) GetCASTableColumns(ctx context.Context, serverID string, caslibName string, tableName string, opts ListOptions) (resp CASTableColumnsResponse, err error) {
	ctx, span := tracer.Start(ctx, "GetCASTableColumns")
	defer span.End()

	r, err := c.collectionRequest(ctx, opts).
		SetResult(&resp).
		Get(fmt.Sprintf("/casManagement/servers/%s/caslibs/%s/tables/%s/columns", serverID, url.PathEscape(caslibName), url.PathEscape(tableName)))
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get CAS table columns, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// GetCASTableRows returns sample rows from a CAS table using dataTables and rowSets.
func (c *Client) GetCASTableRows(ctx context.Context, serverID string, caslibName string, tableName string, opts ListOptions) (resp CASTableRowsResponse, err error) {
	ctx, span := tracer.Start(ctx, "GetCASTableRows")
	defer span.End()

	if opts.Limit <= 0 {
		opts.Limit = 100
	}

	columns, err := c.getAllCASTableDataColumns(ctx, serverID, caslibName, tableName)
	if err != nil {
		return resp, err
	}

	var rowData casRowSetResponse
	r, err := c.collectionRequest(ctx, opts).
		SetResult(&rowData).
		Get(fmt.Sprintf("/rowSets/tables/%s/rows", url.PathEscape(casTableID(serverID, caslibName, tableName))))
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get CAS table rows, status code: %d", r.StatusCode())
	}

	columnNames := make([]string, 0, len(columns))
	for _, column := range columns {
		columnNames = append(columnNames, column.Name)
	}

	rows := make([]map[string]any, 0, len(rowData.Items))
	for _, item := range rowData.Items {
		row := make(map[string]any, len(item.Cells))
		for i, cell := range item.Cells {
			if i >= len(columnNames) {
				break
			}
			row[columnNames[i]] = cell
		}
		rows = append(rows, row)
	}

	return CASTableRowsResponse{
		Columns: columnNames,
		Rows:    rows,
		Count:   rowData.Count,
		Start:   opts.Start,
		Limit:   opts.Limit,
	}, nil
}

// GetCASTableDataColumns returns column metadata from the dataTables API.
func (c *Client) GetCASTableDataColumns(ctx context.Context, serverID string, caslibName string, tableName string, opts ListOptions) (resp CASTableColumnsResponse, err error) {
	ctx, span := tracer.Start(ctx, "GetCASTableDataColumns")
	defer span.End()

	r, err := c.collectionRequest(ctx, opts).
		SetResult(&resp).
		Get(fmt.Sprintf("/dataTables/dataSources/%s/tables/%s/columns", url.PathEscape(casDataSourceID(serverID, caslibName)), url.PathEscape(tableName)))
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get CAS data table columns, status code: %d", r.StatusCode())
	}

	return resp, nil
}

func (c *Client) getAllCASTableDataColumns(ctx context.Context, serverID string, caslibName string, tableName string) ([]CASTableColumn, error) {
	const pageLimit = 1000
	columns := make([]CASTableColumn, 0)
	for start := 0; ; start += pageLimit {
		page, err := c.GetCASTableDataColumns(ctx, serverID, caslibName, tableName, ListOptions{Start: start, Limit: pageLimit})
		if err != nil {
			return nil, err
		}
		columns = append(columns, page.Items...)
		if len(page.Items) < pageLimit || (page.Count > 0 && len(columns) >= page.Count) {
			return columns, nil
		}
	}
}

type casRowSetResponse struct {
	Count int `json:"count"`
	Items []struct {
		Cells []any `json:"cells"`
	} `json:"items"`
}

func (c *Client) collectionRequest(ctx context.Context, opts ListOptions) *resty.Request {
	req := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json, application/vnd.sas.collection+json;version=2, application/vnd.sas.error+json")
	if opts.Start > 0 {
		req.SetQueryParam("start", fmt.Sprintf("%d", opts.Start))
	}
	if opts.Limit > 0 {
		req.SetQueryParam("limit", fmt.Sprintf("%d", opts.Limit))
	}
	return req
}

func casDataSourceID(serverID string, caslibName string) string {
	return fmt.Sprintf("cas~fs~%s~fs~%s", serverID, caslibName)
}

func casTableID(serverID string, caslibName string, tableName string) string {
	return fmt.Sprintf("%s~fs~%s", casDataSourceID(serverID, caslibName), tableName)
}
