package httpclient

import (
	"crypto/tls"
	"strings"

	"github.com/go-resty/resty/v2"
)

var _ Client = (*defaultClient)(nil)

type (
	Client interface {
		Get(urlPath string, opts ...RequestOption) (*resty.Response, error)
		Post(urlPath string, body interface{}, opts ...RequestOption) (*resty.Response, error)
		Put(urlPath string, body interface{}, opts ...RequestOption) (*resty.Response, error)
		Patch(urlPath string, body interface{}, opts ...RequestOption) (*resty.Response, error)
		Delete(urlPath string, body interface{}, opts ...RequestOption) (*resty.Response, error)
		Execute(method, urlPath string, body interface{}, opts ...RequestOption) (*resty.Response, error)
	}

	defaultClient struct {
		Addr        string
		client      *resty.Client
		initOptions *requestOptions
	}

	RequestOption  func(*requestOptions)
	requestOptions struct {
		newClientHook     func(*resty.Client)
		beforeRequestHook func(*resty.Request)
		afterRequestHook  func(*resty.Request, *resty.Response, error)
	}
)

func NewClient(addr string, opts ...RequestOption) Client {
	if addr != "" && !strings.HasPrefix(addr, "http") {
		addr = "http://" + addr
	}

	o := newRequestOptions(opts...)

	rawClient := resty.New().SetBaseURL(addr)
	if o.newClientHook != nil {
		o.newClientHook(rawClient)
	}

	return &defaultClient{
		Addr:        addr,
		client:      rawClient,
		initOptions: o,
	}
}

func WithTLSClientConfig(config *tls.Config) RequestOption {
	return func(o *requestOptions) {
		o.linkNewClientHook(func(c *resty.Client) {
			c.SetTLSClientConfig(config)
		})
	}
}

func WithBody(body interface{}) RequestOption {
	return func(o *requestOptions) {
		o.linkBeforeRequestHook(func(r *resty.Request) {
			r.SetBody(body)
		})
	}
}

func WithQueryParam(param, value string) RequestOption {
	return func(o *requestOptions) {
		o.linkBeforeRequestHook(func(r *resty.Request) {
			r.SetQueryParam(param, value)
		})
	}
}

func WithQueryParams(params map[string]string) RequestOption {
	return func(o *requestOptions) {
		o.linkBeforeRequestHook(func(r *resty.Request) {
			r.SetQueryParams(params)
		})
	}
}

func WithHeader(header, value string) RequestOption {
	return func(o *requestOptions) {
		o.linkBeforeRequestHook(func(r *resty.Request) {
			r.SetHeader(header, value)
		})
	}
}

func WithHeaders(headers map[string]string) RequestOption {
	return func(o *requestOptions) {
		o.linkBeforeRequestHook(func(r *resty.Request) {
			r.SetHeaders(headers)
		})
	}
}

func WithAuthToken(token string) RequestOption {
	return func(o *requestOptions) {
		o.linkBeforeRequestHook(func(r *resty.Request) {
			r.SetAuthToken(token)
		})
	}
}

func WithNewClientHook(fn func(*resty.Client)) RequestOption {
	return func(o *requestOptions) {
		o.linkNewClientHook(fn)
	}
}

func WithBeforeRequestHook(fn func(*resty.Request)) RequestOption {
	return func(o *requestOptions) {
		o.linkBeforeRequestHook(fn)
	}
}

func WithAfterRequestHook(fn func(*resty.Request, *resty.Response, error)) RequestOption {
	return func(o *requestOptions) {
		o.linkAfterRequestHook(fn)
	}
}

func (c *defaultClient) Get(urlPath string, opts ...RequestOption) (*resty.Response, error) {
	return c.Execute(resty.MethodGet, urlPath, nil, opts...)
}

func (c *defaultClient) Post(urlPath string, body interface{}, opts ...RequestOption) (*resty.Response, error) {
	return c.Execute(resty.MethodPost, urlPath, body, opts...)
}

func (c *defaultClient) Put(urlPath string, body interface{}, opts ...RequestOption) (*resty.Response, error) {
	return c.Execute(resty.MethodPut, urlPath, body, opts...)
}

func (c *defaultClient) Patch(urlPath string, body interface{}, opts ...RequestOption) (*resty.Response, error) {
	return c.Execute(resty.MethodPatch, urlPath, body, opts...)
}

func (c *defaultClient) Delete(urlPath string, body interface{}, opts ...RequestOption) (*resty.Response, error) {
	return c.Execute(resty.MethodDelete, urlPath, body, opts...)
}

func (c *defaultClient) Execute(method, urlPath string, body interface{}, opts ...RequestOption) (*resty.Response, error) {
	if body != nil {
		opts = append(opts, WithBody(body))
	}
	return c.doRequest(method, urlPath, opts...)
}

func (c *defaultClient) doRequest(method, urlPath string, opts ...RequestOption) (*resty.Response, error) {
	o := c.initOptions.WithOptions(opts...)

	r := c.client.R()
	if o.beforeRequestHook != nil {
		o.beforeRequestHook(r)
	}

	resp, err := r.Execute(method, urlPath)
	if o.afterRequestHook != nil {
		o.afterRequestHook(r, resp, err)
	}
	return resp, err
}

func newRequestOptions(opts ...RequestOption) *requestOptions {
	return defaultRequestOptions().WithOptions(opts...)
}

func (o *requestOptions) linkNewClientHook(fn func(*resty.Client)) {
	if o.newClientHook == nil {
		o.newClientHook = fn
		return
	}
	preHook := o.newClientHook
	o.newClientHook = func(c *resty.Client) {
		preHook(c)
		fn(c)
	}
}

func (o *requestOptions) linkBeforeRequestHook(fn func(*resty.Request)) {
	if o.beforeRequestHook == nil {
		o.beforeRequestHook = fn
		return
	}
	preHook := o.beforeRequestHook
	o.beforeRequestHook = func(r *resty.Request) {
		preHook(r)
		fn(r)
	}
}

func (o *requestOptions) linkAfterRequestHook(fn func(*resty.Request, *resty.Response, error)) {
	if o.afterRequestHook == nil {
		o.afterRequestHook = fn
		return
	}
	preHook := o.afterRequestHook
	o.afterRequestHook = func(r *resty.Request, resp *resty.Response, err error) {
		preHook(r, resp, err)
		fn(r, resp, err)
	}
}

func (o *requestOptions) WithOptions(opts ...RequestOption) *requestOptions {
	if len(opts) == 0 {
		return o
	}

	cpy := *o
	for _, opt := range opts {
		opt(&cpy)
	}
	return &cpy
}

func defaultRequestOptions() *requestOptions {
	return &requestOptions{}
}
