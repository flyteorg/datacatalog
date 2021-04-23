package models

import "time"

type ReservationKey struct {
	DatasetProject string `gorm:"primary_key"`
	DatasetName    string `gorm:"primary_key"`
	DatasetDomain  string `gorm:"primary_key"`
	DatasetVersion string `gorm:"primary_key"`
	TagName        string `gorm:"primary_key"`
}

type Reservation struct {
	BaseModel
	ReservationKey
	OwnerID            string
	ExpireAt           time.Time
	SerializedMetadata []byte
}
