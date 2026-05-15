package users

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"infiour.local/dms-api-server/internal/domain/user"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByPhone(ctx context.Context, countryCode string, phoneNumber string) (*user.UserEntity, error) {
	var model User
	err := r.db.WithContext(ctx).
		Where("country_code = ? AND phone_number = ?", countryCode, phoneNumber).
		First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, user.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return toDomain(&model), nil
}

func (r *Repository) Create(ctx context.Context, entity *user.UserEntity) (*user.UserEntity, error) {
	model := User{
		Email:       entity.Email,
		PhoneNumber: entity.PhoneNumber,
		CountryCode: entity.CountryCode,
		Name:        entity.Name,
	}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, err
	}
	return toDomain(&model), nil
}

func toDomain(model *User) *user.UserEntity {
	var deletedAt *time.Time
	if model.DeletedAt.Valid {
		deletedAt = &model.DeletedAt.Time
	}

	return &user.UserEntity{
		ID:          model.ID,
		Email:       model.Email,
		PhoneNumber: model.PhoneNumber,
		CountryCode: model.CountryCode,
		Name:        model.Name,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		DeletedAt:   deletedAt,
	}
}
