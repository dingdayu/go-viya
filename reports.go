package viya

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"go.opentelemetry.io/otel/codes"
)

// ReportListOptions configures SAS Viya report collection paging and filtering.
type ReportListOptions struct {
	Start      int
	Limit      int
	FilterName string
}

// Report describes a SAS Visual Analytics report.
type Report struct {
	ID                string         `json:"id"`
	Name              string         `json:"name,omitempty"`
	Description       string         `json:"description,omitempty"`
	CreatedBy         string         `json:"createdBy,omitempty"`
	CreationTimeStamp time.Time      `json:"creationTimeStamp,omitempty"`
	ModifiedBy        string         `json:"modifiedBy,omitempty"`
	ModifiedTimeStamp time.Time      `json:"modifiedTimeStamp,omitempty"`
	Definition        map[string]any `json:"definition,omitempty"`
	Links             []Link         `json:"links,omitempty"`
	Version           int            `json:"version,omitempty"`
}

// ReportsResponse is a collection of SAS Visual Analytics reports.
type ReportsResponse = ListResponse[Report]

// ReportImageJobRequest is the request for rendering a report section image.
type ReportImageJobRequest struct {
	ReportID      string
	SectionIndex  int
	Size          string
	LayoutType    string
	SelectionType string
	RenderLimit   int
}

// ReportImageJob describes a SAS Visual Analytics report image rendering job.
type ReportImageJob struct {
	ID                string         `json:"id"`
	State             string         `json:"state,omitempty"`
	ReportURI         string         `json:"reportUri,omitempty"`
	CreationTimeStamp time.Time      `json:"creationTimeStamp,omitempty"`
	ModifiedTimeStamp time.Time      `json:"modifiedTimeStamp,omitempty"`
	Results           map[string]any `json:"results,omitempty"`
	Links             []Link         `json:"links,omitempty"`
	Version           int            `json:"version,omitempty"`
}

// VisualAnalyticsReportRequest creates or modifies a report through the Visual Analytics operations API.
type VisualAnalyticsReportRequest struct {
	Version            int              `json:"version,omitempty"`
	ResultFolder       string           `json:"resultFolder,omitempty"`
	ResultReportName   string           `json:"resultReportName,omitempty"`
	ResultNameConflict string           `json:"resultNameConflict,omitempty"`
	Operations         []map[string]any `json:"operations,omitempty"`
}

// VisualAnalyticsReportResult describes the result of applying Visual Analytics report operations.
type VisualAnalyticsReportResult struct {
	ID               string           `json:"id,omitempty"`
	Version          int              `json:"version,omitempty"`
	ResultReportID   string           `json:"resultReportId,omitempty"`
	ResultReportName string           `json:"resultReportName,omitempty"`
	ResultReportURI  string           `json:"resultReportUri,omitempty"`
	ResultFolderURI  string           `json:"resultFolderUri,omitempty"`
	Operations       []map[string]any `json:"operations,omitempty"`
	Status           string           `json:"status,omitempty"`
	Links            []Link           `json:"links,omitempty"`
}

// DashboardSpec is an agent-friendly description of a simple Visual Analytics dashboard.
type DashboardSpec struct {
	Description string              `json:"description,omitempty"`
	Operations  []map[string]any    `json:"operations,omitempty"`
	Pages       []DashboardPageSpec `json:"pages,omitempty"`
}

// DashboardPageSpec describes a report page in a dashboard spec.
type DashboardPageSpec struct {
	Name    string                `json:"name"`
	Objects []DashboardObjectSpec `json:"objects,omitempty"`
}

// DashboardObjectSpec describes a visual object in a dashboard spec.
type DashboardObjectSpec struct {
	Type      string         `json:"type"`
	Title     string         `json:"title,omitempty"`
	Category  string         `json:"category,omitempty"`
	Measure   string         `json:"measure,omitempty"`
	Measures  []string       `json:"measures,omitempty"`
	DataRoles map[string]any `json:"dataRoles,omitempty"`
	Options   map[string]any `json:"options,omitempty"`
}

// CreateDashboardRequest creates a Visual Analytics report from a simplified dashboard spec.
type CreateDashboardRequest struct {
	Name               string
	FolderURI          string
	ResultNameConflict string
	ServerID           string
	CaslibName         string
	TableName          string
	Spec               DashboardSpec
}

type reportImageJobBody struct {
	ReportURI     string `json:"reportUri"`
	LayoutType    string `json:"layoutType"`
	SelectionType string `json:"selectionType"`
	SectionIndex  int    `json:"sectionIndex"`
	Size          string `json:"size"`
	RenderLimit   int    `json:"renderLimit"`
}

