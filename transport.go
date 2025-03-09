package harlog

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

// RoundTrip implements http.RoundTripper
func (h *HARLogger) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	harEntry := &HAREntry{
		StartedDateTime: start.Format(time.RFC3339),
	}

	// Record request
	harEntry.Request = h.captureRequest(req)

	// Execute the actual request
	resp, err := h.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// Record response
	body, err := h.captureResponseWithBody(resp)
	if err != nil {
		return nil, err
	}

	harEntry.Response = body
	harEntry.Time = float64(time.Since(start).Milliseconds())

	// Save HAR entry
	h.saveHAR(req, harEntry)

	return resp, nil
}

func (h *HARLogger) captureResponseWithBody(resp *http.Response) (HARResponse, error) {
	headers := make([]HARHeader, 0)
	for name, values := range resp.Header {
		for _, value := range values {
			headers = append(headers, HARHeader{
				Name:  name,
				Value: value,
			})
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return HARResponse{}, err
	}
	resp.Body.Close()

	// Create a new reader with the same body
	resp.Body = io.NopCloser(bytes.NewReader(body))

	return HARResponse{
		Status:      resp.StatusCode,
		StatusText:  resp.Status,
		HTTPVersion: resp.Proto,
		Headers:     headers,
		Content: HARContent{
			Size:     len(body),
			MimeType: resp.Header.Get("Content-Type"),
			Text:     string(body),
		},
		HeadersSize: -1, // Not implemented
		BodySize:    len(body),
	}, nil
}
