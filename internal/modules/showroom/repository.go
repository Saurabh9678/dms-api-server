package showroom

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"infiour.local/dms-api-server/pkg/database"
)

var (
	ErrOwnerRoleNotFound  = errors.New("owner role not found")
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

// userRecord is used for find-or-create membership without importing the user module.
type userRecord struct {
	ID          uint64 `gorm:"column:id;primaryKey"`
	Name        string `gorm:"column:name"`
	CountryCode string `gorm:"column:country_code"`
	PhoneNumber string `gorm:"column:phone_number"`
	database.SoftDeleteableModel
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

// ListByUserID returns active showrooms where the user has an active membership.
func (r *Repository) ListByUserID(ctx context.Context, userID uint64) ([]ShowroomListRecord, error) {
	var records []ShowroomListRecord
	err := r.db.WithContext(ctx).
		Table("user_showroom_relations usr").
		Select("s.id, s.showroom_id, s.name, ur.type AS role").
		Joins("JOIN showrooms s ON s.id = usr.showroom_id AND s.deleted_at IS NULL").
		Joins("JOIN user_roles ur ON ur.id = usr.role_id").
		Where("usr.user_id = ? AND usr.deleted_at IS NULL", userID).
		Scan(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
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

// AddMember find-or-creates an active user by phone, optionally fills an empty name, looks up
// the role, checks for an existing active relation, and inserts a new relation — all in one tx.
// Soft-deleted users with the same phone are ignored (a new active user is created).
// Re-assign after soft-delete: insert succeeds when the unique (user_id, showroom_id, role_id)
// is free (e.g. different role). On unique conflict, restores the soft-deleted matching row.
func (r *Repository) AddMember(ctx context.Context, showroomID uint64, name, countryCode, phoneNumber, roleType string) (uint64, error) {
	var userID uint64
	err := database.RunInTx(ctx, r.db, func(tx *gorm.DB) error {
		u, err := findOrCreateMemberUser(ctx, tx, name, countryCode, phoneNumber)
		if err != nil {
			return err
		}
		userID = u.ID

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
			Where("user_id = ? AND showroom_id = ? AND deleted_at IS NULL", userID, showroomID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrDuplicateMember
		}

		// Soft-deleted same (user, showroom, role) still occupies the unique key — restore first.
		restored, err := tryRestoreSoftDeletedMember(ctx, tx, userID, showroomID, role.ID)
		if err != nil {
			return err
		}
		if restored {
			return nil
		}

		rel := ownerRelation{
			UserID:     userID,
			ShowroomID: showroomID,
			RoleID:     role.ID,
		}
		if err := tx.WithContext(ctx).Create(&rel).Error; err != nil {
			if !isUniqueViolation(err) {
				return err
			}
			// Race or TranslateError path: unique conflict → restore soft-deleted row.
			return restoreSoftDeletedMember(ctx, tx, userID, showroomID, role.ID)
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func isUniqueViolation(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return true
	}
	// Postgres drivers / GORM may wrap without ErrDuplicatedKey when TranslateError is off.
	return strings.Contains(err.Error(), "SQLSTATE 23505")
}

// tryRestoreSoftDeletedMember clears deleted_at when a soft-deleted unique-key row exists.
func tryRestoreSoftDeletedMember(ctx context.Context, tx *gorm.DB, userID, showroomID, roleID uint64) (bool, error) {
	result := tx.WithContext(ctx).
		Table("user_showroom_relations").
		Where("user_id = ? AND showroom_id = ? AND role_id = ? AND deleted_at IS NOT NULL", userID, showroomID, roleID).
		Updates(map[string]any{
			"deleted_at": nil,
			"updated_at": time.Now(),
		})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// restoreSoftDeletedMember clears deleted_at on the soft-deleted relation matching the unique key.
func restoreSoftDeletedMember(ctx context.Context, tx *gorm.DB, userID, showroomID, roleID uint64) error {
	restored, err := tryRestoreSoftDeletedMember(ctx, tx, userID, showroomID, roleID)
	if err != nil {
		return err
	}
	if !restored {
		return ErrDuplicateMember
	}
	return nil
}

func findOrCreateMemberUser(ctx context.Context, tx *gorm.DB, name, countryCode, phoneNumber string) (*userRecord, error) {
	var u userRecord
	err := tx.WithContext(ctx).
		Where("country_code = ? AND phone_number = ?", countryCode, phoneNumber).
		First(&u).Error
	if err == nil {
		if strings.TrimSpace(u.Name) == "" && name != "" {
			if err := tx.WithContext(ctx).Model(&u).Update("name", name).Error; err != nil {
				return nil, err
			}
			u.Name = name
		}
		return &u, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	u = userRecord{
		Name:        name,
		CountryCode: countryCode,
		PhoneNumber: phoneNumber,
	}
	if err := tx.WithContext(ctx).Create(&u).Error; err != nil {
		if !errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, err
		}
		// Concurrent create won — re-fetch the winning active row.
		if err := tx.WithContext(ctx).
			Where("country_code = ? AND phone_number = ?", countryCode, phoneNumber).
			First(&u).Error; err != nil {
			return nil, err
		}
		if strings.TrimSpace(u.Name) == "" && name != "" {
			if err := tx.WithContext(ctx).Model(&u).Update("name", name).Error; err != nil {
				return nil, err
			}
			u.Name = name
		}
	}
	return &u, nil
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
