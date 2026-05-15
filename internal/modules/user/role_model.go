package user

import "infiour.local/dms-api-server/pkg/database"

type UserRole struct {
	ID   uint64       `gorm:"primaryKey;autoIncrement" json:"id"`
	Type UserRoleType `gorm:"type:varchar(255);not null" json:"type"`
	database.SoftDeleteableModel
}

type UserRoleType string

const (
	UserRoleTypeOwner    UserRoleType = "owner"
	UserRoleTypeManager  UserRoleType = "manager"
	UserRoleTypeEmployee UserRoleType = "employee"
)

func (UserRole) TableName() string {
	return "user_roles"
}
