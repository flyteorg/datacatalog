package models

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
	Dataset            Dataset        `gorm:"association_autocreate:false"`
	ArtifactData       []ArtifactData `gorm:"association_foreignkey:DatasetProject,DatasetName,DatasetDomain,DatasetVersion,ArtifactID;foreignkey:DatasetProject,DatasetName,DatasetDomain,DatasetVersion,ArtifactID"`
	SerializedMetadata []byte
}

type ArtifactData struct {
	BaseModel
	ArtifactKey
	Name     string `gorm:"primary_key"`
	Location string
}