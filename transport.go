package harlog

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

// RoundTrip implements http.RoundTripper
func (l *Logger) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	harEntry := &HAREntry{
		StartedDateTime: start.Format(time.RFC3339),
	}

	// Record request
	harEntry.Request = l.captureRequest(req)

	// Execute the actual request
	resp, err := l.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// Record response
	body, err := l.captureResponseWithBody(resp)
	if err != nil {
		return nil, err
	}

	harEntry.Response = body
	harEntry.Time = float64(time.Since(start).Milliseconds())

	// Save HAR entry
	if err := l.saveHAR(req, harEntry); err != nil {
		l.logger.Error("failed to save HAR",
			"error", err,
			"path", req.URL.Path,
			"method", req.Method,
			"host", req.Host,
		)
	}

	return resp, nil
}

func (l *Logger) captureResponseWithBody(resp *http.Response) (HARResponse, error) {
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
	if err := resp.Body.Close(); err != nil {
		return HARResponse{}, err
	}

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
