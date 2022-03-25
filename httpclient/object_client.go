package httpclient

import (
	"encoding/json"

	"github.com/go-resty/resty/v2"
)

type (
	ObjectClient interface {
		Get(urlPath string, responseObj interface{}, opts ...RequestOption) error
		Post(urlPath string, body, responseObj interface{}, opts ...RequestOption) error
		Put(urlPath string, body, responseObj interface{}, opts ...RequestOption) error
		Patch(urlPath string, body, responseObj interface{}, opts ...RequestOption) error
		Delete(urlPath string, body, responseObj interface{}, opts ...RequestOption) error
		Execute(method, urlPath string, body, responseObj interface{}, opts ...RequestOption) error
	}

	defaultObjectClient struct {
		client Client
	}
)

var _ ObjectClient = (*defaultObjectClient)(nil)

func NewObjectClient(addr string, opts ...RequestOption) ObjectClient {
	return NewObjectClientRaw(NewClient(addr, opts...))
}

func NewObjectClientRaw(cli Client) ObjectClient {
	return &defaultObjectClient{
		client: cli,
	}
}

func (c *defaultObjectClient) Get(urlPath string, responseObj interface{}, opts ...RequestOption) error {
	resp, err := c.client.Get(urlPath, opts...)
	return c.convertResponse(responseObj, resp, err)
}

func (c *defaultObjectClient) Post(urlPath string, body, responseObj interface{}, opts ...RequestOption) error {
	resp, err := c.client.Post(urlPath, body, opts...)
	return c.convertResponse(responseObj, resp, err)
}

func (c *defaultObjectClient) Put(urlPath string, body, responseObj interface{}, opts ...RequestOption) error {
	resp, err := c.client.Put(urlPath, body, opts...)
	return c.convertResponse(responseObj, resp, err)
}

func (c *defaultObjectClient) Patch(urlPath string, body, responseObj interface{}, opts ...RequestOption) error {
	resp, err := c.client.Patch(urlPath, body, opts...)
	return c.convertResponse(responseObj, resp, err)
}

func (c *defaultObjectClient) Delete(urlPath string, body, responseObj interface{}, opts ...RequestOption) error {
	resp, err := c.client.Delete(urlPath, body, opts...)
	return c.convertResponse(responseObj, resp, err)
}

func (c *defaultObjectClient) Execute(method, urlPath string, body, responseObj interface{}, opts ...RequestOption) error {
	resp, err := c.client.Execute(method, urlPath, body, opts...)
	return c.convertResponse(responseObj, resp, err)
}

func (*defaultObjectClient) convertResponse(responseObj interface{}, resp *resty.Response, err error) error {
	if err != nil {
		return err
	}

	if err = NewResponseErrorNotSuccess(resp); err != nil {
		return err
	}

	if responseObj == nil {
		return nil
	}

	// Only support json response now
	if err := json.Unmarshal(resp.Body(), responseObj); err != nil {
		return NewResponseError(resp, err)
	}
	return nil
}
