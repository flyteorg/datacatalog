package impl

import (
	"context"

	mockScope "github.com/flyteorg/flytestdlib/promutils"

	"testing"
	"time"

	errors2 "github.com/flyteorg/datacatalog/pkg/errors"
	"github.com/flyteorg/datacatalog/pkg/repositories/mocks"
	"github.com/flyteorg/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
)

var tagName = "tag"
var project = "p"
var name = "n"
var domain = "d"
var version = "v"
var datasetID = datacatalog.DatasetID{
	Project: project,
	Name:    name,
	Domain:  domain,
	Version: version,
}
var heartbeatInterval = time.Second * 5
var heartbeatGracePeriodMultiplier = time.Second * 3
var prevOwner = "prevOwner"
var currentOwner = "currentOwner"

func TestGetOrReserveArtifact_ArtifactExists(t *testing.T) {
	serializedMetadata, err := proto.Marshal(&datacatalog.Metadata{})
	assert.Nil(t, err)
	expectedArtifact := models.Artifact{
		ArtifactKey: models.ArtifactKey{
			ArtifactID: "123",
		},
		SerializedMetadata: serializedMetadata,
	}

	dcRepo := getDatacatalogRepo()

	dcRepo.MockTagRepo.On("Get",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(tagKey models.TagKey) bool {
			return tagKey.DatasetProject == datasetID.Project &&
				tagKey.DatasetName == datasetID.Name &&
				tagKey.DatasetDomain == datasetID.Domain &&
				tagKey.DatasetVersion == datasetID.Version &&
				tagKey.TagName == tagName
		}),
	).Return(models.Tag{
		Artifact: expectedArtifact,
	}, nil)

	reservationManager := NewReservationManager(&dcRepo, heartbeatGracePeriodMultiplier,
		heartbeatInterval, time.Now, mockScope.NewTestScope())

	req := datacatalog.GetOrReserveArtifactRequest{
		DatasetId: &datasetID,
		TagName:   tagName,
		OwnerId:   currentOwner,
	}

	resp, err := reservationManager.GetOrReserveArtifact(context.Background(), &req)
	assert.Nil(t, err)
	artifact := resp.GetArtifact()
	assert.NotNil(t, artifact)
	assert.Equal(t, expectedArtifact.ArtifactKey.ArtifactID, artifact.Id)
}

func getDatacatalogRepo() mocks.DataCatalogRepo {
	return mocks.DataCatalogRepo{
		MockReservationRepo: &mocks.ReservationRepo{},
		MockTagRepo:         &mocks.TagRepo{},
	}
}

func TestGetOrReserveArtifact_CreateReservation(t *testing.T) {
	dcRepo := getDatacatalogRepo()

	setUpTagRepoGetNotFound(&dcRepo)

	dcRepo.MockReservationRepo.On("Get",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(key models.ReservationKey) bool {
			return key.DatasetProject == datasetID.Project &&
				key.DatasetDomain == datasetID.Domain &&
				key.DatasetVersion == datasetID.Version &&
				key.DatasetName == datasetID.Name &&
				key.TagName == tagName
		})).Return(models.Reservation{}, errors2.NewDataCatalogErrorf(codes.NotFound, "entry not found"))

	now := time.Now()

	dcRepo.MockReservationRepo.On("Create",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(reservation models.Reservation) bool {
			return reservation.DatasetProject == datasetID.Project &&
				reservation.DatasetDomain == datasetID.Domain &&
				reservation.DatasetName == datasetID.Name &&
				reservation.DatasetVersion == datasetID.Version &&
				reservation.TagName == tagName &&
				reservation.OwnerID == currentOwner &&
				reservation.ExpiresAt == now.Add(heartbeatInterval*heartbeatGracePeriodMultiplier)
		}),
		mock.MatchedBy(func(now time.Time) bool { return true }),
	).Return(nil)

	reservationManager := NewReservationManager(&dcRepo,
		heartbeatGracePeriodMultiplier, heartbeatInterval,
		func() time.Time { return now }, mockScope.NewTestScope())

	req := datacatalog.GetOrReserveArtifactRequest{
		DatasetId: &datasetID,
		TagName:   tagName,
		OwnerId:   currentOwner,
	}

	resp, err := reservationManager.GetOrReserveArtifact(context.Background(), &req)

	assert.Nil(t, err)
	assert.Equal(t, currentOwner, resp.GetReservationStatus().OwnerId)
	assert.Equal(t, datacatalog.ReservationStatus_ACQUIRED, resp.GetReservationStatus().State)
}

