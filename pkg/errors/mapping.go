package errors

import "net/http"

type Mapper func(err error) (*AppError, bool)

var mappers []Mapper

func RegisterMapper(mapper Mapper) {
	mappers = append(mappers, mapper)
}

func ToAppError(err error) *AppError {
	if err == nil {
		return nil
	}
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	for _, mapper := range mappers {
		if mapped, ok := mapper(err); ok {
			return mapped
		}
	}
	return &AppError{
		Code:       CodeInternal,
		Message:    "internal server error",
		HTTPStatus: http.StatusInternalServerError,
		Wrapped:    err,
	}
}
