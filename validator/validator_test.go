package validator

import (
	"reflect"
	"testing"

	govalidator "github.com/go-playground/validator/v10"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

func TestStruct(t *testing.T) {
	var (
		err error
		ast = assert.New(t)
	)

	type testStruct struct {
		Data int `validate:"min=1,max=10"`
	}
	err = Struct(testStruct{})
	if errs, ok := err.(govalidator.ValidationErrors); ast.True(ok) {
		ast.Len(errs, 1)
	}
	err = Struct(testStruct{Data: -1})
	if errs, ok := err.(govalidator.ValidationErrors); ast.True(ok) {
		ast.Len(errs, 1)
	}
	err = Struct(testStruct{Data: 11})
	if errs, ok := err.(govalidator.ValidationErrors); ast.True(ok) {
		ast.Len(errs, 1)
	}
	err = Struct(testStruct{Data: 3})
	ast.NoError(err)
}

func TestVar(t *testing.T) {
	var (
		err error
		ast = assert.New(t)
	)
	err = Var(0, "min=1,max=10")
	if errs, ok := err.(govalidator.ValidationErrors); ast.True(ok) {
		ast.Len(errs, 1)
	}
	err = Var(-1, "min=1,max=10")
	if errs, ok := err.(govalidator.ValidationErrors); ast.True(ok) {
		ast.Len(errs, 1)
	}
	err = Var(11, "min=1,max=10")
	if errs, ok := err.(govalidator.ValidationErrors); ast.True(ok) {
		ast.Len(errs, 1)
	}
	err = Var(3, "min=1,max=10")
	ast.NoError(err)
}

func TestExtendValidators(t *testing.T) {
	stubs := gostub.Stub(&extendValidators, map[string]govalidator.Func{
		"streq": func(fl govalidator.FieldLevel) bool {
			field := fl.Field()
			param := fl.Param()
			if field.Kind() == reflect.String {
				return field.String() == param
			}
			return false
		},
	})
	defer stubs.Reset()

	var (
		err error
		ast = assert.New(t)
		v   = New()
	)

	err = v.Var("aa", "streq=aaa")
	if errs, ok := err.(govalidator.ValidationErrors); ast.True(ok) {
		ast.Len(errs, 1)
	}
	err = v.Var("aaa", "streq=aaa")
	ast.NoError(err)
}
