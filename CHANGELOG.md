# Changelog

All notable changes to this project will be documented in this file.

Please choose versions by [Semantic Versioning](http://semver.org/).

* MAJOR version when you make incompatible API changes,
* MINOR version when you add functionality in a backwards-compatible manner, and
* PATCH version when you make backwards-compatible bug fixes.

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
