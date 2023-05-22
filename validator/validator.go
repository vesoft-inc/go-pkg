package validator

import (
	"sync"

	govalidator "github.com/go-playground/validator/v10"
)

var (
	gValidator     Validator = (*defaultValidator)(nil)
	gValidatorInit sync.Once
)

type (
	Validator interface {
		RegisterValidation(tag string, fn Func, callValidationEvenIfNull ...bool) error
		Struct(s interface{}) error
		Var(field interface{}, tag string) error
	}

	defaultValidator struct {
		*govalidator.Validate
	}

	// alias
	FieldLevel             = govalidator.FieldLevel
	Func                   = func(fl FieldLevel) bool
	InvalidValidationError = govalidator.InvalidValidationError
	ValidationErrors       = govalidator.ValidationErrors
)

func New() Validator {
	v := &defaultValidator{
		Validate: govalidator.New(),
	}

	for k, val := range extendValidators {
		_ = v.RegisterValidation(k, val)
	}

	return v
}

func RegisterValidation(tag string, fn Func, callValidationEvenIfNull ...bool) error {
	initGValidator()
	return gValidator.RegisterValidation(tag, fn, callValidationEvenIfNull...)
}

func Struct(s interface{}) error {
	initGValidator()
	return gValidator.Struct(s)
}

func Var(field interface{}, tag string) error {
	initGValidator()
	return gValidator.Var(field, tag)
}

func (v *defaultValidator) RegisterValidation(tag string, fn Func, callValidationEvenIfNull ...bool) error {
	return v.Validate.RegisterValidation(tag, fn, callValidationEvenIfNull...)
}

func (v *defaultValidator) Struct(s interface{}) error {
	return v.Validate.Struct(s)
}

func (v *defaultValidator) Var(field interface{}, tag string) error {
	return v.Validate.Var(field, tag)
}

func initGValidator() {
	gValidatorInit.Do(func() {
		gValidator = New()
	})
}
