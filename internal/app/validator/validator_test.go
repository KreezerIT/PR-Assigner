package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequiredStringValidator_Success(t *testing.T) {
	validator := Required("field", "value")
	err := validator.Validate()
	assert.NoError(t, err)
}

func TestRequiredStringValidator_Empty(t *testing.T) {
	validator := Required("field", "")
	err := validator.Validate()
	assert.Error(t, err)

	var valErr ValidationError
	assert.ErrorAs(t, err, &valErr)
	assert.Equal(t, "field", valErr.Field)
	assert.Equal(t, "is required", valErr.Message)
}

func TestRequiredStringValidator_Whitespace(t *testing.T) {
	validator := Required("field", "   ")
	err := validator.Validate()
	assert.Error(t, err)
}

func TestMinLengthValidator_Success(t *testing.T) {
	validator := MinLength("field", "hello", 3)
	err := validator.Validate()
	assert.NoError(t, err)
}

func TestMinLengthValidator_TooShort(t *testing.T) {
	validator := MinLength("field", "hi", 3)
	err := validator.Validate()
	assert.Error(t, err)

	var valErr ValidationError
	assert.ErrorAs(t, err, &valErr)
	assert.Contains(t, valErr.Message, "must be at least 3 characters")
}

func TestMaxLengthValidator_Success(t *testing.T) {
	validator := MaxLength("field", "hello", 10)
	err := validator.Validate()
	assert.NoError(t, err)
}

func TestMaxLengthValidator_TooLong(t *testing.T) {
	validator := MaxLength("field", "very long string", 5)
	err := validator.Validate()
	assert.Error(t, err)

	var valErr ValidationError
	assert.ErrorAs(t, err, &valErr)
	assert.Contains(t, valErr.Message, "must not exceed 5 characters")
}

func TestNotEmptySliceValidator_Success(t *testing.T) {
	validator := NotEmptySlice("items", 5)
	err := validator.Validate()
	assert.NoError(t, err)
}

func TestNotEmptySliceValidator_Empty(t *testing.T) {
	validator := NotEmptySlice("items", 0)
	err := validator.Validate()
	assert.Error(t, err)

	var valErr ValidationError
	assert.ErrorAs(t, err, &valErr)
	assert.Equal(t, "items", valErr.Field)
	assert.Equal(t, "must not be empty", valErr.Message)
}

func TestCustomValidator_Success(t *testing.T) {
	validator := Custom("age", func() bool { return 25 >= 18 }, "must be at least 18")
	err := validator.Validate()
	assert.NoError(t, err)
}

func TestCustomValidator_Failed(t *testing.T) {
	validator := Custom("age", func() bool { return 15 >= 18 }, "must be at least 18")
	err := validator.Validate()
	assert.Error(t, err)

	var valErr ValidationError
	assert.ErrorAs(t, err, &valErr)
	assert.Equal(t, "age", valErr.Field)
	assert.Equal(t, "must be at least 18", valErr.Message)
}

func TestCompositeValidator_AllValid(t *testing.T) {
	cv := NewCompositeValidator()
	cv.Add(Required("name", "John"))
	cv.Add(MinLength("name", "John", 2))
	cv.Add(MaxLength("name", "John", 50))

	err := cv.Validate()
	assert.NoError(t, err)
}

func TestCompositeValidator_MultipleErrors(t *testing.T) {
	cv := NewCompositeValidator()
	cv.Add(Required("name", ""))
	cv.Add(Required("email", ""))

	err := cv.Validate()
	assert.Error(t, err)

	var valErrs ValidationErrors
	assert.ErrorAs(t, err, &valErrs)
	assert.Len(t, valErrs, 2)
}

func TestCompositeValidator_PartialErrors(t *testing.T) {
	cv := NewCompositeValidator()
	cv.Add(Required("name", "John"))
	cv.Add(Required("email", ""))
	cv.Add(MinLength("password", "123", 8))

	err := cv.Validate()
	assert.Error(t, err)

	var valErrs ValidationErrors
	assert.ErrorAs(t, err, &valErrs)
	assert.Len(t, valErrs, 2) // пароль и почта
}

func TestValidationErrors_Error(t *testing.T) {
	errorsList := ValidationErrors{
		ValidationError{Field: "name", Message: "is required"},
		ValidationError{Field: "email", Message: "is invalid"},
	}

	errMsg := errorsList.Error()
	assert.Contains(t, errMsg, "name: is required")
	assert.Contains(t, errMsg, "email: is invalid")
}

func TestValidationErrors_HasErrors(t *testing.T) {
	var errs ValidationErrors
	assert.False(t, errs.HasErrors())

	errs = append(errs, ValidationError{Field: "test", Message: "error"})
	assert.True(t, errs.HasErrors())
}
