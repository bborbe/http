// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/bborbe/errors"
)

// SendJSONResponse writes a JSON-encoded response with the specified status code.
// It sets the Content-Type header to application/json and returns any encoding errors.
func SendJSONResponse(
	ctx context.Context,
	resp http.ResponseWriter,
	data interface{},
	statusCode int,
) error {
	resp.Header().Set(ContentTypeHeaderName, ApplicationJsonContentType)
	resp.WriteHeader(statusCode)
	if err := json.NewEncoder(resp).Encode(data); err != nil {
		return errors.Wrapf(ctx, err, "encode json response failed")
	}
	return nil
}
