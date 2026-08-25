package inventory_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"infiour.local/dms-api-server/pkg/inventory"
)

func TestIsDeadStock(t *testing.T) {
	now := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)

	t.Run("nil buying date", func(t *testing.T) {
		assert.False(t, inventory.IsDeadStock(nil, false, now))
	})

	t.Run("zero buying date", func(t *testing.T) {
		zero := time.Time{}
		assert.False(t, inventory.IsDeadStock(&zero, false, now))
	})

	t.Run("has active sale", func(t *testing.T) {
		old := now.Add(-100 * 24 * time.Hour)
		assert.False(t, inventory.IsDeadStock(&old, true, now))
	})

	t.Run("exactly 90 days is not dead stock", func(t *testing.T) {
		d := now.Add(-90 * 24 * time.Hour)
		assert.False(t, inventory.IsDeadStock(&d, false, now))
	})

	t.Run("over 90 days is dead stock", func(t *testing.T) {
		d := now.Add(-91 * 24 * time.Hour)
		assert.True(t, inventory.IsDeadStock(&d, false, now))
	})
}

func TestDeadStockCaseSQL(t *testing.T) {
	sql := inventory.DeadStockCaseSQL("vp.buying_date")
	assert.Contains(t, sql, "vp.buying_date")
	assert.Contains(t, sql, "> 90")
	assert.Contains(t, sql, "CASE WHEN")
}

func TestDeadStockAgePredicateSQL(t *testing.T) {
	sql := inventory.DeadStockAgePredicateSQL("vp_buying_date")
	assert.Contains(t, sql, "vp_buying_date IS NOT NULL")
	assert.Contains(t, sql, "> 90")
}
