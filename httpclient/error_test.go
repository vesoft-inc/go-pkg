package httpclient

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	ast := assert.New(t)

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == resty.MethodGet {
			if r.URL.Path == "/get/ok" {
				w.WriteHeader(http.StatusOK)
			} else if r.URL.Path == "/get/error" {
				w.WriteHeader(http.StatusNotFound)
				_, err := w.Write([]byte("body"))
				ast.NoError(err)
			}
		}
	}))

	resp, err := resty.New().R().Get(testServer.URL + "/get/ok")
	ast.NoError(err)
	err = NewResponseErrorNotSuccess(resp)
	ast.NoError(err)

	resp, err = resty.New().R().Get(testServer.URL + "/get/error")
	ast.NoError(err)
	err = NewResponseErrorNotSuccess(resp)
	if ast.Error(err) {
		ast.True(IsResponseError(err))
		ast.True(IsResponseError(err, http.StatusNotFound))
		ast.False(IsResponseError(err, http.StatusForbidden))
		ast.False(IsResponseError(err, http.StatusNotFound, http.StatusForbidden))
		respErr, ok := AsResponseError(err)
		ast.True(ok)
		ast.Equal(http.StatusNotFound, respErr.GetResponse().StatusCode())
		ast.True(respErr.IsStatusCode(http.StatusNotFound))

		ast.Equal("404 Not Found", err.Error())
		fields := strings.Split(fmt.Sprintf("%+v", err), "\n")
		if ast.True(len(fields) > 2, "%v", fields) {
			ast.Equal("404 Not Found", fields[0], "%+v", err)
			ast.Contains(fields[1], "httpclient.NewResponseError", "%+v", err)
		}
		ast.Equal("\"404 Not Found\"", fmt.Sprintf("%q", err))
	}

	err = NewResponseError(resp, fmt.Errorf("CauseError"))
	if ast.Error(err) {
		ast.True(IsResponseError(err))
		ast.True(IsResponseError(err, http.StatusNotFound))
		ast.False(IsResponseError(err, http.StatusForbidden))
		ast.False(IsResponseError(err, http.StatusNotFound, http.StatusForbidden))
		respErr, ok := AsResponseError(err)
		ast.True(ok)
		ast.Equal(http.StatusNotFound, respErr.GetResponse().StatusCode())
		ast.True(respErr.IsStatusCode(http.StatusNotFound))

		ast.Equal("404 Not Found: CauseError", err.Error())
		fields := strings.Split(fmt.Sprintf("%+v", err), "\n")
		if ast.True(len(fields) > 3, "%v", fields) {
			ast.Equal("CauseError", fields[0], "%+v", err)
			ast.Equal("404 Not Found: CauseError", fields[1], "%+v", err)
			ast.Contains(fields[2], "httpclient.NewResponseError", "%+v", err)
		}
		ast.Equal("\"404 Not Found: CauseError\"", fmt.Sprintf("%q", err))

		err1 := errors.Unwrap(err)
		ast.Equal("404 Not Found: CauseError", err1.Error(), "%+v", err1)
		err2 := errors.Unwrap(err1)
		ast.Equal("CauseError", err2.Error(), "%+v", err2)
	}

	err = &responseError{resp: resp}
	ast.Equal("\"404 Not Found\"", fmt.Sprintf("%q", err))

	ast.False(IsResponseError(errors.New("err")))
}
