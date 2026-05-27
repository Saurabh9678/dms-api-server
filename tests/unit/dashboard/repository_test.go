package dashboard_test

import (
	"testing"

	"gorm.io/gorm"
	"infiour.local/dms-api-server/internal/modules/dashboard"
)

func TestNewRepository(t *testing.T) {
	repo := dashboard.NewRepository(&gorm.DB{})
	if repo == nil {
		t.Fatal("expected non-nil repository")
	}
}
