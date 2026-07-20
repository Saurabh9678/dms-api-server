package subscription

import (
	"context"
	"errors"

	"github.com/google/uuid"
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

func (r *Repository) CreatePlan(ctx context.Context, plan *SubscriptionPlan) (*SubscriptionPlan, error) {
	model := *plan
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

func (r *Repository) GetPlanByID(ctx context.Context, id uuid.UUID) (*SubscriptionPlan, error) {
	var plan SubscriptionPlan
	if err := r.db.WithContext(ctx).First(&plan, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPlanNotFound
		}
		return nil, err
	}
	return &plan, nil
}

func (r *Repository) GetPlanByCode(ctx context.Context, code string) (*SubscriptionPlan, error) {
	var plan SubscriptionPlan
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&plan).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPlanNotFound
		}
		return nil, err
	}
	return &plan, nil
}

func (r *Repository) ListActivePlans(ctx context.Context, globalOnly bool) ([]SubscriptionPlan, error) {
	q := r.db.WithContext(ctx).Where("is_active = ?", true)
	if globalOnly {
		q = q.Where("is_global = ?", true)
	}
	var plans []SubscriptionPlan
	if err := q.Order("display_order ASC, name ASC").Find(&plans).Error; err != nil {
		return nil, err
	}
	return plans, nil
}

func (r *Repository) CreateFeature(ctx context.Context, feature *SubscriptionFeature) (*SubscriptionFeature, error) {
	model := *feature
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

func (r *Repository) GetFeatureByID(ctx context.Context, id uuid.UUID) (*SubscriptionFeature, error) {
	var feature SubscriptionFeature
	if err := r.db.WithContext(ctx).First(&feature, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFeatureNotFound
		}
		return nil, err
	}
	return &feature, nil
}

func (r *Repository) GetFeatureByKey(ctx context.Context, key string) (*SubscriptionFeature, error) {
	var feature SubscriptionFeature
	if err := r.db.WithContext(ctx).Where("key = ?", key).First(&feature).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFeatureNotFound
		}
		return nil, err
	}
	return &feature, nil
}

func (r *Repository) ListFeatures(ctx context.Context) ([]SubscriptionFeature, error) {
	var features []SubscriptionFeature
	if err := r.db.WithContext(ctx).Order("category ASC NULLS LAST, name ASC").Find(&features).Error; err != nil {
		return nil, err
	}
	return features, nil
}

func (r *Repository) CreatePricing(ctx context.Context, pricing *SubscriptionPricing) (*SubscriptionPricing, error) {
	model := *pricing
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

func (r *Repository) GetPricingByID(ctx context.Context, id uuid.UUID) (*SubscriptionPricing, error) {
	var pricing SubscriptionPricing
	if err := r.db.WithContext(ctx).First(&pricing, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPricingNotFound
		}
		return nil, err
	}
	return &pricing, nil
}

func (r *Repository) ListPricingByPlanID(ctx context.Context, planID uuid.UUID, activeOnly bool) ([]SubscriptionPricing, error) {
	q := r.db.WithContext(ctx).Where("subscription_plan_id = ?", planID)
	if activeOnly {
		q = q.Where("is_active = ?", true)
	}
	var pricing []SubscriptionPricing
	if err := q.Order("duration_days ASC").Find(&pricing).Error; err != nil {
		return nil, err
	}
	return pricing, nil
}

func (r *Repository) UpsertPlanFeature(ctx context.Context, pf *PlanFeature) (*PlanFeature, error) {
	var existing PlanFeature
	err := r.db.WithContext(ctx).
		Where("subscription_plan_id = ? AND subscription_feature_id = ?", pf.SubscriptionPlanID, pf.SubscriptionFeatureID).
		First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		model := *pf
		if createErr := r.db.WithContext(ctx).Create(&model).Error; createErr != nil {
			return nil, createErr
		}
		return &model, nil
	}
	if err != nil {
		return nil, err
	}

	updates := map[string]any{
		"bool_value":    pf.BoolValue,
		"numeric_value": pf.NumericValue,
		"string_value":  pf.StringValue,
	}
	if updateErr := r.db.WithContext(ctx).Model(&existing).Updates(updates).Error; updateErr != nil {
		return nil, updateErr
	}
	return r.GetPlanFeatureByID(ctx, existing.ID)
}

func (r *Repository) GetPlanFeatureByID(ctx context.Context, id uuid.UUID) (*PlanFeature, error) {
	var pf PlanFeature
	if err := r.db.WithContext(ctx).First(&pf, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPlanFeatureNotFound
		}
		return nil, err
	}
	return &pf, nil
}

func (r *Repository) ListPlanFeaturesByPlanID(ctx context.Context, planID uuid.UUID) ([]PlanFeature, error) {
	var features []PlanFeature
	if err := r.db.WithContext(ctx).
		Where("subscription_plan_id = ?", planID).
		Find(&features).Error; err != nil {
		return nil, err
	}
	return features, nil
}

func (r *Repository) CreateDealerSubscription(ctx context.Context, sub *DealerSubscription) (*DealerSubscription, error) {
	model := *sub
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

func (r *Repository) GetDealerSubscriptionByID(ctx context.Context, id uuid.UUID) (*DealerSubscription, error) {
	var sub DealerSubscription
	if err := r.db.WithContext(ctx).First(&sub, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDealerSubscriptionNotFound
		}
		return nil, err
	}
	return &sub, nil
}

func (r *Repository) GetActiveDealerSubscription(ctx context.Context, dealerID uint64) (*DealerSubscription, error) {
	var sub DealerSubscription
	err := r.db.WithContext(ctx).
		Where("dealer_id = ? AND status = ?", dealerID, SubscriptionStatusActive).
		Order("expires_at DESC").
		First(&sub).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrDealerSubscriptionNotFound
	}
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *Repository) UpdateDealerSubscriptionStatus(ctx context.Context, id uuid.UUID, status SubscriptionStatus) error {
	result := r.db.WithContext(ctx).
		Model(&DealerSubscription{}).
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrDealerSubscriptionNotFound
	}
	return nil
}

func (r *Repository) ListDealerSubscriptions(ctx context.Context, dealerID uint64) ([]DealerSubscription, error) {
	var subs []DealerSubscription
	if err := r.db.WithContext(ctx).
		Where("dealer_id = ?", dealerID).
		Order("created_at DESC").
		Find(&subs).Error; err != nil {
		return nil, err
	}
	return subs, nil
}
