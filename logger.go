package harlog

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Logger implements http.Handler, http.RoundTripper and provides middleware functionality
type Logger struct {
	handler    http.Handler
	transport  http.RoundTripper
	outputDir  string
	fileNameFn func(req *http.Request) string
	logger     *slog.Logger
	mu         sync.Mutex
}

// Option represents a configuration option for Logger
type Option func(*Logger)

// WithOutputDir sets the output directory for HAR files
func WithOutputDir(dir string) Option {
	return func(l *Logger) {
		l.outputDir = dir
	}
}

// WithFileNameFn sets the custom filename generator function
func WithFileNameFn(fn func(req *http.Request) string) Option {
	return func(l *Logger) {
		l.fileNameFn = fn
	}
}

// WithHandler sets the initial http.Handler
func WithHandler(handler http.Handler) Option {
	return func(l *Logger) {
		l.handler = handler
	}
}

// WithTransport sets the initial http.RoundTripper
func WithTransport(transport http.RoundTripper) Option {
	return func(l *Logger) {
		l.transport = transport
	}
}

// WithLogger sets the slog.Logger for error logging
func WithLogger(logger *slog.Logger) Option {
	return func(l *Logger) {
		l.logger = logger
	}
}

// defaultFileNameFn generates a unique filename for the HAR file
func (l *Logger) defaultFileNameFn(req *http.Request) string {
	now := time.Now().UTC()
	host := req.Host
	if host == "" {
		host = req.URL.Host
	}
	if host == "" {
		host = "unknown"
	}

	// Create a unique identifier using timestamp, host, method, and path
	// Format: YYYYMMDD-HHMMSS.mmm-{uuid}-{host}-{method}-{path}.har
	return filepath.Join(l.outputDir,
		fmt.Sprintf("%s-%s-%s-%s-%s.har",
			now.Format("20060102-150405.000"),
			uuid.New().String()[:8],
			sanitizeFilename(host),
			req.Method,
			sanitizeFilename(req.URL.Path),
		),
	)
}

// sanitizeFilename removes or replaces characters that might be problematic in filenames
func sanitizeFilename(s string) string {
	// Replace problematic characters with underscore
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		" ", "_",
	)
	return replacer.Replace(s)
}

// New creates a new Logger instance with the given options
func New(opts ...Option) *Logger {
	l := &Logger{
		handler:   http.DefaultServeMux,
		transport: http.DefaultTransport,
		outputDir: ".",
		logger:    slog.New(slog.NewTextHandler(os.Stderr, nil)),
	}
	l.fileNameFn = l.defaultFileNameFn

	// Apply options
	for _, opt := range opts {
		opt(l)
	}

	return l
}

// WrapHandler wraps an existing http.Handler
func (l *Logger) WrapHandler(handler http.Handler) *Logger {
	l.handler = handler
	return l
}

// WrapTransport wraps an existing http.RoundTripper
func (l *Logger) WrapTransport(transport http.RoundTripper) *Logger {
	l.transport = transport
	return l
}
