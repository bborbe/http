// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libhttp "github.com/bborbe/http"
)

var _ = Describe("FileServer", func() {
	var handler http.Handler
	var recorder *httptest.ResponseRecorder
	BeforeEach(func() {
		dir := GinkgoT().TempDir()
		root := filepath.Join(dir, "root")
		Expect(os.Mkdir(root, 0o750)).To(Succeed())
		Expect(
			os.WriteFile(filepath.Join(root, "index.html"), []byte("index"), 0o600),
		).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "app.js"), []byte("app"), 0o600)).To(Succeed())
		Expect(
			os.WriteFile(filepath.Join(dir, "secret.txt"), []byte("secret"), 0o600),
		).To(Succeed())
		handler = libhttp.FileServer(root, "/static")
		recorder = httptest.NewRecorder()
	})
	serve := func(target string) {
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, target, nil))
	}
	It("serves a file below the prefix", func() {
		serve("/static/app.js")
		Expect(recorder.Code).To(Equal(http.StatusOK))
		Expect(recorder.Body.String()).To(Equal("app"))
	})
	It("serves the default file for a missing file", func() {
		serve("/static/missing")
		Expect(recorder.Code).To(Equal(http.StatusOK))
		Expect(recorder.Body.String()).To(Equal("index"))
	})
	It("does not escape root when the prefix is followed by dots", func() {
		serve("/static../")
		Expect(recorder.Body.String()).NotTo(ContainSubstring("secret.txt"))
	})
	It("does not serve a file outside root", func() {
		serve("/static../secret.txt")
		Expect(recorder.Body.String()).NotTo(ContainSubstring("secret"))
	})
})
