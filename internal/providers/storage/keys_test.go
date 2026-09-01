package storage_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	storageprovider "infiour.local/dms-api-server/internal/providers/storage"
)

func TestObjectKeyBuilders(t *testing.T) {
	at := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	assert.Equal(t, "7/showroom/EXT12345/20260101120000.jpg",
		storageprovider.ShowroomMediaObjectKey(7, "EXT12345", ".jpg", at))
	assert.Equal(t, "7/vehicle/18/20260101120000.png",
		storageprovider.VehicleImageObjectKey(7, 18, ".png", at))
	assert.Equal(t, "7/vehicle/18/docs/20260101120000.pdf",
		storageprovider.VehicleDocumentObjectKey(7, 18, ".pdf", at))
}
