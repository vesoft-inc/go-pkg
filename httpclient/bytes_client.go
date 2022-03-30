package httpclient

import "github.com/go-resty/resty/v2"

type (
	BytesClient interface {
		Get(urlPath string, opts ...RequestOption) ([]byte, error)
		Post(urlPath string, body interface{}, opts ...RequestOption) ([]byte, error)
		Put(urlPath string, body interface{}, opts ...RequestOption) ([]byte, error)
		Patch(urlPath string, body interface{}, opts ...RequestOption) ([]byte, error)
		Delete(urlPath string, body interface{}, opts ...RequestOption) ([]byte, error)
		Execute(method, urlPath string, body interface{}, opts ...RequestOption) ([]byte, error)
	}

	defaultBytesClient struct {
		client Client
	}
)

var _ BytesClient = (*defaultBytesClient)(nil)

func NewBytesClient(addr string, opts ...RequestOption) BytesClient {
	return NewBytesClientRaw(NewClient(addr, opts...))
}

func NewBytesClientRaw(cli Client) BytesClient {
	return &defaultBytesClient{
		client: cli,
	}
}

func (c *defaultBytesClient) Get(urlPath string, opts ...RequestOption) ([]byte, error) {
	return c.convertResponse(c.client.Get(urlPath, opts...))
}

func (c *defaultBytesClient) Post(urlPath string, body interface{}, opts ...RequestOption) ([]byte, error) {
	return c.convertResponse(c.client.Post(urlPath, body, opts...))
}

func (c *defaultBytesClient) Put(urlPath string, body interface{}, opts ...RequestOption) ([]byte, error) {
	return c.convertResponse(c.client.Put(urlPath, body, opts...))
}

func (c *defaultBytesClient) Patch(urlPath string, body interface{}, opts ...RequestOption) ([]byte, error) {
	return c.convertResponse(c.client.Patch(urlPath, body, opts...))
}

func (c *defaultBytesClient) Delete(urlPath string, body interface{}, opts ...RequestOption) ([]byte, error) {
	return c.convertResponse(c.client.Delete(urlPath, body, opts...))
}

func (c *defaultBytesClient) Execute(method, urlPath string, body interface{}, opts ...RequestOption) ([]byte, error) {
	return c.convertResponse(c.client.Execute(method, urlPath, body, opts...))
}

func (*defaultBytesClient) convertResponse(resp *resty.Response, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}

	if err = NewResponseErrorNotSuccess(resp); err != nil {
		return nil, err
	}

	return resp.Body(), err
}
