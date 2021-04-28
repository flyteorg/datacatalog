package impl

import (
	"context"
	"errors"
	"time"

	errors2 "github.com/flyteorg/datacatalog/pkg/errors"
	"github.com/flyteorg/datacatalog/pkg/repositories"
	"github.com/flyteorg/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/datacatalog/pkg/repositories/transformers"

	"github.com/flyteorg/datacatalog/pkg/manager/interfaces"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"
)

type reservationManager struct {
	repo               repositories.RepositoryInterface
	reservationTimeout time.Duration
}

func NewReservationManager(reservationTimeout time.Duration) interfaces.ReservationManager {
	return &reservationManager{
		reservationTimeout: reservationTimeout,
	}
}

func (r *reservationManager) GetOrReserveArtifact(ctx context.Context, request *datacatalog.GetOrReserveArtifactRequest) (*datacatalog.GetOrReserveArtifactResponse, error) {
	tagKey := transformers.ToTagKey(request.DatasetId, request.TagName)
	tag, err := r.repo.TagRepo().Get(ctx, tagKey)
	if err != nil {
		if errors2.IsDoesNotExistError(err) {
			// Tag does not exist yet, let's reserve a spot to work on
			// generating the artifact.
			state, err := r.makeReservation(ctx, request)
			if err != nil {
				return nil, err
			}

			return &datacatalog.GetOrReserveArtifactResponse{
				Value: &datacatalog.GetOrReserveArtifactResponse_ReservationStatus{
					ReservationStatus: &datacatalog.ReservationStatus{
						State:                state,
					},
				},
			}, nil
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

func (r *reservationManager) makeReservation(ctx context.Context, request *datacatalog.GetOrReserveArtifactRequest) (datacatalog.ReservationStatus, error) {
	repo := r.repo.ReservationRepo()
	reservationKey := transformers.ToReservationKey(*request.DatasetId, request.TagName)
	rsv, err := repo.Get(ctx, reservationKey)

	if err != nil {
		if errors2.IsDoesNotExistError(err) {
			// Reservation does not exist yet so let's create one
			err := repo.Create(ctx, models.Reservation{
				ReservationKey: reservationKey,
				OwnerID:        request.OwnerId,
				ExpireAt:       time.Now().Add(r.reservationTimeout),
			})

			if err != nil {
				return datacatalog.ReservationStatus{}, err
			}

			return datacatalog.ReservationStatus{
				State:                datacatalog.ReservationStatus_ACQUIRED,
				OwnerId:              request.OwnerId,
			}, nil
		}
		return datacatalog.ReservationStatus{}, err
	}

	// Reservation already exists so there is a task already working on it
	// Let's check if the reservation is expired.
	if rsv.ExpireAt.Before(time.Now()) {
		// The reservation is expired, let's try to grab the reservation
		rowsAffected, err := repo.Update(ctx, reservationKey, time.Now().Add(r.reservationTimeout), request.OwnerId)
		if err != nil {
			return datacatalog.ReservationStatus{}, err
		}

		if rowsAffected > 0 {
			return datacatalog.ReservationStatus{
				State:                datacatalog.ReservationStatus_ACQUIRED,
				OwnerId:              request.OwnerId,
			}, nil
		}
	}

	return datacatalog.ReservationStatus{
		State:                datacatalog.ReservationStatus_ALREADY_IN_PROGRESS,
		OwnerId:              rsv.OwnerID,
	}, nil
}

func (r *reservationManager) ExtendReservation(context.Context, *datacatalog.ExtendReservationRequest) (*datacatalog.ExtendReservationResponse, error) {
	return nil, errors.New("not implemented")
}

func (r *reservationManager) ReleaseReservation(context.Context, *datacatalog.ReleaseReservationRequest) (*datacatalog.ReleaseReservationResponse, error) {
	return nil, errors.New("not implemented")
}
