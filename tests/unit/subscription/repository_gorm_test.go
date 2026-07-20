package subscription_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"infiour.local/dms-api-server/internal/modules/subscription"
)

func newSubscriptionMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	return gormDB, mock
}

func TestNewRepositoryAndWithTx(t *testing.T) {
	gormDB, _ := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)
	assert.NotNil(t, repo)
	txRepo := repo.WithTx(gormDB)
	assert.NotNil(t, txRepo)
}

// ─── Plans ────────────────────────────────────────────────────────────────────

func TestCreatePlan_Success(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)
	planID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "subscription_plans"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(planID))
	mock.ExpectCommit()

	created, err := repo.CreatePlan(context.Background(), &subscription.SubscriptionPlan{
		Code: "basic",
		Name: "Basic",
	})
	assert.NoError(t, err)
	require.NotNil(t, created)
}

func TestCreatePlan_Error(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "subscription_plans"`).
		WillReturnError(gorm.ErrInvalidData)
	mock.ExpectRollback()

	_, err := repo.CreatePlan(context.Background(), &subscription.SubscriptionPlan{Code: "x", Name: "X"})
	assert.Error(t, err)
}

func TestGetPlanByID_Success(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)
	planID := uuid.New()

	mock.ExpectQuery(`SELECT \* FROM "subscription_plans"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name", "is_free", "is_global", "is_active", "display_order"}).
			AddRow(planID, "basic", "Basic", false, true, true, 0))

	plan, err := repo.GetPlanByID(context.Background(), planID)
	assert.NoError(t, err)
	assert.Equal(t, planID, plan.ID)
}

func TestGetPlanByID_NotFound(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "subscription_plans"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := repo.GetPlanByID(context.Background(), uuid.New())
	assert.ErrorIs(t, err, subscription.ErrPlanNotFound)
}

func TestGetPlanByID_DBError(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "subscription_plans"`).
		WillReturnError(gorm.ErrInvalidData)

	_, err := repo.GetPlanByID(context.Background(), uuid.New())
	assert.Error(t, err)
	assert.NotErrorIs(t, err, subscription.ErrPlanNotFound)
}

func TestGetPlanByCode_Success(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)
	planID := uuid.New()

	mock.ExpectQuery(`SELECT \* FROM "subscription_plans"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name"}).AddRow(planID, "pro", "Pro"))

	plan, err := repo.GetPlanByCode(context.Background(), "pro")
	assert.NoError(t, err)
	assert.Equal(t, "pro", plan.Code)
}

func TestGetPlanByCode_NotFound(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "subscription_plans"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := repo.GetPlanByCode(context.Background(), "missing")
	assert.ErrorIs(t, err, subscription.ErrPlanNotFound)
}

func TestGetPlanByCode_DBError(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "subscription_plans"`).
		WillReturnError(gorm.ErrInvalidData)

	_, err := repo.GetPlanByCode(context.Background(), "pro")
	assert.Error(t, err)
}

func TestListActivePlans_GlobalOnly(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "subscription_plans"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name"}).AddRow(uuid.New(), "basic", "Basic"))

	plans, err := repo.ListActivePlans(context.Background(), true)
	assert.NoError(t, err)
	assert.Len(t, plans, 1)
}

func TestListActivePlans_AllActive(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "subscription_plans"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name"}))

	plans, err := repo.ListActivePlans(context.Background(), false)
	assert.NoError(t, err)
	assert.Empty(t, plans)
}

func TestListActivePlans_Error(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "subscription_plans"`).
		WillReturnError(gorm.ErrInvalidData)

	_, err := repo.ListActivePlans(context.Background(), false)
	assert.Error(t, err)
}

// ─── Features ─────────────────────────────────────────────────────────────────

func TestCreateFeature_Success(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "subscription_features"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectCommit()

	created, err := repo.CreateFeature(context.Background(), &subscription.SubscriptionFeature{
		Key:       "vehicles",
		Name:      "Vehicle limit",
		ValueType: subscription.FeatureValueTypeNumber,
	})
	assert.NoError(t, err)
	require.NotNil(t, created)
}

func TestCreateFeature_Error(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "subscription_features"`).
		WillReturnError(gorm.ErrInvalidData)
	mock.ExpectRollback()

	_, err := repo.CreateFeature(context.Background(), &subscription.SubscriptionFeature{
		Key: "x", Name: "X", ValueType: subscription.FeatureValueTypeBoolean,
	})
	assert.Error(t, err)
}

func TestGetFeatureByID_Success(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)
	featureID := uuid.New()

	mock.ExpectQuery(`SELECT \* FROM "subscription_features"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "key", "name", "value_type"}).
			AddRow(featureID, "vehicles", "Vehicle limit", "NUMBER"))

	feature, err := repo.GetFeatureByID(context.Background(), featureID)
	assert.NoError(t, err)
	assert.Equal(t, featureID, feature.ID)
}

