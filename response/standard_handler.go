package response

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/vesoft-inc/go-pkg/errorx"
)

const (
	StandardHandlerBodyJson StandardHandlerBodyType = 0 // nolint:revive
	StandardHandlerBodyNone StandardHandlerBodyType = 1
)

const (
	standardHandlerFieldCode      = "code"
	standardHandlerFieldMessage   = "message"
	standardHandlerFieldData      = "data"
	standardHandlerFieldDebugInfo = "debugInfo"
)

var _ Handler = (*standardHandler)(nil)

type (
	standardHandler struct {
		params StandardHandlerParams
	}

	StandardHandlerBodyType int

	StandardHandlerParams struct {
		// GetErrCode used to parse the error.
		GetErrCode func(error) *errorx.ErrCode
		// CheckBodyType checks the type of body, default is StandardHandlerBodyJson.
		CheckBodyType func(r *http.Request) StandardHandlerBodyType
		// Errorf write the error logs.
		Errorf func(format string, a ...interface{})
		// DebugInfo add debugInfo details in body when error.
		DebugInfo bool
	}

	standardHandlerDataFieldAny struct {
		data interface{}
	}
)

func NewStandardHandler(params StandardHandlerParams) Handler {
	return &standardHandler{
		params: params,
	}
}

// StandardHandlerDataFieldAny is to solve the problem that interface{} cannot be directly returned as the data field.
// For examples:
// 	var data interface{} = ...
// 	return &XxxResp {
//      Data: data,
// 	}
// The response body is:
// 	{
//      "code": 0,
//      "message": "Success",
//      "data": {
//          "data": ...
//      }
//  }
//
// Once you use StandardHandlerDataFieldAny,
// 	var data interface{} = ...
// 	return &XxxResp {
//      Data: StandardHandlerDataFieldAny(data),
// 	}
// The response body is:
// 	{
//      "code": 0,
//      "message": "Success",
//      "data": ...
//  }
func StandardHandlerDataFieldAny(data interface{}) interface{} {
	return &standardHandlerDataFieldAny{data: data}
}

func (h *standardHandler) GetStatusBody(r *http.Request, data interface{}, err error) (httpStatus int, body interface{}) {
	httpStatus = http.StatusOK
	bodyType := StandardHandlerBodyJson

	if r == nil {
		bodyType = StandardHandlerBodyNone
	} else if h.params.CheckBodyType != nil {
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

		if bodyType != StandardHandlerBodyNone {
			resp := map[string]interface{}{
				standardHandlerFieldCode:    e.GetCode(),
				standardHandlerFieldMessage: e.GetMessage(),
			}
			if h.params.DebugInfo {
				resp[standardHandlerFieldDebugInfo] = fmt.Sprintf("%+v", err)
			}
			body = resp
		}
	} else if bodyType != StandardHandlerBodyNone {
		resp := map[string]interface{}{
			standardHandlerFieldCode:    0,
			standardHandlerFieldMessage: "Success",
		}
		data = h.getData(data)
		if data != nil {
			resp[standardHandlerFieldData] = data
		}
		body = resp
	}

	return httpStatus, body
}

func (h *standardHandler) Handle(w http.ResponseWriter, r *http.Request, data interface{}, err error) {
	httpStatus, body := h.GetStatusBody(r, data, err)
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

func (*standardHandler) getData(data interface{}) interface{} {
	if data == nil {
		return nil
	}
	if v, ok := data.(*standardHandlerDataFieldAny); ok {
		return v.data
	}

	reflectType := reflect.TypeOf(data)
	reflectValue := reflect.Indirect(reflect.ValueOf(data))
	if reflectType.Kind() == reflect.Ptr {
		reflectType = reflectType.Elem()
	}
	if reflectType.Kind() != reflect.Struct || reflectType.NumField() != 1 {
		return data
	}
	field := reflectValue.Field(0)
	if !field.CanInterface() {
		return data
	}
	if v, ok := field.Interface().(*standardHandlerDataFieldAny); ok {
		return v.data
	}
	return data
}

func (h *standardHandler) errorf(format string, a ...interface{}) {
	if h.params.Errorf != nil {
		h.params.Errorf(format, a...)
	}
}
