# HTTP Library

A comprehensive Go HTTP utilities library providing robust server and client functionality with extensive middleware capabilities, graceful shutdown support, and production-ready features.

## Features

### 🚀 HTTP Server
- **Graceful shutdown** with context cancellation
- **TLS support** with automatic certificate handling
- **Background request processing** for long-running operations
- **JSON response handlers** with automatic content-type headers
- **Error handling middleware** with structured error responses
- **Profiling endpoints** (CPU, memory, pprof) for debugging
- **File serving** capabilities

### 🔧 HTTP Client & RoundTrippers
- **Retry logic** with configurable delays and skip conditions
- **Rate limiting** to prevent API abuse
- **Authentication** (Basic Auth, Header-based)
- **Request/response logging** with configurable verbosity
- **Metrics collection** (Prometheus compatible)
- **Header manipulation** and path prefix removal
- **Request building utilities**

### 🛡️ Proxy & Middleware
- **Reverse proxy** with error handling
- **Sentry integration** for error reporting
- **Background handlers** for async processing
- **Content type utilities**
- **Request validation** and response checking

## Installation

```bash
go get github.com/bborbe/http
```

## Quick Start

### Basic HTTP Server

```go
package main

import (
    "context"
    "net/http"
    
    bhttp "github.com/bborbe/http"
    "github.com/bborbe/run"
    "github.com/gorilla/mux"
)

func main() {
    router := mux.NewRouter()
    router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })
    
    server := bhttp.NewServerWithPort(8080, router)
    run.CancelOnInterrupt(context.Background(), server)
}
```

### JSON Handler

```go
type Response struct {
    Message string `json:"message"`
    Status  string `json:"status"`
}

func healthHandler(ctx context.Context, req *http.Request) (interface{}, error) {
    return Response{
        Message: "Service is healthy",
        Status:  "OK",
    }, nil
}

func main() {
    router := mux.NewRouter()
    
    // JSON handler automatically sets Content-Type and marshals response
    jsonHandler := bhttp.NewJsonHandler(bhttp.JsonHandlerFunc(healthHandler))
    errorHandler := bhttp.NewErrorHandler(jsonHandler)
    
    router.Handle("/api/health", errorHandler)
    
    server := bhttp.NewServerWithPort(8080, router)
    run.CancelOnInterrupt(context.Background(), server)
}
```

### JSON Error Handler

The JSON error handler returns structured error responses in JSON format instead of plain text, making errors easier to parse and handle programmatically.

```go
type ErrorResponse struct {
    Message string `json:"message"`
}

func apiHandler(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
    // Return an error with status code
    return bhttp.WrapWithCode(
        errors.New(ctx, "validation failed"),
        bhttp.ErrorCodeValidation,
        http.StatusBadRequest,
    )
}

func main() {
    router := mux.NewRouter()

    // JSON error handler returns structured JSON error responses
    handler := bhttp.NewJSONErrorHandler(
        bhttp.WithErrorFunc(apiHandler),
    )

    router.Handle("/api/resource", handler)

    server := bhttp.NewServerWithPort(8080, router)
    run.CancelOnInterrupt(context.Background(), server)
}
```

**Error Response Format:**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "validation failed",
    "details": {
      "field": "email",
      "reason": "invalid_format"
    }
  }
}
```

**Available Error Codes:**
- `ErrorCodeValidation` - For validation errors (400)
- `ErrorCodeNotFound` - For not found errors (404)
- `ErrorCodeUnauthorized` - For authentication errors (401)
- `ErrorCodeForbidden` - For authorization errors (403)
- `ErrorCodeInternal` - For internal server errors (500)

**With Error Details:**
```go
// Add structured details to errors
return bhttp.WrapWithDetails(
    errors.New(ctx, "validation failed"),
    bhttp.ErrorCodeValidation,
    http.StatusBadRequest,
    map[string]string{
        "field": "email",
        "reason": "invalid_format",
    },
)
```

**With Database Transactions:**
```go
// For update operations
handler := bhttp.NewJSONUpdateErrorHandler(db,
    bhttp.WithErrorTxFunc(func(ctx context.Context, tx libkv.Tx, resp http.ResponseWriter, req *http.Request) error {
        // Handle update logic with transaction
        return nil
    }),
)

// For read-only operations
handler := bhttp.NewJSONViewErrorHandler(db,
    bhttp.WithErrorTxFunc(func(ctx context.Context, tx libkv.Tx, resp http.ResponseWriter, req *http.Request) error {
        // Handle read logic with transaction
        return nil
    }),
)
```

### HTTP Client with Retry

```go
package main

