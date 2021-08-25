package interfaces

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"
)

// ReservationManager is the interface to handle reservation requests.
// You can find more details about the APIs in datacatalog service proto
// in flyteidl
type ReservationManager interface {
	GetOrReserveArtifact(context.Context, *datacatalog.GetOrReserveArtifactRequest) (*datacatalog.GetOrReserveArtifactResponse, error)
	ReleaseReservation(context.Context, *datacatalog.ReleaseReservationRequest) (*datacatalog.ReleaseReservationResponse, error)
}
