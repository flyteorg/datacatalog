package fixupmigrations

import (
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

type BaseModel struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}

type ArtifactKey struct {
	DatasetProject string `gorm:"size:64;primary_key"`
	// Can we use a smaller size fo dataset name? This is `flyte-task-<task name>` 
	DatasetName    string `gorm:"size:100;primary_key"`
	DatasetDomain  string `gorm:"size:64;primary_key"`
	DatasetVersion string `gorm:"size:128;primary_key"`
	// This is a UUID
	ArtifactID     string `gorm:"size:36;primary_key"`
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
	Name     string `gorm:"size:32;primary_key"`
	Location string `gorm:"size:2048"`
}


type DatasetKey struct {
	Project string     `gorm:"size:64;primary_key"`                          // part of pkey, no index needed as it is first column in the pkey
	// TODO: figure out what size this should be
	Name    string     `gorm:"size:100;primary_key;index:dataset_name_idx"`    // part of pkey and has separate index for filtering
	Domain  string     `gorm:"size:64;primary_key;index:dataset_domain_idx"`  // part of pkey and has separate index for filtering
	Version string     `gorm:"size:128;primary_key;index:dataset_version_idx"` // part of pkey and has separate index for filtering
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
	// TODO: figure out if this is used.
	Name        string `gorm:"size:100;primary_key"`
}

type TagKey struct {
	DatasetProject string `gorm:"size:64;primary_key"`
	// TODO: figure out what size this should be
	DatasetName    string `gorm:"size:100;primary_key"`
	DatasetDomain  string `gorm:"size:64;primary_key"`
	DatasetVersion string `gorm:"size:128;primary_key"`
	TagName        string `gorm:"size:56;primary_key"`
}

type Tag struct {
	BaseModel
	TagKey
	ArtifactID  string     `gorm:"size:36"`
	DatasetUUID UUIDString `gorm:"index:tags_dataset_uuid_idx"`
	Artifact    Artifact   `gorm:"references:DatasetProject,DatasetName,DatasetDomain,DatasetVersion,ArtifactID;foreignkey:DatasetProject,DatasetName,DatasetDomain,DatasetVersion,ArtifactID"`
}

// TODO: figure out if this is used.
type Partition struct {
	BaseModel
	DatasetUUID UUIDString `gorm:"primary_key;type:uuid"`
	Key         string `gorm:"primary_key"`
	Value       string `gorm:"primary_key"`
	ArtifactID  string `gorm:"primary_key;index"` // index for JOINs with the Tag/Labels table when querying artifacts
}

type ReservationKey struct {
	DatasetProject string `gorm:"size:64;primary_key"`
	// TODO: figure out what size this should be
	DatasetName    string `gorm:"size:100;primary_key"`
	DatasetDomain  string `gorm:"size:64;primary_key"`
	DatasetVersion string `gorm:"size:128;primary_key"`
	// TODO: figure out what size this should be
	TagName        string `gorm:"56;primary_key"`
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
		ID: "2023-04-18-fixup-dataset",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&Dataset{})
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},
	{
		ID: "2023-04-18-fixup-artifact",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&Artifact{})
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},
	{
		ID: "2023-04-18-fixup-artifact-data",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&ArtifactData{})
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},
	{
		ID: "2023-04-18-fixup-tag",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&Tag{})
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},
	{
		ID: "2023-04-18-fixup-partition-key",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&PartitionKey{})
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},
	{
		ID: "2023-04-18-fixup-partition",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&Partition{})
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},
	{
		ID: "2023-04-18-fixup-reservation",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&Reservation{})
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},
}
