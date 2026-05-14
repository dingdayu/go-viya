package viya

// Accept header constants for SAS Viya REST API requests.
// These values are used across endpoint helpers to ensure consistent
// media type negotiation with the SAS Viya REST services.
const (
	// Common Accept header patterns.
	AcceptJSON        = "application/json"
	AcceptOctetStream = "application/octet-stream"
	AcceptTextPlain   = "text/plain"
	AcceptSASError    = "application/vnd.sas.error+json"

	// Combined Accept header values.
	AcceptJSONError  = AcceptJSON + ", " + AcceptSASError
	AcceptOctetError = AcceptOctetStream + ", " + AcceptSASError
	AcceptTextError  = AcceptTextPlain + ", " + AcceptSASError
	AcceptErrorOnly  = AcceptSASError
	AcceptCollection = AcceptJSON + ", application/vnd.sas.collection+json;version=2, " + AcceptSASError
	AcceptJobExecLog = AcceptTextPlain + ", " + AcceptOctetStream + ", " + AcceptSASError

	// Compute API Accept headers.
	AcceptComputeContext = AcceptJSON + ", application/vnd.sas.compute.context+json, " + AcceptSASError
	AcceptComputeSession = AcceptJSON + ", application/vnd.sas.compute.session+json, " + AcceptSASError
	AcceptComputeJob     = AcceptJSON + ", application/vnd.sas.compute.job+json, " + AcceptSASError

	// Batch API Accept headers.
	AcceptBatchContext  = AcceptJSON + ", application/vnd.sas.batch.context+json, " + AcceptSASError
	AcceptBatchFileSet  = AcceptJSON + ", application/vnd.sas.batch.file.set+json, application/vnd.sas.batch.file.set+json;version=1, " + AcceptSASError
	AcceptBatchFile     = AcceptJSON + ", application/vnd.sas.batch.file.set.file+json"
	AcceptBatchServer   = AcceptJSON + ", application/vnd.sas.batch.server+json, " + AcceptSASError
	AcceptBatchJob      = AcceptJSON + ", application/vnd.sas.batch.job+json, application/vnd.sas.batch.job+json;version=1, " + AcceptSASError
	AcceptBatchJobState = AcceptJSON + ", application/vnd.sas.batch.job.state+json, application/vnd.sas.batch.job.state+json;version=1, application/vnd.sas.batch.job+json, application/vnd.sas.batch.job+json;version=1, " + AcceptSASError
)

// Link describes a hypermedia link returned by SAS Viya REST APIs.
//
// Many SAS Viya resources include action links such as "self", "update", or
// "delete"; callers can use these when an API response provides the canonical URL.
type Link struct {
	Method string `json:"method"`
	Rel    string `json:"rel"`
	Href   string `json:"href"`
	URI    string `json:"uri"`
	Type   string `json:"type,omitempty"`
}

// ListResponse represents a standard SAS Viya collection response.
//
// Collection responses usually include paging fields, items, and links. For
// details on SAS Viya REST API conventions, see:
// https://developer.sas.com/docs/rest-apis/getting-started/authentication
type ListResponse[T any] struct {
	Version int    `json:"version"`
	Accept  string `json:"accept"`
	Count   int    `json:"count"`
	Start   int    `json:"start"`
	Limit   int    `json:"limit"`
	Items   []T    `json:"items"`
	Links   []Link `json:"links"`
}

// ErrorResponse represents a structured SAS Viya error response.
type ErrorResponse struct {
	Version        int      `json:"version"`
	Accept         string   `json:"accept,omitempty"`
	HTTPStatusCode int      `json:"httpStatusCode,omitempty"`
	Message        string   `json:"message,omitempty"`
	Details        []string `json:"details,omitempty"`
	ErrorInfo      *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Error returns the most useful message from a SAS Viya error payload.
func (e ErrorResponse) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.ErrorInfo != nil {
		return e.ErrorInfo.Message
	}
	return "unknown error"
}
