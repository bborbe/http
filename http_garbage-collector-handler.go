// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"fmt"
	"io"
	"net/http"
	"runtime"
	"runtime/debug"
)

func asMegabyte(b uint64) uint64 {
	return b / 1024 / 1024
}

func printMemStats(w io.Writer, name string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Fprintf(w, "Memory Stats %s:\n", name)
	fmt.Fprintf(w, "  Allocated (used) memory: %d MB\n", asMegabyte(m.Alloc))
	fmt.Fprintf(w, "  Total memory obtained from OS (reserved): %d MB\n", asMegabyte(m.Sys))
	fmt.Fprintf(w, "  Heap in use: %d MB\n", asMegabyte(m.HeapInuse))
	fmt.Fprintf(w, "  Heap released to OS: %d MB\n", asMegabyte(m.HeapReleased))
	fmt.Fprintf(w, "  Heap idle (could be released): %d MB\n", asMegabyte(m.HeapIdle))
}

func NewGarbageCollectorHandler() http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "text/plain")
		printMemStats(resp, "Before GC")

		runtime.GC()
		printMemStats(resp, "After GC")

		debug.FreeOSMemory()
		printMemStats(resp, "After FreeOSMemory")
	})
}
