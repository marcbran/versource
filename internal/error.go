package internal

import (
	"fmt"
)

type UserError struct {
	Message string
	Cause   error
}

func (e *UserError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *UserError) Unwrap() error {
	return e.Cause
}

func UserErr(message string) error {
	return &UserError{Message: message}
}

func UserErrf(format string, args ...any) error {
	return &UserError{Message: fmt.Sprintf(format, args...)}
}

func UserErrE(message string, cause error) error {
	return &UserError{Message: message, Cause: cause}
}

func IsUserError(err error) bool {
	_, ok := err.(*UserError)
	return ok
}

type InternalError struct {
	Message string
	Cause   error
}

func (e *InternalError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *InternalError) Unwrap() error {
	return e.Cause
}

func InternalErr(message string) error {
	return &InternalError{Message: message}
}

func InternalErrf(format string, args ...any) error {
	return &InternalError{Message: fmt.Sprintf(format, args...)}
}

func InternalErrE(message string, cause error) error {
	return &InternalError{Message: message, Cause: cause}
}

func IsInternalError(err error) bool {
	_, ok := err.(*InternalError)
	return ok
}
