package viya

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"go.opentelemetry.io/otel/codes"
)

// Report describes a SAS Visual Analytics report returned by the reports service.
type Report struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"description,omitempty"`
	ModifiedBy        string    `json:"modifiedBy,omitempty"`
	ModifiedTimeStamp time.Time `json:"modifiedTimeStamp,omitempty"`
	CreatedBy         string    `json:"createdBy,omitempty"`
	CreationTimeStamp time.Time `json:"creationTimeStamp,omitempty"`
	Version           int       `json:"version,omitempty"`
	Links             []Link    `json:"links,omitempty"`
}

// ReportsResponse is a collection of SAS Visual Analytics reports.
type ReportsResponse = ListResponse[Report]

// ReportDefinitionResponse contains report metadata and its definition payload.
type ReportDefinitionResponse map[string]any

// ReportImageOptions configures Visual Analytics report image rendering.
type ReportImageOptions struct {
	SectionIndex  *int
	Size          string
	LayoutType    string
	SelectionType string
	RenderLimit   int
}

// ReportImageJob describes a Visual Analytics report image rendering job.
type ReportImageJob struct {
	ID                string        `json:"id"`
	State             string        `json:"state,omitempty"`
	CreatedBy         string        `json:"createdBy,omitempty"`
	CreationTimeStamp time.Time     `json:"creationTimeStamp,omitempty"`
	ModifiedBy        string        `json:"modifiedBy,omitempty"`
	ModifiedTimeStamp time.Time     `json:"modifiedTimeStamp,omitempty"`
	Version           int           `json:"version,omitempty"`
	Images            []ReportImage `json:"images,omitempty"`
	Links             []Link        `json:"links,omitempty"`
}

// ReportImage describes one rendered Visual Analytics report image result.
type ReportImage struct {
	SectionIndex      int       `json:"sectionIndex"`
	SectionName       string    `json:"sectionName,omitempty"`
	SectionLabel      string    `json:"sectionLabel,omitempty"`
	ElementName       string    `json:"elementName,omitempty"`
	ModifiedTimeStamp time.Time `json:"modifiedTimeStamp,omitempty"`
	VisualType        string    `json:"visualType,omitempty"`
	Size              string    `json:"size,omitempty"`
	State             string    `json:"state,omitempty"`
	Links             []Link    `json:"links,omitempty"`
}

type reportImageJobRequest struct {
	ReportURI     string `json:"reportUri"`
	LayoutType    string `json:"layoutType"`
	SelectionType string `json:"selectionType"`
	SectionIndex  int    `json:"sectionIndex"`
	Size          string `json:"size"`
	RenderLimit   int    `json:"renderLimit"`
}

// GetReports returns Visual Analytics reports visible to the caller.
func (c *Client) GetReports(ctx context.Context, opts ListOptions) (resp ReportsResponse, err error) {
	ctx, span := tracer.Start(ctx, "GetReports")
	defer span.End()

	r, err := c.collectionRequest(ctx, opts).
		SetResult(&resp).
		Get("/reports/reports")
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get reports, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// GetReport returns metadata and definition for a Visual Analytics report.
func (c *Client) GetReport(ctx context.Context, reportID string) (resp ReportDefinitionResponse, err error) {
	if reportID == "" {
		return resp, &ErrInvalidParameter{Parameter: "reportID", Reason: "must not be empty"}
	}

	ctx, span := tracer.Start(ctx, "GetReport")
	defer span.End()

	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", AcceptJSONError).
		SetResult(&resp).
		Get(fmt.Sprintf("/reports/reports/%s", url.PathEscape(reportID)))
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get report, status code: %d", r.StatusCode())
	}

	var definition map[string]any
	contentResp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", AcceptJSONError).
		SetResult(&definition).
		Get(fmt.Sprintf("/reports/reports/%s/content", url.PathEscape(reportID)))
	if err != nil {
		return resp, err
	}
	if !contentResp.IsSuccess() {
		span.SetStatus(codes.Error, contentResp.String())
		return resp, fmt.Errorf("failed to get report content, status code: %d", contentResp.StatusCode())
	}
	if resp == nil {
		return resp, fmt.Errorf("failed to decode report metadata: empty response")
	}
	resp["definition"] = definition

	return resp, nil
}

// GetReportImage creates a Visual Analytics report image rendering job.
func (c *Client) GetReportImage(ctx context.Context, reportID string, opts ReportImageOptions) (resp ReportImageJob, err error) {
	if reportID == "" {
		return resp, &ErrInvalidParameter{Parameter: "reportID", Reason: "must not be empty"}
	}

	ctx, span := tracer.Start(ctx, "GetReportImage")
	defer span.End()

	req := reportImageJobRequest{
		ReportURI:     fmt.Sprintf("/reports/reports/%s", url.PathEscape(reportID)),
		LayoutType:    defaultReportImageLayoutType(opts),
		SelectionType: defaultString(opts.SelectionType, "perSection"),
		SectionIndex:  defaultReportImageSectionIndex(opts),
		Size:          defaultString(opts.Size, "800x600"),
		RenderLimit:   defaultReportRenderLimit(opts.RenderLimit),
	}

	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", AcceptReportImageJob).
		SetContentType(ContentTypeReportImageJob).
		SetBody(req).
		SetResult(&resp).
		Post("/reportImages/jobs")
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get report image, status code: %d", r.StatusCode())
	}

	return resp, nil
}

func defaultString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func defaultReportImageLayoutType(opts ReportImageOptions) string {
	if opts.LayoutType != "" {
		return opts.LayoutType
	}
	if opts.SectionIndex != nil {
		return "entireSection"
	}
	return "thumbnail"
}

func defaultReportImageSectionIndex(opts ReportImageOptions) int {
	if opts.SectionIndex != nil {
		return *opts.SectionIndex
	}
	return 0
}

func defaultReportRenderLimit(value int) int {
	if value == 0 {
		return -1
	}
	return value
}
