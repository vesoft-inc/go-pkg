package response

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/vesoft-inc/go-pkg/errorx"
)

const (
	StandardHandlerBodyJson StandardHandlerBodyType = 0 //nolint:revive
	StandardHandlerBodyNone StandardHandlerBodyType = 1

	StandardHandlerDetailsNone      StandardHandlerDetailsType = 0
	StandardHandlerDetailsNormal    StandardHandlerDetailsType = 1
	StandardHandlerDetailsWithError StandardHandlerDetailsType = 2
	StandardHandlerDetailsFull      StandardHandlerDetailsType = 3
)

const (
	standardHandlerFieldCode    = "code"
	standardHandlerFieldMessage = "message"
	standardHandlerFieldData    = "data"
	standardHandlerFieldDetails = "details"
)

var _ Handler = (*standardHandler)(nil)

type (
	standardHandler struct {
		params StandardHandlerParams
	}

	StandardHandlerBodyType int

	StandardHandlerDetailsType int

	StandardHandlerParams struct {
		// GetErrCode used to parse the error.
		GetErrCode func(error) *errorx.ErrCode
		// CheckBodyType checks the type of body, default is StandardHandlerBodyJson.
		CheckBodyType func(r *http.Request) StandardHandlerBodyType
		// Errorf write the error logs.
		// Deprecated: Use ContextErrorf instead.
		Errorf func(format string, a ...interface{})
		// ContextErrorf write the error logs.
		ContextErrorf func(ctx context.Context, format string, a ...interface{})
		// DetailsType is the type for details field, default is StandardHandlerDetailsDisable.
		DetailsType StandardHandlerDetailsType
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
//
//	var data interface{} = ...
//	return &XxxResp {
//	    Data: data,
//	}
//
// The response body is:
//
//	{
//	    "code": 0,
//	    "message": "Success",
//	    "data": {
//	        "data": ...
//	    }
//	}
//
// Once you use StandardHandlerDataFieldAny,
//
//	var data interface{} = ...
//	return &XxxResp {
//	    Data: StandardHandlerDataFieldAny(data),
//	}
//
// The response body is:
//
//	{
//	    "code": 0,
//	    "message": "Success",
//	    "data": ...
//	}
func StandardHandlerDataFieldAny(data interface{}) interface{} {
	return &standardHandlerDataFieldAny{data: data}
}

func GetStandardHandlerDataFieldAnyData(data interface{}) (interface{}, bool) {
	if v, ok := data.(*standardHandlerDataFieldAny); ok {
		return v.data, true
	}
	return nil, false
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
		
		if httpStatus >= http.StatusBadRequest {
			switch httpStatus {
			case http.StatusUnauthorized:
			case http.StatusForbidden:
			case http.StatusNotFound:
			default:
				h.errorf(r, "request failed %+v", err)
			}
		}

		if bodyType != StandardHandlerBodyNone {
			resp := map[string]interface{}{
				standardHandlerFieldCode:    e.GetCode(),
				standardHandlerFieldMessage: e.GetMessage(),
			}
			if details := h.getDetails(e); details != "" {
				resp[standardHandlerFieldDetails] = details
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
		h.errorf(r, "write response json.Marshal failed, error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	if n, err := w.Write(bs); err != nil {
		if err != http.ErrHandlerTimeout {
			h.errorf(r, "write response failed, error: %s", err)
		}
	} else if n < len(bs) {
		h.errorf(r, "write response failed, actual bytes: %d, written bytes: %d", len(bs), n)
	}
}

func (*standardHandler) getData(data interface{}) interface{} {
	if isInterfaceNil(data) {
		return nil
	}
	if v, ok := GetStandardHandlerDataFieldAnyData(data); ok {
		return v
	}

	reflectValue := reflect.Indirect(reflect.ValueOf(data))
	if reflectValue.Kind() != reflect.Struct || reflectValue.NumField() != 1 {
		return data
	}
	field := reflectValue.Field(0)
	if !field.CanInterface() {
		return data
	}
	if v, ok := GetStandardHandlerDataFieldAnyData(field.Interface()); ok {
		return v
	}
	return data
}

func (h *standardHandler) errorf(r *http.Request, format string, a ...interface{}) {
	var requestInfo string
	if r != nil && r.URL != nil {
		requestInfo = fmt.Sprintf("[%s] ", r.URL.String())
	}
	if h.params.ContextErrorf != nil {
		ctx := context.Background()
		if r != nil {
			ctx = r.Context()
		}
		h.params.ContextErrorf(ctx, requestInfo+format, a...)
	} else if h.params.Errorf != nil {
		h.params.Errorf(requestInfo+format, a...)
	}
}

func (h *standardHandler) getDetails(e errorx.CodeError) string {
	switch h.params.DetailsType {
	case StandardHandlerDetailsNone:
	case StandardHandlerDetailsNormal:
		return e.Error()
	case StandardHandlerDetailsWithError:
		if internalError := errors.Unwrap(e); internalError != nil {
			return fmt.Sprintf("%s:%s", e.Error(), internalError.Error())
		}
		return e.Error()
	case StandardHandlerDetailsFull:
		return fmt.Sprintf("%+v", e)
	}
	return ""
}

func isInterfaceNil(i interface{}) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() { //nolint:exhaustive
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}
