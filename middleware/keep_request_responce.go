package middleware

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"sync"
)

type (
	ReserveRequestConfig struct {
		Skipper Skipper
	}

	ReserveResponseWriterConfig struct {
		Skipper Skipper
	}

	reservedRequestCtxKey        struct{}
	reservedResponseWriterCtxKey struct{}

	reservedRequest struct {
		r            *http.Request
		rawBody      io.ReadCloser
		reservedBody *reservedRequestBody
	}

	reservedResponseWriter struct {
		w http.ResponseWriter
	}

	reservedRequestBody struct {
		body      io.ReadCloser
		teeReader io.Reader
		buff      bytes.Buffer
		once      sync.Once
		isRead    bool
	}
)

func ReserveRequest(config ReserveRequestConfig) func(next http.Handler) http.Handler {
	if config.Skipper == nil {
		config.Skipper = DefaultSkipper
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.Skipper(r) {
				next.ServeHTTP(w, r)
				return
			}

			reservedReq := newReservedRequest(r)
			reservedReq.next(next, w)
		})
	}
}

func GetRequest(ctx context.Context) (*http.Request, bool) {
	v := ctx.Value(reservedRequestCtxKey{})
	if v == nil {
		return nil, false
	}
	return v.(*reservedRequest).getRequest(), true
}

func ReserveResponseWriter(config ReserveResponseWriterConfig) func(next http.Handler) http.Handler {
	if config.Skipper == nil {
		config.Skipper = DefaultSkipper
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.Skipper(r) {
				next.ServeHTTP(w, r)
				return
			}

			reservedResp := newReservedResponseWriter(w)
			reservedResp.next(next, r)
		})
	}
}

func GetResponseWriter(ctx context.Context) (http.ResponseWriter, bool) {
	v := ctx.Value(reservedResponseWriterCtxKey{})
	if v == nil {
		return nil, false
	}
	return v.(*reservedResponseWriter).getResponseWriter(), true
}

func newReservedRequest(r *http.Request) *reservedRequest {
	return &reservedRequest{
		r:            r,
		rawBody:      r.Body,
		reservedBody: newReservedRequestBody(r.Body),
	}
}

func newReservedResponseWriter(w http.ResponseWriter) *reservedResponseWriter {
	return &reservedResponseWriter{
		w: w,
	}
}

func newReservedRequestBody(body io.ReadCloser) *reservedRequestBody {
	rc := &reservedRequestBody{
		body: body,
	}
	rc.teeReader = io.TeeReader(body, &rc.buff)
	return rc
}

func (rr *reservedRequest) next(next http.Handler, w http.ResponseWriter) {
	r := rr.r
	r.Body = rr.reservedBody
	r = r.WithContext(context.WithValue(r.Context(), reservedRequestCtxKey{}, rr))
	next.ServeHTTP(w, r)
}

func (rr *reservedRequest) getRequest() *http.Request {
	r := rr.r
	if rr.reservedBody.isRead {
		r = r.Clone(r.Context())
		r.Body = io.NopCloser(&rr.reservedBody.buff)
	} else {
		r.Body = rr.rawBody
	}
	return r
}

func (rr *reservedResponseWriter) next(next http.Handler, r *http.Request) {
	r = r.WithContext(context.WithValue(r.Context(), reservedResponseWriterCtxKey{}, rr))
	next.ServeHTTP(rr.w, r)
}

func (rr *reservedResponseWriter) getResponseWriter() http.ResponseWriter {
	return rr.w
}

func (rrb *reservedRequestBody) Read(p []byte) (n int, err error) {
	rrb.once.Do(func() {
		rrb.isRead = true
	})
	return rrb.teeReader.Read(p)
}

func (rrb *reservedRequestBody) Close() error {
	return rrb.body.Close()
}
