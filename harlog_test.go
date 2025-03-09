package harlog

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestHARLogger_Handler(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "harlog-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"message": "Hello, World!"}`)); err != nil {
			t.Error("failed to write response:", err)
		}
	})

	// Create logger with custom options
	logger := New(
		WithOutputDir(tmpDir),
		WithFileNameFn(func(req *http.Request) string {
			return filepath.Join(tmpDir, "handler-"+req.URL.Path[1:]+".har")
		}),
		WithHandler(handler),
	)

	// Create test server
	server := httptest.NewServer(logger)
	defer server.Close()

	// Make test request
	resp, err := http.Get(server.URL + "/test")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// Check if HAR file was created
	harFile := filepath.Join(tmpDir, "handler-test.har")
	if _, err := os.Stat(harFile); os.IsNotExist(err) {
		t.Errorf("HAR file was not created: %s", harFile)
	}
}

func TestHARLogger_Middleware(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "harlog-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"message": "Hello, World!"}`)); err != nil {
			t.Error("failed to write response:", err)
		}
	})

	// Create logger with custom options
	logger := New(
		WithOutputDir(tmpDir),
		WithFileNameFn(func(req *http.Request) string {
			return filepath.Join(tmpDir, "middleware-"+req.URL.Path[1:]+".har")
		}),
	)

	// Create test server with middleware
	server := httptest.NewServer(logger.Middleware(handler))
	defer server.Close()

	// Make test request
	resp, err := http.Get(server.URL + "/test")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// Check if HAR file was created
	harFile := filepath.Join(tmpDir, "middleware-test.har")
	if _, err := os.Stat(harFile); os.IsNotExist(err) {
		t.Errorf("HAR file was not created: %s", harFile)
	}
}

func TestHARLogger_RoundTripper(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "harlog-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"message": "Hello, World!"}`)); err != nil {
			t.Error("failed to write response:", err)
		}
	}))
	defer server.Close()

	// Create logger with custom options
	logger := New(
		WithOutputDir(tmpDir),
		WithFileNameFn(func(req *http.Request) string {
			return filepath.Join(tmpDir, "roundtrip-"+time.Now().Format("20060102-150405")+".har")
		}),
	)

	// Create HTTP client with logger as transport
	client := &http.Client{
		Transport: logger,
	}

	// Make test request
	resp, err := client.Get(server.URL + "/test")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// Check if any HAR files were created
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) == 0 {
		t.Error("No HAR files were created")
	}
}
