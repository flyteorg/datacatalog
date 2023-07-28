package models

type TagKey struct {
	DatasetProject string `gorm:"size:128;primary_key"`
	DatasetName    string `gorm:"size:128;primary_key"`
	DatasetDomain  string `gorm:"size:128;primary_key"`
	DatasetVersion string `gorm:"size:128;primary_key"`
	TagName        string `gorm:"size:128;primary_key"`
}

type Tag struct {
	BaseModel
	TagKey
	ArtifactID  string
	DatasetUUID UUIDString `gorm:"index:tags_dataset_uuid_idx"`
	Artifact    Artifact   `gorm:"references:DatasetProject,DatasetName,DatasetDomain,DatasetVersion,ArtifactID;foreignkey:DatasetProject,DatasetName,DatasetDomain,DatasetVersion,ArtifactID"`
}
