# PRD: Standardized JSON Error Handler

**Status**: Implemented ✅
**Created**: 2025-12-05
**Implemented**: 2025-12-05
**Obsidian Task**: [[Implement Standardized JSON Error Responses]]
**Target**: github.com/bborbe/http library

## Summary

Add standardized JSON error response handlers (`NewJSONErrorHandler` and `NewJSONErrorHandlerTx`) to the http library, enabling services to return structured error details (code, message, requestId) instead of plain text. This resolves debugging friction where clients receive generic 500 errors and must check server logs to understand failures.

## Problem Statement

### Current Behavior

HTTP services using `NewErrorHandler()` return plain text errors via `http.Error()`:

```go
// Current: http_error-handler.go line 59
http.Error(resp, fmt.Sprintf("request failed: %v", err), statusCode)
```

**Result**: Clients receive unhelpful plain text:
```
HTTP/1.1 500 Internal Server Error
Content-Type: text/plain

request failed: db view failed: parse strategy group by failed: columnGroup '' is unknown
```

### Real-World Impact (2025-12-05)

During API client implementation:
- Backend API returned `500 Internal Server Error`
- Response body: plain text "request failed"
- Actual error: "invalid parameter value" (only in server logs)
- Developer had to SSH to production, tail logs, find error
- Took 15+ minutes to diagnose simple validation error

### Why Plain Text Fails

1. **No structure** - Clients can't parse error details programmatically
2. **No error codes** - Can't distinguish validation vs internal errors
3. **No trace IDs** - Can't correlate client error to server logs
4. **No context** - Can't show which field failed or why

## Goals

### Primary Goals

1. **Structured Errors** - Return JSON error responses with code, message, details, requestId
2. **Backward Compatible** - Keep existing `NewErrorHandler()`, add new variants
3. **Consistent Pattern** - Follow existing library conventions (middleware, transactions, interfaces)
4. **Simple Migration** - One-line change to use JSON errors: `NewErrorHandler` → `NewJSONErrorHandler`

### Non-Goals

- ❌ Modify existing `NewErrorHandler()` behavior (breaking change)
- ❌ Error response localization (English only)
- ❌ Stacktraces in error responses (security risk)
- ❌ Automatic migration of all services (gradual rollout)

## User Stories

### US-1: Service Developer Adds JSON Error Handling

**As a** service developer
**I want to** replace `NewErrorHandler` with `NewJSONErrorHandler`
**So that** my API returns structured JSON errors instead of plain text

**Acceptance Criteria**:
- Change one line: `libhttp.NewErrorHandler(handler)` → `libhttp.NewJSONErrorHandler(handler)`
- Errors return JSON with structure: `{"error": {"code": "...", "message": "..."}}`
- Status codes still extracted from `ErrorWithStatusCode` interface
- Backward compatible - services using old handler unchanged

### US-2: Client Parses Error Details

**As a** MCP tool developer
**I want to** parse JSON error responses
**So that** I can show meaningful error messages to users

**Acceptance Criteria**:
- Error response is valid JSON
- Contains `error.message` field with human-readable text
- Contains `error.code` field for programmatic handling
- Optional `error.requestId` for log correlation

### US-3: Service Adds Custom Error Codes

**As a** service developer
**I want to** return typed error codes (VALIDATION_ERROR, NOT_FOUND)
**So that** clients can handle different error types appropriately

**Acceptance Criteria**:
- Interface `ErrorWithCode` extracts error code from error
- Default code is "INTERNAL_ERROR" if not specified
- Service can wrap errors: `libhttp.WrapWithCode(err, "VALIDATION_ERROR")`
- HTTP status codes correctly mapped (400 for validation, 404 for not found, 500 for internal)

### US-4: Transaction Handler Returns JSON Errors

**As a** service developer using database transactions
**I want to** use `NewJSONErrorHandlerTx()`
**So that** transaction-aware handlers also return JSON errors

