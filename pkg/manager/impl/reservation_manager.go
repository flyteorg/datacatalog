package impl

import (
	"context"
	"errors"

	"github.com/flyteorg/datacatalog/pkg/manager/interfaces"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"
)

type reservationManager struct{}

func NewReservationManager() interfaces.ReservationManager {
	return &reservationManager{}
}

func (r *reservationManager) GetOrReserveArtifact(context.Context, *datacatalog.GetOrReserveArtifactRequest) (*datacatalog.GetOrReserveArtifactResponse, error) {
	return nil, errors.New("not implemented")
}

func (r *reservationManager) ExtendReservation(context.Context, *datacatalog.ExtendReservationRequest) (*datacatalog.ExtendReservationResponse, error) {
	return nil, errors.New("not implemented")
}

func (r *reservationManager) ReleaseReservation(context.Context, *datacatalog.ReleaseReservationRequest) (*datacatalog.ReleaseReservationResponse, error) {
	return nil, errors.New("not implemented")
}
