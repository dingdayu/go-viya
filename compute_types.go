package viya

import "time"

// SAS Viya Compute API reference:
// https://developer.sas.com/rest-apis/compute

// ComputeEnvironment provides SAS options and initialization code for Compute sessions and jobs.
type ComputeEnvironment struct {
	Options  []string `json:"options,omitempty"`
	InitCode []string `json:"initCode,omitempty"`
}

// ComputeExternalResource describes a resource requested for a Compute session or job.
type ComputeExternalResource struct {
	Name    string         `json:"name,omitempty"`
	URI     string         `json:"uri,omitempty"`
	Type    string         `json:"type,omitempty"`
	Scope   string         `json:"scope,omitempty"`
	Options map[string]any `json:"options,omitempty"`
}

// ComputeOutputStatistics describes line-oriented Compute output metadata.
type ComputeOutputStatistics struct {
	LineCount         int       `json:"lineCount"`
	ModifiedTimeStamp time.Time `json:"modifiedTimeStamp"`
}

// ComputeLogLine describes one line of Compute log or listing output.
type ComputeLogLine struct {
	Line string `json:"line"`
	Type string `json:"type"`
}

// ComputeLogLinesResponse is a collection of Compute log or listing lines.
type ComputeLogLinesResponse = ListResponse[ComputeLogLine]