**Acceptance Criteria**:
- `NewJSONErrorHandlerTx()` works like `NewErrorHandlerTx()` but returns JSON
- Supports both `NewJSONUpdateErrorHandlerTx()` and `NewJSONViewErrorHandlerTx()`
- Transaction rollback behavior unchanged
- Error format identical to non-Tx variant

## Technical Specification

### Error Response Structure

Following industry standards (Stripe API, Google Cloud, JSON:API):

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "columnGroup '' is unknown",
    "details": {
      "field": "columnGroup",
      "received": "",
      "expected": "day|week|month|year"
    },
    "requestId": "c7f43b1f-8a3d-4e2b-9c1a-5d4e3f2a1b0c"
  }
}
```

### Go Types

**Core types** (new file: `http_error-response.go`):

```go
// ErrorResponse wraps error details in standard JSON format
type ErrorResponse struct {
    Error ErrorDetails `json:"error"`
}

// ErrorDetails contains structured error information
type ErrorDetails struct {
    Code      string            `json:"code"`                 // Error type (VALIDATION_ERROR, NOT_FOUND, etc.)
    Message   string            `json:"message"`              // Human-readable error message
    Details   map[string]string `json:"details,omitempty"`    // Optional structured data from errors.HasData
    RequestID string            `json:"requestId,omitempty"`  // Trace ID for log correlation
}
```

**Error interfaces** (add to existing files):

```go
// ErrorWithCode extracts error code for categorization
type ErrorWithCode interface {
    error
    Code() string
}

// Note: ErrorWithDetails is NOT needed - use existing github.com/bborbe/errors.HasData
// HasData interface (from errors library):
//   type HasData interface {
//       Data() map[string]string
//   }

// WrapWithCode creates error with both status code and error code
func WrapWithCode(err error, code string, statusCode int) error

// WrapWithDetails convenience helper - uses errors.AddToContext internally
func WrapWithDetails(err error, code string, statusCode int, details map[string]string) error
```

**Standard error codes** (constants):

```go
const (
    ErrorCodeValidation  = "VALIDATION_ERROR"   // 400 Bad Request
    ErrorCodeNotFound    = "NOT_FOUND"          // 404 Not Found
    ErrorCodeUnauthorized = "UNAUTHORIZED"      // 401 Unauthorized
    ErrorCodeForbidden   = "FORBIDDEN"          // 403 Forbidden
    ErrorCodeInternal    = "INTERNAL_ERROR"     // 500 Internal Server Error
)
```

### Handler Functions

**New file: `http_json-error-handler.go`**:

```go
// NewJSONErrorHandler wraps WithError handler, returns JSON errors
func NewJSONErrorHandler(withError WithError) http.Handler {
    return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
        ctx := req.Context()

        if err := withError.ServeHTTP(ctx, resp, req); err != nil {
            glog.V(2).Infof("handle %s request to %s failed: %v", req.Method, req.URL.Path, err)

            // Extract status code (existing pattern)
            statusCode := http.StatusInternalServerError
            var errorWithStatusCode ErrorWithStatusCode
            if errors.As(err, &errorWithStatusCode) {
                statusCode = errorWithStatusCode.StatusCode()
            }

            // Extract error code (new pattern)
            errorCode := ErrorCodeInternal
            var errorWithCode ErrorWithCode
            if errors.As(err, &errorWithCode) {
                errorCode = errorWithCode.Code()
            }

            // Extract structured details from error chain (optional)
            // Uses existing github.com/bborbe/errors.HasData interface
            var details map[string]string
            details = errors.DataFromError(err)

            // Extract request ID from context (if available)
            requestID := extractRequestID(ctx)

            // Build error response
            errorResponse := ErrorResponse{
                Error: ErrorDetails{
                    Code:      errorCode,
                    Message:   err.Error(),
                    Details:   details,
                    RequestID: requestID,
                },
            }

            // Send JSON response
            if err := SendJSONResponse(ctx, resp, errorResponse, statusCode); err != nil {
                glog.Warningf("failed to send JSON error response: %v", err)
                http.Error(resp, "internal server error", http.StatusInternalServerError)
            }
        }
    })
}
```

**New file: `http_json-error-handler-tx.go`**:

```go
// NewJSONUpdateErrorHandlerTx wraps WithErrorTx for updates, returns JSON errors
func NewJSONUpdateErrorHandlerTx(
    txProvider libkv.TxProvider,
    withErrorTx WithErrorTx,
) http.Handler {
    return NewJSONErrorHandler(NewUpdateHandlerTx(txProvider, withErrorTx))
}