func TestGetFeatureByID_NotFound(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "subscription_features"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := repo.GetFeatureByID(context.Background(), uuid.New())
	assert.ErrorIs(t, err, subscription.ErrFeatureNotFound)
}

func TestGetFeatureByID_DBError(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "subscription_features"`).
		WillReturnError(gorm.ErrInvalidData)

	_, err := repo.GetFeatureByID(context.Background(), uuid.New())
	assert.Error(t, err)
}

func TestGetFeatureByKey_Success(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "subscription_features"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "key", "name", "value_type"}).
			AddRow(uuid.New(), "vehicles", "Vehicle limit", "NUMBER"))

	feature, err := repo.GetFeatureByKey(context.Background(), "vehicles")
	assert.NoError(t, err)
	assert.Equal(t, "vehicles", feature.Key)
}

func TestGetFeatureByKey_NotFound(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "subscription_features"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := repo.GetFeatureByKey(context.Background(), "missing")
	assert.ErrorIs(t, err, subscription.ErrFeatureNotFound)
}

func TestGetFeatureByKey_DBError(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "subscription_features"`).
		WillReturnError(gorm.ErrInvalidData)

	_, err := repo.GetFeatureByKey(context.Background(), "vehicles")
	assert.Error(t, err)
}

func TestListFeatures_Success(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "subscription_features"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "key", "name"}))

	features, err := repo.ListFeatures(context.Background())
	assert.NoError(t, err)
	assert.Empty(t, features)
}

func TestListFeatures_Error(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "subscription_features"`).
		WillReturnError(gorm.ErrInvalidData)

	_, err := repo.ListFeatures(context.Background())
	assert.Error(t, err)
}

// ─── Pricing ──────────────────────────────────────────────────────────────────

func TestCreatePricing_Success(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "subscription_pricing"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectCommit()

	created, err := repo.CreatePricing(context.Background(), &subscription.SubscriptionPricing{
		SubscriptionPlanID: uuid.New(),
		BillingCycle:       subscription.BillingCycleMonthly,
		DurationDays:       30,
		Price:              999,
	})
	assert.NoError(t, err)
	require.NotNil(t, created)
}

func TestCreatePricing_Error(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "subscription_pricing"`).
		WillReturnError(gorm.ErrInvalidData)
	mock.ExpectRollback()

	_, err := repo.CreatePricing(context.Background(), &subscription.SubscriptionPricing{
		SubscriptionPlanID: uuid.New(),
		BillingCycle:       subscription.BillingCycleMonthly,
		DurationDays:       30,
		Price:              999,
	})
	assert.Error(t, err)
}

func TestGetPricingByID_Success(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)
	pricingID := uuid.New()

	mock.ExpectQuery(`SELECT \* FROM "subscription_pricing"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "billing_cycle", "duration_days", "price"}).
			AddRow(pricingID, "MONTHLY", 30, 999))

	pricing, err := repo.GetPricingByID(context.Background(), pricingID)
	assert.NoError(t, err)
	assert.Equal(t, pricingID, pricing.ID)
}

func TestGetPricingByID_NotFound(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "subscription_pricing"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := repo.GetPricingByID(context.Background(), uuid.New())
	assert.ErrorIs(t, err, subscription.ErrPricingNotFound)
}

func TestGetPricingByID_DBError(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "subscription_pricing"`).
		WillReturnError(gorm.ErrInvalidData)

	_, err := repo.GetPricingByID(context.Background(), uuid.New())
	assert.Error(t, err)
}

func TestListPricingByPlanID_ActiveOnly(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "subscription_pricing"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "subscription_plan_id"}))

	pricing, err := repo.ListPricingByPlanID(context.Background(), uuid.New(), true)
	assert.NoError(t, err)
	assert.Empty(t, pricing)
}

func TestListPricingByPlanID_All(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "subscription_pricing"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "subscription_plan_id"}).
			AddRow(uuid.New(), uuid.New()))

	pricing, err := repo.ListPricingByPlanID(context.Background(), uuid.New(), false)
	assert.NoError(t, err)
	assert.Len(t, pricing, 1)
}

