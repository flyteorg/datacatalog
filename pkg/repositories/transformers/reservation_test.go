package transformers

import (
	"testing"
	"time"

	"github.com/flyteorg/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/magiconair/properties/assert"
)

func TestFromReservationID(t *testing.T) {
	reservationID := datacatalog.ReservationID{
		DatasetId: &datacatalog.DatasetID{
			Project: "p",
			Name:    "n",
			Domain:  "d",
			Version: "v",
		},
		TagName: "t",
	}

	reservationKey := FromReservationID(&reservationID)
	assert.Equal(t, reservationKey.DatasetProject, reservationID.DatasetId.Project)
	assert.Equal(t, reservationKey.DatasetName, reservationID.DatasetId.Name)
	assert.Equal(t, reservationKey.DatasetDomain, reservationID.DatasetId.Domain)
	assert.Equal(t, reservationKey.DatasetVersion, reservationID.DatasetId.Version)
	assert.Equal(t, reservationKey.TagName, reservationID.TagName)
}

func TestCreateReservationStatus(t *testing.T) {
	now := time.Now()
	heartbeatInterval := time.Duration(time.Second * 5)
	reservation := models.Reservation {
		ReservationKey: models.ReservationKey {
			DatasetProject: "p",
			DatasetName:    "n",
			DatasetDomain:  "d",
			DatasetVersion: "v",
			TagName:        "t",
		},
		OwnerID: "o",
		ExpiresAt: now,
	}

	reservationStatus, err := CreateReservationStatus(&reservation, heartbeatInterval, datacatalog.ReservationStatus_ACQUIRED)

	assert.Equal(t, err, nil)
	assert.Equal(t, reservationStatus.OwnerId, reservation.OwnerID)
	assert.Equal(t, reservationStatus.ExpiresAt.AsTime(), reservation.ExpiresAt.UTC())

	reservationID := reservationStatus.ReservationId
	assert.Equal(t, reservationID.TagName, reservation.TagName)

	datasetID := reservationID.DatasetId
	assert.Equal(t, datasetID.Project, reservation.DatasetProject)
	assert.Equal(t, datasetID.Name, reservation.DatasetName)
	assert.Equal(t, datasetID.Domain, reservation.DatasetDomain)
	assert.Equal(t, datasetID.Version, reservation.DatasetVersion)
}
