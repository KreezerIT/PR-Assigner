package validator

import (
	"errors"
	"fmt"
	"strings"
)

// ValidationError ошибка валидации
type ValidationError struct {
	Field   string
	Message string
}

// ValidationErrors набор ошибок валидации
type ValidationErrors []ValidationError

// Error возвращает объединённое строковое представление ошибки/ошибок валидации

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}

	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

func (e ValidationErrors) HasErrors() bool {
	return len(e) > 0
}

type Validator interface {
	Validate() error
}

// CompositeValidator хранит список валидаторов и выполняет их последовательно
type CompositeValidator struct {
	validators []Validator
}

func NewCompositeValidator() *CompositeValidator {
	return &CompositeValidator{
		validators: make([]Validator, 0),
	}
}

func (cv *CompositeValidator) Add(validator Validator) *CompositeValidator {
	cv.validators = append(cv.validators, validator)
	return cv
}

// Validate выполняет все зарегистрированные валидаторы и собирает ошибки
func (cv *CompositeValidator) Validate() error {
	var validationErrs ValidationErrors

	for _, validator := range cv.validators {
		if err := validator.Validate(); err != nil {
			var valErr ValidationError
			if errors.As(err, &valErr) {
				validationErrs = append(validationErrs, valErr)
			} else {
				var valErrs ValidationErrors
				if errors.As(err, &valErrs) {
					validationErrs = append(validationErrs, valErrs...)
				} else {
					validationErrs = append(validationErrs, ValidationError{
						Field:   "unknown",
						Message: err.Error(),
					})
				}
			}
		}
	}

	if validationErrs.HasErrors() {
		return validationErrs
	}

	return nil
}

// RequiredStringValidator проверяет, что строка не пустая
type RequiredStringValidator struct {
	Field string
	Value string
}

func Required(field, value string) Validator {
	return &RequiredStringValidator{Field: field, Value: value}
}

func (v *RequiredStringValidator) Validate() error {
	if strings.TrimSpace(v.Value) == "" {
		return ValidationError{
			Field:   v.Field,
			Message: "is required",
		}
	}
	return nil
}

// MinLengthValidator проверяет минимальную длину строки
type MinLengthValidator struct {
	Field     string
	Value     string
	MinLength int
}

func MinLength(field, value string, minLength int) Validator {
	return &MinLengthValidator{
		Field:     field,
		Value:     value,
		MinLength: minLength,
	}
}

func (v *MinLengthValidator) Validate() error {
	if len(strings.TrimSpace(v.Value)) < v.MinLength {
		return ValidationError{
			Field:   v.Field,
			Message: fmt.Sprintf("must be at least %d characters", v.MinLength),
		}
	}
	return nil
}

// MaxLengthValidator проверяет максимальную длину строки
type MaxLengthValidator struct {
	Field     string
	Value     string
	MaxLength int
}

func MaxLength(field, value string, maxLength int) Validator {
	return &MaxLengthValidator{
		Field:     field,
		Value:     value,
		MaxLength: maxLength,
	}
}

func (v *MaxLengthValidator) Validate() error {
	if len(v.Value) > v.MaxLength {
		return ValidationError{
			Field:   v.Field,
			Message: fmt.Sprintf("must not exceed %d characters", v.MaxLength),
		}
	}
	return nil
}

// NotEmptySliceValidator проверяет, что срез не пустой
type NotEmptySliceValidator struct {
	Field string
	Size  int
}

func NotEmptySlice(field string, size int) Validator {
	return &NotEmptySliceValidator{
		Field: field,
		Size:  size,
	}
}

func (v *NotEmptySliceValidator) Validate() error {
	if v.Size == 0 {
		return ValidationError{
			Field:   v.Field,
			Message: "must not be empty",
		}
	}
	return nil
}

// CustomValidator для кастомной логики
type CustomValidator struct {
	Field     string
	CheckFunc func() bool
	Message   string
}

func Custom(field string, checkFunc func() bool, message string) Validator {
	return &CustomValidator{
		Field:     field,
		CheckFunc: checkFunc,
		Message:   message,
	}
}

func (v *CustomValidator) Validate() error {
	if !v.CheckFunc() {
		return ValidationError{
			Field:   v.Field,
			Message: v.Message,
		}
	}
	return nil
}
