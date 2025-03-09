package harlog

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestParseHARFile(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "harlog-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test HAR data
	har := HAR{
		Log: HARLog{
			Version: "1.2",
			Creator: HARCreator{
				Name:    "harlog-test",
				Version: "1.0",
			},
			Entries: []HAREntry{
				{
					Request: HARRequest{
						Method:      "POST",
						URL:         "https://api.example.com/users?id=123",
						HTTPVersion: "HTTP/1.1",
						Headers: []HARHeader{
							{Name: "Content-Type", Value: "application/json"},
							{Name: "Authorization", Value: "Bearer token123"},
						},
						QueryString: []HARQuery{
							{Name: "id", Value: "123"},
						},
						PostData: &HARPostData{
							MimeType: "application/json",
							Text:     `{"name": "test user"}`,
						},
					},
					Response: HARResponse{
						Status:      200,
						StatusText:  "OK",
						HTTPVersion: "HTTP/1.1",
						Headers: []HARHeader{
							{Name: "Content-Type", Value: "application/json"},
						},
						Content: HARContent{
							Size:     19,
							MimeType: "application/json",
							Text:     `{"status": "success"}`,
						},
					},
				},
			},
		},
	}

	// Write test HAR file
	harFile := filepath.Join(tmpDir, "test.har")
	harData, err := json.Marshal(har)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(harFile, harData, 0600); err != nil {
		t.Fatal(err)
	}

	// Test ParseHARFile
	messages, err := ParseHARFile(harFile)
	if err != nil {
		t.Fatal(err)
	}

	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}

	msg := messages[0]

	// Verify request
	if msg.Request.Method != "POST" {
		t.Errorf("expected method POST, got %s", msg.Request.Method)
	}
	if msg.Request.URL.String() != "https://api.example.com/users?id=123" {
		t.Errorf("expected URL https://api.example.com/users?id=123, got %s", msg.Request.URL.String())
	}
	if msg.Request.Header.Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", msg.Request.Header.Get("Content-Type"))
	}
	if msg.Request.Header.Get("Authorization") != "Bearer token123" {
		t.Errorf("expected Authorization Bearer token123, got %s", msg.Request.Header.Get("Authorization"))
	}

	// Verify request body
	if msg.Request.Body != nil {
		body, err := io.ReadAll(msg.Request.Body)
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != `{"name": "test user"}` {
			t.Errorf("expected request body {\"name\": \"test user\"}, got %s", string(body))
		}
	}

	// Verify response
	if msg.Response.StatusCode != 200 {
		t.Errorf("expected status code 200, got %d", msg.Response.StatusCode)
	}
	if msg.Response.Header.Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", msg.Response.Header.Get("Content-Type"))
	}

	// Verify response body
	respBody, err := io.ReadAll(msg.Response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(respBody) != `{"status": "success"}` {
		t.Errorf("expected response body {\"status\": \"success\"}, got %s", string(respBody))
	}
}

func TestParseHARData(t *testing.T) {
	// Test data
	harData := []byte(`{
		"log": {
			"version": "1.2",
			"creator": {
				"name": "harlog-test",
				"version": "1.0"
			},
			"entries": [
				{
					"request": {
						"method": "GET",
						"url": "https://api.example.com/status",
						"httpVersion": "HTTP/1.1",
						"headers": [
							{
								"name": "Accept",
								"value": "application/json"
							}
						],
						"queryString": []
					},
					"response": {
						"status": 200,
						"statusText": "OK",
						"httpVersion": "HTTP/1.1",
						"headers": [
							{
								"name": "Content-Type",
								"value": "application/json"
							}
						],
						"content": {
							"size": 16,
							"mimeType": "application/json",
							"text": "{\"status\":\"ok\"}"
						}
					}
				}
			]
		}
	}`)

	// Test ParseHARData
	messages, err := ParseHARData(harData)
	if err != nil {
		t.Fatal(err)
	}

	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}

	msg := messages[0]

	// Verify request
	if msg.Request.Method != "GET" {
		t.Errorf("expected method GET, got %s", msg.Request.Method)
	}
	if msg.Request.URL.String() != "https://api.example.com/status" {
		t.Errorf("expected URL https://api.example.com/status, got %s", msg.Request.URL.String())
	}
	if msg.Request.Header.Get("Accept") != "application/json" {
		t.Errorf("expected Accept application/json, got %s", msg.Request.Header.Get("Accept"))
	}

	// Verify response
	if msg.Response.StatusCode != 200 {
		t.Errorf("expected status code 200, got %d", msg.Response.StatusCode)
	}
	if msg.Response.Header.Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", msg.Response.Header.Get("Content-Type"))
	}

	// Verify response body
	respBody, err := io.ReadAll(msg.Response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(respBody) != "{\"status\":\"ok\"}" {
		t.Errorf("expected response body {\"status\":\"ok\"}, got %s", string(respBody))
	}
}

