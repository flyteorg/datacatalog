package gormimpl

import (
	"context"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/flyteorg/datacatalog/pkg/repositories/interfaces"

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
	reservation := GetReservation()

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	reservationCreated := false
	GlobalMock.NewMock().WithQuery(
		`INSERT  INTO "reservations" ("created_at","updated_at","deleted_at","dataset_project","dataset_name","dataset_domain","dataset_version","tag_name","owner_id","expire_at","serialized_metadata") VALUES (?,?,?,?,?,?,?,?,?,?,?)`).WithCallback(
		func(s string, values []driver.NamedValue) {
			reservationCreated = true
		},
	)

	reservationRepo := getReservationRepo(t)

	err := reservationRepo.Create(context.Background(), reservation)
	assert.Nil(t, err)

	assert.True(t, reservationCreated)
}

func TestCreateAlreadyExists(t *testing.T) {
	reservation := GetReservation()

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	GlobalMock.NewMock().WithQuery(
		`INSERT  INTO "reservations" ("created_at","updated_at","deleted_at","dataset_project","dataset_name","dataset_domain","dataset_version","tag_name","owner_id","expire_at","serialized_metadata") VALUES (?,?,?,?,?,?,?,?,?,?,?)`).WithError(
		getAlreadyExistsErr(),
	)

	reservationRepo := getReservationRepo(t)

	err := reservationRepo.Create(context.Background(), reservation)
	assert.NotNil(t, err)
	dcErr, ok := err.(apiErrors.DataCatalogError)
	assert.True(t, ok)
	assert.Equal(t, dcErr.Code(), codes.AlreadyExists)
}

func TestGet(t *testing.T) {
	expectedReservation := GetReservation()

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "reservations"  WHERE "reservations"."deleted_at" IS NULL AND (("reservations"."dataset_project" = testProject) AND ("reservations"."dataset_name" = testDataset) AND ("reservations"."dataset_domain" = testDomain) AND ("reservations"."dataset_version" = testVersion) AND ("reservations"."tag_name" = testTag)) LIMIT 1`,
	).WithReply(getDBResponse(expectedReservation))

	reservationRepo := getReservationRepo(t)
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
	expectedReservation := GetReservation()

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	GlobalMock.NewMock().WithError(gorm.ErrRecordNotFound)

	reservationRepo := getReservationRepo(t)
	_, err := reservationRepo.Get(context.Background(), expectedReservation.ReservationKey)
	assert.Error(t, err)
	dcErr, ok := err.(apiErrors.DataCatalogError)
	assert.True(t, ok)
	assert.Equal(t, dcErr.Code(), codes.NotFound)

}

func TestUpdate(t *testing.T) {
	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	GlobalMock.NewMock().WithQuery(
		`UPDATE "" SET "expire_at" = ?, "owner_id" = ?  WHERE ("reservations"."dataset_project" = ?) AND ("reservations"."dataset_name" = ?) AND ("reservations"."dataset_domain" = ?) AND ("reservations"."dataset_version" = ?) AND ("reservations"."tag_name" = ?) AND ("reservations"."expire_at" = ?)`,
	).WithRowsNum(1)

	reservationRepo := getReservationRepo(t)

	reservationKey := GetReservationKey()
	prevExpireAt := time.Now()
	expireAt := prevExpireAt.Add(time.Second * 50)
	ownerID := "hello"

	rows, err := reservationRepo.Update(context.Background(), reservationKey, prevExpireAt, expireAt, ownerID)

	assert.Nil(t, err)
	assert.Equal(t, rows, int64(1))
}

func getReservationRepo(t *testing.T) interfaces.ReservationRepo {
	return NewReservationRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer(), promutils.NewTestScope())
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

func GetReservationKey() models.ReservationKey {
	return models.ReservationKey{
		DatasetProject: "testProject",
		DatasetName:    "testDataset",
		DatasetDomain:  "testDomain",
		DatasetVersion: "testVersion",
		TagName:        "testTag",
	}
}

func GetReservation() models.Reservation {
	reservation := models.Reservation{
		ReservationKey: GetReservationKey(),
		OwnerID:        "batman",
		ExpireAt:       time.Unix(1, 1),
	}
	return reservation
}
