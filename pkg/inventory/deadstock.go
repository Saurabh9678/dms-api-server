package inventory

import (
	"fmt"
	"time"
)

// DeadStockThresholdDays is the age (from buying_date) after which an unsold
// vehicle is counted as dead stock. Shared by dashboard and vehicle listing.
const DeadStockThresholdDays = 90

// IsDeadStock reports whether a vehicle is dead stock using the dashboard rule:
// no active sale, buying_date present, and age in days strictly greater than
// DeadStockThresholdDays.
func IsDeadStock(buyingDate *time.Time, hasActiveSale bool, now time.Time) bool {
	if hasActiveSale || buyingDate == nil || buyingDate.IsZero() {
		return false
	}
	ageDays := now.Sub(*buyingDate).Hours() / 24
	return ageDays > float64(DeadStockThresholdDays)
}

// DeadStockCaseSQL returns a SQL CASE expression that yields 1 when buyingDateCol
// exceeds the dead-stock age threshold, else 0. Callers must ensure the row is
// already unsold (e.g. dashboard inventory WHERE excludes sales).
func DeadStockCaseSQL(buyingDateCol string) string {
	return fmt.Sprintf(
		`CASE WHEN %s IS NOT NULL AND EXTRACT(EPOCH FROM (NOW() - %s)) / 86400 > %d THEN 1 ELSE 0 END`,
		buyingDateCol, buyingDateCol, DeadStockThresholdDays,
	)
}

// DeadStockAgePredicateSQL returns a boolean SQL predicate for dead-stock age
// on buyingDateCol (does not check sale status).
func DeadStockAgePredicateSQL(buyingDateCol string) string {
	return fmt.Sprintf(
		`%s IS NOT NULL AND EXTRACT(EPOCH FROM (NOW() - %s)) / 86400 > %d`,
		buyingDateCol, buyingDateCol, DeadStockThresholdDays,
	)
}
