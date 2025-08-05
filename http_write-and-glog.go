// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"fmt"
	"io"

	"github.com/golang/glog"
)

// WriteAndGlog writes formatted text to both a writer and the glog at verbose level 2.
// It formats the message using fmt.Printf-style formatting and writes it to the writer with a newline.
// The same message is also logged using glog.V(2).InfoDepthf for debugging purposes.
func WriteAndGlog(w io.Writer, format string, a ...any) (n int, err error) {
	glog.V(2).InfoDepthf(1, format, a...)
	return fmt.Fprintf(w, format+"\n", a...)
}
