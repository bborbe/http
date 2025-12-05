# PRD: Request ID Middleware and Tracing

**Status**: Draft 📝
**Created**: 2025-12-05
**Related**: JSON Error Handler (implemented)
**Target**: github.com/bborbe/http library

## Summary

Add request ID middleware for distributed tracing across HTTP services. Enables request correlation through logs, errors, and service boundaries by extracting/generating request IDs and propagating them via context and response headers.

## Problem Statement

### Current State

JSON error handler has a `requestId` field in error responses, but:
- No middleware to extract request IDs from headers
- No middleware to generate request IDs if missing
- No way to propagate request IDs to downstream services
- Request ID context key is private to json-error-handler.go

**Result**: `requestId` field is always empty in error responses.

### Real-World Problem

When debugging issues across microservices:
1. Client makes request to Service A → receives error
2. Service A calls Service B → Service B fails
3. Developer has error message but can't find related logs
4. Must manually correlate timestamps across multiple services
5. Takes 15+ minutes to trace issue through service chain

### Example Scenario

**Without Request IDs**:
```
[Service A logs] 2025-12-05 10:23:15 ERROR: database query failed
[Service B logs] 2025-12-05 10:23:14 ERROR: connection timeout
[Client error]   {"error": {"code": "INTERNAL_ERROR", "message": "database query failed"}}
```
**Question**: Which Service A error corresponds to which Service B error?

**With Request IDs**:
```
[Service A logs] 2025-12-05 10:23:15 req=abc123 ERROR: database query failed
[Service B logs] 2025-12-05 10:23:14 req=abc123 ERROR: connection timeout
[Client error]   {"error": {"code": "INTERNAL_ERROR", "message": "...", "requestId": "abc123"}}
```
**Answer**: grep for `req=abc123` finds entire request chain instantly.

## Goals

### Primary Goals

1. **Extract Request IDs** - Read from standard headers (X-Request-ID, X-Correlation-ID, Trace-Id)
2. **Generate Request IDs** - Create UUID if no header present
3. **Context Propagation** - Store in context for handlers/loggers to access
4. **Response Headers** - Echo request ID back to client
5. **Downstream Propagation** - Inject into outbound HTTP requests

### Non-Goals

- ❌ Full OpenTelemetry tracing (just request IDs)
- ❌ Trace sampling or performance monitoring
- ❌ Request ID validation/format enforcement
- ❌ Database transaction correlation (use separate tools)

## User Stories

### US-1: Service Developer Adds Request ID Middleware

**As a** service developer
**I want to** add request ID middleware to my HTTP server
**So that** all requests get a unique ID for tracing

**Acceptance Criteria**:
- One-line middleware addition: `NewRequestIDMiddleware()`
- Extracts from `X-Request-ID` header if present
- Generates UUID v4 if header missing
- Stores in context via public context key
- Adds `X-Request-ID` response header

### US-2: Logger Includes Request ID

**As a** developer debugging issues
**I want** request IDs in all log entries
**So that** I can trace a request through the entire codebase

**Acceptance Criteria**:
- Context contains request ID accessible via `RequestIDFromContext(ctx)`
- Logger can extract and include in structured logs
- Works with glog, zap, logrus, etc.

### US-3: Error Responses Include Request ID

**As a** client/frontend developer
**I want** error responses to include request ID
**So that** I can report issues with a trace ID

**Acceptance Criteria**:
- JSON error handler automatically includes request ID
- Extracted via `RequestIDFromContext(ctx)`
- Field is populated (not empty) when middleware enabled

### US-4: Service Propagates Request ID to Dependencies

**As a** service making downstream HTTP calls
**I want to** propagate request IDs to dependencies
**So that** distributed traces are connected

**Acceptance Criteria**:
- RoundTripper middleware adds `X-Request-ID` header
- Extracts from context via `RequestIDFromContext(ctx)`
- Works with existing retry/auth/logging RoundTrippers

## Technical Specification

### Middleware Implementation

**New file: `http_request-id-middleware.go`**

