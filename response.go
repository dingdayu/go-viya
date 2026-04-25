package viya

type Link struct {
	Method string `json:"method"`
	Rel    string `json:"rel"`
	Href   string `json:"href"`
	URI    string `json:"uri"`
	Type   string `json:"type,omitempty"`
}

type ListResponse[T any] struct {
	Version int    `json:"version"`
	Accept  string `json:"accept"`
	Count   int    `json:"count"`
	Start   int    `json:"start"`
	Limit   int    `json:"limit"`
	Items   []T    `json:"items"`
	Links   []Link `json:"links"`
}

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

func (e ErrorResponse) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.ErrorInfo != nil {
		return e.ErrorInfo.Message
	}
	return "unknown error"
}
