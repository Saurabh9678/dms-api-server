package user

import "infiour.local/dms-api-server/pkg/database"

type User struct {
	ID          uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	Email       string `gorm:"type:varchar(100);null" json:"email"`
	PhoneNumber string `gorm:"type:varchar(50);not null" json:"phone_number"`
	CountryCode string `gorm:"type:varchar(5);not null" json:"country_code"`
	Name        string `gorm:"type:varchar(100);null" json:"name"`
	database.SoftDeleteableModel
}

func (User) TableName() string {
	return "users"
}
