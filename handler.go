package harlog

import (
	"net/http"
	"time"
)

// responseWriter is a wrapper for http.ResponseWriter that captures the response
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body = append(rw.body, b...)
	return rw.ResponseWriter.Write(b)
}

// ServeHTTP implements http.Handler interface for backward compatibility
func (l *Logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if l.handler == nil {
		l.handler = http.DefaultServeMux
	}
	l.Middleware(l.handler).ServeHTTP(w, r)
}

// Middleware creates a new middleware handler
func (l *Logger) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		harEntry := &HAREntry{
			StartedDateTime: start.Format(time.RFC3339),
		}

		// Create a response wrapper to capture the response
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			body:           make([]byte, 0),
		}

		// Record request
		harEntry.Request = l.captureRequest(r)

		// Call the next handler
		next.ServeHTTP(rw, r)

		// Record response
		harEntry.Response = l.captureResponse(rw)
		harEntry.Time = float64(time.Since(start).Milliseconds())

		// Save HAR entry
		if err := l.saveHAR(r, harEntry); err != nil {
			l.logger.Error("failed to save HAR",
				"error", err,
				"path", r.URL.Path,
				"method", r.Method,
				"host", r.Host,
			)
		}
	})
}

func (l *Logger) captureRequest(r *http.Request) HARRequest {
	headers := make([]HARHeader, 0)
	for name, values := range r.Header {
		for _, value := range values {
			headers = append(headers, HARHeader{
				Name:  name,
				Value: value,
			})
		}
	}

	queryString := make([]HARQuery, 0)
	for name, values := range r.URL.Query() {
		for _, value := range values {
			queryString = append(queryString, HARQuery{
				Name:  name,
				Value: value,
			})
		}
	}

	return HARRequest{
		Method:      r.Method,
		URL:         r.URL.String(),
		HTTPVersion: r.Proto,
		Headers:     headers,
		QueryString: queryString,
		HeadersSize: -1, // Not implemented
		BodySize:    -1, // Not implemented
	}
}

func (l *Logger) captureResponse(rw *responseWriter) HARResponse {
	headers := make([]HARHeader, 0)
	for name, values := range rw.Header() {
		for _, value := range values {
			headers = append(headers, HARHeader{
				Name:  name,
				Value: value,
			})
		}
	}

	return HARResponse{
		Status:      rw.statusCode,
		StatusText:  http.StatusText(rw.statusCode),
		HTTPVersion: "HTTP/1.1",
		Headers:     headers,
		Content: HARContent{
			Size:     len(rw.body),
			MimeType: rw.Header().Get("Content-Type"),
			Text:     string(rw.body),
		},
		HeadersSize: -1, // Not implemented
		BodySize:    len(rw.body),
	}
}
