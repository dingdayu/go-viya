package viya

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"go.opentelemetry.io/otel/codes"
)

// MLProject describes a SAS Viya AutoML pipeline automation project.
type MLProject struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"description,omitempty"`
	State             string    `json:"state,omitempty"`
	Type              string    `json:"type,omitempty"`
	CreatedBy         string    `json:"createdBy,omitempty"`
	CreationTimeStamp time.Time `json:"creationTimeStamp,omitempty"`
	ModifiedBy        string    `json:"modifiedBy,omitempty"`
	ModifiedTimeStamp time.Time `json:"modifiedTimeStamp,omitempty"`
	Version           int       `json:"version,omitempty"`
	Links             []Link    `json:"links,omitempty"`
}

// MLProjectsResponse is a collection of SAS Viya AutoML projects.
type MLProjectsResponse = ListResponse[MLProject]

// CreateMLProjectRequest is the request body for creating an AutoML project.
type CreateMLProjectRequest struct {
	Name                       string                  `json:"name"`
	Description                string                  `json:"description,omitempty"`
	Type                       string                  `json:"type"`
	DataTableURI               string                  `json:"dataTableUri"`
	PipelineBuildMethod        string                  `json:"pipelineBuildMethod"`
	Settings                   MLProjectSettings       `json:"settings"`
	AnalyticsProjectAttributes MLProjectAnalyticsAttrs `json:"analyticsProjectAttributes"`
}

// MLProjectSettings configures automatic pipeline project creation.
type MLProjectSettings struct {
	ApplyGlobalMetadata bool `json:"applyGlobalMetadata"`
	AutoRun             bool `json:"autoRun"`
	NumberOfModels      int  `json:"numberOfModels"`
}

// MLProjectAnalyticsAttrs configures target metadata for an AutoML project.
type MLProjectAnalyticsAttrs struct {
	TargetVariable          string `json:"targetVariable"`
	TargetLevel             string `json:"targetLevel"`
	TargetEventLevel        string `json:"targetEventLevel,omitempty"`
	PartitionEnabled        bool   `json:"partitionEnabled"`
	ClassSelectionStatistic string `json:"classSelectionStatistic"`
}

// RegisteredModel describes a model in the SAS Viya model repository.
type RegisteredModel struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"description,omitempty"`
	Author            string    `json:"author,omitempty"`
	ModelVersionName  string    `json:"modelVersionName,omitempty"`
	CreatedBy         string    `json:"createdBy,omitempty"`
	CreationTimeStamp time.Time `json:"creationTimeStamp,omitempty"`
	ModifiedBy        string    `json:"modifiedBy,omitempty"`
	ModifiedTimeStamp time.Time `json:"modifiedTimeStamp,omitempty"`
	Version           int       `json:"version,omitempty"`
	Links             []Link    `json:"links,omitempty"`
}

// RegisteredModelsResponse is a collection of registered models.
type RegisteredModelsResponse = ListResponse[RegisteredModel]

