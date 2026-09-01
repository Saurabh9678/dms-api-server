package storage

import (
	"fmt"
	"time"
)

// Object key timestamp segment format (YYYYMMDDHHmmss).
const ObjectKeyTimestampLayout = "20060102150405"

// Object key path format templates. Use the *ObjectKey helpers to build keys.
const (
	// ShowroomMediaObjectKeyFormat is {userID}/showroom/{externalShowroomID}/{timestamp}{ext}.
	ShowroomMediaObjectKeyFormat = "%d/showroom/%s/%s%s"
	// VehicleImageObjectKeyFormat is {userID}/vehicle/{vehicleID}/{timestamp}{ext}.
	VehicleImageObjectKeyFormat = "%d/vehicle/%d/%s%s"
	// VehicleDocumentObjectKeyFormat is {userID}/vehicle/{vehicleID}/docs/{timestamp}{ext}.
	VehicleDocumentObjectKeyFormat = "%d/vehicle/%d/docs/%s%s"
)

// ShowroomMediaObjectKey builds a showroom logo/banner object key.
func ShowroomMediaObjectKey(userID uint64, externalShowroomID, ext string, at time.Time) string {
	return fmt.Sprintf(ShowroomMediaObjectKeyFormat, userID, externalShowroomID, at.Format(ObjectKeyTimestampLayout), ext)
}

// VehicleImageObjectKey builds a vehicle photo object key.
func VehicleImageObjectKey(userID, vehicleID uint64, ext string, at time.Time) string {
	return fmt.Sprintf(VehicleImageObjectKeyFormat, userID, vehicleID, at.Format(ObjectKeyTimestampLayout), ext)
}

// VehicleDocumentObjectKey builds a vehicle document object key.
func VehicleDocumentObjectKey(userID, vehicleID uint64, ext string, at time.Time) string {
	return fmt.Sprintf(VehicleDocumentObjectKeyFormat, userID, vehicleID, at.Format(ObjectKeyTimestampLayout), ext)
}
