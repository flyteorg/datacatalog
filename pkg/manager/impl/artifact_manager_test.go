package impl

import (
	"context"
	"testing"
	"time"

	"fmt"

	"github.com/flyteorg/datacatalog/pkg/common"
	"github.com/flyteorg/datacatalog/pkg/errors"
	"github.com/flyteorg/datacatalog/pkg/repositories/mocks"
	"github.com/flyteorg/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/flyteorg/flytestdlib/contextutils"
	mockScope "github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func init() {
	labeled.SetMetricKeys(contextutils.AppNameKey)
}

func createInmemoryDataStore(t testing.TB, scope mockScope.Scope) *storage.DataStore {
	cfg := storage.Config{
		Type: storage.TypeMemory,
	}
	d, err := storage.NewDataStore(&cfg, scope)
	assert.NoError(t, err)
	return d
}

func getTestStringLiteral() *core.Literal {
	return &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Primitive{
					Primitive: &core.Primitive{Value: &core.Primitive_StringValue{StringValue: "value1"}},
				},
			},
		},
	}
}

func getTestTimestamp() time.Time {
	timestamp, _ := time.Parse(time.RFC3339, "2019-12-26T00:00:00+00:00")
	return timestamp
}

func getTestArtifact() *datacatalog.Artifact {
	datasetID := &datacatalog.DatasetID{
		Project: "test-project",
		Domain:  "test-domain",
		Name:    "test-name",
		Version: "test-version",
		UUID:    "test-uuid",
	}
	createdAt, _ := ptypes.TimestampProto(getTestTimestamp())

	return &datacatalog.Artifact{
		Id:      "test-id",
		Dataset: datasetID,
		Metadata: &datacatalog.Metadata{
			KeyMap: map[string]string{"key1": "value1"},
		},
		Data: []*datacatalog.ArtifactData{
			{
				Name:  "data1",
				Value: getTestStringLiteral(),
			},
		},
		Partitions: []*datacatalog.Partition{
			{Key: "key1", Value: "value1"},
			{Key: "key2", Value: "value2"},
		},
		Tags: []*datacatalog.Tag{
			{Name: "test-tag", Dataset: datasetID, ArtifactId: "test-id"},
		},
		CreatedAt: createdAt,
	}
}

func newMockDataCatalogRepo() *mocks.DataCatalogRepo {
	return &mocks.DataCatalogRepo{
		MockDatasetRepo:     &mocks.DatasetRepo{},
		MockArtifactRepo:    &mocks.ArtifactRepo{},
		MockReservationRepo: &mocks.ReservationRepo{},
	}
}

func getExpectedDatastoreLocation(ctx context.Context, store *storage.DataStore, prefix storage.DataReference, artifact *datacatalog.Artifact, idx int) (storage.DataReference, error) {
	dataset := artifact.Dataset
	return store.ConstructReference(ctx, prefix, dataset.Project, dataset.Domain, dataset.Name, dataset.Version, artifact.Id, artifact.Data[idx].Name, artifactDataFile)
}

func getExpectedArtifactModel(ctx context.Context, t *testing.T, datastore *storage.DataStore, artifact *datacatalog.Artifact) models.Artifact {
	expectedDataset := artifact.Dataset
	// Write sample artifact data to the expected location and see if the retrieved data matches
	testStoragePrefix, err := datastore.ConstructReference(ctx, datastore.GetBaseContainerFQN(ctx), "test")
	assert.NoError(t, err)
	dataLocation, err := getExpectedDatastoreLocation(ctx, datastore, testStoragePrefix, artifact, 0)
	assert.NoError(t, err)
	err = datastore.WriteProtobuf(ctx, dataLocation, storage.Options{}, getTestStringLiteral())
	assert.NoError(t, err)

	// construct the artifact model we will return on the queries
	serializedMetadata, err := proto.Marshal(artifact.Metadata)
	assert.NoError(t, err)
	datasetKey := models.DatasetKey{
		Project: expectedDataset.Project,
		Domain:  expectedDataset.Domain,
		Version: expectedDataset.Version,
		Name:    expectedDataset.Name,
		UUID:    expectedDataset.UUID,
	}
	return models.Artifact{
		ArtifactKey: models.ArtifactKey{
			DatasetProject: expectedDataset.Project,
			DatasetDomain:  expectedDataset.Domain,
			DatasetVersion: expectedDataset.Version,
			DatasetName:    expectedDataset.Name,
			ArtifactID:     artifact.Id,
		},
		DatasetUUID: expectedDataset.UUID,
		ArtifactData: []models.ArtifactData{
			{Name: "data1", Location: dataLocation.String()},
		},
		Dataset: models.Dataset{
			DatasetKey:         datasetKey,
			SerializedMetadata: serializedMetadata,
		},
		SerializedMetadata: serializedMetadata,
		Partitions: []models.Partition{
			{Key: "key1", Value: "value1"},
			{Key: "key2", Value: "value2"},
		},
		Tags: []models.Tag{
			{TagKey: models.TagKey{TagName: "test-tag"}, DatasetUUID: expectedDataset.UUID, ArtifactID: artifact.Id},
		},
		BaseModel: models.BaseModel{
			CreatedAt: getTestTimestamp(),
		},
	}
}