```go
package http

import (
    "context"
    "net/http"

    "github.com/google/uuid"
)

// Context key for request ID (exported)
type contextKey string

const RequestIDKey contextKey = "request-id"

// Standard headers to check (in priority order)
var requestIDHeaders = []string{
    "X-Request-ID",
    "X-Correlation-ID",
    "Trace-Id",
}

// RequestIDFromContext extracts request ID from context
func RequestIDFromContext(ctx context.Context) string {
    if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
        return requestID
    }
    return ""
}

// NewRequestIDMiddleware creates middleware that extracts/generates request IDs
func NewRequestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract from headers (check all standard headers)
        requestID := extractRequestIDFromHeaders(r)

        // Generate if missing
        if requestID == "" {
            requestID = uuid.New().String()
        }

        // Store in context
        ctx := context.WithValue(r.Context(), RequestIDKey, requestID)

        // Echo back in response header
        w.Header().Set("X-Request-ID", requestID)

        // Call next handler with updated context
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func extractRequestIDFromHeaders(r *http.Request) string {
    for _, header := range requestIDHeaders {
        if id := r.Header.Get(header); id != "" {
            return id
        }
    }
    return ""
}
```

### RoundTripper for Downstream Propagation

**New file: `http_roundtripper-request-id.go`**

```go
// NewRoundTripperRequestID creates a RoundTripper that propagates request IDs
func NewRoundTripperRequestID(transport http.RoundTripper) http.RoundTripper {
    return &roundTripperRequestID{
        transport: transport,
    }
}

type roundTripperRequestID struct {
    transport http.RoundTripper
}

func (r *roundTripperRequestID) RoundTrip(req *http.Request) (*http.Response, error) {
    // Extract request ID from context
    if requestID := RequestIDFromContext(req.Context()); requestID != "" {
        // Clone request and add header
        reqClone := req.Clone(req.Context())
        reqClone.Header.Set("X-Request-ID", requestID)
        return r.transport.RoundTrip(reqClone)
    }

    return r.transport.RoundTrip(req)
}
```

### JSON Error Handler Integration

**Modify: `http_json-error-handler.go`**

```go
// Remove private contextKey and requestIDKey
// Remove extractRequestID() function

// Update NewJSONErrorHandler:
func NewJSONErrorHandler(withError WithError) http.Handler {
    // ...

    // Extract request ID from context (now uses public function)
    requestID := RequestIDFromContext(ctx)

    // Build error response
    errorResponse := ErrorResponse{
        Error: ErrorDetails{
            Code:      errorCode,
            Message:   err.Error(),
            Details:   details,
            RequestID: requestID, // Now populated when middleware present
        },
    }
    // ...
}
```

## Usage Examples

### Server Setup

```go
func main() {
    router := mux.NewRouter()

    // Add routes
    router.Handle("/api/users", userHandler)

    // Wrap with request ID middleware
    handler := libhttp.NewRequestIDMiddleware(router)

    server := libhttp.NewServerWithPort(8080, handler)
    run.CancelOnInterrupt(context.Background(), server)
}
```

### Logging with Request ID

```go
func myHandler(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
    requestID := libhttp.RequestIDFromContext(ctx)

    // Include in logs
    glog.V(2).Infof("req=%s processing user request", requestID)

    // Use in business logic
    if err := processUser(ctx); err != nil {
        glog.Errorf("req=%s failed to process user: %v", requestID, err)
        return err
    }

    return nil
}
```

### Client Propagation

```go
func main() {
    // Build HTTP client with request ID propagation
    transport := http.DefaultTransport
    transport = libhttp.NewRoundTripperRequestID(transport)
    transport = libhttp.NewRoundTripperLog(transport)

    client := &http.Client{Transport: transport}

    // Request ID automatically propagated
    req, _ := http.NewRequestWithContext(ctx, "GET", "http://api/users", nil)
    resp, err := client.Do(req)
}
```

### Error Response with Request ID

