package interfaces

import (
	"context"
	"time"

	"github.com/flyteorg/datacatalog/pkg/repositories/models"
)

type ReservationRepo interface {
	Get(ctx context.Context, reservationKey models.ReservationKey) (models.Reservation, error)
	CreateOrUpdate(ctx context.Context, reservation models.Reservation, now time.Time) error
}
