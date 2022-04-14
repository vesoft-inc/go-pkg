package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vesoft-inc/go-pkg/errorx"

	"github.com/stretchr/testify/assert"
)

type testRecorder struct {
	HeaderMap   http.Header
	Code        int
	Body        *bytes.Buffer
	write       func([]byte) (int, error)
	wroteHeader bool
}

func newTestRecorder(write func([]byte) (int, error)) *testRecorder {
	return &testRecorder{
		write:     write,
		HeaderMap: make(http.Header),
		Code:      200,
		Body:      new(bytes.Buffer),
	}
}

func (rec *testRecorder) Header() http.Header {
	return rec.HeaderMap
}

func (rec *testRecorder) Write(buf []byte) (int, error) {
	n, err := rec.write(buf)
	rec.Body.Write(buf[:n])
	return n, err
}

func (rec *testRecorder) WriteHeader(statusCode int) {
	if rec.wroteHeader {
		return
	}
	rec.wroteHeader = true
	rec.Code = statusCode
}

func TestGetStandardResponseHandler(t *testing.T) {
	handler := GetStandardResponseHandler(context.Background())
	assert.Nil(t, handler)
}

func TestStandardResponseHandler(t *testing.T) {
	tests := []struct {
		name           string
		params         StandardResponseHandlerParams
		r              *http.Request
		data           interface{}
		err            error
		expectedStatus int
		expectedBody   interface{}
	}{{
		name:           "data:no",
		params:         StandardResponseHandlerParams{},
		expectedStatus: 200,
		expectedBody: map[string]interface{}{
			"code":    0,
			"message": "Success",
		},
	}, {
		name:           "data:no:error:normal",
		params:         StandardResponseHandlerParams{},
		err:            errors.New("testError"),
		expectedStatus: 500,
		expectedBody: map[string]interface{}{
			"code":    50000000,
			"message": "ErrInternalServer",
		},
	}, {
		name:           "data:no:error:errorx",
		params:         StandardResponseHandlerParams{},
		err:            errorx.WithCode(errorx.NewErrCode(403, 1, 2, "testError"), nil),
		expectedStatus: 403,
		expectedBody: map[string]interface{}{
			"code":    40301002,
			"message": "testError",
		},
	}, {
		name: "data:no:error:GetErrCode",
		params: StandardResponseHandlerParams{
			GetErrCode: func(err error) *errorx.ErrCode {
				return errorx.NewErrCode(403, 1, 2, "testError")
			},
		},
		err:            errors.New("testError0"),
		expectedStatus: 403,
		expectedBody: map[string]interface{}{
			"code":    40301002,
			"message": "testError",
		},
	}, {
		name: "data:no:DebugInfo",
		params: StandardResponseHandlerParams{
			DebugInfo: true,
		},
		err:            errors.New("testError"),
		expectedStatus: 500,
		expectedBody: map[string]interface{}{
			"code":      50000000,
			"message":   "ErrInternalServer",
			"debugInfo": "testError\n50000000: ErrInternalServer",
		},
	}, {
		name: "data:no:CheckBodyType:none",
		params: StandardResponseHandlerParams{
			CheckBodyType: func(r *http.Request) StandardResponseBodyType { return StandardResponseBodyNone },
		},
		err:            errors.New("testError"),
		expectedStatus: 500,
		expectedBody:   nil,
	}, {
		name: "data:no:CheckBodyType:json",
		params: StandardResponseHandlerParams{
			CheckBodyType: func(r *http.Request) StandardResponseBodyType { return StandardResponseBodyJson },
		},
		err:            errors.New("testError"),
		expectedStatus: 500,
		expectedBody: map[string]interface{}{
			"code":    50000000,
			"message": "ErrInternalServer",
		},
	}, {
		name:           "data:yes",
		params:         StandardResponseHandlerParams{},
		data:           "data",
		expectedStatus: 200,
		expectedBody: map[string]interface{}{
			"code":    0,
			"message": "Success",
			"data":    "data",
		},
	}, {
		name:           "data:yes:error",
		params:         StandardResponseHandlerParams{},
		data:           "data",
		err:            errors.New("testError"),
		expectedStatus: 500,
		expectedBody: map[string]interface{}{
			"code":    50000000,
			"message": "ErrInternalServer",
		},
	}, {
		name: "data:yes:CheckBodyType:none",
		params: StandardResponseHandlerParams{
			CheckBodyType: func(r *http.Request) StandardResponseBodyType { return StandardResponseBodyNone },
		},
		data:           "data",
		expectedStatus: 200,
		expectedBody:   nil,
	}, {
		name: "data:yes:CheckBodyType:json",
		params: StandardResponseHandlerParams{
			CheckBodyType: func(r *http.Request) StandardResponseBodyType { return StandardResponseBodyJson },
		},
		data:           "data",
		expectedStatus: 200,
		expectedBody: map[string]interface{}{
			"code":    0,
			"message": "Success",
			"data":    "data",
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fn := func(httpStatus int, body interface{}) {
				assert.Equal(t, test.expectedStatus, httpStatus)
				if test.err != nil && test.params.DebugInfo {
					assert.Contains(t, body.(map[string]interface{})["debugInfo"], test.expectedBody.(map[string]interface{})["debugInfo"])
					assert.Equal(t, map[string]interface{}{
						"code":    test.expectedBody.(map[string]interface{})["code"],
						"message": test.expectedBody.(map[string]interface{})["message"],
					}, map[string]interface{}{
						"code":    body.(map[string]interface{})["code"],
						"message": body.(map[string]interface{})["message"],
					})
				} else {
					assert.Equal(t, test.expectedBody, body)
				}
			}

			h := NewStandardResponseHandler(test.params)
			{
				httpStatus, body := h.GetHandleBody(test.r, test.data, test.err)
				fn(httpStatus, body)
			}
			{
				rec := httptest.NewRecorder()
				h.Handle(rec, test.r, test.data, test.err)
				if rec.Body.String() == "" {
					fn(rec.Code, nil)
				} else {
					body := struct {
						Code      int
						Message   string
						Data      interface{}
						DebugInfo string
					}{}
					err := json.Unmarshal(rec.Body.Bytes(), &body)
					assert.NoError(t, err)
					cmpBody := map[string]interface{}{
						"code":    body.Code,
						"message": body.Message,
					}
					if body.Data != nil {
						cmpBody["data"] = body.Data
					}
					if body.DebugInfo != "" {
						cmpBody["debugInfo"] = body.DebugInfo
					}
					fn(rec.Code, cmpBody)
				}
			}
		})
	}
}

func TestStandardResponseMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		CheckBodyType  func(r *http.Request) StandardResponseBodyType
		data           interface{}
		err            error
		expectedStatus int
		expectedBody   string
	}{{
		name:           "CheckBodyType:json",
		CheckBodyType:  func(r *http.Request) StandardResponseBodyType { return StandardResponseBodyJson },
		data:           "data",
		expectedStatus: 200,
		expectedBody:   "{\"code\":0,\"data\":\"data\",\"message\":\"Success\"}",
	}, {
		name:           "CheckBodyType:none",
		CheckBodyType:  func(r *http.Request) StandardResponseBodyType { return StandardResponseBodyNone },
		data:           "data",
		err:            nil,
		expectedStatus: 200,
		expectedBody:   "",
	}, {
		name:           "error:normal",
		data:           "data",
		err:            errors.New("testError"),
		expectedStatus: 500,
		expectedBody:   "{\"code\":50000000,\"message\":\"ErrInternalServer\"}",
	}, {
		name:           "error:errorx",
		data:           "data",
		err:            errorx.WithCode(errorx.NewErrCode(403, 1, 2, "testError"), nil),
		expectedStatus: 403,
		expectedBody:   "{\"code\":40301002,\"message\":\"testError\"}",
	}, {
		name:           "data:unsupported:type",
		data:           make(chan struct{}),
		err:            nil,
		expectedStatus: 500,
		expectedBody:   "",
	}}

	for _, test := range tests {
		if test.name != "data:unsupported:type" {
			continue
		}
		t.Run(test.name, func(t *testing.T) {
			m := NewStandardResponse(StandardResponseParams{
				Handle: NewStandardResponseHandler(StandardResponseHandlerParams{
					CheckBodyType: test.CheckBodyType,
					Errorf:        func(format string, a ...interface{}) {},
				}).Handle,
			})

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				f := r.Context().Value(standardResponseCtxKey{})
				assert.NotNil(t, f)
				handle := GetStandardResponseHandler(r.Context())
				assert.NotNil(t, handle)
				handle(test.data, test.err)
			})

			h := m.Handler(nextHandler)
			req := httptest.NewRequest("GET", "http://localhost", nil)

			{
				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)
				assert.Equal(t, test.expectedStatus, rec.Code)
				assert.Equal(t, test.expectedBody, rec.Body.String())
			}

			{
				rec := newTestRecorder(func(buf []byte) (int, error) {
					return 0, errors.New("testError")
				})
				h.ServeHTTP(rec, req)
				assert.Equal(t, test.expectedStatus, rec.Code)
				assert.Equal(t, "", rec.Body.String())
			}

			{
				rec := newTestRecorder(func(buf []byte) (int, error) {
					l := len(buf)
					if l == 0 {
						return 0, nil
					}
					return l - 1, nil
				})
				h.ServeHTTP(rec, req)
				assert.Equal(t, test.expectedStatus, rec.Code)
				if test.expectedBody == "" {
					assert.Equal(t, test.expectedBody, rec.Body.String())
				} else {
					assert.Equal(t, test.expectedBody[:len(test.expectedBody)-1], rec.Body.String())
				}
			}
		})
	}
}
