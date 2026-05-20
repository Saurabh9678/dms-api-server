package vehicle

import (
	"context"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, vehicle *Vehicle) (*Vehicle, error) {
	if err := r.db.WithContext(ctx).Create(vehicle).Error; err != nil {
		return nil, err
	}
	return vehicle, nil
}