func TestCreateArtifact(t *testing.T) {
	ctx := context.Background()
	datastore := createInmemoryDataStore(t, mockScope.NewTestScope())
	testStoragePrefix, err := datastore.ConstructReference(ctx, datastore.GetBaseContainerFQN(ctx), "test")
	assert.NoError(t, err)

	// Mock dataset to return for artifact lookups
	expectedDataset := getTestDataset()
	mockDatasetModel := models.Dataset{
		DatasetKey: models.DatasetKey{
			Project: expectedDataset.Id.Project,
			Domain:  expectedDataset.Id.Domain,
			Name:    expectedDataset.Id.Name,
			Version: expectedDataset.Id.Version,
			UUID:    expectedDataset.Id.UUID,
		},
		PartitionKeys: []models.PartitionKey{
			{Name: expectedDataset.PartitionKeys[0]},
			{Name: expectedDataset.PartitionKeys[1]},
		},
	}

	t.Run("HappyPath", func(t *testing.T) {
		datastore := createInmemoryDataStore(t, mockScope.NewTestScope())
		expectedDataset := getTestDataset()

		ctx := context.Background()
		dcRepo := newMockDataCatalogRepo()
		dcRepo.MockDatasetRepo.On("Get", mock.Anything,
			mock.MatchedBy(func(dataset models.DatasetKey) bool {
				return dataset.Project == expectedDataset.Id.Project &&
					dataset.Domain == expectedDataset.Id.Domain &&
					dataset.Name == expectedDataset.Id.Name &&
					dataset.Version == expectedDataset.Id.Version
			})).Return(mockDatasetModel, nil)

		dcRepo.MockArtifactRepo.On("Create",
			mock.MatchedBy(func(ctx context.Context) bool { return true }),
			mock.MatchedBy(func(artifact models.Artifact) bool {
				expectedArtifact := getTestArtifact()
				return artifact.ArtifactID == expectedArtifact.Id &&
					artifact.SerializedMetadata != nil &&
					len(artifact.ArtifactData) == len(expectedArtifact.Data) &&
					artifact.ArtifactKey.DatasetProject == expectedArtifact.Dataset.Project &&
					artifact.ArtifactKey.DatasetDomain == expectedArtifact.Dataset.Domain &&
					artifact.ArtifactKey.DatasetName == expectedArtifact.Dataset.Name &&
					artifact.ArtifactKey.DatasetVersion == expectedArtifact.Dataset.Version &&
					artifact.DatasetUUID == expectedArtifact.Dataset.UUID &&
					artifact.Partitions[0].Key == expectedArtifact.Partitions[0].Key &&
					artifact.Partitions[0].Value == expectedArtifact.Partitions[0].Value &&
					artifact.Partitions[0].DatasetUUID == expectedDataset.Id.UUID &&
					artifact.Partitions[1].Key == expectedArtifact.Partitions[1].Key &&
					artifact.Partitions[1].Value == expectedArtifact.Partitions[1].Value &&
					artifact.Partitions[1].DatasetUUID == expectedDataset.Id.UUID
			})).Return(nil)

		request := &datacatalog.CreateArtifactRequest{Artifact: getTestArtifact()}
		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		artifactResponse, err := artifactManager.CreateArtifact(ctx, request)
		assert.NoError(t, err)
		assert.NotNil(t, artifactResponse)

		// check that the datastore has the artifactData
		dataRef, err := getExpectedDatastoreLocation(ctx, datastore, testStoragePrefix, getTestArtifact(), 0)
		assert.NoError(t, err)
		var value core.Literal
		err = datastore.ReadProtobuf(ctx, dataRef, &value)
		assert.NoError(t, err)
		assert.Equal(t, value, *getTestArtifact().Data[0].Value)
	})

	t.Run("Dataset does not exist", func(t *testing.T) {
		dcRepo := newMockDataCatalogRepo()
		dcRepo.MockDatasetRepo.On("Get", mock.Anything, mock.Anything).Return(models.Dataset{}, status.Error(codes.NotFound, "not found"))

		request := &datacatalog.CreateArtifactRequest{Artifact: getTestArtifact()}
		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		artifactResponse, err := artifactManager.CreateArtifact(ctx, request)
		assert.Error(t, err)
		assert.Nil(t, artifactResponse)
		responseCode := status.Code(err)
		assert.Equal(t, codes.NotFound, responseCode)
	})

	t.Run("Artifact missing ID", func(t *testing.T) {
		request := &datacatalog.CreateArtifactRequest{
			Artifact: &datacatalog.Artifact{
				// missing artifact id
				Dataset: getTestDataset().Id,
			},
		}

		artifactManager := NewArtifactManager(&mocks.DataCatalogRepo{}, createInmemoryDataStore(t, mockScope.NewTestScope()), testStoragePrefix, mockScope.NewTestScope())
		_, err := artifactManager.CreateArtifact(ctx, request)
		assert.Error(t, err)
		responseCode := status.Code(err)
		assert.Equal(t, codes.InvalidArgument, responseCode)
	})

	t.Run("Artifact missing artifact data", func(t *testing.T) {
		request := &datacatalog.CreateArtifactRequest{
			Artifact: &datacatalog.Artifact{
				Id:      "test",
				Dataset: getTestDataset().Id,
				// missing artifactData
			},
		}

		artifactManager := NewArtifactManager(&mocks.DataCatalogRepo{}, datastore, testStoragePrefix, mockScope.NewTestScope())
		_, err := artifactManager.CreateArtifact(ctx, request)
		assert.Error(t, err)
		responseCode := status.Code(err)
		assert.Equal(t, codes.InvalidArgument, responseCode)
	})

	t.Run("Already exists", func(t *testing.T) {
		dcRepo := newMockDataCatalogRepo()

		dcRepo.MockDatasetRepo.On("Get", mock.Anything, mock.Anything).Return(mockDatasetModel, nil)

		dcRepo.MockArtifactRepo.On("Create",
			mock.MatchedBy(func(ctx context.Context) bool { return true }),
			mock.MatchedBy(func(artifact models.Artifact) bool {
				expectedArtifact := getTestArtifact()
				return artifact.ArtifactID == expectedArtifact.Id &&
					artifact.SerializedMetadata != nil &&
					len(artifact.ArtifactData) == len(expectedArtifact.Data) &&
					artifact.ArtifactKey.DatasetProject == expectedArtifact.Dataset.Project &&
					artifact.ArtifactKey.DatasetDomain == expectedArtifact.Dataset.Domain &&
					artifact.ArtifactKey.DatasetName == expectedArtifact.Dataset.Name &&
					artifact.ArtifactKey.DatasetVersion == expectedArtifact.Dataset.Version
			})).Return(status.Error(codes.AlreadyExists, "test already exists"))

		request := &datacatalog.CreateArtifactRequest{Artifact: getTestArtifact()}
		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		artifactResponse, err := artifactManager.CreateArtifact(ctx, request)
		assert.Error(t, err)
		assert.Nil(t, artifactResponse)

		responseCode := status.Code(err)
		assert.Equal(t, codes.AlreadyExists, responseCode)
	})

	t.Run("Missing Partitions", func(t *testing.T) {
		dcRepo := newMockDataCatalogRepo()
		dcRepo.MockDatasetRepo.On("Get", mock.Anything, mock.Anything).Return(mockDatasetModel, nil)
		artifact := getTestArtifact()
		artifact.Partitions = nil
		dcRepo.MockArtifactRepo.On("Create",
			mock.MatchedBy(func(ctx context.Context) bool { return true }),
			mock.MatchedBy(func(artifact models.Artifact) bool {
				return false
			})).Return(fmt.Errorf("Validation should happen before this happens"))

		request := &datacatalog.CreateArtifactRequest{Artifact: artifact}
		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		artifactResponse, err := artifactManager.CreateArtifact(ctx, request)
		assert.Error(t, err)
		assert.Nil(t, artifactResponse)

		responseCode := status.Code(err)
		assert.Equal(t, codes.InvalidArgument, responseCode)
	})

	t.Run("No Partitions", func(t *testing.T) {
		dcRepo := newMockDataCatalogRepo()
		mockDatasetModel := models.Dataset{
			DatasetKey: models.DatasetKey{
				Project: expectedDataset.Id.Project,
				Domain:  expectedDataset.Id.Domain,
				Name:    expectedDataset.Id.Name,
				Version: expectedDataset.Id.Version,
			},
		}
		dcRepo.MockDatasetRepo.On("Get", mock.Anything, mock.Anything).Return(mockDatasetModel, nil)
		artifact := getTestArtifact()
		artifact.Partitions = []*datacatalog.Partition{}
		dcRepo.MockArtifactRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

		request := &datacatalog.CreateArtifactRequest{Artifact: artifact}
		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		_, err := artifactManager.CreateArtifact(ctx, request)
		assert.NoError(t, err)
	})

	t.Run("Invalid Partition", func(t *testing.T) {
		dcRepo := newMockDataCatalogRepo()
		dcRepo.MockDatasetRepo.On("Get", mock.Anything, mock.Anything).Return(mockDatasetModel, nil)
		artifact := getTestArtifact()
		artifact.Partitions = append(artifact.Partitions, &datacatalog.Partition{Key: "invalidKey", Value: "invalid"})
		dcRepo.MockArtifactRepo.On("Create",
			mock.MatchedBy(func(ctx context.Context) bool { return true }),
			mock.MatchedBy(func(artifact models.Artifact) bool {
				return false
			})).Return(fmt.Errorf("Validation should happen before this happens"))

		request := &datacatalog.CreateArtifactRequest{Artifact: artifact}
		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		artifactResponse, err := artifactManager.CreateArtifact(ctx, request)
		assert.Error(t, err)
		assert.Nil(t, artifactResponse)

		responseCode := status.Code(err)
		assert.Equal(t, codes.InvalidArgument, responseCode)
	})

}

