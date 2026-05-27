package user

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) WithTx(tx *gorm.DB) *Repository {
	return &Repository{db: tx}
}

func (r *Repository) FindByID(ctx context.Context, userID uint64) (*User, error) {
	var model User
	err := r.db.WithContext(ctx).First(&model, userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &model, nil
}

func (r *Repository) FindByPhone(ctx context.Context, countryCode string, phoneNumber string) (*User, error) {
	var model User
	err := r.db.WithContext(ctx).
		Where("country_code = ? AND phone_number = ?", countryCode, phoneNumber).
		First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &model, nil
}

func (r *Repository) Create(ctx context.Context, record *User) (*User, error) {
	model := *record
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

func (r *Repository) FindShowroomRolesByUserID(ctx context.Context, userID uint64) ([]ShowroomRole, error) {
	var results []ShowroomRole
	err := r.db.WithContext(ctx).
		Table("user_showroom_relations usr").
		Select("usr.showroom_id, s.name AS showroom_name, ur.type AS role").
		Joins("JOIN showrooms s ON s.id = usr.showroom_id").
		Joins("JOIN user_roles ur ON ur.id = usr.role_id").
		Where("usr.user_id = ? AND usr.deleted_at IS NULL", userID).
		Scan(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (r *Repository) UpdateName(ctx context.Context, userID uint64, name string) error {
	result := r.db.WithContext(ctx).Model(&User{}).Where("id = ?", userID).Update("name", name)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}
