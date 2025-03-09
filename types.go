package harlog

import (
	"net/http"
)

// HAR represents the root of a HAR log
type HAR struct {
	Log HARLog `json:"log"`
}

// HARLog represents the main content of a HAR file
type HARLog struct {
	Version string     `json:"version"`
	Creator HARCreator `json:"creator"`
	Entries []HAREntry `json:"entries"`
}

// HARCreator represents the creator of the HAR file
type HARCreator struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// HAREntry represents a single HTTP request/response pair
type HAREntry struct {
	StartedDateTime string      `json:"startedDateTime"`
	Time            float64     `json:"time"`
	Request         HARRequest  `json:"request"`
	Response        HARResponse `json:"response"`
	Cache           HARCache    `json:"cache"`
	Timings         HARTimings  `json:"timings"`
}

// HARRequest represents an HTTP request
type HARRequest struct {
	Method      string       `json:"method"`
	URL         string       `json:"url"`
	HTTPVersion string       `json:"httpVersion"`
	Headers     []HARHeader  `json:"headers"`
	QueryString []HARQuery   `json:"queryString"`
	PostData    *HARPostData `json:"postData,omitempty"`
	HeadersSize int          `json:"headersSize"`
	BodySize    int          `json:"bodySize"`
}

// HARResponse represents an HTTP response
type HARResponse struct {
	Status      int         `json:"status"`
	StatusText  string      `json:"statusText"`
	HTTPVersion string      `json:"httpVersion"`
	Headers     []HARHeader `json:"headers"`
	Content     HARContent  `json:"content"`
	HeadersSize int         `json:"headersSize"`
	BodySize    int         `json:"bodySize"`
}

// HARHeader represents an HTTP header
type HARHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// HARQuery represents a query string parameter
type HARQuery struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// HARPostData represents POST data
type HARPostData struct {
	MimeType string `json:"mimeType"`
	Text     string `json:"text"`
}

// HARContent represents response content
type HARContent struct {
	Size     int    `json:"size"`
	MimeType string `json:"mimeType"`
	Text     string `json:"text"`
}

// HARCache represents cache information
type HARCache struct {
	// Usually empty, but required by the HAR spec
}

// HARTimings represents various timing measurements
type HARTimings struct {
	Send    float64 `json:"send"`
	Wait    float64 `json:"wait"`
	Receive float64 `json:"receive"`
}

// HTTPMessage represents a pair of HTTP request and response
type HTTPMessage struct {
	Request  *http.Request
	Response *http.Response
}

// HTTPMessages represents a collection of HTTP messages
type HTTPMessages []HTTPMessage
