package interfaces

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"
)

type ReservationManager interface {
	GetOrReserveArtifact(context.Context, *datacatalog.GetOrReserveArtifactRequest) (*datacatalog.GetOrReserveArtifactResponse, error)
	ExtendReservation(context.Context, *datacatalog.ExtendReservationRequest) (*datacatalog.ExtendReservationResponse, error)
	ReleaseReservation(context.Context, *datacatalog.ReleaseReservationRequest) (*datacatalog.ReleaseReservationResponse, error)
}
