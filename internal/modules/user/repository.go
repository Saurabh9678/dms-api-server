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
