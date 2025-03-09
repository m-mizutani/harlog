package harlog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// ParseHARFile reads a HAR file and converts it to HTTP messages
func ParseHARFile(filename string) (HTTPMessages, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read HAR file: %w", err)
	}
	return ParseHARData(data)
}

// ParseHARData parses HAR data from bytes and converts it to HTTP messages
func ParseHARData(data []byte) (HTTPMessages, error) {
	var har HAR
	if err := json.Unmarshal(data, &har); err != nil {
		return nil, fmt.Errorf("failed to parse HAR data: %w", err)
	}

	messages := make(HTTPMessages, 0, len(har.Log.Entries))
	for _, entry := range har.Log.Entries {
		req, err := convertHARRequestToHTTP(&entry.Request)
		if err != nil {
			return nil, fmt.Errorf("failed to convert HAR request: %w", err)
		}

		resp, err := convertHARResponseToHTTP(&entry.Response)
		if err != nil {
			return nil, fmt.Errorf("failed to convert HAR response: %w", err)
		}

		messages = append(messages, HTTPMessage{
			Request:  req,
			Response: resp,
		})
	}

	return messages, nil
}

func convertHARRequestToHTTP(harReq *HARRequest) (*http.Request, error) {
	// Parse URL
	reqURL, err := url.Parse(harReq.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse request URL: %w", err)
	}

	// Create body if POST data exists
	var body io.Reader
	if harReq.PostData != nil {
		body = strings.NewReader(harReq.PostData.Text)
	}

	// Create request
	req, err := http.NewRequest(harReq.Method, reqURL.String(), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set protocol version
	req.Proto = harReq.HTTPVersion
	if strings.HasPrefix(harReq.HTTPVersion, "HTTP/") {
		version := strings.TrimPrefix(harReq.HTTPVersion, "HTTP/")
		parts := strings.Split(version, ".")
		if len(parts) >= 1 {
			if major, err := fmt.Sscanf(parts[0], "%d", &req.ProtoMajor); err == nil {
				req.ProtoMajor = major
			}
			if len(parts) >= 2 {
				if minor, err := fmt.Sscanf(parts[1], "%d", &req.ProtoMinor); err == nil {
					req.ProtoMinor = minor
				}
			} else {
				req.ProtoMinor = 0
			}
		}
	}

	// Set headers
	for _, header := range harReq.Headers {
		req.Header.Add(header.Name, header.Value)
	}

	// Set query parameters only if they are not already in the URL
	if len(harReq.QueryString) > 0 && len(reqURL.RawQuery) == 0 {
		q := req.URL.Query()
		for _, query := range harReq.QueryString {
			q.Add(query.Name, query.Value)
		}
		req.URL.RawQuery = q.Encode()
	}

	return req, nil
}

func convertHARResponseToHTTP(harResp *HARResponse) (*http.Response, error) {
	// Create response
	resp := &http.Response{
		StatusCode: harResp.Status,
		Status:     harResp.StatusText,
		Proto:      harResp.HTTPVersion,
		Header:     make(http.Header),
	}

	// Set protocol version
	if strings.HasPrefix(harResp.HTTPVersion, "HTTP/") {
		version := strings.TrimPrefix(harResp.HTTPVersion, "HTTP/")
		parts := strings.Split(version, ".")
		if len(parts) >= 1 {
			if major, err := fmt.Sscanf(parts[0], "%d", &resp.ProtoMajor); err == nil {
				resp.ProtoMajor = major
			}
			if len(parts) >= 2 {
				if minor, err := fmt.Sscanf(parts[1], "%d", &resp.ProtoMinor); err == nil {
					resp.ProtoMinor = minor
				}
			} else {
				resp.ProtoMinor = 0
			}
		}
	}

	// Set headers
	for _, header := range harResp.Headers {
		resp.Header.Add(header.Name, header.Value)
	}

	// Set body
	resp.Body = io.NopCloser(bytes.NewReader([]byte(harResp.Content.Text)))
	resp.ContentLength = int64(len(harResp.Content.Text))

	return resp, nil
}
