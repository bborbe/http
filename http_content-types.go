// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

const (
	// ApplicationJSONContentType is the MIME type for JSON responses.
	ApplicationJSONContentType = "application/json"
	// TextHTML is the MIME type for HTML responses.
	TextHTML = "text/html"

	// ApplicationJsonContentType is deprecated. Use ApplicationJSONContentType instead.
	//
	// Deprecated: Use ApplicationJSONContentType for correct Go naming conventions.
	//
	//nolint:revive
	ApplicationJsonContentType = ApplicationJSONContentType

	// TextHtml is deprecated. Use TextHTML instead.
	//
	// Deprecated: Use TextHTML for correct Go naming conventions.
	//
	//nolint:revive
	TextHtml = TextHTML
)