func TestGetOrReserveArtifact_TakeOverReservation(t *testing.T) {
	dcRepo := getDatacatalogRepo()

	setUpTagRepoGetNotFound(&dcRepo)

	now := time.Now()
	prevExpiresAt := now.Add(time.Second * 10 * time.Duration(-1))

	setUpReservationRepoGet(&dcRepo, prevExpiresAt)

	dcRepo.MockReservationRepo.On("Update",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(reservation models.Reservation) bool {
			return reservation.DatasetProject == datasetID.Project &&
				reservation.DatasetDomain == datasetID.Domain &&
				reservation.DatasetName == datasetID.Name &&
				reservation.DatasetVersion == datasetID.Version &&
				reservation.TagName == tagName &&
				reservation.OwnerID == currentOwner &&
				reservation.ExpiresAt == now.Add(heartbeatInterval*heartbeatGracePeriodMultiplier)
		}),
		mock.MatchedBy(func(now time.Time) bool { return true }),
	).Return(nil)

	reservationManager := NewReservationManager(&dcRepo,
		heartbeatGracePeriodMultiplier, heartbeatInterval,
		func() time.Time { return now }, mockScope.NewTestScope())

	req := datacatalog.GetOrReserveArtifactRequest{
		DatasetId: &datasetID,
		TagName:   tagName,
		OwnerId:   currentOwner,
	}

	resp, err := reservationManager.GetOrReserveArtifact(context.Background(), &req)

	assert.Nil(t, err)
	assert.Equal(t, currentOwner, resp.GetReservationStatus().OwnerId)
	assert.Equal(t, datacatalog.ReservationStatus_ACQUIRED, resp.GetReservationStatus().State)
}

func setUpReservationRepoGet(dcRepo *mocks.DataCatalogRepo, prevExpiresAt time.Time) {
	dcRepo.MockReservationRepo.On("Get",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(key models.ReservationKey) bool {
			return key.DatasetProject == datasetID.Project &&
				key.DatasetDomain == datasetID.Domain &&
				key.DatasetVersion == datasetID.Version &&
				key.DatasetName == datasetID.Name &&
				key.TagName == tagName
		})).Return(
		models.Reservation{
			ReservationKey: getReservationKey(),
			OwnerID:        prevOwner,
			ExpiresAt:      prevExpiresAt,
		}, nil,
	)
}

func setUpTagRepoGetNotFound(dcRepo *mocks.DataCatalogRepo) {
	dcRepo.MockTagRepo.On("Get",
		mock.Anything,
		mock.Anything,
	).Return(models.Tag{}, errors2.NewDataCatalogErrorf(codes.NotFound, "entry not found"))
}

func getReservationKey() models.ReservationKey {
	return models.ReservationKey{
		DatasetProject: project,
		DatasetName:    name,
		DatasetDomain:  domain,
		DatasetVersion: version,
		TagName:        tagName,
	}
}

func TestGetOrReserveArtifact_AlreadyInProgress(t *testing.T) {
	dcRepo := getDatacatalogRepo()

	setUpTagRepoGetNotFound(&dcRepo)

	now := time.Now()
	prevExpiresAt := now.Add(time.Second * 10)

	setUpReservationRepoGet(&dcRepo, prevExpiresAt)

	reservationManager := NewReservationManager(&dcRepo,
		heartbeatGracePeriodMultiplier, heartbeatInterval,
		func() time.Time { return now }, mockScope.NewTestScope())

	req := datacatalog.GetOrReserveArtifactRequest{
		DatasetId: &datasetID,
		TagName:   tagName,
		OwnerId:   currentOwner,
	}

	resp, err := reservationManager.GetOrReserveArtifact(context.Background(), &req)

	assert.Nil(t, err)
	assert.Equal(t, prevOwner, resp.GetReservationStatus().OwnerId)
	assert.Equal(t, datacatalog.ReservationStatus_ALREADY_IN_PROGRESS, resp.GetReservationStatus().State)
}
