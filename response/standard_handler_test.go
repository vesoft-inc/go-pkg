package response

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
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

func TestStandardHandler(t *testing.T) {
	type testStruct struct {
		F int
	}
	var nilStruct *testStruct
	tests := []struct {
		name            string
		params          StandardHandlerParams
		r               *http.Request
		data            interface{}
		unsupportedData bool
		err             error
		expectedStatus  int
		expectedBody    interface{}
	}{{
		name:           "data:no",
		params:         StandardHandlerParams{},
		r:              httptest.NewRequest("GET", "http://localhost", nil),
		expectedStatus: 200,
		expectedBody: map[string]interface{}{
			"code":    0,
			"message": "Success",
		},
	}, {
		name:           "data:no:error:normal",
		params:         StandardHandlerParams{},
		r:              httptest.NewRequest("GET", "http://localhost", nil),
		err:            errors.New("testError"),
		expectedStatus: 500,
		expectedBody: map[string]interface{}{
			"code":    50000000,
			"message": "ErrInternalServer",
		},
	}, {
		name:           "data:no:error:errorx",
		params:         StandardHandlerParams{},
		r:              httptest.NewRequest("GET", "http://localhost", nil),
		err:            errorx.WithCode(errorx.NewErrCode(403, 1, 2, "testError"), nil),
		expectedStatus: 403,
		expectedBody: map[string]interface{}{
			"code":    40301002,
			"message": "testError",
		},
	}, {
		name: "data:no:error:GetErrCode",
		params: StandardHandlerParams{
			GetErrCode: func(err error) *errorx.ErrCode {
				return errorx.NewErrCode(403, 1, 2, "testError")
			},
		},
		r:              httptest.NewRequest("GET", "http://localhost", nil),
		err:            errors.New("testError0"),
		expectedStatus: 403,
		expectedBody: map[string]interface{}{
			"code":    40301002,
			"message": "testError",
		},
	}, {
		name: "data:no:DebugInfo",
		params: StandardHandlerParams{
			DebugInfo: true,
		},
		r:              httptest.NewRequest("GET", "http://localhost", nil),
		err:            errors.New("testError"),
		expectedStatus: 500,
		expectedBody: map[string]interface{}{
			"code":      50000000,
			"message":   "ErrInternalServer",
			"debugInfo": "50000000(ErrInternalServer):testError\n",
		},
	}, {
		name: "data:no:CheckBodyType:none",
		params: StandardHandlerParams{
			CheckBodyType: func(r *http.Request) StandardHandlerBodyType { return StandardHandlerBodyNone },
		},
		r:              httptest.NewRequest("GET", "http://localhost", nil),
		expectedStatus: 200,
		expectedBody:   nil,
	}, {
		name: "data:no:CheckBodyType:none:error",
		params: StandardHandlerParams{
			CheckBodyType: func(r *http.Request) StandardHandlerBodyType { return StandardHandlerBodyNone },
		},
		r:              httptest.NewRequest("GET", "http://localhost", nil),
		err:            errors.New("testError"),
		expectedStatus: 500,
		expectedBody:   nil,
	}, {
		name: "data:no:CheckBodyType:json",
		params: StandardHandlerParams{
			CheckBodyType: func(r *http.Request) StandardHandlerBodyType { return StandardHandlerBodyJson },
			Errorf:        func(format string, a ...interface{}) {},
		},
		r:              httptest.NewRequest("GET", "http://localhost", nil),
		err:            errors.New("testError"),
		expectedStatus: 500,
		expectedBody: map[string]interface{}{
			"code":    50000000,
			"message": "ErrInternalServer",
		},
	}, {
		name:           "data:yes",
		params:         StandardHandlerParams{},
		r:              httptest.NewRequest("GET", "http://localhost", nil),
		data:           "data",
		expectedStatus: 200,
		expectedBody: map[string]interface{}{
			"code":    0,
			"message": "Success",
			"data":    "data",
		},
	}, {
		name:           "data:yes:nil:struct",
		params:         StandardHandlerParams{},
		r:              httptest.NewRequest("GET", "http://localhost", nil),
		data:           nilStruct,
		expectedStatus: 200,
		expectedBody: map[string]interface{}{
			"code":    0,
			"message": "Success",
		},
	}, {
		name:           "data:yes:any0:nil",
		params:         StandardHandlerParams{},
		r:              httptest.NewRequest("GET", "http://localhost", nil),
		data:           StandardHandlerDataFieldAny(nil),
		expectedStatus: 200,
		expectedBody: map[string]interface{}{
			"code":    0,
			"message": "Success",
		},
	}, {
		name:           "data:yes:any0:str",
		params:         StandardHandlerParams{},
		r:              httptest.NewRequest("GET", "http://localhost", nil),
		data:           StandardHandlerDataFieldAny("data"),
		expectedStatus: 200,
		expectedBody: map[string]interface{}{
			"code":    0,
			"message": "Success",
			"data":    "data",
		},
	}, {
		name:           "data:yes:any1:nil",
		params:         StandardHandlerParams{},
		r:              httptest.NewRequest("GET", "http://localhost", nil),
		data:           struct{ D interface{} }{D: StandardHandlerDataFieldAny(nil)},
		expectedStatus: 200,
		expectedBody: map[string]interface{}{
			"code":    0,
			"message": "Success",
		},
	}, {
		name:           "data:yes:any1:str",
		params:         StandardHandlerParams{},
		r:              httptest.NewRequest("GET", "http://localhost", nil),
		data:           struct{ D interface{} }{D: StandardHandlerDataFieldAny("data")},
		expectedStatus: 200,
		expectedBody: map[string]interface{}{
			"code":    0,
			"message": "Success",
			"data":    "data",
		},
	}, {
		name:           "data:yes:any1:struct",
		params:         StandardHandlerParams{},
		r:              httptest.NewRequest("GET", "http://localhost", nil),
		data:           struct{ D interface{} }{D: StandardHandlerDataFieldAny("data")},
		expectedStatus: 200,
		expectedBody: map[string]interface{}{
			"code":    0,
			"message": "Success",
			"data":    "data",
		},
	}, {
		name:           "data:yes:any1:struct:pointer",
		params:         StandardHandlerParams{},
		r:              httptest.NewRequest("GET", "http://localhost", nil),
		data:           &struct{ D interface{} }{D: StandardHandlerDataFieldAny("data")},
		expectedStatus: 200,
		expectedBody: map[string]interface{}{
			"code":    0,
			"message": "Success",
			"data":    "data",
		},
	}, {
		name:           "data:yes:any1:other",
		params:         StandardHandlerParams{},
		r:              httptest.NewRequest("GET", "http://localhost", nil),
		data:           struct{ D interface{} }{D: "data"},
		expectedStatus: 200,
		expectedBody: map[string]interface{}{
			"code":    0,
			"message": "Success",
			"data":    map[string]interface{}{"D": "data"},
		},
	}, {
		name:           "data:yes:any1:struct:2",
		params:         StandardHandlerParams{},
		r:              httptest.NewRequest("GET", "http://localhost", nil),
		data:           struct{ D, D2 interface{} }{D: StandardHandlerDataFieldAny("data")},
		expectedStatus: 200,
		expectedBody: map[string]interface{}{
			"code":    0,
			"message": "Success",
			"data":    map[string]interface{}{"D": map[string]interface{}{}, "D2": interface{}(nil)},
		},
	}, {
		name:           "data:yes:any1:struct:unexported",
		params:         StandardHandlerParams{},
		r:              httptest.NewRequest("GET", "http://localhost", nil),
		data:           struct{ d interface{} }{d: StandardHandlerDataFieldAny("data")},
		expectedStatus: 200,
		expectedBody: map[string]interface{}{
			"code":    0,
			"message": "Success",
			"data":    map[string]interface{}{},
		},
	}, {
		name:           "data:yes:error",
		params:         StandardHandlerParams{},
		r:              httptest.NewRequest("GET", "http://localhost", nil),
		data:           "data",
		err:            errors.New("testError"),
		expectedStatus: 500,
		expectedBody: map[string]interface{}{
			"code":    50000000,
			"message": "ErrInternalServer",
		},
	}, {
		name: "data:yes:CheckBodyType:none",
		params: StandardHandlerParams{
			CheckBodyType: func(r *http.Request) StandardHandlerBodyType { return StandardHandlerBodyNone },
		},
		r:              httptest.NewRequest("GET", "http://localhost", nil),
		data:           "data",
		expectedStatus: 200,
		expectedBody:   nil,
	}, {
		name: "data:yes:CheckBodyType:json",
		params: StandardHandlerParams{
			CheckBodyType: func(r *http.Request) StandardHandlerBodyType { return StandardHandlerBodyJson },
		},
		r:              httptest.NewRequest("GET", "http://localhost", nil),
		data:           "data",
		expectedStatus: 200,
		expectedBody: map[string]interface{}{
			"code":    0,
			"message": "Success",
			"data":    "data",
		},
	}, {
		name:            "data:unsupported:type",
		params:          StandardHandlerParams{},
		r:               httptest.NewRequest("GET", "http://localhost", nil),
		data:            complex(0, 0),
		unsupportedData: true,
		expectedStatus:  200,
		expectedBody: map[string]interface{}{
			"code":    0,
			"message": "Success",
			"data":    complex(0, 0),
		},
	}, {
		name: "data:yes:r:nil",
		params: StandardHandlerParams{
			CheckBodyType: func(r *http.Request) StandardHandlerBodyType { return StandardHandlerBodyJson },
		},
		data:           "data",
		expectedStatus: 200,
		expectedBody:   nil,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			checkFunc := func(expectedStatus, httpStatus int, expectedBody, body interface{}) {
				assert.Equal(t, expectedStatus, httpStatus)
				if test.err != nil && test.params.DebugInfo {
					assert.Contains(t, body.(map[string]interface{})["debugInfo"], expectedBody.(map[string]interface{})["debugInfo"])
					assert.Equal(t, map[string]interface{}{
						"code":    expectedBody.(map[string]interface{})["code"],
						"message": expectedBody.(map[string]interface{})["message"],
					}, map[string]interface{}{
						"code":    body.(map[string]interface{})["code"],
						"message": body.(map[string]interface{})["message"],
					})
				} else {
					assert.Equal(t, expectedBody, body)
				}
			}
			checkRespFunc := func(expectedStatus, httpStatus int, expectedBody interface{}, bodyBytes []byte) {
				var body interface{}
				if len(bodyBytes) > 0 {
					bodyStruct := struct {
						Code      int
						Message   string
						Data      interface{}
						DebugInfo string
					}{}
					err := json.Unmarshal(bodyBytes, &bodyStruct)
					assert.NoError(t, err)
					cmpBody := map[string]interface{}{
						"code":    bodyStruct.Code,
						"message": bodyStruct.Message,
					}
					if bodyStruct.Data != nil {
						cmpBody["data"] = bodyStruct.Data
					}
					if bodyStruct.DebugInfo != "" {
						cmpBody["debugInfo"] = bodyStruct.DebugInfo
					}
					body = cmpBody
				}
				checkFunc(expectedStatus, httpStatus, expectedBody, body)
			}

			h := NewStandardHandler(test.params)
			{
				httpStatus, body := h.GetStatusBody(test.r, test.data, test.err)
				if body != nil {
					if data := body.(map[string]interface{})["data"]; data != nil && !test.unsupportedData {
						typeOfData := reflect.TypeOf(data)
						if typeOfData.Kind() == reflect.Ptr {
							typeOfData = typeOfData.Elem()
						}
						if typeOfData.Kind() == reflect.Struct {
							bs, err := json.Marshal(data)
							assert.NoError(t, err)
							var m map[string]interface{}
							err = json.Unmarshal(bs, &m)
							assert.NoError(t, err)
							body.(map[string]interface{})["data"] = m
						}
					}
				}
				checkFunc(test.expectedStatus, httpStatus, test.expectedBody, body)
			}
			{
				rec := httptest.NewRecorder()
				h.Handle(rec, test.r, test.data, test.err)
				if test.unsupportedData {
					assert.Equal(t, 500, rec.Code)
					assert.Equal(t, "", rec.Body.String())
				} else {
					checkRespFunc(test.expectedStatus, rec.Code, test.expectedBody, rec.Body.Bytes())
				}
			}
			{
				rec := newTestRecorder(func(buf []byte) (int, error) {
					return 0, errors.New("testError")
				})
				h.Handle(rec, test.r, test.data, test.err)
				if test.unsupportedData {
					assert.Equal(t, 500, rec.Code)
					assert.Equal(t, "", rec.Body.String())
				} else {
					assert.Equal(t, test.expectedStatus, rec.Code)
					assert.Equal(t, "", rec.Body.String())
				}
			}

			{
				rec := newTestRecorder(func(buf []byte) (int, error) {
					l := len(buf)
					if l == 0 {
						return 0, nil
					}
					return l - 1, nil
				})
				h.Handle(rec, test.r, test.data, test.err)
				if test.unsupportedData {
					assert.Equal(t, 500, rec.Code)
					assert.Equal(t, "", rec.Body.String())
				} else if test.expectedBody == nil {
					checkRespFunc(test.expectedStatus, rec.Code, test.expectedBody, rec.Body.Bytes())
				} else {
					checkRespFunc(test.expectedStatus, rec.Code, test.expectedBody, append(rec.Body.Bytes(), byte('}')))
				}
			}
		})
	}
}
