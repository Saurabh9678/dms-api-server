package user

import "time"

type UserEntity struct {
	ID uint64 
	Email string 
	PhoneNumber string 
	CountryCode string
	Name string
	CreatedAt time.Time
	UpdatedAt time.Time 
	DeletedAt time.Time
}

type UserRoleEntity struct {
	ID uint64
	Type string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

type UserShowroomEntity struct {
	ID uint64
	UserID uint64
	ShowroomID uint64
	RoleID uint64
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}