package httpx

import "errors"

// TrustedError is an error for returning information to the consumer of the API
// since errors by default of logged in the errors middleware and generic HTTP status
// text is returned in the response. TrustedError.Msg can contain the trusted message.
// The default behavior of preventing errors from being returned is to avoid leaking
// implementation details.
type TrustedError struct {
	Msg string
	Err error
}

func NewTrustedError(msg string, err error) error {
	return &TrustedError{msg, err}
}

func (te *TrustedError) Error() string {
	return te.Msg
}

func (te *TrustedError) Unwrap() error {
	return te.Err
}

func IsTrustedError(err error) bool {
	var te *TrustedError
	val := errors.As(err, &te)
	return val
}

func GetTrustedError(err error) *TrustedError {
	var te *TrustedError
	if !errors.As(err, &te) {
		return nil
	}
	return te
}
