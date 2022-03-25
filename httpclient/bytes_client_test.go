package httpclient

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestBytesClient(t *testing.T) {
	reqBody := []byte("testReqBody")
	respBody := []byte("testRespBody")
	reqFuncMap := map[string]func(BytesClient) ([]byte, error){
		resty.MethodGet: func(c BytesClient) ([]byte, error) {
			return c.Get("")
		},
		resty.MethodPost: func(c BytesClient) ([]byte, error) {
			return c.Post("", reqBody)
		},
		resty.MethodPut: func(c BytesClient) ([]byte, error) {
			return c.Put("", reqBody)
		},
		resty.MethodPatch: func(c BytesClient) ([]byte, error) {
			return c.Patch("", reqBody)
		},
		resty.MethodDelete: func(c BytesClient) ([]byte, error) {
			return c.Delete("", reqBody)
		},
	}

	for _, statusCode := range []int{200, 301, 404, 500, 502} {
		for _, method := range []string{resty.MethodGet, resty.MethodPost, resty.MethodPut, resty.MethodPatch, resty.MethodDelete} {
			t.Run(fmt.Sprintf("%s:%d", method, statusCode), func(t *testing.T) {
				ast := assert.New(t)
				testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					ast.Equal(method, r.Method)
					if r.Method != resty.MethodGet {
						body, err := ioutil.ReadAll(r.Body)
						ast.NoError(err)
						ast.Equal(reqBody, body)
					}
					ast.Equal(url.Values{
						"q0": []string{"qv0"},
						"q1": []string{"qv1"},
						"q2": []string{"qv2"},
					}, r.URL.Query())
					ast.Equal([]string{"hv0", "hv1", "hv2"}, []string{r.Header.Get("h0"), r.Header.Get("h1"), r.Header.Get("h2")})
					ast.Equal("Bearer AuthToken", r.Header.Get("Authorization"))

					w.WriteHeader(statusCode)
					w.Write(respBody)
				}))

				defer testServer.Close()

				checkHookFunc := func(resp *resty.Response, err error) {
					if statusCode == 301 {
						ast.Error(err)
						ast.Contains(err.Error(), "301 response missing Location header")
					} else {
						ast.NoError(err)
						ast.Equal(statusCode, resp.StatusCode())
						ast.Equal(respBody, resp.Body())
					}
				}
				checkFunc := func(respBytes []byte, err error) {
					if statusCode == 301 {
						ast.Error(err)
						ast.Contains(err.Error(), "301 response missing Location header")
					} else if statusCode != 200 {
						ast.Nil(respBytes)
						respErr, ok := AsResponseError(err)
						ast.True(ok)
						resp := respErr.GetResponse()
						ast.NotNil(resp)
						ast.Equal(statusCode, resp.StatusCode())
						ast.Equal(respBody, resp.Body())
					} else {
						ast.NoError(err)
						ast.Equal(respBody, respBytes)
					}
				}

				c := NewBytesClient(testServer.URL,
					WithQueryParam("q0", "qv0"),
					WithQueryParams(map[string]string{
						"q1": "qv1",
						"q2": "qv2",
					}),
					WithHeader("h0", "hv0"),
					WithHeaders(map[string]string{
						"h1": "hv1",
						"h2": "hv2",
					}),
					WithAuthToken("AuthToken"),
					WithAfterRequestHook(func(_ *resty.Request, resp *resty.Response, err error) {
						checkHookFunc(resp, err)
					}),
					WithAfterRequestHook(func(_ *resty.Request, resp *resty.Response, err error) {
						checkHookFunc(resp, err)
					}),
				)
				f := reqFuncMap[method]
				if ast.NotNil(f) {
					resp, err := f(c)
					checkFunc(resp, err)
				}
				var executeBody interface{}
				if method != resty.MethodGet {
					executeBody = reqBody
				}
				resp, err := c.Execute(method, "", executeBody)
				checkFunc(resp, err)
			})
		}
	}
}
