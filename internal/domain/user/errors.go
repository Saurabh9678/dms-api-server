package user

import "errors"

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserRoleNotFound = errors.New("user role not found")
	ErrUserShowroomNotFound = errors.New("user showroom not found")
)

