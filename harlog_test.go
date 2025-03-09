package harlog

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHARLogger_Handler(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "harlog-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	expectedReqBody := `{"request": "test data"}`
	expectedRespBody := `{"message": "Hello, World!"}`

	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read and verify request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error("failed to read request body:", err)
		}
		if r.Method == http.MethodPost && string(body) != expectedReqBody {
			t.Errorf("Request body mismatch\nwant: %s\ngot: %s", expectedReqBody, string(body))
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(expectedRespBody)); err != nil {
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

	// Make test request with body
	req, err := http.NewRequest(http.MethodPost, server.URL+"/test", strings.NewReader(expectedReqBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// Check if HAR file was created
	harFile := filepath.Join(tmpDir, "handler-test.har")
	if _, err := os.Stat(harFile); os.IsNotExist(err) {
		t.Errorf("HAR file was not created: %s", harFile)
	}

	// Read and parse HAR file
	harData, err := os.ReadFile(harFile)
	if err != nil {
		t.Fatal(err)
	}

	var har HAR
	if err := json.Unmarshal(harData, &har); err != nil {
		t.Fatal(err)
	}

	if len(har.Log.Entries) == 0 {
		t.Fatal("No entries in HAR file")
	}

	entry := har.Log.Entries[0]
	// Verify request body in HAR
	if entry.Request.PostData == nil {
		t.Error("PostData is nil in HAR file")
	} else if entry.Request.PostData.Text != expectedReqBody {
		t.Errorf("HAR request body mismatch\nwant: %s\ngot: %s", expectedReqBody, entry.Request.PostData.Text)
	}

	// Verify response body in HAR
	if entry.Response.Content.Text != expectedRespBody {
		t.Errorf("HAR response body mismatch\nwant: %s\ngot: %s", expectedRespBody, entry.Response.Content.Text)
	}

	if entry.Response.Content.MimeType != "application/json" {
		t.Errorf("Response content type mismatch\nwant: application/json\ngot: %s", entry.Response.Content.MimeType)
	}
}

func TestHARLogger_Middleware(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "harlog-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	expectedReqBody := `{"request": "test data"}`
	expectedRespBody := `{"message": "Hello, World!"}`

	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read and verify request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error("failed to read request body:", err)
		}
		if r.Method == http.MethodPost && string(body) != expectedReqBody {
			t.Errorf("Request body mismatch\nwant: %s\ngot: %s", expectedReqBody, string(body))
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(expectedRespBody)); err != nil {
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

	// Make test request with body
	req, err := http.NewRequest(http.MethodPost, server.URL+"/test", strings.NewReader(expectedReqBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// Check if HAR file was created
	harFile := filepath.Join(tmpDir, "middleware-test.har")
	if _, err := os.Stat(harFile); os.IsNotExist(err) {
		t.Errorf("HAR file was not created: %s", harFile)
	}

	// Read and parse HAR file
	harData, err := os.ReadFile(harFile)
	if err != nil {
		t.Fatal(err)
	}

	var har HAR
	if err := json.Unmarshal(harData, &har); err != nil {
		t.Fatal(err)
	}

	if len(har.Log.Entries) == 0 {
		t.Fatal("No entries in HAR file")
	}

	entry := har.Log.Entries[0]
	// Verify request body in HAR
	if entry.Request.PostData == nil {
		t.Error("PostData is nil in HAR file")
	} else if entry.Request.PostData.Text != expectedReqBody {
		t.Errorf("HAR request body mismatch\nwant: %s\ngot: %s", expectedReqBody, entry.Request.PostData.Text)
	}

	// Verify response body in HAR
	if entry.Response.Content.Text != expectedRespBody {
		t.Errorf("HAR response body mismatch\nwant: %s\ngot: %s", expectedRespBody, entry.Response.Content.Text)
	}

	if entry.Response.Content.MimeType != "application/json" {
		t.Errorf("Response content type mismatch\nwant: application/json\ngot: %s", entry.Response.Content.MimeType)
	}
}

func TestHARLogger_RoundTripper(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "harlog-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	expectedReqBody := `{"request": "test data"}`
	expectedRespBody := `{"message": "Hello, World!"}`

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read and verify request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error("failed to read request body:", err)
		}
		if string(body) != expectedReqBody {
			t.Errorf("Request body mismatch\nwant: %s\ngot: %s", expectedReqBody, string(body))
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(expectedRespBody)); err != nil {
			t.Error("failed to write response:", err)
		}
	}))
	defer server.Close()

	// Create logger with custom options
	harFilename := "roundtrip-test.har"
	logger := New(
		WithOutputDir(tmpDir),
		WithFileNameFn(func(req *http.Request) string {
			return filepath.Join(tmpDir, harFilename)
		}),
	)

	// Create HTTP client with logger as transport
	client := &http.Client{
		Transport: logger,
	}

	// Make test request with body
	req, err := http.NewRequest(http.MethodPost, server.URL+"/test", strings.NewReader(expectedReqBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// Check if HAR file was created
	harFile := filepath.Join(tmpDir, harFilename)
	if _, err := os.Stat(harFile); os.IsNotExist(err) {
		t.Errorf("HAR file was not created: %s", harFile)
	}

	// Read and parse HAR file
	harData, err := os.ReadFile(harFile)
	if err != nil {
		t.Fatal(err)
	}

	var har HAR
	if err := json.Unmarshal(harData, &har); err != nil {
		t.Fatal(err)
	}

	if len(har.Log.Entries) == 0 {
		t.Fatal("No entries in HAR file")
	}

	entry := har.Log.Entries[0]
	// Verify request body in HAR
	if entry.Request.PostData == nil {
		t.Error("PostData is nil in HAR file")
	} else if entry.Request.PostData.Text != expectedReqBody {
		t.Errorf("HAR request body mismatch\nwant: %s\ngot: %s", expectedReqBody, entry.Request.PostData.Text)
	}

	// Verify response body in HAR
	if entry.Response.Content.Text != expectedRespBody {
		t.Errorf("HAR response body mismatch\nwant: %s\ngot: %s", expectedRespBody, entry.Response.Content.Text)
	}

	if entry.Response.Content.MimeType != "application/json" {
		t.Errorf("Response content type mismatch\nwant: application/json\ngot: %s", entry.Response.Content.MimeType)
	}
}