func TestParseHARFileWithTestData(t *testing.T) {
	messages, err := ParseHARFile("testdata/github.com_m-mizutani_harlog.har")
	if err != nil {
		t.Fatal(err)
	}

	if len(messages) == 0 {
		t.Fatal("no messages found in HAR file")
	}

	// Verify the first message
	msg := messages[0]

	// Basic request checks
	if msg.Request == nil {
		t.Fatal("request is nil")
	}

	// Check request fields
	expectedReqFields := map[string]string{
		"Method":     "GET",
		"URL":        "https://github.com/m-mizutani/harlog",
		"Proto":      "HTTP/2",
		"Host":       "github.com",
		"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:136.0) Gecko/20100101 Firefox/136.0",
	}

	if msg.Request.Method != expectedReqFields["Method"] {
		t.Errorf("request method: expected %s, got %s", expectedReqFields["Method"], msg.Request.Method)
	}
	if msg.Request.URL.String() != expectedReqFields["URL"] {
		t.Errorf("request URL: expected %s, got %s", expectedReqFields["URL"], msg.Request.URL.String())
	}
	if msg.Request.Proto != expectedReqFields["Proto"] {
		t.Errorf("request protocol: expected %s, got %s", expectedReqFields["Proto"], msg.Request.Proto)
	}
	if msg.Request.Host != expectedReqFields["Host"] {
		t.Errorf("request host: expected %s, got %s", expectedReqFields["Host"], msg.Request.Host)
	}
	if msg.Request.Header.Get("User-Agent") != expectedReqFields["User-Agent"] {
		t.Errorf("request User-Agent: expected %s, got %s", expectedReqFields["User-Agent"], msg.Request.Header.Get("User-Agent"))
	}

	// Check request headers
	expectedHeaders := map[string]string{
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language":           "ja,en-US;q=0.7,en;q=0.3",
		"Accept-Encoding":           "gzip, deflate, br, zstd",
		"DNT":                       "1",
		"Sec-GPC":                   "1",
		"Connection":                "keep-alive",
		"Upgrade-Insecure-Requests": "1",
	}

	for header, expectedValue := range expectedHeaders {
		if value := msg.Request.Header.Get(header); value != expectedValue {
			t.Errorf("request header %s: expected %s, got %s", header, expectedValue, value)
		}
	}

	// Basic response checks
	if msg.Response == nil {
		t.Fatal("response is nil")
	}

	// Check response fields
	expectedRespFields := map[string]interface{}{
		"StatusCode":  200,
		"Proto":       "HTTP/2",
		"ContentType": "text/html; charset=utf-8",
	}

	if msg.Response.StatusCode != expectedRespFields["StatusCode"].(int) {
		t.Errorf("response status code: expected %d, got %d", expectedRespFields["StatusCode"].(int), msg.Response.StatusCode)
	}
	if msg.Response.Proto != expectedRespFields["Proto"].(string) {
		t.Errorf("response protocol: expected %s, got %s", expectedRespFields["Proto"].(string), msg.Response.Proto)
	}
	if contentType := msg.Response.Header.Get("Content-Type"); contentType != expectedRespFields["ContentType"].(string) {
		t.Errorf("response Content-Type: expected %s, got %s", expectedRespFields["ContentType"].(string), contentType)
	}

	// Check response headers
	expectedRespHeaders := map[string]string{
		"Server":                    "GitHub.com",
		"X-Frame-Options":           "deny",
		"X-Content-Type-Options":    "nosniff",
		"X-XSS-Protection":          "0",
		"Strict-Transport-Security": "max-age=31536000; includeSubdomains; preload",
	}

	for header, expectedValue := range expectedRespHeaders {
		if value := msg.Response.Header.Get(header); value != expectedValue {
			t.Errorf("response header %s: expected %s, got %s", header, expectedValue, value)
		}
	}

	// Log the total number of messages
	t.Logf("Total messages in HAR file: %d", len(messages))

	// Check a few messages to ensure they are properly parsed
	for i, msg := range messages {
		if i >= 3 { // Only check first 3 messages to keep output manageable
			break
		}
		t.Logf("Message %d:", i)
		t.Logf("  Method: %s", msg.Request.Method)
		t.Logf("  URL: %s", msg.Request.URL)
		t.Logf("  Status: %d", msg.Response.StatusCode)
		t.Logf("  Content-Type: %s", msg.Response.Header.Get("Content-Type"))
	}
}
