// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import "net/http"

// HandlerFunc is an alias for http.HandlerFunc to enable mock generation.
// It provides the same interface as the standard library's HandlerFunc for HTTP request handling.
type HandlerFunc http.HandlerFunc
