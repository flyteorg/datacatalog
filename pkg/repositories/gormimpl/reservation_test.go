package gormimpl

import (
	"context"
	"database/sql/driver"
	"testing"
	"time"

	apiErrors "github.com/flyteorg/datacatalog/pkg/errors"
	"github.com/jinzhu/gorm"
	"google.golang.org/grpc/codes"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/flyteorg/datacatalog/pkg/repositories/errors"
	"github.com/flyteorg/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/datacatalog/pkg/repositories/utils"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	reservation := getReservation()

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	reservationCreated := false
	GlobalMock.NewMock().WithQuery(
		`INSERT  INTO "reservations" ("created_at","updated_at","deleted_at","dataset_project","dataset_name","dataset_domain","dataset_version","tag_name","owner_id","expire_at","serialized_metadata") VALUES (?,?,?,?,?,?,?,?,?,?,?)`).WithCallback(
		func(s string, values []driver.NamedValue) {
			reservationCreated = true
		},
	)

	reservationRepo := NewReservationRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer(), promutils.NewTestScope())

	err := reservationRepo.Create(context.Background(), reservation)
	assert.Nil(t, err)

	assert.True(t, reservationCreated)
}

func TestCreateAlreadyExists(t *testing.T) {
	reservation := getReservation()

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	GlobalMock.NewMock().WithQuery(
		`INSERT  INTO "reservations" ("created_at","updated_at","deleted_at","dataset_project","dataset_name","dataset_domain","dataset_version","tag_name","owner_id","expire_at","serialized_metadata") VALUES (?,?,?,?,?,?,?,?,?,?,?)`).WithError(
		getAlreadyExistsErr(),
	)

	reservationRepo := NewReservationRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer(), promutils.NewTestScope())

	err := reservationRepo.Create(context.Background(), reservation)
	assert.NotNil(t, err)
	dcErr, ok := err.(apiErrors.DataCatalogError)
	assert.True(t, ok)
	assert.Equal(t, dcErr.Code(), codes.AlreadyExists)
}

func TestGet(t *testing.T) {
	expectedReservation := getReservation()

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "reservations"  WHERE "reservations"."deleted_at" IS NULL AND (("reservations"."dataset_project" = testProject) AND ("reservations"."dataset_name" = testDataset) AND ("reservations"."dataset_domain" = testDomain) AND ("reservations"."dataset_version" = testVersion) AND ("reservations"."tag_name" = testTag)) ORDER BY "reservations"."dataset_project" ASC LIMIT 1`,
	).WithReply(getDBResponse(expectedReservation))

	reservationRepo := NewReservationRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer(), promutils.NewTestScope())
	reservation, err := reservationRepo.Get(context.Background(), expectedReservation.ReservationKey)
	assert.Nil(t, err)
	assert.Equal(t, expectedReservation.DatasetProject, reservation.DatasetProject)
	assert.Equal(t, expectedReservation.DatasetDomain, reservation.DatasetDomain)
	assert.Equal(t, expectedReservation.DatasetName, reservation.DatasetName)
	assert.Equal(t, expectedReservation.DatasetVersion, reservation.DatasetVersion)
	assert.Equal(t, expectedReservation.TagName, reservation.TagName)
	assert.Equal(t, expectedReservation.ExpireAt, reservation.ExpireAt)
}

func TestGetNotFound(t *testing.T) {
	expectedReservation := getReservation()

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	GlobalMock.NewMock().WithError(gorm.ErrRecordNotFound)

	reservationRepo := NewReservationRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer(), promutils.NewTestScope())
	_, err := reservationRepo.Get(context.Background(), expectedReservation.ReservationKey)
	assert.Error(t, err)
	dcErr, ok := err.(apiErrors.DataCatalogError)
	assert.True(t, ok)
	assert.Equal(t, dcErr.Code(), codes.NotFound)

}

func TestUpdate(t *testing.T) {
	assert.FailNow(t, "not implemented yet")
}

func getDBResponse(reservation models.Reservation) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"dataset_project": reservation.DatasetProject,
			"dataset_name":    reservation.DatasetName,
			"dataset_domain":  reservation.DatasetDomain,
			"dataset_version": reservation.DatasetVersion,
			"tag_name":        reservation.TagName,
			"owner_id":        reservation.OwnerID,
			"expire_at":       reservation.ExpireAt,
		},
	}
}

func getReservationKey() models.ReservationKey {
	return models.ReservationKey{
		DatasetProject: "testProject",
		DatasetName:    "testDataset",
		DatasetDomain:  "testDomain",
		DatasetVersion: "testVersion",
		TagName:        "testTag",
	}
}

func getReservation() models.Reservation {
	reservation := models.Reservation{
		ReservationKey: getReservationKey(),
		OwnerID:        "batman",
		ExpireAt:       time.Unix(1, 1),
	}
	return reservation
}