func TestListPricingByPlanID_Error(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "subscription_pricing"`).
		WillReturnError(gorm.ErrInvalidData)

	_, err := repo.ListPricingByPlanID(context.Background(), uuid.New(), false)
	assert.Error(t, err)
}

// ─── Plan features ────────────────────────────────────────────────────────────

func TestUpsertPlanFeature_Create(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)
	planID := uuid.New()
	featureID := uuid.New()
	pfID := uuid.New()
	val := float64(10)

	mock.ExpectQuery(`SELECT \* FROM "plan_features"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "plan_features"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(pfID))
	mock.ExpectCommit()

	created, err := repo.UpsertPlanFeature(context.Background(), &subscription.PlanFeature{
		SubscriptionPlanID:    planID,
		SubscriptionFeatureID: featureID,
		NumericValue:          &val,
	})
	assert.NoError(t, err)
	require.NotNil(t, created)
}

func TestUpsertPlanFeature_Update(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)
	pfID := uuid.New()
	planID := uuid.New()
	featureID := uuid.New()
	val := float64(20)

	mock.ExpectQuery(`SELECT \* FROM "plan_features"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "subscription_plan_id", "subscription_feature_id"}).
			AddRow(pfID, planID, featureID))
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "plan_features"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectQuery(`SELECT \* FROM "plan_features"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "numeric_value"}).AddRow(pfID, val))

	updated, err := repo.UpsertPlanFeature(context.Background(), &subscription.PlanFeature{
		SubscriptionPlanID:    planID,
		SubscriptionFeatureID: featureID,
		NumericValue:          &val,
	})
	assert.NoError(t, err)
	require.NotNil(t, updated)
}

func TestUpsertPlanFeature_LookupError(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "plan_features"`).
		WillReturnError(gorm.ErrInvalidData)

	_, err := repo.UpsertPlanFeature(context.Background(), &subscription.PlanFeature{
		SubscriptionPlanID:    uuid.New(),
		SubscriptionFeatureID: uuid.New(),
	})
	assert.Error(t, err)
}

func TestUpsertPlanFeature_CreateError(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "plan_features"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "plan_features"`).
		WillReturnError(gorm.ErrInvalidData)
	mock.ExpectRollback()

	_, err := repo.UpsertPlanFeature(context.Background(), &subscription.PlanFeature{
		SubscriptionPlanID:    uuid.New(),
		SubscriptionFeatureID: uuid.New(),
	})
	assert.Error(t, err)
}

func TestUpsertPlanFeature_UpdateError(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)
	pfID := uuid.New()

	mock.ExpectQuery(`SELECT \* FROM "plan_features"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "subscription_plan_id", "subscription_feature_id"}).
			AddRow(pfID, uuid.New(), uuid.New()))
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "plan_features"`).
		WillReturnError(gorm.ErrInvalidData)
	mock.ExpectRollback()

	_, err := repo.UpsertPlanFeature(context.Background(), &subscription.PlanFeature{
		SubscriptionPlanID:    uuid.New(),
		SubscriptionFeatureID: uuid.New(),
	})
	assert.Error(t, err)
}

func TestGetPlanFeatureByID_Success(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)
	pfID := uuid.New()

	mock.ExpectQuery(`SELECT \* FROM "plan_features"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(pfID))

	pf, err := repo.GetPlanFeatureByID(context.Background(), pfID)
	assert.NoError(t, err)
	assert.Equal(t, pfID, pf.ID)
}

func TestGetPlanFeatureByID_NotFound(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "plan_features"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := repo.GetPlanFeatureByID(context.Background(), uuid.New())
	assert.ErrorIs(t, err, subscription.ErrPlanFeatureNotFound)
}

func TestGetPlanFeatureByID_DBError(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "plan_features"`).
		WillReturnError(gorm.ErrInvalidData)

	_, err := repo.GetPlanFeatureByID(context.Background(), uuid.New())
	assert.Error(t, err)
}

func TestListPlanFeaturesByPlanID_Success(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "plan_features"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "subscription_plan_id"}))

	features, err := repo.ListPlanFeaturesByPlanID(context.Background(), uuid.New())
	assert.NoError(t, err)
	assert.Empty(t, features)
}

