package models

import (
	"fmt"

	"github.com/gofrs/uuid"

	"database/sql/driver"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type UUIDString string

// As per https://gorm.io/docs/data_types.html, custom data types need to implement the following interfaces
func (s *UUIDString) Scan(value interface{}) error {
	switch src := value.(type) {
		case string:
			*s = UUIDString(src)
			return nil
		case []byte:
			*s = UUIDString(src[:])
			return nil
		default:
			return fmt.Errorf("failed to scan UUID. Type = %v", src)
	}
}

func (s UUIDString) Value() (driver.Value, error) {
	return string(s), nil
}

type DatasetKey struct {
	Project string     `gorm:"primary_key;"`                          // part of pkey, no index needed as it is first column in the pkey
	Name    string     `gorm:"primary_key;index:dataset_name_idx"`    // part of pkey and has separate index for filtering
	Domain  string     `gorm:"primary_key;index:dataset_domain_idx"`  // part of pkey and has separate index for filtering
	Version string     `gorm:"primary_key;index:dataset_version_idx"` // part of pkey and has separate index for filtering
	UUID    UUIDString `gorm:"unique;type:uuid"`
}

type Dataset struct {
	BaseModel
	DatasetKey
	SerializedMetadata []byte
	PartitionKeys      []PartitionKey `gorm:"references:UUID;foreignkey:DatasetUUID"`
}

type PartitionKey struct {
	BaseModel
	DatasetUUID UUIDString `gorm:"type:uuid;primary_key"`
	Name        string `gorm:"primary_key"`
}

// BeforeCreate so that we set the UUID in golang rather than from a DB function call
func (dataset *Dataset) BeforeCreate(tx *gorm.DB) error {
	if dataset.UUID == "" {
		generated, err := uuid.NewV4()
		if err != nil {
			return err
		}
		tx.Model(dataset).Update("UUID", generated)
	}
	return nil
}

func (dataset UUIDString) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	// use field.Tag, field.TagSettings gets field's tags
	// checkout https://github.com/go-gorm/gorm/blob/master/schema/field.go for all options
	if db.Dialector.Name() == "mysql" {
		return "varchar(36)"
	}
	return "uuid"
}
