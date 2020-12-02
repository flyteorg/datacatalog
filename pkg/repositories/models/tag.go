package models

type TagKey struct {
	DatasetProject string
	DatasetName    string
	DatasetDomain  string
	DatasetVersion string
	TagName        string `gorm:"primary_key"`
}

type Tag struct {
	BaseModel
	TagKey
	ArtifactID  string   `gorm:"primary_key"`
	DatasetUUID string   `gorm:"type:uuid;index:tags_dataset_uuid_idx"`
	Artifact    Artifact `gorm:"association_foreignkey:DatasetProject,DatasetName,DatasetDomain,DatasetVersion,ArtifactID;foreignkey:DatasetProject,DatasetName,DatasetDomain,DatasetVersion,ArtifactID"`
}
