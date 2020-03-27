package autha

import (
	"errors"
	"strings"
	"unicode"
)

var (
	ErrTryAgain = errors.New("Try again")
)

type WrappedError struct {
	Message string
	Err     error
}

func (w WrappedError) Error() string {
	return w.Message
}

func (w WrappedError) Unwrap() error {
	return w.Err
}

func NewWrapped(message string, err error) error {
	m := strings.TrimFunc(message, unicode.IsSpace)
	if len(m) == 0 {
		return err
	}
	return &WrappedError{
		Message: message,
		Err:     err,
	}
}
