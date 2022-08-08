package httpclient

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		addr     string
		expected string
	}{{
		name:     "empty",
		addr:     "",
		expected: "",
	}, {
		name:     "localhost",
		addr:     "localhost",
		expected: "http://localhost",
	}, {
		name:     "localhost:8080",
		addr:     "localhost:8080",
		expected: "http://localhost:8080",
	}, {
		name:     "localhost:8080/",
		addr:     "localhost:8080/",
		expected: "http://localhost:8080/",
	}, {
		name:     "http://localhost:8080",
		addr:     "http://localhost:8080",
		expected: "http://localhost:8080",
	}, {
		name:     "https://localhost:8080",
		addr:     "https://localhost:8080",
		expected: "https://localhost:8080",
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := NewClient(test.addr)
			cli, ok := c.(*defaultClient)
			assert.True(t, ok)
			assert.Equal(t, test.expected, cli.Addr)
			assert.Equal(t, resty.New().SetBaseURL(test.expected).HostURL, cli.client.HostURL)
		})
	}
}

func TestWithBeforeRequestHook(t *testing.T) {
	tests := []struct {
		name      string
		updateFns []func(*resty.Request)
		checkFn   func(t *testing.T, r *resty.Request)
	}{{
		name: "set 1 field",
		updateFns: []func(*resty.Request){
			func(r *resty.Request) {
				r.SetBody(1)
			},
		},
		checkFn: func(t *testing.T, r *resty.Request) {
			assert.Equal(t, 1, r.Body)
		},
	}, {
		name: "set 2 field",
		updateFns: []func(*resty.Request){
			func(r *resty.Request) {
				r.SetBody(1)
			}, func(r *resty.Request) {
				r.SetAuthToken("token")
			},
		},
		checkFn: func(t *testing.T, r *resty.Request) {
			assert.Equal(t, 1, r.Body)
			assert.Equal(t, "token", r.Token)
		},
	}, {
		name: "set 2 field with rewrite",
		updateFns: []func(*resty.Request){
			func(r *resty.Request) {
				r.SetBody(1)
				r.SetAuthToken("token")
			}, func(r *resty.Request) {
				r.SetBody("body")
			},
		},
		checkFn: func(t *testing.T, r *resty.Request) {
			assert.Equal(t, "body", r.Body)
			assert.Equal(t, "token", r.Token)
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var opts []RequestOption
			for _, fn := range test.updateFns {
				opts = append(opts, WithBeforeRequestHook(fn))
			}

			o := &requestOptions{}
			for _, opt := range opts {
				opt(o)
			}
			r := new(resty.Request)
			if o.beforeRequestHook != nil {
				o.beforeRequestHook(r)
			}
			test.checkFn(t, r)
		})
	}
}

func TestWithBeforeRequestHookSerial(t *testing.T) {
	tests := []struct {
		name     string
		ns       []int
		expected []int
	}{{
		name:     "nil",
		ns:       nil,
		expected: nil,
	}, {
		name:     "empty",
		ns:       []int{},
		expected: nil,
	}, {
		name:     "1, 2, 3, 4, 5",
		ns:       []int{1, 2, 3, 4, 5},
		expected: []int{1, 2, 3, 4, 5},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var actual []int
			var opts []RequestOption

			genOptionFn := func(n int) func(*resty.Request) {
				return func(*resty.Request) {
					actual = append(actual, n)
				}
			}

			for _, n := range test.ns {
				opts = append(opts, WithBeforeRequestHook(genOptionFn(n)))
			}

			o := &requestOptions{}
			for _, opt := range opts {
				opt(o)
			}
			if o.beforeRequestHook != nil {
				o.beforeRequestHook(nil)
			}
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestWithBody(t *testing.T) {
	tests := []struct {
		name     string
		bodyList []interface{}
		expected interface{}
	}{{
		name:     "nil",
		bodyList: nil,
		expected: nil,
	}, {
		name:     "a",
		bodyList: []interface{}{"a"},
		expected: "a",
	}, {
		name:     "a,nil",
		bodyList: []interface{}{"a", nil},
		expected: nil,
	}, {
		name:     "a,nil,1",
		bodyList: []interface{}{"a", nil, 1},
		expected: 1,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var opts []RequestOption

			for _, body := range test.bodyList {
				opts = append(opts, WithBody(body))
			}

			o := &requestOptions{}
			for _, opt := range opts {
				opt(o)
			}
			r := new(resty.Request)
			if o.beforeRequestHook != nil {
				o.beforeRequestHook(r)
			}
			assert.Equal(t, test.expected, r.Body)
		})
	}
}

func TestClient(t *testing.T) {
	reqBody := []byte("testReqBody")
	respBody := []byte("testRespBody")
	reqFuncMap := map[string]func(Client) (*resty.Response, error){
		resty.MethodGet: func(c Client) (*resty.Response, error) {
			return c.Get("")
		},
		resty.MethodPost: func(c Client) (*resty.Response, error) {
			return c.Post("", reqBody)
		},
		resty.MethodPut: func(c Client) (*resty.Response, error) {
			return c.Put("", reqBody)
		},
		resty.MethodPatch: func(c Client) (*resty.Response, error) {
			return c.Patch("", reqBody)
		},
		resty.MethodDelete: func(c Client) (*resty.Response, error) {
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
						body, err := io.ReadAll(r.Body)
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

				checkFunc := func(resp *resty.Response, err error) {
					if statusCode == 301 {
						ast.Error(err)
						ast.Contains(err.Error(), "301 response missing Location header")
					} else {
						ast.NoError(err)
						ast.Equal(statusCode, resp.StatusCode())
						ast.Equal(respBody, resp.Body())
					}
				}

				c := NewClient(testServer.URL,
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
						checkFunc(resp, err)
					}),
					WithAfterRequestHook(func(_ *resty.Request, resp *resty.Response, err error) {
						checkFunc(resp, err)
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
