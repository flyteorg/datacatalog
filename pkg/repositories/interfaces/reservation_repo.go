package interfaces

import (
	"context"
	"time"

	"github.com/flyteorg/datacatalog/pkg/repositories/models"
)

type ReservationRepo interface {
	Create(ctx context.Context, reservation models.Reservation) error
	Get(ctx context.Context, reservationKey models.ReservationKey) (models.Reservation, error)
	Update(ctx context.Context, reservationKey models.ReservationKey, expirationDate time.Time, OwnerID string) (int64, error)
}
