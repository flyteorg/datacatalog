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

	// Create the reservation. If the reservation already exists, we try to take over the
	// reservation via update.
	CreateOrUpdate(ctx context.Context, reservation models.Reservation, now time.Time) error
}