// NewJSONViewErrorHandlerTx wraps WithErrorTx for reads, returns JSON errors
func NewJSONViewErrorHandlerTx(
    txProvider libkv.TxProvider,
    withErrorTx WithErrorTx,
) http.Handler {
    return NewJSONErrorHandler(NewViewHandlerTx(txProvider, withErrorTx))
}
```

**Helper functions**:

```go
// extractRequestID gets request ID from context (X-Request-ID header)
func extractRequestID(ctx context.Context) string {
    // Implementation: extract from context or generate UUID
}

// WrapWithCode creates error with code and status
func WrapWithCode(err error, code string, statusCode int) error {
    return &errorWithCodeAndStatus{
        err:        err,
        code:       code,
        statusCode: statusCode,
    }
}

// WrapWithDetails creates error with code, status, and structured details
// Convenience helper that uses errors.AddDataToError internally
func WrapWithDetails(err error, code string, statusCode int, details map[string]string) error {
    wrappedErr := WrapWithCode(err, code, statusCode)
    return errors.AddDataToError(wrappedErr, details)
}
```

**Common details patterns**:

```go
// Validation error details
map[string]string{
    "field":    "columnGroup",
    "received": "",
    "expected": "day|week|month|year",
    "reason":   "missing_required_parameter",
}

// Not found error details
map[string]string{
    "resource": "user",
    "id":       "user-12345",
    "reason":   "resource_not_found",
}

// Multiple values (join or use multiple keys)
map[string]string{
    "field_1": "from",
    "field_1_reason": "invalid_date_format",
    "field_2": "until",
    "field_2_reason": "invalid_date_format",
}
```

**Note**: Details are `map[string]string` (from `errors.HasData` interface), not `interface{}`. This is intentional - keeps error data simple and serializable.

### Migration Path

**Before (plain text errors)**:
```go
handler := libhttp.NewErrorHandler(myHandler)
```

**After (JSON errors)**:
```go
handler := libhttp.NewJSONErrorHandler(myHandler)
```

**With transactions (before)**:
```go
handler := libhttp.NewUpdateErrorHandlerTx(txProvider, myHandler)
```

**With transactions (after)**:
```go
handler := libhttp.NewJSONUpdateErrorHandlerTx(txProvider, myHandler)
```

### Error Code and Details Usage

**Basic error with code**:

```go
// In handler
if columnGroup == "" {
    return libhttp.WrapWithCode(
        errors.New(ctx, "columnGroup '' is unknown"),
        libhttp.ErrorCodeValidation,
        http.StatusBadRequest,
    )
}
```

**Error with structured details (Option 1: Via context)**:

```go
// Add details to context BEFORE creating error
ctx = errors.AddToContext(ctx, "field", "columnGroup")
ctx = errors.AddToContext(ctx, "received", columnGroup)
ctx = errors.AddToContext(ctx, "expected", "day|week|month|year")

