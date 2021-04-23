package gormimpl

import (
	"context"
	"errors"

	errors2 "github.com/flyteorg/datacatalog/pkg/repositories/errors"
	errors3 "github.com/flyteorg/datacatalog/pkg/repositories/errors"
	"github.com/flyteorg/datacatalog/pkg/repositories/interfaces"
	"github.com/flyteorg/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/jinzhu/gorm"

	"time"
)

type reservationRepo struct {
	db               *gorm.DB
	repoMetrics      gormMetrics
	errorTransformer errors2.ErrorTransformer
}

func NewReservationRepo(db *gorm.DB, errorTransformer errors3.ErrorTransformer, scope promutils.Scope) interfaces.ReservationRepo {
	return &reservationRepo{
		db:               db,
		errorTransformer: errorTransformer,
		repoMetrics:      newGormMetrics(scope),
	}
}

func (r *reservationRepo) Create(ctx context.Context, reservation models.Reservation) error {
	timer := r.repoMetrics.CreateDuration.Start(ctx)
	defer timer.Stop()
	result := r.db.Create(reservation)
	if result.Error != nil {
		return r.errorTransformer.ToDataCatalogError(result.Error)
	}
	return nil
}

func (r *reservationRepo) Get(ctx context.Context, reservationKey models.ReservationKey) (models.Reservation, error) {
	return models.Reservation{}, errors.New("not implemented")
}

func (r *reservationRepo) Update(ctx context.Context, reservationKey models.ReservationKey, expirationDate time.Time) (int, error) {
	return 0, errors.New("not implemented")
}
