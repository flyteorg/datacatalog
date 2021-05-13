package gormimpl

import (
	"context"

	errors2 "github.com/flyteorg/datacatalog/pkg/repositories/errors"
	errors3 "github.com/flyteorg/datacatalog/pkg/repositories/errors"
	"github.com/flyteorg/datacatalog/pkg/repositories/interfaces"
	"github.com/flyteorg/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flytestdlib/promutils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	timer := r.repoMetrics.GetDuration.Start(ctx)
	defer timer.Stop()

	var reservation models.Reservation

	result := r.db.Where(&models.Reservation{
		ReservationKey: reservationKey,
	}).Take(&reservation)

	if result.Error != nil {
		return reservation, r.errorTransformer.ToDataCatalogError(result.Error)
	}

	return reservation, nil
}

func (r *reservationRepo) Update(ctx context.Context, reservationKey models.ReservationKey, prevExpireAt time.Time, expireAt time.Time, ownerID string) (int64, error) {
	timer := r.repoMetrics.UpdateDuration.Start(ctx)
	defer timer.Stop()

	result := r.db.Where(
		&models.Reservation{
			ReservationKey: reservationKey,
			ExpireAt:       prevExpireAt,
		},
	).Updates(
		models.Reservation{
			OwnerID:  ownerID,
			ExpireAt: expireAt,
		})
	if result.Error != nil {
		return 0, r.errorTransformer.ToDataCatalogError(result.Error)
	}

	return result.RowsAffected, nil
}

func (r *reservationRepo) CreateOrUpdate(ctx context.Context, reservation models.Reservation, now time.Time) (int64, error) {

	expressions := make([]clause.Expression, 0)
	expressions = append(expressions, clause.Lte{Column: "expire_at", Value: now})

	result := r.db.Clauses(
		clause.OnConflict{
			Where: clause.Where{Exprs: expressions},
			UpdateAll: true,
		},
		).Create(&reservation)
	if result.Error != nil {
		return 0, r.errorTransformer.ToDataCatalogError(result.Error)
	}

	return result.RowsAffected, nil
}