return libhttp.WrapWithCode(
    errors.New(ctx, "columnGroup '' is unknown"),
    libhttp.ErrorCodeValidation,
    http.StatusBadRequest,
)
```

**Error with structured details (Option 2: Convenience helper)**:

```go
// All-in-one helper
return libhttp.WrapWithDetails(
    errors.New(ctx, "columnGroup '' is unknown"),
    libhttp.ErrorCodeValidation,
    http.StatusBadRequest,
    map[string]string{
        "field":    "columnGroup",
        "received": columnGroup,
        "expected": "day|week|month|year",
    },
)
```

**Result JSON**:
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "columnGroup '' is unknown",
    "details": {
      "field": "columnGroup",
      "received": "",
      "expected": "day|week|month|year"
    },
    "requestId": "c7f43b1f"
  }
}
```

**Or use default (INTERNAL_ERROR, no details)**:
```go
// Existing code unchanged - gets default INTERNAL_ERROR code, no details
return errors.Wrap(ctx, err, "database query failed")
```

**Note**: Uses existing `github.com/bborbe/errors.HasData` interface. The JSON handler calls `errors.DataFromError()` to extract all data from the error chain.

## Implementation Plan

### Phase 1: Core Types and Interfaces (2 hours)

**Files to create**:
- `http_error-response.go` - ErrorResponse, ErrorDetails types
- `http_error-response_test.go` - Unit tests for types

**Files to modify**:
- `http_error.go` - Add ErrorWithCode interface, WrapWithCode function

**Deliverables**:
- [ ] ErrorResponse and ErrorDetails types
- [ ] ErrorWithCode interface
- [ ] WrapWithCode helper function
- [ ] Standard error code constants
- [ ] Unit tests with 100% coverage

### Phase 2: JSON Error Handler (3 hours)

**Files to create**:
- `http_json-error-handler.go` - NewJSONErrorHandler implementation
- `http_json-error-handler_test.go` - Unit tests
- `http_json-error-handler-tx.go` - Transaction variants
- `http_json-error-handler-tx_test.go` - Transaction tests

**Reuse existing**:
- `SendJSONResponse()` from `http_send-json-response.go`
- Status code extraction pattern from `http_error-handler.go`

**Deliverables**:
- [ ] NewJSONErrorHandler() function
- [ ] NewJSONUpdateErrorHandlerTx() function
- [ ] NewJSONViewErrorHandlerTx() function
- [ ] Request ID extraction helper
- [ ] Unit tests with mocks for all handlers
- [ ] Integration tests with real HTTP calls

### Phase 3: Documentation (1 hour)

**Files to modify**:
- `README.md` - Add JSON error handler section with examples

**Documentation includes**:
- [ ] When to use JSON vs plain text errors
- [ ] Migration guide (one-line change)
- [ ] Error code conventions
- [ ] Example error responses
- [ ] Transaction handler usage

**Total Estimate**: 6 hours

## Testing Strategy

### Unit Tests

**http_error-response_test.go**:
```go
func TestErrorResponse_JSON(t *testing.T) {
    // Test JSON marshaling of error response
    // Verify field presence (code, message, requestId)
    // Test with/without optional details field
}
```

**http_json-error-handler_test.go**:
```go
func TestNewJSONErrorHandler_Success(t *testing.T) {
    // Handler returns no error → no response written
}

func TestNewJSONErrorHandler_ErrorWithStatusCode(t *testing.T) {
    // Handler returns error with StatusCode() → JSON error with correct status
}

func TestNewJSONErrorHandler_ErrorWithCode(t *testing.T) {
    // Handler returns error with Code() → JSON includes error code
}

func TestNewJSONErrorHandler_DefaultCode(t *testing.T) {
    // Handler returns plain error → JSON includes "INTERNAL_ERROR"
}

func TestNewJSONErrorHandler_RequestID(t *testing.T) {
    // Context has request ID → JSON includes requestId field
}
```

**http_json-error-handler-tx_test.go**:
```go
func TestNewJSONUpdateErrorHandlerTx(t *testing.T) {
    // Transaction handler returns error → JSON error + rollback
}

func TestNewJSONViewErrorHandlerTx(t *testing.T) {
    // Read-only transaction handler returns error → JSON error
}
```

### Integration Tests

