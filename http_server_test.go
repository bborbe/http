// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/bborbe/run"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libhttp "github.com/bborbe/http"
)

var _ = Describe("Http Server", func() {
	var ctx context.Context
	var httpServer run.Func
	var err error
	var port int
	var cancel context.CancelFunc
	BeforeEach(func() {
		ctx, cancel = context.WithCancel(context.Background())

		port, err = freePort()
		Expect(err).To(BeNil())

		httpServer = libhttp.NewServer(
			fmt.Sprintf("localhost:%d", port),
			http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				fmt.Fprint(writer, "ok")
			}),
		)
		go func() {
			defer GinkgoRecover()
			Expect(httpServer.Run(ctx)).To(BeNil())
		}()
	})
	AfterEach(func() {
		cancel()
	})
	It("successfull get call", func() {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d", port))
		Expect(err).To(BeNil())
		Expect(resp).NotTo(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		content, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		Expect(string(content)).To(Equal("ok"))
	})
})

func freePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	tcpAddr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		return 0, fmt.Errorf("expected *net.TCPAddr, got %T", l.Addr())
	}
	return tcpAddr.Port, nil
}
