# Package `errorx`

`errorx` is an error code helper package.

## How To Use

Create an error code package `ecode`, which contains two files:

- alias.go
- codes.go

The `alias.go` files:

```golang
package ecode

import (
	"github.com/vesoft-inc/go-pkg/errorx"
)

const (
	// CodeCategory
	CCBadRequest     = errorx.CCBadRequest     // 400
	CCUnauthorized   = errorx.CCUnauthorized   // 401
	CCForbidden      = errorx.CCForbidden      // 403
	CCNotFound       = errorx.CCNotFound       // 404
	CCInternalServer = errorx.CCInternalServer // 500
	CCNotImplemented = errorx.CCNotImplemented // 501
	CCUnknown        = errorx.CCUnknown        // 900
)

var (
	// WithCode return error warps with codeError.
	// c is the code. err is the real err. formatWithArgs is format string with args.
	// For example:
	//  WithCode(ErrBadRequest, nil)
	//  WithCode(ErrBadRequest, err)
	//  WithCode(ErrBadRequest, err, "message")
	//  WithCode(ErrBadRequest, err, "message %s", "id")
	WithCode         = errorx.WithCode
	AsCodeError      = errorx.AsCodeError
	IsCodeError      = errorx.IsCodeError
	SeparateCode     = errorx.SeparateCode
	TakeCodePriority = errorx.TakeCodePriority

	// newErrCode is create an new *ErrCode, it's only used for global initialization.
	// Do not export so that it cannot be used outside of this package.
	newErrCode = errorx.NewErrCode
)

type (
	// ErrCode is the error code for app
	// 0 indicates success, others indicate failure.
	// It is combined of error category code, platform code, and specific code via CodeCombiner.
	// The default CodeCombiner's rules are as follows:
	// - The first three digits represent the category code, analogous to the http status code.
	// - The next two digits indicate the platform code.
	// - The last three digits indicate the specific code.
	//   For example:
	//     4041001:
	// 	     404 is the error category code
	// 	      10 is the error platform code
	// 	     001 is the error specific code
	ErrCode   = errorx.ErrCode
	CodeError = errorx.CodeError
)
```

The `codes.go` files:

```golang
package ecode

import (
	"net/http"
)

const (
	// TODO: Please modify it to your own platform code.
	PlatformCode = 0
)

// Define you error code here
var (
	ErrBadRequest                  = newErrCode(CCBadRequest, PlatformCode, 0, "ErrBadRequest")                  // 4000000
	ErrParam                       = newErrCode(CCBadRequest, PlatformCode, 1, "ErrParam")                       // 4000001
	ErrUnauthorized                = newErrCode(CCUnauthorized, PlatformCode, 0, "ErrUnauthorized")              // 4010000
	ErrForbidden                   = newErrCode(CCForbidden, PlatformCode, 0, "ErrForbidden")                    // 4030000
	ErrNotFound                    = newErrCode(CCNotFound, PlatformCode, 0, "ErrNotFound")                      // 4040000
	ErrInternalServer              = newErrCode(CCInternalServer, PlatformCode, 0, "ErrInternalServer")          // 5000000
	ErrNotImplemented              = newErrCode(CCNotImplemented, PlatformCode, 0, "ErrNotImplemented")          // 5010000
	ErrUnknown                     = newErrCode(CCUnknown, PlatformCode, 0, "ErrUnknown")                        // 9000000
)

var statusCodeErrorMapping = map[int]*ErrCode{
	http.StatusBadRequest:          ErrBadRequest,
	http.StatusUnauthorized:        ErrUnauthorized,
	http.StatusForbidden:           ErrForbidden,
	http.StatusNotFound:            ErrNotFound,
	http.StatusInternalServerError: ErrInternalServer,
	http.StatusNotImplemented:      ErrNotImplemented,
}

func GetErrCodeByHTTPStatus(httpStatus int) *ErrCode {
	if c, ok := statusCodeErrorMapping[httpStatus]; ok {
		return c
	}
	return ErrInternalServer
}

func WithBadRequest(err error, formatWithArgs ...interface{}) error {
	return WithCode(ErrBadRequest, err, formatWithArgs...)
}

func WithUnauthorized(err error, formatWithArgs ...interface{}) error {
	return WithCode(ErrUnauthorized, err, formatWithArgs...)
}

func WithForbidden(err error, formatWithArgs ...interface{}) error {
	return WithCode(ErrForbidden, err, formatWithArgs...)
}

func WithNotFound(err error, formatWithArgs ...interface{}) error {
	return WithCode(ErrNotFound, err, formatWithArgs...)
}

func WithInternalServer(err error, formatWithArgs ...interface{}) error {
	return WithCode(ErrInternalServer, err, formatWithArgs...)
}

func WithNotImplemented(err error, formatWithArgs ...interface{}) error {
	return WithCode(ErrNotImplemented, err, formatWithArgs...)
}

func WithUnknown(err error, formatWithArgs ...interface{}) error {
	return WithCode(ErrUnknown, err, formatWithArgs...)
}

func IsBadRequest(err error) bool {
	return IsCodeError(err, ErrBadRequest)
}

func IsUnauthorized(err error) bool {
	return IsCodeError(err, ErrUnauthorized)
}

func IsForbidden(err error) bool {
	return IsCodeError(err, ErrForbidden)
}

func IsNotFound(err error) bool {
	return IsCodeError(err, ErrNotFound)
}

func IsInternalServer(err error) bool {
	return IsCodeError(err, ErrInternalServer)
}

func IsNotImplemented(err error) bool {
	return IsCodeError(err, ErrNotImplemented)
}

func IsUnknown(err error) bool {
	return IsCodeError(err, ErrUnknown)
}
```
