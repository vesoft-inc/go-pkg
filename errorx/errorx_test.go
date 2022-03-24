package errorx

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

const testCPCloudServer = 1

type (
	testCodeCombiner struct{}
)

var (
	testErrBadRequest     = NewErrCode(CCBadRequest, testCPCloudServer, 0, "testErrBadRequest")         // 40001000
	testErrParam          = NewErrCode(CCBadRequest, testCPCloudServer, 1, "testErrParam")              // 40001001
	testErrUnauthorized   = NewErrCode(CCUnauthorized, testCPCloudServer, 0, "testErrUnauthorized")     // 40101000
	testErrForbidden      = NewErrCode(CCForbidden, testCPCloudServer, 0, "testErrForbidden")           // 40301000
	testErrNotFound       = NewErrCode(CCNotFound, testCPCloudServer, 0, "testErrNotFound")             // 40401000
	testErrInternalServer = NewErrCode(CCInternalServer, testCPCloudServer, 0, "testErrInternalServer") // 50001000
	testErrNotImplemented = NewErrCode(CCNotImplemented, testCPCloudServer, 0, "testErrNotImplemented") // 50101000
	testErrUnknown        = NewErrCode(CCUnknown, testCPCloudServer, 0, "testErrUnknown")               // 90001000
)

func Test_TakeCodePriority(t *testing.T) {
	assert.Equal(t, (*ErrCode)(nil), TakeCodePriority())
	assert.Equal(t, (*ErrCode)(nil), TakeCodePriority(func() *ErrCode {
		return nil
	}))
	assert.Equal(t, testErrInternalServer, TakeCodePriority(func() *ErrCode {
		return testErrInternalServer
	}))
}

func Test_getMessage(t *testing.T) {
	kcs := map[string]*ErrCode{
		"testErrBadRequest":     testErrBadRequest,
		"testErrParam":          testErrParam,
		"testErrUnauthorized":   testErrUnauthorized,
		"testErrForbidden":      testErrForbidden,
		"testErrNotFound":       testErrNotFound,
		"testErrInternalServer": testErrInternalServer,
		"testErrNotImplemented": testErrNotImplemented,
		"testErrUnknown":        testErrUnknown,
	}
	for k, c := range kcs {
		assert.Equal(t, k, c.GetMessage())
	}
}

func Test_NewErrCode(t *testing.T) {
	c := NewErrCode(123, 45, 67, "msg")
	assert.Equal(t, 12345067, c.code)
	assert.Equal(t, "msg", c.message)
	assert.Equal(t, c, c.GetErrCode())
	assert.Equal(t, 12345067, c.GetCode())
	assert.Equal(t, 123, c.GetCategoryCode())
	assert.Equal(t, 45, c.GetPlatformCode())
	assert.Equal(t, 67, c.GetSpecificCode())
	assert.Equal(t, 123, c.GetHTTPStatus())

	categoryCode, platformCode, specificCode := SeparateCode(c.GetCode())
	assert.Equal(t, 123, categoryCode)
	assert.Equal(t, 45, platformCode)
	assert.Equal(t, 67, specificCode)

	c = NewErrCode(CCUnknown, 45, 67, "msg")
	assert.Equal(t, http.StatusInternalServerError, c.GetHTTPStatus())
}

func Test_SetCodeCombiner(t *testing.T) {
	curCodeCombiner := codeCombiner
	defer SetCodeCombiner(curCodeCombiner)

	SetCodeCombiner(testCodeCombiner{})
	c := NewErrCode(1, 2, 3, "msg")
	assert.Equal(t, 123, c.code)
	assert.Equal(t, 1, c.GetCategoryCode())
	assert.Equal(t, 2, c.GetPlatformCode())
	assert.Equal(t, 3, c.GetSpecificCode())
	assert.Equal(t, 1, c.GetHTTPStatus())

	categoryCode, platformCode, specificCode := SeparateCode(c.GetCode())
	assert.Equal(t, 1, categoryCode)
	assert.Equal(t, 2, platformCode)
	assert.Equal(t, 3, specificCode)
}

