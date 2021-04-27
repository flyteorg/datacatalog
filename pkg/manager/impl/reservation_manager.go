package impl

import (
	"context"
	"errors"

	errors2 "github.com/flyteorg/datacatalog/pkg/errors"
	"github.com/flyteorg/datacatalog/pkg/repositories"
	"github.com/flyteorg/datacatalog/pkg/repositories/transformers"

	"github.com/flyteorg/datacatalog/pkg/manager/interfaces"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"
)

type reservationManager struct {
	repo repositories.RepositoryInterface
}

func NewReservationManager() interfaces.ReservationManager {
	return &reservationManager{}
}

func (r *reservationManager) GetOrReserveArtifact(ctx context.Context, request *datacatalog.GetOrReserveArtifactRequest) (*datacatalog.GetOrReserveArtifactResponse, error) {
	tagKey := transformers.ToTagKey(request.DatasetId, request.TagName)
	tag, err := r.repo.TagRepo().Get(ctx, tagKey)
	if err != nil {
		if errors2.IsDoesNotExistError(err) {
			// Tag does not exist yet, let's reserve a spot to work on
			// generating the artifact.
			r.makeReservation(ctx, request)
		}
		return nil, err
	}

	artifact, err := transformers.FromArtifactModel(tag.Artifact)
	if err != nil {
		return nil, err
	}

	return &datacatalog.GetOrReserveArtifactResponse{
		Value: &datacatalog.GetOrReserveArtifactResponse_Artifact{
			Artifact: artifact,
		},
	}, nil
}

func (r *reservationManager) makeReservation(ctx context.Context, request *datacatalog.GetOrReserveArtifactRequest) {
	_, err := r.repo.ReservationRepo().Get(ctx, transformers.ToReservationKey(
		*request.DatasetId,
		request.TagName,
	))
	if err != nil {
		if errors2.IsDoesNotExistError(err) {
			// Reservation does not exist yet, let's create one
		}
	}

	// Reservation already exists so there is a task already start working on it
	// Let's check if the reservation is expired.
}

func (r *reservationManager) ExtendReservation(context.Context, *datacatalog.ExtendReservationRequest) (*datacatalog.ExtendReservationResponse, error) {
	return nil, errors.New("not implemented")
}

func (r *reservationManager) ReleaseReservation(context.Context, *datacatalog.ReleaseReservationRequest) (*datacatalog.ReleaseReservationResponse, error) {
	return nil, errors.New("not implemented")
}
