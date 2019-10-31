package models

type TagKey struct {
	DatasetProject string `gorm:"primary_key"`
	DatasetName    string `gorm:"primary_key"`
	DatasetDomain  string `gorm:"primary_key"`
	DatasetVersion string `gorm:"primary_key"`
	TagName        string `gorm:"primary_key"`
	DatasetUUID    string `gorm:"type:uuid"`
}

type Tag struct {
	BaseModel
	TagKey
	ArtifactID string   `gorm:"primary_key"`
	Artifact   Artifact `gorm:"association_foreignkey:DatasetProject,DatasetName,DatasetDomain,DatasetVersion,ArtifactID,DatasetUUID;foreignkey:DatasetProject,DatasetName,DatasetDomain,DatasetVersion,ArtifactID,DatasetUUID"`
}
