// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libhttp "github.com/bborbe/http"
)

var _ = Describe("SendJSONResponse", func() {
	var ctx context.Context
	var response *httptest.ResponseRecorder
	var data interface{}
	var statusCode int
	var err error

	BeforeEach(func() {
		ctx = context.Background()
		response = httptest.NewRecorder()
	})

	JustBeforeEach(func() {
		err = libhttp.SendJSONResponse(ctx, response, data, statusCode)
	})

	Context("with valid data", func() {
		BeforeEach(func() {
			data = map[string]interface{}{
				"message": "test",
				"count":   42,
			}
			statusCode = http.StatusOK
		})

		It("returns no error", func() {
			Expect(err).To(BeNil())
		})

		It("sets correct content type", func() {
			Expect(response.Header().Get("Content-Type")).To(Equal("application/json"))
		})

		It("sets correct status code", func() {
			Expect(response.Code).To(Equal(http.StatusOK))
		})

		It("writes JSON data to response", func() {
			var result map[string]interface{}
			err := json.Unmarshal(response.Body.Bytes(), &result)
			Expect(err).To(BeNil())
			Expect(result["message"]).To(Equal("test"))
			Expect(result["count"]).To(BeNumerically("==", 42))
		})
	})

	Context("with nil data", func() {
		BeforeEach(func() {
			data = nil
			statusCode = http.StatusOK
		})

		It("returns no error", func() {
			Expect(err).To(BeNil())
		})

		It("writes null to response", func() {
			Expect(response.Body.String()).To(Equal("null\n"))
		})
	})

	Context("with empty struct", func() {
		BeforeEach(func() {
			data = struct{}{}
			statusCode = http.StatusCreated
		})

		It("returns no error", func() {
			Expect(err).To(BeNil())
		})

		It("sets correct status code", func() {
			Expect(response.Code).To(Equal(http.StatusCreated))
		})

		It("writes empty JSON object", func() {
			Expect(response.Body.String()).To(Equal("{}\n"))
		})
	})

	Context("with slice data", func() {
		BeforeEach(func() {
			data = []string{"item1", "item2", "item3"}
			statusCode = http.StatusAccepted
		})

		It("returns no error", func() {
			Expect(err).To(BeNil())
		})

		It("sets correct status code", func() {
			Expect(response.Code).To(Equal(http.StatusAccepted))
		})

		It("writes JSON array", func() {
			var result []string
			err := json.Unmarshal(response.Body.Bytes(), &result)
			Expect(err).To(BeNil())
			Expect(result).To(Equal([]string{"item1", "item2", "item3"}))
		})
	})

	Context("with different status codes", func() {
		BeforeEach(func() {
			data = "test"
		})

		Context("400 Bad Request", func() {
			BeforeEach(func() {
				statusCode = http.StatusBadRequest
			})

			It("sets correct status code", func() {
				Expect(response.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("500 Internal Server Error", func() {
			BeforeEach(func() {
				statusCode = http.StatusInternalServerError
			})

			It("sets correct status code", func() {
				Expect(response.Code).To(Equal(http.StatusInternalServerError))
			})
		})
	})

	Context("with unencodable data", func() {
		BeforeEach(func() {
			// channels cannot be JSON-encoded
			data = make(chan int)
			statusCode = http.StatusOK
		})

		It("returns error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("json: unsupported type"))
		})

		It("still sets headers and status code", func() {
			Expect(response.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(response.Code).To(Equal(http.StatusOK))
		})
	})
})

var _ = Describe("SendJSONFileResponse", func() {
	var ctx context.Context
	var response *httptest.ResponseRecorder
	var data interface{}
	var fileName string
	var statusCode int
	var err error

	BeforeEach(func() {
		ctx = context.Background()
		response = httptest.NewRecorder()
	})

	JustBeforeEach(func() {
		err = libhttp.SendJSONFileResponse(ctx, response, data, fileName, statusCode)
	})

	Context("with valid data and filename", func() {
		BeforeEach(func() {
			data = map[string]interface{}{
				"message": "test",
				"count":   42,
			}
			fileName = "download.json"
			statusCode = http.StatusOK
		})

		It("returns no error", func() {
			Expect(err).To(BeNil())
		})

		It("sets correct content type", func() {
			Expect(response.Header().Get("Content-Type")).To(Equal("application/json"))
		})

		It("sets correct status code", func() {
			Expect(response.Code).To(Equal(http.StatusOK))
		})

		It("sets content-disposition header with attachment", func() {
			disposition := response.Header().Get("Content-Disposition")
			Expect(disposition).To(Equal(`attachment; filename="download.json"`))
		})

		It("sets content-length header", func() {
			contentLength := response.Header().Get("Content-Length")
			Expect(contentLength).NotTo(BeEmpty())
			// Verify it matches the actual body length
			Expect(contentLength).To(Equal("29"))
		})

		It("writes correct JSON data", func() {
			var result map[string]interface{}
			err := json.Unmarshal(response.Body.Bytes(), &result)
			Expect(err).To(BeNil())
			Expect(result["message"]).To(Equal("test"))
			Expect(result["count"]).To(BeNumerically("==", 42))
		})
	})

	Context("with invalid filenames", func() {
		BeforeEach(func() {
			data = "test"
			statusCode = http.StatusOK
		})

		Context("with quotes in filename", func() {
			BeforeEach(func() {
				fileName = `test"file.json`
			})

			It("returns validation error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("filename contains quotes"))
			})
		})

		Context("with path traversal in filename", func() {
			BeforeEach(func() {
				fileName = "../../etc/passwd"
			})

			It("returns validation error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("filename contains '..'"))
			})
		})

		Context("with absolute path in filename", func() {
			BeforeEach(func() {
				fileName = "/etc/secret/config.json"
			})

			It("returns validation error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("filename starts with a slash"))
			})
		})

		Context("with newline in filename (header injection attempt)", func() {
			BeforeEach(func() {
				fileName = "test\r\nX-Injected: malicious"
			})

			It("returns validation error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("filename contains control character"))
			})
		})

		Context("with control characters in filename", func() {
			BeforeEach(func() {
				fileName = "test\x00\x01\x1ffile.json"
			})

			It("returns validation error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("filename contains control character"))
			})
		})

		Context("with empty filename", func() {
			BeforeEach(func() {
				fileName = ""
			})

			It("returns validation error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("filename cannot be empty"))
			})
		})

		Context("with path separator in filename", func() {
			BeforeEach(func() {
				fileName = "test/file.json"
			})

			It("returns validation error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("filename contains a path separator"))
			})
		})

		Context("with backslash in filename", func() {
			BeforeEach(func() {
				fileName = `test\file.json`
			})

			It("returns validation error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("filename contains a path separator"))
			})
		})
	})

	Context("with nil data", func() {
		BeforeEach(func() {
			data = nil
			fileName = "null.json"
			statusCode = http.StatusOK
		})

		It("returns no error", func() {
			Expect(err).To(BeNil())
		})

		It("writes null to response", func() {
			Expect(response.Body.String()).To(Equal("null"))
		})
	})

	Context("with empty struct", func() {
		BeforeEach(func() {
			data = struct{}{}
			fileName = "empty.json"
			statusCode = http.StatusCreated
		})

		It("returns no error", func() {
			Expect(err).To(BeNil())
		})

		It("sets correct status code", func() {
			Expect(response.Code).To(Equal(http.StatusCreated))
		})

		It("writes empty JSON object", func() {
			Expect(response.Body.String()).To(Equal("{}"))
		})
	})

	Context("with slice data", func() {
		BeforeEach(func() {
			data = []string{"item1", "item2", "item3"}
			fileName = "items.json"
			statusCode = http.StatusAccepted
		})

		It("returns no error", func() {
			Expect(err).To(BeNil())
		})

		It("sets correct status code", func() {
			Expect(response.Code).To(Equal(http.StatusAccepted))
		})

		It("writes JSON array", func() {
			var result []string
			err := json.Unmarshal(response.Body.Bytes(), &result)
			Expect(err).To(BeNil())
			Expect(result).To(Equal([]string{"item1", "item2", "item3"}))
		})
	})

	Context("with different status codes", func() {
		BeforeEach(func() {
			data = "test"
			fileName = "test.json"
		})

		Context("400 Bad Request", func() {
			BeforeEach(func() {
				statusCode = http.StatusBadRequest
			})

			It("sets correct status code", func() {
				Expect(response.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("500 Internal Server Error", func() {
			BeforeEach(func() {
				statusCode = http.StatusInternalServerError
			})

			It("sets correct status code", func() {
				Expect(response.Code).To(Equal(http.StatusInternalServerError))
			})
		})
	})

	Context("with unencodable data", func() {
		BeforeEach(func() {
			// channels cannot be JSON-encoded
			data = make(chan int)
			fileName = "test.json"
			statusCode = http.StatusOK
		})

		It("returns error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("json: unsupported type"))
		})
	})

	Context("with unicode filename", func() {
		BeforeEach(func() {
			data = "test"
			fileName = "测试文件.json"
			statusCode = http.StatusOK
		})

		It("preserves unicode characters", func() {
			disposition := response.Header().Get("Content-Disposition")
			Expect(disposition).To(Equal(`attachment; filename="测试文件.json"`))
		})
	})

	Context("with very long filename", func() {
		BeforeEach(func() {
			data = "test"
			fileName = "very_long_filename_that_exceeds_normal_length_limits_but_should_still_work.json"
			statusCode = http.StatusOK
		})

		It("preserves long filename", func() {
			disposition := response.Header().Get("Content-Disposition")
			Expect(disposition).To(ContainSubstring("very_long_filename"))
		})
	})
})
