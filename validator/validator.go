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
		Struct(s interface{}) error
		Var(field interface{}, tag string) error
	}

	defaultValidator struct {
		*govalidator.Validate
	}

	// alias
	InvalidValidationError = govalidator.InvalidValidationError
	ValidationErrors       = govalidator.ValidationErrors
)

func New() Validator {
	validate := govalidator.New()
	for k, val := range extendValidators {
		_ = validate.RegisterValidation(k, val)
	}
	return &defaultValidator{
		Validate: validate,
	}
}

func Struct(s interface{}) error {
	initGValidator()
	return gValidator.Struct(s)
}

func Var(field interface{}, tag string) error {
	initGValidator()
	return gValidator.Var(field, tag)
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
