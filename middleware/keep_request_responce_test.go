package middleware

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReserveRequest(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		withMiddleware bool
		config         ReserveRequestConfig
		readTwice      bool
	}{{
		name:           "get:middleware:false",
		method:         http.MethodGet,
		withMiddleware: false,
	}, {
		name:           "get:middleware:true",
		method:         http.MethodGet,
		withMiddleware: true,
	}, {
		name:           "get:middleware:true:skipper",
		method:         http.MethodGet,
		withMiddleware: true,
		config: ReserveRequestConfig{
			Skipper: func(*http.Request) bool {
				return true
			},
		},
	}, {
		name:           "get:middleware:true:readTwice",
		method:         http.MethodGet,
		withMiddleware: true,
		readTwice:      true,
	}, {
		name:           "post:middleware:false",
		method:         http.MethodPost,
		withMiddleware: false,
	}, {
		name:           "post:middleware:true",
		method:         http.MethodPost,
		withMiddleware: true,
	}, {
		name:           "post:middleware:true",
		method:         http.MethodPost,
		withMiddleware: true,
	}, {
		name:           "post:middleware:true:skipper",
		method:         http.MethodPost,
		withMiddleware: true,
		config: ReserveRequestConfig{
			Skipper: func(*http.Request) bool {
				return true
			},
		},
	}, {
		name:           "post:middleware:true:readTwice",
		method:         http.MethodPost,
		withMiddleware: true,
		readTwice:      true,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bodyString := "test body"

			var req *http.Request
			if test.method == http.MethodPost {
				req = httptest.NewRequest(http.MethodPost, "http://localhost", strings.NewReader(bodyString))
			} else {
				req = httptest.NewRequest(http.MethodGet, "http://localhost", nil)
			}

			rec := httptest.NewRecorder()
			var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer func() {
					assert.NoError(t, r.Body.Close())
				}()

				checkBody := func(readCloser io.ReadCloser) {
					bodyBytes, err := ioutil.ReadAll(readCloser)
					assert.NoError(t, err)
					if test.method == http.MethodPost {
						assert.Equal(t, bodyString, string(bodyBytes))
					} else {
						assert.Equal(t, "", string(bodyBytes))
					}
				}

				if test.readTwice {
					checkBody(r.Body)
				}

				httpReq, ok := GetRequest(r.Context())
				if test.withMiddleware && (test.config.Skipper == nil || !test.config.Skipper(r)) {
					assert.True(t, ok)
					assert.NotNil(t, httpReq)
					checkBody(httpReq.Body)
				} else {
					assert.False(t, ok)
					assert.Nil(t, httpReq)
				}
			})
			if test.withMiddleware {
				m := ReserveRequest(test.config)
				h = m(h)
			}
			h.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}

func TestReserveResponseWriter(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		withMiddleware bool
		config         ReserveResponseWriterConfig
	}{{
		name:           "middleware:false",
		method:         http.MethodGet,
		withMiddleware: false,
	}, {
		name:           "middleware:true",
		method:         http.MethodGet,
		withMiddleware: true,
	}, {
		name:           "middleware:true:skipper",
		method:         http.MethodGet,
		withMiddleware: true,
		config: ReserveResponseWriterConfig{
			Skipper: func(*http.Request) bool {
				return true
			},
		},
	}, {
		name:           "middleware:true:readTwice",
		method:         http.MethodGet,
		withMiddleware: true,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bodyString := "test body"
			req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)

			rec := httptest.NewRecorder()
			var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer func() {
					assert.NoError(t, r.Body.Close())
				}()

				httpResp, ok := GetResponseWriter(r.Context())
				if test.withMiddleware && (test.config.Skipper == nil || !test.config.Skipper(r)) {
					assert.True(t, ok)
					assert.NotNil(t, httpResp)
					httpResp.Write([]byte(bodyString))
				} else {
					assert.False(t, ok)
					assert.Nil(t, httpResp)
				}
			})
			if test.withMiddleware {
				m := ReserveResponseWriter(test.config)
				h = m(h)
			}
			h.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
			if test.withMiddleware && (test.config.Skipper == nil || !test.config.Skipper(req)) {
				assert.Equal(t, bodyString, rec.Body.String())
			} else {
				assert.Equal(t, "", rec.Body.String())
			}
		})
	}
}
