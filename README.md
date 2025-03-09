# harlog

[![Go Reference](https://pkg.go.dev/badge/github.com/m-mizutani/harlog.svg)](https://pkg.go.dev/github.com/m-mizutani/harlog)
[![Test](https://github.com/m-mizutani/harlog/actions/workflows/test.yml/badge.svg)](https://github.com/m-mizutani/harlog/actions/workflows/test.yml)
[![Lint](https://github.com/m-mizutani/harlog/actions/workflows/lint.yml/badge.svg)](https://github.com/m-mizutani/harlog/actions/workflows/lint.yml)

harlog is a Go library that provides HTTP middleware and RoundTripper implementations for logging HTTP traffic in HAR (HTTP Archive) format. It can be used to capture and analyze HTTP requests and responses in both server and client applications.

## Features

- Multiple server-side integration options:
  - Standard middleware pattern (`func(http.Handler) http.Handler`)
  - Traditional `http.Handler` interface
- Client-side support via `http.RoundTripper` interface
- Saves each request/response pair as a separate HAR file
- Customizable output directory and file naming
- Thread-safe file writing
- Captures full request and response details including headers, body, and timing information
- Flexible configuration using functional options pattern

## Installation

```bash
go get github.com/m-mizutani/harlog
```

## Usage

### Server-side (Middleware Pattern)

The recommended way to use harlog on the server side is with the standard middleware pattern. Here's an example using [chi](https://github.com/go-chi/chi) router:

```go
package main

import (
    "net/http"
    "github.com/go-chi/chi/v5"
    "github.com/m-mizutani/harlog"
)

func main() {
    // Create a new router
    r := chi.NewRouter()

    // Create a new logger
    logger := harlog.New()

    // Use harlog middleware for all routes
    r.Use(logger.Middleware)

    // Add your routes
    r.Get("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    })

    // Start the server
    http.ListenAndServe(":8080", r)
}
```

### Server-side (Traditional Handler)

For backward compatibility, you can also use harlog as a traditional http.Handler:

```go
package main

import (
    "net/http"
    "github.com/m-mizutani/harlog"
)

func main() {
    // Create a new logger with initial handler
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    })

    logger := harlog.New(
        harlog.WithHandler(handler),
    )

    // Use as http.Handler
    http.ListenAndServe(":8080", logger)
}
```

### Client-side (RoundTripper)

```go
package main

import (
    "net/http"
    "github.com/m-mizutani/harlog"
)

func main() {
    // Create a new logger
    logger := harlog.New()

    // Create an HTTP client
    client := &http.Client{
        Transport: logger,
    }

    // Make HTTP requests - they will be logged automatically
    resp, err := client.Get("https://api.example.com/data")
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
}
```

## Configuration Options

harlog uses the functional options pattern for configuration. The following options are available:

```go
// Set the output directory for HAR files (default: ".")
harlog.WithOutputDir("custom_logs")

// Set a custom filename generator function
harlog.WithFileNameFn(func(req *http.Request) string {
    return fmt.Sprintf("logs/%s_%s.har", time.Now().Format("20060102-150405"), req.URL.Path)
})

// Set an initial handler for http.Handler usage
harlog.WithHandler(yourHandler)

// Set a custom transport for client usage
harlog.WithTransport(customTransport)
```

If no options are provided, harlog will use these defaults:
- Output directory: "." (current directory)
- Filename: timestamp-based format (`YYYYMMDD-HHMMSS.SSS.har`)
- Handler: `http.DefaultServeMux`
- Transport: `http.DefaultTransport`

## HAR File Format

The generated HAR files follow the standard HAR 1.2 specification and include:

- Request details (method, URL, headers, query parameters)
- Response details (status, headers, body)
- Timing information
- HTTP version information

## License

Apache License 2.0
