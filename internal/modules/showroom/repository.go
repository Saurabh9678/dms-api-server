package showroom

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"infiour.local/dms-api-server/pkg/database"
)

var (
	ErrOwnerRoleNotFound  = errors.New("owner role not found")
	ErrTargetUserNotFound = errors.New("target user not found")
	ErrDuplicateMember    = errors.New("user already a member")
	ErrMemberNotFound     = errors.New("member not found")
	ErrMemberRoleNotFound = errors.New("role not found")
	ErrShowroomNotFound   = errors.New("showroom not found")
)

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

// userRecord is used to verify that a user exists without importing the user module.
type userRecord struct {
	ID uint64 `gorm:"column:id;primaryKey"`
}

func (userRecord) TableName() string { return "users" }

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

// GetByID returns an active (non-deleted) showroom by ID, or ErrShowroomNotFound.
func (r *Repository) GetByID(ctx context.Context, showroomID uint64) (*Showroom, error) {
	var s Showroom
	if err := r.db.WithContext(ctx).First(&s, showroomID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrShowroomNotFound
		}
		return nil, err
	}
	return &s, nil
}

// UpdateShowroomFields applies a partial update to an active showroom using a field map.
// Nil values in the map are written as SQL NULL. GORM automatically sets updated_at.
func (r *Repository) UpdateShowroomFields(ctx context.Context, showroomID uint64, updates map[string]any) error {
	return r.db.WithContext(ctx).
		Model(&Showroom{}).
		Where("id = ?", showroomID).
		Updates(updates).Error
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

// AddMember verifies the target user exists, looks up the role, checks for an existing active
// relation, and inserts a new relation — all in a single transaction.
func (r *Repository) AddMember(ctx context.Context, showroomID, targetUserID uint64, roleType string) error {
	return database.RunInTx(ctx, r.db, func(tx *gorm.DB) error {
		var u userRecord
		if err := tx.WithContext(ctx).First(&u, targetUserID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTargetUserNotFound
			}
			return err
		}

		var role userRole
		if err := tx.WithContext(ctx).Where("type = ?", roleType).First(&role).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrMemberRoleNotFound
			}
			return err
		}

		var count int64
		if err := tx.WithContext(ctx).
			Table("user_showroom_relations").
			Where("user_id = ? AND showroom_id = ? AND deleted_at IS NULL", targetUserID, showroomID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrDuplicateMember
		}

		rel := ownerRelation{
			UserID:     targetUserID,
			ShowroomID: showroomID,
			RoleID:     role.ID,
		}
		return tx.WithContext(ctx).Create(&rel).Error
	})
}

// ListMembers returns paginated members of a showroom with their user and role details.
func (r *Repository) ListMembers(ctx context.Context, showroomID uint64, page, limit int) ([]MemberRecord, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).
		Table("user_showroom_relations").
		Where("showroom_id = ? AND deleted_at IS NULL", showroomID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var records []MemberRecord
	offset := (page - 1) * limit
	err := r.db.WithContext(ctx).
		Table("user_showroom_relations usr").
		Select("usr.user_id, u.name, u.country_code, u.phone_number, ur.type AS role").
		Joins("JOIN users u ON u.id = usr.user_id").
		Joins("JOIN user_roles ur ON ur.id = usr.role_id").
		Where("usr.showroom_id = ? AND usr.deleted_at IS NULL", showroomID).
		Limit(limit).
		Offset(offset).
		Scan(&records).Error
	if err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

// GetMemberRole returns the active role of a user in a showroom, or ErrMemberNotFound.
func (r *Repository) GetMemberRole(ctx context.Context, showroomID, targetUserID uint64) (string, error) {
	type roleResult struct {
		Role string `gorm:"column:role"`
	}
	var results []roleResult
	err := r.db.WithContext(ctx).
		Table("user_showroom_relations usr").
		Select("ur.type AS role").
		Joins("JOIN user_roles ur ON ur.id = usr.role_id").
		Where("usr.user_id = ? AND usr.showroom_id = ? AND usr.deleted_at IS NULL", targetUserID, showroomID).
		Limit(1).
		Scan(&results).Error
	if err != nil {
		return "", err
	}
	if len(results) == 0 {
		return "", ErrMemberNotFound
	}
	return results[0].Role, nil
}

// RemoveMember soft-deletes the active relation for a user in a showroom.
func (r *Repository) RemoveMember(ctx context.Context, showroomID, targetUserID uint64) error {
	result := r.db.WithContext(ctx).
		Model(&ownerRelation{}).
		Where("user_id = ? AND showroom_id = ? AND deleted_at IS NULL", targetUserID, showroomID).
		Update("deleted_at", time.Now())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrMemberNotFound
	}
	return nil
}

// UpdateMemberRole changes the role of an existing active member in a single transaction.
func (r *Repository) UpdateMemberRole(ctx context.Context, showroomID, targetUserID uint64, newRoleType string) error {
	return database.RunInTx(ctx, r.db, func(tx *gorm.DB) error {
		var role userRole
		if err := tx.WithContext(ctx).Where("type = ?", newRoleType).First(&role).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrMemberRoleNotFound
			}
			return err
		}

		result := tx.WithContext(ctx).
			Model(&ownerRelation{}).
			Where("user_id = ? AND showroom_id = ? AND deleted_at IS NULL", targetUserID, showroomID).
			Update("role_id", role.ID)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrMemberNotFound
		}
		return nil
	})
}