func TestGetArtifact(t *testing.T) {
	ctx := context.Background()
	datastore := createInmemoryDataStore(t, mockScope.NewTestScope())
	testStoragePrefix, err := datastore.ConstructReference(ctx, datastore.GetBaseContainerFQN(ctx), "test")
	assert.NoError(t, err)

	dcRepo := &mocks.DataCatalogRepo{
		MockDatasetRepo:  &mocks.DatasetRepo{},
		MockArtifactRepo: &mocks.ArtifactRepo{},
		MockTagRepo:      &mocks.TagRepo{},
	}

	expectedArtifact := getTestArtifact()
	mockArtifactModel := getExpectedArtifactModel(ctx, t, datastore, expectedArtifact)

	t.Run("Get by Id", func(t *testing.T) {

		dcRepo.MockArtifactRepo.On("Get", mock.Anything,
			mock.MatchedBy(func(artifactKey models.ArtifactKey) bool {
				return artifactKey.ArtifactID == expectedArtifact.Id &&
					artifactKey.DatasetProject == expectedArtifact.Dataset.Project &&
					artifactKey.DatasetDomain == expectedArtifact.Dataset.Domain &&
					artifactKey.DatasetVersion == expectedArtifact.Dataset.Version &&
					artifactKey.DatasetName == expectedArtifact.Dataset.Name
			})).Return(mockArtifactModel, nil)

		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		artifactResponse, err := artifactManager.GetArtifact(ctx, &datacatalog.GetArtifactRequest{
			Dataset:     getTestDataset().Id,
			QueryHandle: &datacatalog.GetArtifactRequest_ArtifactId{ArtifactId: expectedArtifact.Id},
		})
		assert.NoError(t, err)

		assert.True(t, proto.Equal(expectedArtifact, artifactResponse.Artifact))
	})

	t.Run("Get by Artifact Tag", func(t *testing.T) {
		expectedTag := getTestTag()

		dcRepo.MockTagRepo.On("Get", mock.Anything,
			mock.MatchedBy(func(tag models.TagKey) bool {
				return tag.TagName == expectedTag.TagName &&
					tag.DatasetProject == expectedTag.DatasetProject &&
					tag.DatasetDomain == expectedTag.DatasetDomain &&
					tag.DatasetVersion == expectedTag.DatasetVersion &&
					tag.DatasetName == expectedTag.DatasetName
			})).Return(models.Tag{
			TagKey: models.TagKey{
				DatasetProject: expectedTag.DatasetProject,
				DatasetDomain:  expectedTag.DatasetDomain,
				DatasetName:    expectedTag.DatasetName,
				DatasetVersion: expectedTag.DatasetVersion,
				TagName:        expectedTag.TagName,
			},
			DatasetUUID: expectedTag.DatasetUUID,
			Artifact:    mockArtifactModel,
			ArtifactID:  mockArtifactModel.ArtifactID,
		}, nil)

		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		artifactResponse, err := artifactManager.GetArtifact(ctx, &datacatalog.GetArtifactRequest{
			Dataset:     getTestDataset().Id,
			QueryHandle: &datacatalog.GetArtifactRequest_TagName{TagName: expectedTag.TagName},
		})
		assert.NoError(t, err)
		assert.True(t, proto.Equal(expectedArtifact, artifactResponse.Artifact))
	})

	t.Run("Get missing input", func(t *testing.T) {
		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		artifactResponse, err := artifactManager.GetArtifact(ctx, &datacatalog.GetArtifactRequest{Dataset: getTestDataset().Id})
		assert.Error(t, err)
		assert.Nil(t, artifactResponse)
		responseCode := status.Code(err)
		assert.Equal(t, codes.InvalidArgument, responseCode)
	})

	t.Run("Get does not exist", func(t *testing.T) {
		dcRepo.MockTagRepo.On("Get", mock.Anything, mock.Anything).Return(
			models.Tag{}, errors.NewDataCatalogError(codes.NotFound, "tag with artifact does not exist"))
		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		artifactResponse, err := artifactManager.GetArtifact(ctx, &datacatalog.GetArtifactRequest{Dataset: getTestDataset().Id, QueryHandle: &datacatalog.GetArtifactRequest_TagName{TagName: "test"}})
		assert.Error(t, err)
		assert.Nil(t, artifactResponse)
		responseCode := status.Code(err)
		assert.Equal(t, codes.NotFound, responseCode)
	})
}

