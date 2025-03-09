package harlog

import (
	"net/http"
	"path/filepath"
	"sync"
	"time"
)

// HARLogger implements http.Handler, http.RoundTripper and provides middleware functionality
type HARLogger struct {
	handler    http.Handler
	transport  http.RoundTripper
	outputDir  string
	fileNameFn func(req *http.Request) string
	mu         sync.Mutex
}

// Option represents a configuration option for HARLogger
type Option func(*HARLogger)

// WithOutputDir sets the output directory for HAR files
func WithOutputDir(dir string) Option {
	return func(h *HARLogger) {
		h.outputDir = dir
	}
}

// WithFileNameFn sets the custom filename generator function
func WithFileNameFn(fn func(req *http.Request) string) Option {
	return func(h *HARLogger) {
		h.fileNameFn = fn
	}
}

// WithHandler sets the initial http.Handler
func WithHandler(handler http.Handler) Option {
	return func(h *HARLogger) {
		h.handler = handler
	}
}

// WithTransport sets the initial http.RoundTripper
func WithTransport(transport http.RoundTripper) Option {
	return func(h *HARLogger) {
		h.transport = transport
	}
}

// New creates a new HARLogger instance with the given options
func New(opts ...Option) *HARLogger {
	h := &HARLogger{
		handler:   http.DefaultServeMux,
		transport: http.DefaultTransport,
		outputDir: "har_logs",
		fileNameFn: func(req *http.Request) string {
			timestamp := time.Now().Format("20060102-150405.000")
			return filepath.Join("har_logs", timestamp+".har")
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(h)
	}

	return h
}

// WrapHandler wraps an existing http.Handler
func (h *HARLogger) WrapHandler(handler http.Handler) *HARLogger {
	h.handler = handler
	return h
}

// WrapTransport wraps an existing http.RoundTripper
func (h *HARLogger) WrapTransport(transport http.RoundTripper) *HARLogger {
	h.transport = transport
	return h
}
