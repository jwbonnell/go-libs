package validate

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

var (
	uni      *ut.UniversalTranslator
	trans    ut.Translator
	validate *validator.Validate
)

func init() {
	uni = ut.New(en.New(), en.New())
	var ok bool
	if trans, ok = uni.GetTranslator("en"); !ok {
		panic("validate: failed to find 'en' translator")
	}

	validate = validator.New()
	if err := en_translations.RegisterDefaultTranslations(validate, trans); err != nil {
		panic("validate: failed to register translations: " + err.Error())
	}

	// Set field names to JSON struct tags if found
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// Check validates the provided value
func Check(val any) error {
	if err := validate.Struct(val); err != nil {
		ve := createValidationErrors(err)
		if ve != nil {
			return ve
		}
		return err
	}
	return nil
}

// createValidationErrors creates VErrors from validator.ValidationErrors
func createValidationErrors(err error) VErrors {
	var verrors validator.ValidationErrors
	if !errors.As(err, &verrors) {
		return nil
	}

	var ves VErrors
	for _, verror := range verrors {
		ve := VError{
			Field: verror.Field(),
			Err:   verror.Translate(trans),
		}
		ves = append(ves, ve)
	}
	return ves
}

// VError represents a single validation error
type VError struct {
	Field string `json:"field"`
	Err   string `json:"error"`
}

// VErrors is a slice of VError
type VErrors []VError

// Error implements the error interface
func (ves VErrors) Error() string {
	msgs := make([]string, len(ves))
	for i, ve := range ves {
		msgs[i] = ve.Field + ": " + ve.Err
	}
	return "validation failed: " + strings.Join(msgs, "; ")
}
