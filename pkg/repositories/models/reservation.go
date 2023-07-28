package models

import "time"

// ReservationKey uniquely identifies a reservation
type ReservationKey struct {
	DatasetProject string `gorm:"size:100;primary_key"`
	DatasetName    string `gorm:"size:100;primary_key"`
	DatasetDomain  string `gorm:"size:100;primary_key"`
	DatasetVersion string `gorm:"size:100;primary_key"`
	TagName        string `gorm:"size:100;primary_key"`
}

// Reservation tracks the metadata needed to allow
// task cache serialization
type Reservation struct {
	BaseModel
	ReservationKey

	// Identifies who owns the reservation
	OwnerID string `gorm:"size:100"`

	// When the reservation will expire
	ExpiresAt          time.Time
	SerializedMetadata []byte
}
