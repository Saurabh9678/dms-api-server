package vehicle_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"infiour.local/dms-api-server/internal/modules/vehicle"
)

func TestNewRepository(t *testing.T) {
	repo := vehicle.NewRepository(nil)
	assert.NotNil(t, repo)
}

func TestErrVehicleNotFound(t *testing.T) {
	assert.EqualError(t, vehicle.ErrVehicleNotFound, "vehicle not found")
}
