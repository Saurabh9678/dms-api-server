# Subscription Module

## Ownership

- Subscription catalog: plans, features, pricing tiers, and plan-feature mappings.
- Dealer subscription lifecycle records tied to authenticated users (`users.id` as `dealer_id`).

## Location

- `internal/modules/subscription/`

## Layer Structure

| Layer | Files | Responsibility |
|-------|-------|----------------|
| Models | `plan_model.go`, `feature_model.go`, `pricing_model.go`, `plan_feature_model.go`, `dealer_subscription_model.go` | GORM entities and enum types |
| Repository | `repository.go` | Persistence for catalog and dealer subscriptions |
| Errors | `errors.go` | Module-level sentinel errors |

## Database Tables

- `subscription_plans` — plan catalog (`is_global`, optional `customer_id` for custom plans)
- `subscription_features` — feature definitions with typed values
- `subscription_pricing` — billing-cycle pricing per plan
- `plan_features` — plan-to-feature value mappings
- `dealer_subscriptions` — active/historical dealer subscriptions

See `docs/database/tables/subscription_*.md` and `docs/database/tables/dealer_subscriptions.md`.

## Migrations

- `000021_create_subscription_plans_table`
- `000022_create_subscription_features_table`
- `000023_create_subscription_pricing_table`
- `000024_create_plan_features_table`
- `000025_create_dealer_subscriptions_table`

## API Endpoints

None yet. HTTP handlers and routes will be added when subscription APIs are defined.

## Repository Methods

### Plans

- `CreatePlan`, `GetPlanByID`, `GetPlanByCode`, `ListActivePlans(globalOnly bool)`

### Features

- `CreateFeature`, `GetFeatureByID`, `GetFeatureByKey`, `ListFeatures`

### Pricing

- `CreatePricing`, `GetPricingByID`, `ListPricingByPlanID(planID, activeOnly bool)`

### Plan features

- `UpsertPlanFeature`, `GetPlanFeatureByID`, `ListPlanFeaturesByPlanID`

### Dealer subscriptions

- `CreateDealerSubscription`, `GetDealerSubscriptionByID`, `GetActiveDealerSubscription(dealerID)`
- `UpdateDealerSubscriptionStatus`, `ListDealerSubscriptions(dealerID)`

## Notes

- `dealer_id` is `BIGINT` and references `users.id` (dealers are authenticated users).
- Plan primary keys and most subscription entities use `UUID`; user IDs remain `BIGSERIAL`.
- `subscription_plans.customer_id` is `UUID` with no FK yet — `customers.id` is still `BIGSERIAL`.
