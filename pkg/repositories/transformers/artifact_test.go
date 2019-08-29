package transformers

import (
	"testing"

	"github.com/lyft/datacatalog/pkg/repositories/models"
	datacatalog "github.com/lyft/datacatalog/protos/gen"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
)

var testInteger = &core.Literal{
	Value: &core.Literal_Scalar{
		Scalar: &core.Scalar{
			Value: &core.Scalar_Primitive{
				Primitive: &core.Primitive{Value: &core.Primitive_Integer{Integer: 1}},
			},
		},
	},
}

func TestCreateArtifactModel(t *testing.T) {
	artifactDataList := []*datacatalog.ArtifactData{
		{Name: "data1", Value: testInteger},
		{Name: "data2", Value: testInteger},
	}

	createArtifactRequest := datacatalog.CreateArtifactRequest{
		Artifact: &datacatalog.Artifact{
			Id:       "artifactID-1",
			Dataset:  &datasetID,
			Data:     artifactDataList,
			Metadata: &metadata,
		},
	}

	testArtifactData := []models.ArtifactData{
		{Name: "data1", Location: "s3://test1"},
		{Name: "data3", Location: "s3://test2"},
	}
	artifactModel, err := CreateArtifactModel(createArtifactRequest, testArtifactData)
	assert.NoError(t, err)
	assert.Equal(t, artifactModel.ArtifactID, createArtifactRequest.Artifact.Id)
	assert.Equal(t, artifactModel.ArtifactKey.DatasetProject, datasetID.Project)
	assert.Equal(t, artifactModel.ArtifactKey.DatasetDomain, datasetID.Domain)
	assert.Equal(t, artifactModel.ArtifactKey.DatasetName, datasetID.Name)
	assert.Equal(t, artifactModel.ArtifactKey.DatasetVersion, datasetID.Version)
	assert.EqualValues(t, testArtifactData, artifactModel.ArtifactData)
}

func TestCreateArtifactModelNoMetdata(t *testing.T) {
	artifactDataList := []*datacatalog.ArtifactData{
		{Name: "data1", Value: testInteger},
		{Name: "data2", Value: testInteger},
	}

	createArtifactRequest := datacatalog.CreateArtifactRequest{
		Artifact: &datacatalog.Artifact{
			Id:      "artifactID-1",
			Dataset: &datasetID,
			Data:    artifactDataList,
		},
	}

	testArtifactData := []models.ArtifactData{
		{Name: "data1", Location: "s3://test1"},
		{Name: "data3", Location: "s3://test2"},
	}
	artifactModel, err := CreateArtifactModel(createArtifactRequest, testArtifactData)
	assert.NoError(t, err)
	assert.Equal(t, []byte{}, artifactModel.SerializedMetadata)
}

func TestFromArtifactModel(t *testing.T) {
	artifactModel := models.Artifact{
		ArtifactKey: models.ArtifactKey{
			DatasetProject: "project1",
			DatasetDomain:  "domain1",
			DatasetName:    "name1",
			DatasetVersion: "version1",
			ArtifactID:     "id1",
		},
		SerializedMetadata: []byte{},
	}

	actual, err := FromArtifactModel(artifactModel)
	assert.NoError(t, err)
	assert.Equal(t, artifactModel.ArtifactID, actual.Id)
	assert.Equal(t, artifactModel.DatasetProject, actual.Dataset.Project)
	assert.Equal(t, artifactModel.DatasetDomain, actual.Dataset.Domain)
	assert.Equal(t, artifactModel.DatasetName, actual.Dataset.Name)
	assert.Equal(t, artifactModel.DatasetVersion, actual.Dataset.Version)
}

func TestToArtifactKey(t *testing.T) {
	artifactKey := ToArtifactKey(datasetID, "artifactID-1")
	assert.Equal(t, datasetID.Project, artifactKey.DatasetProject)
	assert.Equal(t, datasetID.Domain, artifactKey.DatasetDomain)
	assert.Equal(t, datasetID.Name, artifactKey.DatasetName)
	assert.Equal(t, datasetID.Version, artifactKey.DatasetVersion)
	assert.Equal(t, artifactKey.ArtifactID, "artifactID-1")
}