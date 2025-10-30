# Changelog

All notable changes to this project will be documented in this file.

Please choose versions by [Semantic Versioning](http://semver.org/).

* MAJOR version when you make incompatible API changes,
* MINOR version when you add functionality in a backwards-compatible manner, and
* PATCH version when you make backwards-compatible bug fixes.

## v1.19.0
- Fix critical context bug: replace context.Background() with context.WithoutCancel(ctx) in server shutdown to preserve trace context
- Add ErrNotFound sentinel error for 404 responses (exported for errors.Is comparisons)
- Add ErrTooManyRedirects sentinel error for redirect limit exceeded
- Fix error wrapping: replace errors.Wrapf with errors.Wrap where no format arguments used (15 locations)
- Add deprecation wrappers for Go naming conventions: Json→JSON, Http→HTTP, Tls→TLS
- Fix WithInsecureSkipVerify bug: now correctly returns builder instead of nil
- Fix redirect limit checking: use configured h.maxRedirect instead of hardcoded 10
- Enable golangci-lint in Makefile check target
- Add tests for ErrNotFound sentinel error wrapping
- Improve API consistency with proper error naming (ST1012 compliance)
- Maintain test coverage at 52.3%

## v1.18.0
- Update Go version from 1.25.2 to 1.25.3 (fixes OSV vulnerability GO-2025-4007)
- Add comprehensive timeout configuration to ServerOptions (ReadTimeout, WriteTimeout, IdleTimeout, ShutdownTimeout, MaxHeaderBytes)
- Fix critical security issue: set TLS MinVersion to TLS 1.2 in http_client-builder.go and http_roundtripper-default.go
- Fix Slowloris attack vulnerability: correctly use ReadHeaderTimeout in http_server.go
- Implement graceful shutdown with configurable timeout using separate context
- Refactor HTTP server creation with CreateHttpServer and CreateServerOptions helper functions
- Add comprehensive GoDoc documentation for ServerOptions struct
- Fix error handling: use errors.Wrap instead of errors.Wrapf when no format arguments needed
- Add security suppressions with justification for legitimate file operations (CA cert loading, file downloader)
- Pass all gosec security checks (0 issues, 2 documented suppressions)
- Increase test coverage from 38.1% to 39.7%
- Set production-ready default timeouts: ReadHeaderTimeout=10s, ReadTimeout=30s, WriteTimeout=30s, IdleTimeout=60s, ShutdownTimeout=5s, MaxHeaderBytes=1MB

## v1.17.0
- Add ValidateFilename function for secure filename validation
- Add SendJSONFileResponse for JSON file downloads with Content-Disposition header
- Implement comprehensive security checks to prevent header injection and path traversal attacks
- Add extensive test coverage for filename validation and file download functionality
- Increase test coverage from 33.8% to 38.1%

## v1.16.0
- Add SendJSONResponse helper function for writing JSON responses
- Add comprehensive test coverage for SendJSONResponse

## v1.15.2
- Update Go version from 1.25.1 to 1.25.2

## v1.15.1
- Update github.com/google/osv-scanner from v1.9.2 to v2.2.3
- Add support for .osv-scanner.toml configuration file in Makefile
- Update transitive dependencies

## v1.15.0
- Upgrade Go version from 1.24.5 to 1.25.1
- Add golangci-lint integration with .golangci.yml configuration
- Add security scanning tools: osv-scanner, gosec, and trivy
- Add golines for consistent line length formatting (max 100 chars)
- Update goimports-reviser to v3 with improved formatting
- Update multiple dependencies to latest versions
- Add Trivy installation to CI workflow
- Improve Makefile with additional quality checks and security tools
- Update import formatting across codebase

## v1.14.2

- Improve godoc for BuildRequest function to clarify parameters handling

## v1.14.1

- Add comprehensive package documentation with examples and usage guides
- Enhance README with detailed feature descriptions and code examples
- Add license headers to all Go source files
- Improve code documentation and formatting

## v1.14.0

- add NewBackgroundRunRequestHandler for background processing with request access

## v1.13.2

- add github workflow
- go mod update

## v1.13.1

- go mod update
- add tests

## v1.13.0

- add RoundTripperMetrics

## v1.12.0

- add GarbageCollectorHandler

## v1.11.1

- add RoundTripperFunc

## v1.11.0

- add http.Handler mock
- add Handler and HandlerFunc

## v1.10.3

- RoundTripperRetry retry http request on io.EOF error

## v1.10.2

- improve WithRedirects

## v1.10.1

- add WithRetry and WithoutRetry to HttpClientBuilder
- go mod update

## v1.10.0

- remove vendor
- go mod update

## v1.9.0

- allow define error code error use by ErrorHandler
- go mod update

## v1.8.1

- allow define skip statusCodes in RetryRoundTripper
- go mod update

## v1.8.0

- add UpdateErrorHandler
- add ViewErrorHandler
- add WithErrorTx
- add JsonHandler
- add JsonHandlerTx

## v1.7.1

- add missing license
- go mod update

## v1.7.0

- add CheckResponseIsSuccessful
- go mod update

## v1.6.0

- add file download handler
- add pprof handler

## v1.5.6

- skip error: http: TLS handshake error from

## v1.5.5

- allow HttpClientBuilder with client cert

## v1.5.4

- add CreateTlsClientConfig
- add CreateDefaultRoundTripperTls

## v1.5.3

- fix NewServerTLS

## v1.5.2

- add NewServerTLS
- go mod update

## v1.5.1

- add basic auth roundtripper

## v1.5.0

- add helper to register pprof handler

## v1.4.0

- add remove prefix roundTripper
- go mod update

## v1.3.0

- add proxy error handler sentry
- go mod update

## v1.2.0

- go mod update
- remove ratelimiter from default http client

## v1.1.0

- add HttpClientBuilder
- go mod update

## v1.0.0

- Initial Version
