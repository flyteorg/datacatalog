package models

type ArtifactKey struct {
	DatasetProject string `gorm:"size:128;primary_key"`
	DatasetName    string `gorm:"size:128;primary_key"`
	DatasetDomain  string `gorm:"size:128;primary_key"`
	DatasetVersion string `gorm:"size:128;primary_key"`
	ArtifactID     string `gorm:"size:128;primary_key"`
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
	Name     string `gorm:"size:256;primary_key"`
	Location string
}
