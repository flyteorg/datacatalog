package models

type ArtifactKey struct {
	DatasetProject string `gorm:"primary_key"`
	DatasetName    string `gorm:"primary_key"`
	DatasetDomain  string `gorm:"primary_key"`
	DatasetVersion string `gorm:"primary_key"`
	ArtifactID     string `gorm:"primary_key"`
	DatasetUUID    string `gorm:"type:uuid"`
}

type Artifact struct {
	BaseModel
	ArtifactKey
	Dataset            Dataset        `gorm:"association_autocreate:false"`
	ArtifactData       []ArtifactData `gorm:"association_foreignkey:ArtifactID;foreignkey:ArtifactID"`
	Partitions         []Partition    `gorm:"association_foreignkey:ArtifactID,DatasetUUID;foreignkey:ArtifactID,DatasetUUID"`
	SerializedMetadata []byte
}

type ArtifactData struct {
	BaseModel
	ArtifactID string `gorm:"primary_key"`
	Name       string `gorm:"primary_key"`
	Location   string
}
