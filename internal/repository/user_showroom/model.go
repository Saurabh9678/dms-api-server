package usershowroom

import (
	"infiour.local/dms-api-server/internal/repository"
	"infiour.local/dms-api-server/internal/repository/showroom"
	"infiour.local/dms-api-server/internal/repository/user_role"
	"infiour.local/dms-api-server/internal/repository/users"
)

type UserShowroom struct {
	ID         uint64            `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint64            `gorm:"not null" json:"user_id"`
	ShowroomID uint64            `gorm:"not null" json:"showroom_id"`
	RoleID     uint64            `gorm:"not null" json:"role_id"`
	User       users.User        `gorm:"foreignKey:UserID;references:ID" json:"user"`
	Showroom   showroom.Showroom `gorm:"foreignKey:ShowroomID;references:ID" json:"showroom"`
	Role       userrole.UserRole `gorm:"foreignKey:RoleID;references:ID" json:"role"`
	repository.SoftDeleteableModel
}

func (UserShowroom) TableName() string {
	return "user_showroom_relations"
}