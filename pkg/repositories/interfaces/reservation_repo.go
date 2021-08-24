package interfaces

import (
	"context"
	"time"

	"github.com/flyteorg/datacatalog/pkg/repositories/models"
)

// Interface to interact with Reservation Table
type ReservationRepo interface {

	// Get reservation
	Get(ctx context.Context, reservationKey models.ReservationKey) (models.Reservation, error)

	// TODO - comment
	Create(ctx context.Context, reservation models.Reservation, now time.Time) error

	// TODO - comment
	Update(ctx context.Context, reservation models.Reservation, now time.Time) error

	// Create the reservation. If the reservation already exists, we try to take over the
	// reservation via update when the reservation has expired. Note: Each reservation has its own
	// expire date which is tracked in expire_at column in the reservation table. And the
	// reservation expires when the date stored in expire_at column is in the past.
	//CreateOrUpdate(ctx context.Context, reservation models.Reservation, now time.Time) error
}