import (
    "context"
    "net/http"
    "time"

    bhttp "github.com/bborbe/http"
)

func main() {
    // Create HTTP client with retry logic
    transport := bhttp.NewRoundTripperRetry(
        http.DefaultTransport,
        3,                    // retry limit
        time.Second * 2,      // retry delay
    )

    client := &http.Client{
        Transport: transport,
        Timeout:   time.Second * 30,
    }

    // Make request - automatically retries on failure
    resp, err := client.Get("https://api.example.com/data")
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
}
```

### Background Request Handler

```go
func longRunningTask(ctx context.Context, req *http.Request) error {
    // Simulate long-running task
    time.Sleep(10 * time.Second)
    
    // Your background processing logic here
    return nil
}

func main() {
    router := mux.NewRouter()
    
    // Background handler processes requests asynchronously
    bgHandler := bhttp.NewBackgroundRequestHandler(
        bhttp.BackgroundRequestHandlerFunc(longRunningTask),
    )
    errorHandler := bhttp.NewErrorHandler(bgHandler)
    
    router.Handle("/api/process", errorHandler)
    
    server := bhttp.NewServerWithPort(8080, router)
    run.CancelOnInterrupt(context.Background(), server)
}
```

### HTTP Proxy

```go
func main() {
    targetURL, _ := url.Parse("https://api.backend.com")
    
    // Create proxy with error handling
    errorHandler := bhttp.NewProxyErrorHandler()
    transport := http.DefaultTransport
    
    proxy := bhttp.NewProxy(transport, targetURL, errorHandler)
    
    server := bhttp.NewServerWithPort(8080, proxy)
    run.CancelOnInterrupt(context.Background(), server)
}
```

### Advanced Client with Middleware Stack

```go
func main() {
    // Build client with multiple middleware layers
    transport := http.DefaultTransport
    
    // Add retry logic
    transport = bhttp.NewRoundTripperRetry(transport, 3, time.Second*2)
    
    // Add authentication
    transport = bhttp.NewRoundTripperBasicAuth(transport, "username", "password")
    
    // Add logging
    transport = bhttp.NewRoundTripperLog(transport)
    
    // Add rate limiting
    transport = bhttp.NewRoundTripperRateLimit(transport, 10) // 10 req/sec
    
    client := &http.Client{
        Transport: transport,
        Timeout:   time.Second * 30,
    }
    
    // Client now has retry, auth, logging, and rate limiting
    resp, err := client.Get("https://api.example.com/protected")
    // Handle response...
}
```

## Testing

The library includes comprehensive test coverage and mock generation using Counterfeiter:

```bash
# Run tests
make test

# Run all quality checks (format, test, lint, etc.)
make precommit

# Generate mocks for testing
make generate
```

### Pre-generated Mocks

The library provides pre-generated mocks in the `mocks/` package for easy testing:

```go
import "github.com/bborbe/http/mocks"

// Available mocks:
// - HttpHandler
// - HttpJsonHandler  
// - HttpJsonHandlerTx
// - HttpProxyErrorHandler
// - HttpRoundtripper
// - HttpRoundtripperMetrics
// - HttpWithError
```

Example usage in tests:

```go
func TestMyService(t *testing.T) {
    mockHandler := &mocks.HttpJsonHandler{}
    mockHandler.ServeHTTPReturns(map[string]string{"status": "ok"}, nil)
    
    // Use mock in your test...
}
```

## Advanced Features

### Profiling Support
Built-in handlers for performance profiling:
- CPU profiling: `/debug/pprof/profile`
- Memory profiling: `/debug/pprof/heap`
- Goroutine profiling: `/debug/pprof/goroutine`

### Metrics Integration
Prometheus-compatible metrics collection for monitoring request performance, error rates, and more.

### Error Handling
Structured error handling with context preservation and optional Sentry integration for production error tracking.

### Graceful Shutdown
All server components support graceful shutdown with proper resource cleanup when receiving termination signals.

## API Documentation

Complete API documentation is available on [pkg.go.dev](https://pkg.go.dev/github.com/bborbe/http).

View the documentation locally with:
```bash
go doc -all github.com/bborbe/http
```

## Dependencies

This library uses minimal external dependencies:
- Standard `net/http` package
- `github.com/bborbe/errors` for enhanced error handling
- `github.com/bborbe/run` for graceful lifecycle management
- Optional Prometheus metrics support
- Optional Sentry error reporting

## License

BSD-style license. See LICENSE file for details.
