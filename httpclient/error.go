package httpclient

import (
	"fmt"
	"io"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
)

var _ ResponseError = (*responseError)(nil)

type (
	ResponseError interface {
		error
		GetResponse() *resty.Response
		IsStatusCode(statusCode int) bool
	}

	responseError struct {
		error
		resp *resty.Response
	}
)

func NewResponseError(resp *resty.Response, err error) error {
	return errors.WithStack(&responseError{resp: resp, error: err})
}

func NewResponseErrorNotSuccess(resp *resty.Response) error {
	if resp == nil || resp.IsSuccess() {
		return nil
	}
	return NewResponseError(resp, nil)
}

func AsResponseError(err error) (ResponseError, bool) {
	if e := new(responseError); errors.As(err, &e) {
		return e, true
	}
	return nil, false
}

func IsResponseError(err error, statusCode ...int) bool {
	hce, ok := AsResponseError(err)
	if !ok {
		return false
	}

	switch len(statusCode) {
	case 0:
		return true
	case 1:
		return hce.IsStatusCode(statusCode[0])
	default:
		return false
	}
}

func (e *responseError) GetResponse() *resty.Response {
	return e.resp
}

func (e *responseError) IsStatusCode(statusCode int) bool {
	return e.GetResponse().StatusCode() == statusCode
}

func (e *responseError) Error() string {
	if e.error == nil {
		return e.resp.Status()
	}
	return fmt.Sprintf("%s: %s", e.resp.Status(), e.error.Error())
}

func (e *responseError) Cause() error { return e.error }

func (e *responseError) Unwrap() error { return e.error }

func (e *responseError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') && e.Cause() != nil {
			_, _ = fmt.Fprintf(s, "%+v\n", e.Cause())
			_, _ = io.WriteString(s, e.Error())
			return
		}
		fallthrough
	case 's':
		_, _ = io.WriteString(s, e.Error())
	case 'q':
		_, _ = fmt.Fprintf(s, "%q", e.Error())
	}
}