```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "database connection failed",
    "requestId": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

## Implementation Plan

### Phase 1: Middleware (2 hours)

**Files to create**:
- `http_request-id-middleware.go` - Middleware implementation
- `http_request-id-middleware_test.go` - Unit tests

**Deliverables**:
- [ ] RequestIDKey constant (exported)
- [ ] RequestIDFromContext() helper
- [ ] NewRequestIDMiddleware() implementation
- [ ] Header extraction (X-Request-ID, X-Correlation-ID, Trace-Id)
- [ ] UUID generation for missing IDs
- [ ] Response header injection
- [ ] Unit tests with 100% coverage

### Phase 2: RoundTripper (1 hour)

**Files to create**:
- `http_roundtripper-request-id.go` - RoundTripper implementation
- `http_roundtripper-request-id_test.go` - Unit tests

**Deliverables**:
- [ ] NewRoundTripperRequestID() implementation
- [ ] Context extraction
- [ ] Header injection for outbound requests
- [ ] Unit tests with mocked transport

### Phase 3: Integration (1 hour)

**Files to modify**:
- `http_json-error-handler.go` - Update to use public RequestIDFromContext()
- `README.md` - Add middleware examples

**Deliverables**:
- [ ] Remove private context key from json-error-handler.go
- [ ] Update extractRequestID() to use RequestIDFromContext()
- [ ] Add README section for request ID middleware
- [ ] Add example showing server setup with middleware
- [ ] Add example showing client propagation

**Total Estimate**: 4 hours

## Testing Strategy

### Unit Tests

**http_request-id-middleware_test.go**:
```go
func TestRequestIDMiddleware_ExtractsFromHeader(t *testing.T) {
    // Test extraction from X-Request-ID header
}

func TestRequestIDMiddleware_GeneratesUUID(t *testing.T) {
    // Test UUID generation when header missing
}

func TestRequestIDMiddleware_SetsResponseHeader(t *testing.T) {
    // Test X-Request-ID response header
}

func TestRequestIDMiddleware_HeaderPriority(t *testing.T) {
    // Test X-Request-ID > X-Correlation-ID > Trace-Id
}
```

**http_roundtripper-request-id_test.go**:
```go
func TestRoundTripperRequestID_PropagatesFromContext(t *testing.T) {
    // Test header added from context
}

func TestRoundTripperRequestID_SkipsWhenMissing(t *testing.T) {
    // Test no header added when context empty
}
```

### Integration Tests

**Test end-to-end flow**:
1. Client sends request with `X-Request-ID: abc123`
2. Middleware extracts and stores in context
3. Handler logs with request ID
4. Error handler includes request ID in response
5. Response header echoes `X-Request-ID: abc123`

## Dependencies

### Internal
- `http_json-error-handler.go` - Update to use public context key
- `http_send-json-response.go` - No changes needed

### External
- `github.com/google/uuid` - For UUID v4 generation (already in go.mod)
- Standard library: `net/http`, `context`

### Backward Compatibility

**Breaking changes**: None
- New middleware is opt-in
- Existing code works without middleware (requestId field empty)
- Public context key doesn't break existing private usage

**Migration path**:
- Services add middleware when ready
- Old services continue working
- Request ID field in errors populated when middleware present

## Rollout Plan

### Stage 1: Library Release
1. Implement middleware and RoundTripper
2. Update JSON error handler to use public context key
3. Write comprehensive tests
4. Update README with examples
5. Tag library release (e.g., v1.24.0)

### Stage 2: Service Adoption
1. Document migration steps
2. Update one service as pilot
3. Verify request IDs appear in logs and errors
4. Verify downstream propagation works
5. Roll out to other services gradually

### Success Metrics
- Request IDs present in 100% of error responses (when middleware enabled)
- Reduced time to correlate logs (from 15min → <1min)
- Request traces visible across service boundaries

## Open Questions

- [ ] Should we support custom request ID generators? (vs always UUID v4)
- [ ] Should we validate request ID format? (e.g., max length, allowed chars)
- [ ] Should we add `X-Parent-Request-ID` for distributed traces?
- [ ] Should middleware be configurable (header names, generator)?

## Related Resources

- **RFC**: https://www.w3.org/TR/trace-context/ (W3C Trace Context)
- **Similar implementations**:
  - Gorilla mux: https://github.com/gorilla/mux
  - Chi middleware: https://github.com/go-chi/chi
  - Echo middleware: https://echo.labstack.com/middleware/request-id/
- **OpenTelemetry**: https://opentelemetry.io/docs/instrumentation/go/

## Updates Log

**2025-12-05**: Initial PRD created
- Extracted from JSON error handler review
- Identified incomplete request ID implementation
- Designed complete middleware solution
- Ready for implementation in future iteration
