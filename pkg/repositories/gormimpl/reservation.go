package gormimpl

import (
	"context"

	datacatalog_error "github.com/flyteorg/datacatalog/pkg/errors"
	"google.golang.org/grpc/codes"

	"time"

	errors2 "github.com/flyteorg/datacatalog/pkg/repositories/errors"
	"github.com/flyteorg/datacatalog/pkg/repositories/interfaces"
	"github.com/flyteorg/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flytestdlib/promutils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type reservationRepo struct {
	db               *gorm.DB
	repoMetrics      gormMetrics
	errorTransformer errors2.ErrorTransformer
}

func NewReservationRepo(db *gorm.DB, errorTransformer errors2.ErrorTransformer, scope promutils.Scope) interfaces.ReservationRepo {
	return &reservationRepo{
		db:               db,
		errorTransformer: errorTransformer,
		repoMetrics:      newGormMetrics(scope),
	}
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

func (r *reservationRepo) CreateOrUpdate(ctx context.Context, reservation models.Reservation, now time.Time) error {
	timer := r.repoMetrics.CreateOrUpdateDuration.Start(ctx)
	defer timer.Stop()

	expressions := make([]clause.Expression, 0)
	expressions = append(expressions, clause.Lte{Column: "expire_at", Value: now})

	result := r.db.Clauses(
		clause.OnConflict{
			Where:     clause.Where{Exprs: expressions},
			UpdateAll: true,
		},
	).Create(&reservation)
	if result.Error != nil {
		return r.errorTransformer.ToDataCatalogError(result.Error)
	}

	if result.RowsAffected == 0 {
		return datacatalog_error.NewDataCatalogError(
			codes.FailedPrecondition,
			errors2.ReservationAlreadyInProgress,
		)
	}

	return nil
}