func TestListPlanFeaturesByPlanID_Error(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "plan_features"`).
		WillReturnError(gorm.ErrInvalidData)

	_, err := repo.ListPlanFeaturesByPlanID(context.Background(), uuid.New())
	assert.Error(t, err)
}

// ─── Dealer subscriptions ─────────────────────────────────────────────────────

func TestCreateDealerSubscription_Success(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "dealer_subscriptions"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectCommit()

	created, err := repo.CreateDealerSubscription(context.Background(), &subscription.DealerSubscription{
		DealerID:              1,
		SubscriptionPlanID:    uuid.New(),
		SubscriptionPricingID: uuid.New(),
		Status:                subscription.SubscriptionStatusPending,
		StartsAt:              now,
		ExpiresAt:             now.AddDate(0, 1, 0),
	})
	assert.NoError(t, err)
	require.NotNil(t, created)
}

func TestCreateDealerSubscription_Error(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "dealer_subscriptions"`).
		WillReturnError(gorm.ErrInvalidData)
	mock.ExpectRollback()

	_, err := repo.CreateDealerSubscription(context.Background(), &subscription.DealerSubscription{
		DealerID: 1, Status: subscription.SubscriptionStatusPending,
	})
	assert.Error(t, err)
}

func TestGetDealerSubscriptionByID_Success(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)
	subID := uuid.New()

	mock.ExpectQuery(`SELECT \* FROM "dealer_subscriptions"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "dealer_id", "status"}).
			AddRow(subID, 1, "ACTIVE"))

	sub, err := repo.GetDealerSubscriptionByID(context.Background(), subID)
	assert.NoError(t, err)
	assert.Equal(t, subID, sub.ID)
}

func TestGetDealerSubscriptionByID_NotFound(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "dealer_subscriptions"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := repo.GetDealerSubscriptionByID(context.Background(), uuid.New())
	assert.ErrorIs(t, err, subscription.ErrDealerSubscriptionNotFound)
}

func TestGetDealerSubscriptionByID_DBError(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "dealer_subscriptions"`).
		WillReturnError(gorm.ErrInvalidData)

	_, err := repo.GetDealerSubscriptionByID(context.Background(), uuid.New())
	assert.Error(t, err)
}

func TestGetActiveDealerSubscription_Success(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)
	subID := uuid.New()

	mock.ExpectQuery(`SELECT \* FROM "dealer_subscriptions"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "dealer_id", "status"}).
			AddRow(subID, 1, "ACTIVE"))

	sub, err := repo.GetActiveDealerSubscription(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, subscription.SubscriptionStatusActive, sub.Status)
}

func TestGetActiveDealerSubscription_NotFound(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "dealer_subscriptions"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := repo.GetActiveDealerSubscription(context.Background(), 1)
	assert.ErrorIs(t, err, subscription.ErrDealerSubscriptionNotFound)
}

func TestGetActiveDealerSubscription_DBError(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "dealer_subscriptions"`).
		WillReturnError(gorm.ErrInvalidData)

	_, err := repo.GetActiveDealerSubscription(context.Background(), 1)
	assert.Error(t, err)
}

func TestUpdateDealerSubscriptionStatus_Success(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "dealer_subscriptions"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.UpdateDealerSubscriptionStatus(context.Background(), uuid.New(), subscription.SubscriptionStatusCancelled)
	assert.NoError(t, err)
}

func TestUpdateDealerSubscriptionStatus_NotFound(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "dealer_subscriptions"`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := repo.UpdateDealerSubscriptionStatus(context.Background(), uuid.New(), subscription.SubscriptionStatusCancelled)
	assert.ErrorIs(t, err, subscription.ErrDealerSubscriptionNotFound)
}

func TestUpdateDealerSubscriptionStatus_DBError(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "dealer_subscriptions"`).
		WillReturnError(gorm.ErrInvalidData)
	mock.ExpectRollback()

	err := repo.UpdateDealerSubscriptionStatus(context.Background(), uuid.New(), subscription.SubscriptionStatusCancelled)
	assert.Error(t, err)
}

func TestListDealerSubscriptions_Success(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "dealer_subscriptions"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "dealer_id"}).AddRow(uuid.New(), 1))

	subs, err := repo.ListDealerSubscriptions(context.Background(), 1)
	assert.NoError(t, err)
	assert.Len(t, subs, 1)
}

func TestListDealerSubscriptions_Error(t *testing.T) {
	gormDB, mock := newSubscriptionMockDB(t)
	repo := subscription.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "dealer_subscriptions"`).
		WillReturnError(gorm.ErrInvalidData)

	_, err := repo.ListDealerSubscriptions(context.Background(), 1)
	assert.Error(t, err)
}

func TestSubscriptionErrors(t *testing.T) {
	assert.Error(t, subscription.ErrPlanNotFound)
	assert.Error(t, subscription.ErrFeatureNotFound)
	assert.Error(t, subscription.ErrPricingNotFound)
	assert.Error(t, subscription.ErrPlanFeatureNotFound)
	assert.Error(t, subscription.ErrDealerSubscriptionNotFound)
}
