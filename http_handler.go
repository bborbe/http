// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import "net/http"

//counterfeiter:generate -o mocks/http-handler.go --fake-name HttpHandler . Handler

// Handler is an alias for http.Handler to enable mock generation.
// It provides the same interface as the standard library's Handler for HTTP request handling.
type Handler http.Handler