// ModelDecision describes a published MAS module for a model or decision.
type ModelDecision struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"description,omitempty"`
	Type              string    `json:"type,omitempty"`
	CreatedBy         string    `json:"createdBy,omitempty"`
	CreationTimeStamp time.Time `json:"creationTimeStamp,omitempty"`
	ModifiedBy        string    `json:"modifiedBy,omitempty"`
	ModifiedTimeStamp time.Time `json:"modifiedTimeStamp,omitempty"`
	Version           int       `json:"version,omitempty"`
	Links             []Link    `json:"links,omitempty"`
}

// ModelsAndDecisionsResponse is a collection of published MAS modules.
type ModelsAndDecisionsResponse = ListResponse[ModelDecision]

// ScoreDataRequest is the request body for scoring through a MAS module step.
type ScoreDataRequest struct {
	Inputs []ScoreInput `json:"inputs"`
}

// ScoreInput describes one named scoring input value.
type ScoreInput struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}

// NewCreateMLProjectRequest builds an AutoML project request matching SAS MCP defaults.
func NewCreateMLProjectRequest(projectName string, dataTableURI string, targetVariable string) CreateMLProjectRequest {
	return CreateMLProjectRequest{
		Name:                projectName,
		Type:                "predictive",
		DataTableURI:        dataTableURI,
		PipelineBuildMethod: "automatic",
		Settings: MLProjectSettings{
			ApplyGlobalMetadata: true,
			AutoRun:             true,
			NumberOfModels:      5,
		},
		AnalyticsProjectAttributes: MLProjectAnalyticsAttrs{
			TargetVariable:          targetVariable,
			TargetLevel:             "binary",
			TargetEventLevel:        "1",
			PartitionEnabled:        true,
			ClassSelectionStatistic: "ks",
		},
	}
}

// NewScoreDataRequest converts a map of input values into the MAS scoring request shape.
func NewScoreDataRequest(inputData map[string]any) ScoreDataRequest {
	inputs := make([]ScoreInput, 0, len(inputData))
	for name, value := range inputData {
		inputs = append(inputs, ScoreInput{Name: name, Value: value})
	}
	return ScoreDataRequest{Inputs: inputs}
}

// GetMLProjects returns AutoML projects visible to the caller.
func (c *Client) GetMLProjects(ctx context.Context, opts ListOptions) (resp MLProjectsResponse, err error) {
	ctx, span := tracer.Start(ctx, "GetMLProjects")
	defer span.End()

	r, err := c.collectionRequest(ctx, opts).
		SetResult(&resp).
		Get("/mlPipelineAutomation/projects")
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get ML projects, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// CreateMLProject creates a new AutoML project.
func (c *Client) CreateMLProject(ctx context.Context, req CreateMLProjectRequest) (resp MLProject, err error) {
	if req.Name == "" {
		return resp, &ErrInvalidParameter{Parameter: "Name", Reason: "must not be empty"}
	}
	if req.DataTableURI == "" {
		return resp, &ErrInvalidParameter{Parameter: "DataTableURI", Reason: "must not be empty"}
	}
	if req.AnalyticsProjectAttributes.TargetVariable == "" {
		return resp, &ErrInvalidParameter{Parameter: "AnalyticsProjectAttributes.TargetVariable", Reason: "must not be empty"}
	}

	ctx, span := tracer.Start(ctx, "CreateMLProject")
	defer span.End()

	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", AcceptJSONError).
		SetContentType("application/json").
		SetBody(req).
		SetResult(&resp).
		Post("/mlPipelineAutomation/projects")
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to create ML project, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// RunMLProject runs pipeline automation for an AutoML project.
func (c *Client) RunMLProject(ctx context.Context, projectID string) (resp MLProject, err error) {
	if projectID == "" {
		return resp, &ErrInvalidParameter{Parameter: "projectID", Reason: "must not be empty"}
	}

	ctx, span := tracer.Start(ctx, "RunMLProject")
	defer span.End()

	projectPath := fmt.Sprintf("/mlPipelineAutomation/projects/%s", url.PathEscape(projectID))
	getResp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", AcceptMLProject).
		SetResult(&resp).
		Get(projectPath)
	if err != nil {
		return resp, err
	}
	if !getResp.IsSuccess() {
		span.SetStatus(codes.Error, getResp.String())
		return resp, fmt.Errorf("failed to get ML project, status code: %d", getResp.StatusCode())
	}

	putReq := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", AcceptMLProject).
		SetHeader("Accept-Language", "en").
		SetContentType(AcceptMLProject).
		SetQueryParam("action", "retrainProject").
		SetBody(resp).
		SetResult(&resp)
	if etag := getResp.Header().Get("etag"); etag != "" {
		putReq.SetHeader("If-Match", etag)
	}

	putResp, err := putReq.Put(projectPath)
	if err != nil {
		return resp, err
	}
	if !putResp.IsSuccess() {
		span.SetStatus(codes.Error, putResp.String())
		return resp, fmt.Errorf("failed to run ML project pipeline, status code: %d", putResp.StatusCode())
	}

	return resp, nil
}

// GetRegisteredModels returns models in the repository visible to the caller.
func (c *Client) GetRegisteredModels(ctx context.Context, opts ListOptions) (resp RegisteredModelsResponse, err error) {
	ctx, span := tracer.Start(ctx, "GetRegisteredModels")
	defer span.End()

	r, err := c.collectionRequest(ctx, opts).
		SetResult(&resp).
		Get("/modelRepository/models")
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get registered models, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// GetModelsAndDecisions returns published MAS modules visible to the caller.
func (c *Client) GetModelsAndDecisions(ctx context.Context, opts ListOptions) (resp ModelsAndDecisionsResponse, err error) {
	ctx, span := tracer.Start(ctx, "GetModelsAndDecisions")
	defer span.End()

	r, err := c.collectionRequest(ctx, opts).
		SetResult(&resp).
		Get("/microanalyticScore/modules")
	if err != nil {
		return resp, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return resp, fmt.Errorf("failed to get models and decisions, status code: %d", r.StatusCode())
	}

	return resp, nil
}

// ScoreData scores input data through a MAS module step.
func (c *Client) ScoreData(ctx context.Context, moduleID string, stepID string, req ScoreDataRequest) (resp map[string]any, err error) {
	if moduleID == "" {
		return nil, &ErrInvalidParameter{Parameter: "moduleID", Reason: "must not be empty"}
	}
	if stepID == "" {
		return nil, &ErrInvalidParameter{Parameter: "stepID", Reason: "must not be empty"}
	}

	ctx, span := tracer.Start(ctx, "ScoreData")
	defer span.End()

	r, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", AcceptJSONError).
		SetContentType("application/json").
		SetBody(req).
		SetResult(&resp).
		Post(fmt.Sprintf("/microanalyticScore/modules/%s/steps/%s", url.PathEscape(moduleID), url.PathEscape(stepID)))
	if err != nil {
		return nil, err
	}
	if !r.IsSuccess() {
		span.SetStatus(codes.Error, r.String())
		return nil, fmt.Errorf("failed to score data, status code: %d", r.StatusCode())
	}

	return resp, nil
}
