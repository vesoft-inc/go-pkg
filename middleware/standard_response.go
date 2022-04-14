package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/vesoft-inc/go-pkg/errorx"
)

const (
	StandardResponseBodyJson StandardResponseBodyType = 0 // nolint:revive
	StandardResponseBodyNone StandardResponseBodyType = 1
)

const (
	standardResponseFieldCode      = "code"
	standardResponseFieldMessage   = "message"
	standardResponseFieldData      = "data"
	standardResponseFieldDebugInfo = "debugInfo"
)

type (
	StandardResponseHandler struct {
		params StandardResponseHandlerParams
	}

	StandardResponseBodyType int

	StandardResponseHandlerParams struct {
		// GetErrCode used to parse the error.
		GetErrCode func(error) *errorx.ErrCode
		// CheckBodyType checks the type of body, default is StandardResponseBodyJson.
		CheckBodyType func(r *http.Request) StandardResponseBodyType
		// Errorf write the error logs.
		Errorf func(format string, a ...interface{})
		// DebugInfo add debugInfo details in body when error.
		DebugInfo bool
	}

	StandardResponse struct {
		params StandardResponseParams
	}

	StandardResponseParams struct {
		Handle func(w http.ResponseWriter, r *http.Request, data interface{}, err error)
	}

	standardResponseCtxKey struct{}
)

func NewStandardResponseHandler(params StandardResponseHandlerParams) *StandardResponseHandler {
	return &StandardResponseHandler{
		params: params,
	}
}

func NewStandardResponse(params StandardResponseParams) *StandardResponse {
	return &StandardResponse{
		params: params,
	}
}

func GetStandardResponseHandler(ctx context.Context) func(interface{}, error) {
	v := ctx.Value(standardResponseCtxKey{})
	if v == nil {
		return nil
	}
	return v.(func(data interface{}, err error))
}

func (h *StandardResponseHandler) GetHandleBody(
	r *http.Request,
	data interface{},
	err error,
) (httpStatus int, body interface{}) {
	httpStatus = http.StatusOK
	bodyType := StandardResponseBodyJson
	if h.params.CheckBodyType != nil {
		bodyType = h.params.CheckBodyType(r)
	}

	if err != nil {
		e, ok := errorx.AsCodeError(err)
		if !ok {
			err = errorx.WithCode(errorx.TakeCodePriority(func() *errorx.ErrCode {
				if h.params.GetErrCode == nil {
					return nil
				}
				return h.params.GetErrCode(err)
			}, func() *errorx.ErrCode {
				return errorx.NewErrCode(errorx.CCInternalServer, 0, 0, "ErrInternalServer")
			}), err)
			e, _ = errorx.AsCodeError(err)
		}
		httpStatus = e.GetHTTPStatus()

		if bodyType != StandardResponseBodyNone {
			resp := map[string]interface{}{
				standardResponseFieldCode:    e.GetCode(),
				standardResponseFieldMessage: e.GetMessage(),
			}
			if h.params.DebugInfo {
				resp[standardResponseFieldDebugInfo] = fmt.Sprintf("%+v", err)
			}
			body = resp
		}
	} else if bodyType != StandardResponseBodyNone {
		resp := map[string]interface{}{
			standardResponseFieldCode:    0,
			standardResponseFieldMessage: "Success",
		}
		if data != nil {
			resp[standardResponseFieldData] = data
		}
		body = resp
	}

	return httpStatus, body
}

func (h *StandardResponseHandler) Handle(
	w http.ResponseWriter,
	r *http.Request,
	data interface{},
	err error,
) {
	httpStatus, body := h.GetHandleBody(r, data, err)
	if body == nil {
		w.WriteHeader(httpStatus)
		return
	}

	bs, err := json.Marshal(body)
	if err != nil {
		h.errorf("write response json.Marshal failed, error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	if n, err := w.Write(bs); err != nil {
		if err != http.ErrHandlerTimeout {
			h.errorf("write response failed, error: %s", err)
		}
	} else if n < len(bs) {
		h.errorf("write response failed, actual bytes: %d, written bytes: %d", len(bs), n)
	}
}

func (m *StandardResponse) Handler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		next(w, r.WithContext(context.WithValue(r.Context(),
			standardResponseCtxKey{},
			func(data interface{}, err error) {
				m.params.Handle(w, r, data, err)
			}),
		))
	}
}

func (h *StandardResponseHandler) errorf(format string, a ...interface{}) {
	if h.params.Errorf != nil {
		h.params.Errorf(format, a...)
	}
}