func TestListArtifact(t *testing.T) {
	ctx := context.Background()
	datastore := createInmemoryDataStore(t, mockScope.NewTestScope())
	testStoragePrefix, err := datastore.ConstructReference(ctx, datastore.GetBaseContainerFQN(ctx), "test")
	assert.NoError(t, err)

	dcRepo := &mocks.DataCatalogRepo{
		MockDatasetRepo:  &mocks.DatasetRepo{},
		MockArtifactRepo: &mocks.ArtifactRepo{},
		MockTagRepo:      &mocks.TagRepo{},
	}

	expectedDataset := getTestDataset()
	mockDatasetModel := models.Dataset{
		DatasetKey: models.DatasetKey{
			Project: expectedDataset.Id.Project,
			Domain:  expectedDataset.Id.Domain,
			Name:    expectedDataset.Id.Name,
			Version: expectedDataset.Id.Version,
			UUID:    expectedDataset.Id.UUID,
		},
	}

	expectedArtifact := getTestArtifact()
	mockArtifactModel := getExpectedArtifactModel(ctx, t, datastore, expectedArtifact)

	t.Run("List Artifact on invalid filter", func(t *testing.T) {
		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		filter := &datacatalog.FilterExpression{
			Filters: []*datacatalog.SinglePropertyFilter{
				{
					PropertyFilter: &datacatalog.SinglePropertyFilter_DatasetFilter{
						DatasetFilter: &datacatalog.DatasetPropertyFilter{
							Property: &datacatalog.DatasetPropertyFilter_Project{
								Project: "test",
							},
						},
					},
				},
			},
		}

		artifactResponse, err := artifactManager.ListArtifacts(ctx, &datacatalog.ListArtifactsRequest{Dataset: getTestDataset().Id, Filter: filter})
		assert.Error(t, err)
		assert.Nil(t, artifactResponse)
		responseCode := status.Code(err)
		assert.Equal(t, codes.InvalidArgument, responseCode)
	})

	t.Run("List Artifacts with Partition and Tag", func(t *testing.T) {
		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		filter := &datacatalog.FilterExpression{
			Filters: []*datacatalog.SinglePropertyFilter{
				{
					PropertyFilter: &datacatalog.SinglePropertyFilter_PartitionFilter{
						PartitionFilter: &datacatalog.PartitionPropertyFilter{
							Property: &datacatalog.PartitionPropertyFilter_KeyVal{
								KeyVal: &datacatalog.KeyValuePair{Key: "key1", Value: "val1"},
							},
						},
					},
				},
				{
					PropertyFilter: &datacatalog.SinglePropertyFilter_PartitionFilter{
						PartitionFilter: &datacatalog.PartitionPropertyFilter{
							Property: &datacatalog.PartitionPropertyFilter_KeyVal{
								KeyVal: &datacatalog.KeyValuePair{Key: "key2", Value: "val2"},
							},
						},
					},
				},
				{
					PropertyFilter: &datacatalog.SinglePropertyFilter_TagFilter{
						TagFilter: &datacatalog.TagPropertyFilter{
							Property: &datacatalog.TagPropertyFilter_TagName{
								TagName: "special",
							},
						},
					},
				},
			},
		}

		dcRepo.MockDatasetRepo.On("Get", mock.Anything,
			mock.MatchedBy(func(dataset models.DatasetKey) bool {
				return dataset.Project == expectedDataset.Id.Project &&
					dataset.Domain == expectedDataset.Id.Domain &&
					dataset.Name == expectedDataset.Id.Name &&
					dataset.Version == expectedDataset.Id.Version
			})).Return(mockDatasetModel, nil)

		mockArtifacts := []models.Artifact{
			mockArtifactModel,
			mockArtifactModel,
		}

		dcRepo.MockArtifactRepo.On("List", mock.Anything,
			mock.MatchedBy(func(dataset models.DatasetKey) bool {
				return dataset.Project == expectedDataset.Id.Project &&
					dataset.Domain == expectedDataset.Id.Domain &&
					dataset.Name == expectedDataset.Id.Name &&
					dataset.Version == expectedDataset.Id.Version
			}),
			mock.MatchedBy(func(listInput models.ListModelsInput) bool {
				return len(listInput.ModelFilters) == 3 &&
					listInput.ModelFilters[0].Entity == common.Partition &&
					len(listInput.ModelFilters[0].ValueFilters) == 2 &&
					listInput.ModelFilters[1].Entity == common.Partition &&
					len(listInput.ModelFilters[1].ValueFilters) == 2 &&
					listInput.ModelFilters[2].Entity == common.Tag &&
					len(listInput.ModelFilters[2].ValueFilters) == 1 &&
					listInput.Limit == 50 &&
					listInput.Offset == 0
			})).Return(mockArtifacts, nil)

		artifactResponse, err := artifactManager.ListArtifacts(ctx, &datacatalog.ListArtifactsRequest{Dataset: expectedDataset.Id, Filter: filter})
		assert.NoError(t, err)
		assert.NotEmpty(t, artifactResponse)
	})

	t.Run("List Artifacts with No Partition", func(t *testing.T) {
		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		filter := &datacatalog.FilterExpression{Filters: nil}

		dcRepo.MockDatasetRepo.On("Get", mock.Anything,
			mock.MatchedBy(func(dataset models.DatasetKey) bool {
				return dataset.Project == expectedDataset.Id.Project &&
					dataset.Domain == expectedDataset.Id.Domain &&
					dataset.Name == expectedDataset.Id.Name &&
					dataset.Version == expectedDataset.Id.Version
			})).Return(mockDatasetModel, nil)

		mockArtifacts := []models.Artifact{
			mockArtifactModel,
			mockArtifactModel,
		}
		dcRepo.MockArtifactRepo.On("List", mock.Anything,
			mock.MatchedBy(func(dataset models.DatasetKey) bool {
				return dataset.Project == expectedDataset.Id.Project &&
					dataset.Domain == expectedDataset.Id.Domain &&
					dataset.Name == expectedDataset.Id.Name &&
					dataset.Version == expectedDataset.Id.Version
			}),
			mock.MatchedBy(func(listInput models.ListModelsInput) bool {
				return len(listInput.ModelFilters) == 0
			})).Return(mockArtifacts, nil)

		artifactResponse, err := artifactManager.ListArtifacts(ctx, &datacatalog.ListArtifactsRequest{Dataset: expectedDataset.Id, Filter: filter})
		assert.NoError(t, err)
		assert.NotEmpty(t, artifactResponse)
	})
}