// GetReports returns SAS Visual Analytics reports visible to the caller.
func (c *Client) GetReports(ctx context.Context, opts ReportListOptions) (resp ReportsResponse, err error) {
	ctx, span := tracer.Start(ctx, "GetReports")
	defer span.End()

	req := c.collectionRequest(ctx, ListOptions{Start: opts.Start, Limit: opts.Limit}).
		SetResult(&resp)
	if opts.FilterName != "" {
		req.SetQueryParam("filter", fmt.Sprintf("contains(name,'%s')", strings.ReplaceAll(opts.FilterName, "'", "''")))
	}

	r, err := req.Get("/reports/reports")
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get reports, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// GetReport returns metadata and definition for a SAS Visual Analytics report.
func (c *Client) GetReport(ctx context.Context, reportID string) (resp Report, err error) {
	ctx, span := tracer.Start(ctx, "GetReport")
	defer span.End()

	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json, application/vnd.sas.error+json").
		SetResult(&resp).
		Get(reportPath(reportID))
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get report, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// CreateReportImageJob requests image rendering for a SAS Visual Analytics report section.
func (c *Client) CreateReportImageJob(ctx context.Context, req ReportImageJobRequest) (resp ReportImageJob, err error) {
	ctx, span := tracer.Start(ctx, "CreateReportImageJob")
	defer span.End()

	if req.LayoutType == "" {
		req.LayoutType = "thumbnail"
	}
	if req.SelectionType == "" {
		req.SelectionType = "perSection"
	}
	if req.Size == "" {
		req.Size = "800x600"
	}
	if req.RenderLimit <= 0 {
		req.RenderLimit = 1
	}

	body := reportImageJobBody{
		ReportURI:     reportPath(req.ReportID),
		LayoutType:    req.LayoutType,
		SelectionType: req.SelectionType,
		SectionIndex:  req.SectionIndex,
		Size:          req.Size,
		RenderLimit:   req.RenderLimit,
	}

	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", "application/vnd.sas.report.images.job+json, application/json, application/vnd.sas.error+json").
		SetContentType("application/vnd.sas.report.images.job.request+json").
		SetBody(body).
		SetResult(&resp).
		Post("/reportImages/jobs")
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to create report image job, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// CreateVisualAnalyticsReport creates a report by applying Visual Analytics operations.
func (c *Client) CreateVisualAnalyticsReport(ctx context.Context, req VisualAnalyticsReportRequest) (resp VisualAnalyticsReportResult, err error) {
	ctx, span := tracer.Start(ctx, "CreateVisualAnalyticsReport")
	defer span.End()

	if req.Version == 0 {
		req.Version = 1
	}

	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json, application/vnd.sas.report.operations.results+json, application/vnd.sas.report.operations.error+json, application/vnd.sas.error+json").
		SetContentType("application/vnd.sas.visual.analytics.reports.request+json").
		SetBody(req).
		SetResult(&resp).
		Post("/visualAnalytics/reports")
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to create Visual Analytics report, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// CreateDashboard creates a Visual Analytics report from a simple dashboard spec.
func (c *Client) CreateDashboard(ctx context.Context, req CreateDashboardRequest) (VisualAnalyticsReportResult, error) {
	operations := BuildDashboardOperations(req)
	conflict := req.ResultNameConflict
	if conflict == "" {
		conflict = "rename"
	}
	return c.CreateVisualAnalyticsReport(ctx, VisualAnalyticsReportRequest{
		Version:            1,
		ResultFolder:       req.FolderURI,
		ResultReportName:   req.Name,
		ResultNameConflict: conflict,
		Operations:         operations,
	})
}

// BuildDashboardOperations converts an agent-friendly dashboard spec into Visual Analytics operations.
func BuildDashboardOperations(req CreateDashboardRequest) []map[string]any {
	operations := make([]map[string]any, 0, 1+len(req.Spec.Pages)+len(req.Spec.Operations))
	dataSource := req.TableName
	if req.ServerID != "" && req.CaslibName != "" && req.TableName != "" {
		operations = append(operations, map[string]any{
			"addData": map[string]any{
				"cas": map[string]any{
					"server":  req.ServerID,
					"library": req.CaslibName,
					"table":   req.TableName,
				},
			},
		})
	}

	for _, page := range req.Spec.Pages {
		pageName := page.Name
		if pageName == "" {
			pageName = "Page"
		}
		operations = append(operations, map[string]any{
			"addPage": map[string]any{
				"label": pageName,
			},
		})

		for i, object := range page.Objects {
			objectType := object.Type
			if objectType == "" {
				objectType = "barChart"
			}
			roles := object.DataRoles
			if roles == nil {
				roles = dashboardObjectDataRoles(object)
			}
			options := object.Options
			if options == nil {
				options = map[string]any{}
			}
			if object.Title != "" {
				options["title"] = object.Title
			}

			objectBody := map[string]any{}
			if dataSource != "" {
				objectBody["dataSource"] = dataSource
			}
			if len(roles) > 0 {
				objectBody["dataRoles"] = roles
			}
			if len(options) > 0 {
				objectBody["options"] = options
			}

			operations = append(operations, map[string]any{
				"operationId":             fmt.Sprintf("%s-%d", objectType, i+1),
				"includeObjectInResponse": true,
				"addObject": map[string]any{
					"object": map[string]any{
						objectType: objectBody,
					},
					"placement": map[string]any{
						"page": map[string]any{
							"target":   pageName,
							"position": "end",
						},
					},
				},
			})
		}
	}

	operations = append(operations, req.Spec.Operations...)
	return operations
}

func dashboardObjectDataRoles(object DashboardObjectSpec) map[string]any {
	roles := map[string]any{}
	if object.Category != "" {
		roles["category"] = object.Category
	}
	measures := object.Measures
	if len(measures) == 0 && object.Measure != "" {
		measures = []string{object.Measure}
	}
	if len(measures) > 0 {
		roles["measures"] = measures
	}
	return roles
}

func reportPath(reportID string) string {
	return fmt.Sprintf("/reports/reports/%s", url.PathEscape(reportID))
}