func TestWithCode(t *testing.T) {
	tests := []struct {
		name           string
		code           *ErrCode
		err            error
		args           []interface{}
		expectedErrStr string
	}{{
		name:           "nil err",
		code:           testErrBadRequest,
		err:            nil,
		args:           nil,
		expectedErrStr: "40001000: testErrBadRequest",
	}, {
		name:           "with err",
		code:           testErrBadRequest,
		err:            errors.New("myError"),
		args:           nil,
		expectedErrStr: "40001000: testErrBadRequest",
	}, {
		name:           "with message",
		code:           testErrBadRequest,
		err:            nil,
		args:           []interface{}{"myMessage"},
		expectedErrStr: "myMessage: 40001000: testErrBadRequest",
	}, {
		name:           "with messagef",
		code:           testErrBadRequest,
		err:            nil,
		args:           []interface{}{"myMessage %d", 10},
		expectedErrStr: "myMessage 10: 40001000: testErrBadRequest",
	}, {
		name:           "with message not string",
		code:           testErrBadRequest,
		err:            nil,
		args:           []interface{}{0},
		expectedErrStr: "40001000: testErrBadRequest",
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := WithCode(test.code, test.err, test.args...)
			assert.Equal(t, test.expectedErrStr, err.Error())
			assert.Contains(t, fmt.Sprintf("%q", err), test.expectedErrStr)
			stackCount := strings.Count(fmt.Sprintf("%+v", err), "errorx.TestWithCode")
			assert.Equal(t, 1, stackCount)
		})
	}
}

func Test_AsCodeError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		errCode  *ErrCode
		expected bool
	}{{
		name:     "nil err",
		err:      nil,
		errCode:  nil,
		expected: false,
	}, {
		name:     "other error",
		err:      errors.New("otherError"),
		errCode:  nil,
		expected: false,
	}, {
		name:     "newCodeErrorInternal",
		err:      newCodeErrorInternal(testErrBadRequest),
		errCode:  testErrBadRequest,
		expected: true,
	}, {
		name:     "WithCode",
		err:      WithCode(testErrNotFound, errors.New("otherError"), "myMessage %d", 10),
		errCode:  testErrNotFound,
		expected: true,
	}, {
		name:     "WithCode twice",
		err:      WithCode(testErrInternalServer, WithCode(testErrNotFound, errors.New("otherError"), "myMessage %d", 10), "myMessage %d", 10),
		errCode:  testErrInternalServer,
		expected: true,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e, ok := AsCodeError(test.err)
			assert.Equal(t, test.expected, ok)
			assert.True(t, ok == (test.errCode != nil))
			if ok {
				assert.Equal(t, e.GetCode(), test.errCode.GetCode())
				assert.Equal(t, e.GetCategoryCode(), test.errCode.GetCategoryCode())
				assert.Equal(t, e.GetPlatformCode(), test.errCode.GetPlatformCode())
				assert.Equal(t, e.GetSpecificCode(), test.errCode.GetSpecificCode())
				assert.Equal(t, e.GetMessage(), test.errCode.GetMessage())
				assert.True(t, e.IsErrCode(test.errCode))
			}
		})
	}
}

func Test_IsCodeError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		errCode  []*ErrCode
		expected bool
	}{{
		name:     "nil err",
		err:      nil,
		errCode:  nil,
		expected: false,
	}, {
		name:     "other error",
		err:      errors.New("otherError"),
		errCode:  nil,
		expected: false,
	}, {
		name:     "other error with code",
		err:      errors.New("otherError"),
		errCode:  []*ErrCode{testErrBadRequest},
		expected: false,
	}, {
		name:     "multi error code",
		err:      newCodeErrorInternal(testErrBadRequest),
		errCode:  []*ErrCode{testErrBadRequest, testErrNotFound},
		expected: false,
	}, {
		name:     "newCodeErrorInternal",
		err:      newCodeErrorInternal(testErrBadRequest),
		errCode:  nil,
		expected: true,
	}, {
		name:     "newCodeErrorInternal with code",
		err:      newCodeErrorInternal(testErrBadRequest),
		errCode:  []*ErrCode{testErrBadRequest},
		expected: true,
	}, {
		name:     "WithCode",
		err:      WithCode(testErrNotFound, errors.New("otherError"), "myMessage %d", 10),
		errCode:  []*ErrCode{testErrNotFound},
		expected: true,
	}, {
		name:     "WithCode twice",
		err:      WithCode(testErrInternalServer, WithCode(testErrNotFound, errors.New("otherError"), "myMessage %d", 10), "myMessage %d", 10),
		errCode:  []*ErrCode{testErrInternalServer},
		expected: true,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, IsCodeError(test.err, test.errCode...), test.name)
		})
	}
}

func (testCodeCombiner) Combine(categoryCode, platformCode, specificCode int) int {
	return categoryCode*100 + platformCode*10 + specificCode
}

func (testCodeCombiner) Separate(code int) (categoryCode, platformCode, specificCode int) {
	return code / 100, code / 10 % 10, code % 10
}
