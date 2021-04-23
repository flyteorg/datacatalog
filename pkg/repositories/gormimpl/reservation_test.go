package gormimpl

import (
	"context"
	"database/sql/driver"
	"testing"
	"time"

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

func getReservation() models.Reservation {
	reservation := models.Reservation{
		ReservationKey: models.ReservationKey{
			DatasetProject: "testProject",
			DatasetName:    "testDataset",
			DatasetDomain:  "testDomain",
			DatasetVersion: "testVersion",
			TagName:        "testTag",
		},
		OwnerID:  "batman",
		ExpireAt: time.Time{},
	}
	return reservation
}