**Test with real HTTP server**:
```go
func TestJSONErrorHandler_Integration(t *testing.T) {
    handler := libhttp.NewJSONErrorHandler(
        libhttp.WithErrorFunc(func(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
            return libhttp.WrapWithCode(
                errors.New(ctx, "test error"),
                libhttp.ErrorCodeValidation,
                http.StatusBadRequest,
            )
        }),
    )

    server := httptest.NewServer(handler)
    defer server.Close()

    resp, _ := http.Get(server.URL)

    assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
    assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

    var errorResp libhttp.ErrorResponse
    json.NewDecoder(resp.Body).Decode(&errorResp)

    assert.Equal(t, "VALIDATION_ERROR", errorResp.Error.Code)
    assert.Equal(t, "test error", errorResp.Error.Message)
}
```

### Manual Testing Checklist

- [ ] Verify JSON response format with curl
- [ ] Check Content-Type: application/json header
- [ ] Confirm status codes match error types (400, 404, 500)
- [ ] Verify error codes appear correctly
- [ ] Test with/without request ID in context
- [ ] Test transaction variants (update and view)

## Dependencies

### Internal Dependencies
- `http_error-handler.go` - Status code extraction pattern
- `http_send-json-response.go` - JSON encoding logic
- `http_handler-tx.go` - Transaction handler wrapping
- `github.com/bborbe/errors` - Error wrapping, context data, `HasData` interface, `DataFromError()` function
- `github.com/bborbe/kv` - Transaction interface (for Tx variants)

### External Dependencies
- Standard library: `encoding/json`, `net/http`, `context`
- `github.com/golang/glog` - Logging
- No new external dependencies

### Backward Compatibility

**Existing handlers unchanged**:
- `NewErrorHandler()` continues returning plain text
- `NewErrorHandlerTx()` continues returning plain text
- Services using old handlers work identically

**Migration is opt-in**:
- Services choose when to migrate
- No breaking changes to existing APIs
- Both JSON and plain text errors coexist

## Rollout Plan

### Stage 1: Library Implementation
1. Implement in github.com/bborbe/http
2. Write unit and integration tests
3. Update README with examples
4. Tag library release (e.g., v1.x.0)

### Stage 2: Pilot Service
1. Update a service to use NewJSONErrorHandler
2. Deploy to dev environment
3. Test with API clients - verify JSON error parsing
4. Monitor logs for issues
5. Deploy to prod after validation

### Stage 3: Documentation & Rollout
1. Document migration process
2. Create service migration guide
3. Announce to team with examples
4. Services migrate gradually (non-breaking)

### Success Metrics
- Library tests pass with 100% coverage
- Pilot service successfully returns JSON errors
- API clients can parse and display error messages
- No regressions in existing services

## Open Questions

- [x] Should we support error localization? → No, English only for now
- [x] Should we include stacktraces? → No, security risk
- [x] Should we auto-generate request IDs? → Yes, if not in context
- [ ] Should we add error details helper functions? (e.g., for field validation errors)
- [ ] Should we add middleware to inject request IDs automatically?

## Related Resources

- **Obsidian Task**: `~/Documents/Obsidian/23 Tasks/Implement Standardized JSON Error Responses.md`
- **Library**: `~/Documents/workspaces/http/`
- **Similar APIs**: Stripe API error format, Google Cloud API errors, JSON:API spec

## Updates Log

**2025-12-05**: Initial PRD created based on API debugging experience
**2025-12-05**: PRD completed with full technical specification
- Error response structure defined (ErrorResponse, ErrorDetails types)
- Integration with existing `github.com/bborbe/errors.HasData` interface
- Both handler variants specified: `NewJSONErrorHandler()` and `NewJSONErrorHandlerTx()`
- Details field uses `map[string]string` (from errors.DataFromError)
- Two usage patterns documented: context-based and convenience helper
- Testing strategy and rollout plan complete
- Status: Ready for implementation (Phase 1 estimated 6 hours)
