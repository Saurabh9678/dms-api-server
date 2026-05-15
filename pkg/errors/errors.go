package errors

import "fmt"

type AppError struct {
	Code       string
	Message    string
	HTTPStatus int
	Wrapped    error
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	if e.Wrapped != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Wrapped)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Wrapped
}

func NewAppError(code string, message string, httpStatus int, wrapped error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Wrapped:    wrapped,
	}
}
