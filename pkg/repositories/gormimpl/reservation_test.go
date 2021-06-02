package gormimpl

import (
	"context"
	"database/sql"
	"fmt"

	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/flyteorg/datacatalog/pkg/repositories/interfaces"

	apiErrors "github.com/flyteorg/datacatalog/pkg/errors"
	"google.golang.org/grpc/codes"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/flyteorg/datacatalog/pkg/repositories/errors"
	"github.com/flyteorg/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	expectedReservation := GetReservation()

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "reservations" WHERE "reservations"."dataset_project" = $1 AND "reservations"."dataset_name" = $2 AND "reservations"."dataset_domain" = $3 AND "reservations"."dataset_version" = $4 AND "reservations"."tag_name" = $5 LIMIT 1%!!(string=testTag)!(string=testVersion)!(string=testDomain)!(string=testDataset)(EXTRA string=testProject)`,
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

func TestCreateOrUpdate(t *testing.T) {
	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	expectedReservation := GetReservation()

	GlobalMock.NewMock().WithQuery(
		`INSERT INTO "reservations" ("created_at","updated_at","deleted_at","dataset_project","dataset_name","dataset_domain","dataset_version","tag_name","owner_id","expire_at","serialized_metadata") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) ON CONFLICT ("dataset_project","dataset_name","dataset_domain","dataset_version","tag_name") DO UPDATE SET "updated_at"="excluded"."updated_at","deleted_at"="excluded"."deleted_at","owner_id"="excluded"."owner_id","expire_at"="excluded"."expire_at","serialized_metadata"="excluded"."serialized_metadata"WHERE "expire_at" <= $12`,
	).WithRowsNum(1)

	reservationRepo := getReservationRepo(t)

	err := reservationRepo.CreateOrUpdate(context.Background(), expectedReservation, time.Now())
	assert.NoError(t, err)
}

func getReservationRepo(t *testing.T) interfaces.ReservationRepo {
	mocket.Catcher.Register()
	sqlDB, err := sql.Open(mocket.DriverName, "blah")
	assert.Nil(t, err)

	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}))
	if err != nil {
		t.Fatal(fmt.Sprintf("Failed to open mock db with err %v", err))
	}

	return NewReservationRepo(db, errors.NewPostgresErrorTransformer(), promutils.NewTestScope())
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
