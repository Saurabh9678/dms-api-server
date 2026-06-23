package showroom

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"infiour.local/dms-api-server/pkg/database"
)

var ErrOwnerRoleNotFound = errors.New("owner role not found")

type userRole struct {
	ID   uint64 `gorm:"column:id"`
	Type string `gorm:"column:type"`
}

func (userRole) TableName() string { return "user_roles" }

type ownerRelation struct {
	UserID     uint64 `gorm:"column:user_id"`
	ShowroomID uint64 `gorm:"column:showroom_id"`
	RoleID     uint64 `gorm:"column:role_id"`
}

func (ownerRelation) TableName() string { return "user_showroom_relations" }

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// CreateWithOwner inserts a showroom and assigns the caller as owner in a single transaction.
func (r *Repository) CreateWithOwner(ctx context.Context, userID uint64, s *Showroom) (*Showroom, error) {
	var created Showroom
	err := database.RunInTx(ctx, r.db, func(tx *gorm.DB) error {
		model := *s
		if err := tx.WithContext(ctx).Create(&model).Error; err != nil {
			return err
		}
		created = model

		var role userRole
		if err := tx.WithContext(ctx).Where("type = ?", "owner").First(&role).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrOwnerRoleNotFound
			}
			return err
		}

		rel := ownerRelation{
			UserID:     userID,
			ShowroomID: created.ID,
			RoleID:     role.ID,
		}
		return tx.WithContext(ctx).Create(&rel).Error
	})
	if err != nil {
		return nil, err
	}
	return &created, nil
}

// UpdateFilePaths sets logo and/or banner paths on an existing showroom record.
func (r *Repository) UpdateFilePaths(ctx context.Context, showroomID uint64, logoPath, bannerPath *string) error {
	updates := map[string]any{}
	if logoPath != nil {
		updates["showroom_logo"] = *logoPath
	}
	if bannerPath != nil {
		updates["showroom_banner"] = *bannerPath
	}
	if len(updates) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&Showroom{}).Where("id = ?", showroomID).Updates(updates).Error
}
