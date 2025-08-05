// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package http provides comprehensive HTTP utilities for building robust server and client applications.
//
// The package offers three main categories of functionality:
//
// # HTTP Server Components
//
// Server utilities with graceful shutdown support:
//   - NewServer and NewServerWithPort for creating HTTP servers
//   - NewServerTLS for HTTPS servers with certificate handling
//   - Background request handlers for async processing
//   - JSON response handlers with automatic content-type management
//   - Error handling middleware with structured responses
//   - Profiling endpoints for debugging (CPU, memory, pprof)
//
// # HTTP Client Components
//
// RoundTripper middleware for HTTP clients:
//   - Retry logic with configurable delays and skip conditions
//   - Rate limiting to prevent API abuse
//   - Authentication (Basic Auth, Header-based)
//   - Request/response logging with configurable verbosity
//   - Metrics collection (Prometheus compatible)
//   - Header manipulation and path prefix removal
//
// # Proxy and Middleware
//
// Proxy utilities and additional middleware:
//   - Reverse proxy with error handling
//   - Sentry integration for error reporting
//   - Content type utilities
//   - Request validation and response checking
//
// # Example Usage
//
// Basic HTTP server:
//
//	router := mux.NewRouter()
//	router.HandleFunc("/health", healthHandler)
//	server := http.NewServerWithPort(8080, router)
//	run.CancelOnInterrupt(context.Background(), server)
//
// HTTP client with retry:
//
//	transport := http.NewRoundTripperRetry(http.DefaultTransport, 3, time.Second*2)
//	client := &http.Client{Transport: transport}
//
// JSON handler:
//
//	jsonHandler := http.NewJsonHandler(http.JsonHandlerFunc(myHandler))
//	errorHandler := http.NewErrorHandler(jsonHandler)
//	router.Handle("/api/data", errorHandler)
package http
