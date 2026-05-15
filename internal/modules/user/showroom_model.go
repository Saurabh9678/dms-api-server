package user

import (
	"infiour.local/dms-api-server/internal/modules/showroom"
	"infiour.local/dms-api-server/pkg/database"
)

type UserShowroom struct {
	ID         uint64            `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint64            `gorm:"not null" json:"user_id"`
	ShowroomID uint64            `gorm:"not null" json:"showroom_id"`
	RoleID     uint64            `gorm:"not null" json:"role_id"`
	User       User              `gorm:"foreignKey:UserID;references:ID" json:"user"`
	Showroom   showroom.Showroom `gorm:"foreignKey:ShowroomID;references:ID" json:"showroom"`
	Role       UserRole          `gorm:"foreignKey:RoleID;references:ID" json:"role"`
	database.SoftDeleteableModel
}

func (UserShowroom) TableName() string {
	return "user_showroom_relations"
}
