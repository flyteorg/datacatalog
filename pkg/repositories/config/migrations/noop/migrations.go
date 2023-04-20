package noopmigrations

import (
	"database/sql/driver"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type UUIDString string

func (uuidString UUIDString) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	// use field.Tag, field.TagSettings gets field's tags
	// checkout https://github.com/go-gorm/gorm/blob/master/schema/field.go for all options
	if db.Dialector.Name() == "mysql" {
		return "varchar(36)"
	}
	return "uuid"
}

func (uuidString *UUIDString) Scan(value interface{}) error {
	return nil
}

// Value return json value, implement driver.Valuer interface
func (uuidString UUIDString) Value() (driver.Value, error) {
	return uuidString, nil
}

type BaseModel struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}

type ArtifactKey struct {
	DatasetProject string `gorm:"primary_key"`
	DatasetName    string `gorm:"primary_key"`
	DatasetDomain  string `gorm:"primary_key"`
	DatasetVersion string `gorm:"primary_key"`
	ArtifactID     string `gorm:"primary_key"`
}

type Artifact struct {
	BaseModel
	ArtifactKey
	DatasetUUID        UUIDString     `gorm:"index:artifacts_dataset_uuid_idx"`
	Dataset            Dataset        `gorm:"association_autocreate:false"`
	ArtifactData       []ArtifactData `gorm:"references:DatasetProject,DatasetName,DatasetDomain,DatasetVersion,ArtifactID;foreignkey:DatasetProject,DatasetName,DatasetDomain,DatasetVersion,ArtifactID"`
	Partitions         []Partition    `gorm:"references:ArtifactID;foreignkey:ArtifactID"`
	Tags               []Tag          `gorm:"references:ArtifactID,DatasetUUID;foreignkey:ArtifactID,DatasetUUID"`
	SerializedMetadata []byte
}

type ArtifactData struct {
	BaseModel
	ArtifactKey
	Name     string `gorm:"primary_key"`
	Location string
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
	Name        string     `gorm:"primary_key"`
}

type TagKey struct {
	DatasetProject string `gorm:"primary_key"`
	DatasetName    string `gorm:"primary_key"`
	DatasetDomain  string `gorm:"primary_key"`
	DatasetVersion string `gorm:"primary_key"`
	TagName        string `gorm:"primary_key"`
}

type Tag struct {
	BaseModel
	TagKey
	ArtifactID  string
	DatasetUUID UUIDString `gorm:"index:tags_dataset_uuid_idx"`
	Artifact    Artifact   `gorm:"references:DatasetProject,DatasetName,DatasetDomain,DatasetVersion,ArtifactID;foreignkey:DatasetProject,DatasetName,DatasetDomain,DatasetVersion,ArtifactID"`
}

type Partition struct {
	BaseModel
	DatasetUUID UUIDString `gorm:"primary_key;type:uuid"`
	Key         string     `gorm:"primary_key"`
	Value       string     `gorm:"primary_key"`
	ArtifactID  string     `gorm:"primary_key;index"` // index for JOINs with the Tag/Labels table when querying artifacts
}

type ReservationKey struct {
	DatasetProject string `gorm:"primary_key"`
	DatasetName    string `gorm:"primary_key"`
	DatasetDomain  string `gorm:"primary_key"`
	DatasetVersion string `gorm:"primary_key"`
	TagName        string `gorm:"primary_key"`
}

// Reservation tracks the metadata needed to allow
// task cache serialization
type Reservation struct {
	BaseModel
	ReservationKey

	// Identifies who owns the reservation
	OwnerID string

	// When the reservation will expire
	ExpiresAt          time.Time
	SerializedMetadata []byte
}

var Migrations = []*gormigrate.Migration{
	{
		ID: "2023-04-18-noop-dataset",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&Dataset{})
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},
	{
		ID: "2023-04-18-noop-artifact",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&Artifact{})
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},
	{
		ID: "2023-04-18-noop-artifact-data",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&ArtifactData{})
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},
	{
		ID: "2023-04-18-noop-tag",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&Tag{})
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},
	{
		ID: "2023-04-18-noop-partition-key",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&PartitionKey{})
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},
	{
		ID: "2023-04-18-noop-partition",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&Partition{})
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},
	{
		ID: "2023-04-18-noop-reservation",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&Reservation{})
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},
}
