package impl

import (
	"context"
	"errors"
	"time"

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"

	errors2 "github.com/flyteorg/datacatalog/pkg/errors"
	"github.com/flyteorg/datacatalog/pkg/repositories"
	"github.com/flyteorg/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/datacatalog/pkg/repositories/transformers"

	"github.com/flyteorg/datacatalog/pkg/manager/interfaces"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"
)

type reservationMetrics struct {
	scope                        promutils.Scope
	reservationAcquiredViaCreate labeled.Counter
	reservationAcquiredViaUpdate labeled.Counter
	reservationAlreadyInProgress labeled.Counter
	makeReservationFailure       labeled.Counter
	getTagFailure                labeled.Counter
}

type NowFunc func() time.Time

type reservationManager struct {
	repo               repositories.RepositoryInterface
	reservationTimeout time.Duration
	now                NowFunc
	systemMetrics      reservationMetrics
}

func NewReservationManager(
	repo repositories.RepositoryInterface,
	reservationTimeout time.Duration,
	nowFunc NowFunc, // Easier to mock time.Time for testing
	reservationScope promutils.Scope,
) interfaces.ReservationManager {
	systemMetrics := reservationMetrics{
		scope: reservationScope,
		reservationAcquiredViaCreate: labeled.NewCounter(
			"reservation_acquired_via_create",
			"Number of times a reservation was acquired via create",
			reservationScope),
		reservationAcquiredViaUpdate: labeled.NewCounter(
			"reservation_acquired_via_update",
			"Number of times a reservation was acquired via update",
			reservationScope),
		reservationAlreadyInProgress: labeled.NewCounter(
			"reservation_already_in_progress",
			"Number of times we try of acquire a reservation but the reservation is in progress",
			reservationScope,
		),
		makeReservationFailure: labeled.NewCounter(
			"make_reservation_failure",
			"Number of times we failed to make reservation",
			reservationScope,
		),
		getTagFailure: labeled.NewCounter(
			"get_tag_failure",
			"Number of times we failed to get tag",
			reservationScope,
		),
	}

	return &reservationManager{
		repo:               repo,
		reservationTimeout: reservationTimeout,
		now:                nowFunc,
		systemMetrics:      systemMetrics,
	}
}

func (r *reservationManager) GetOrReserveArtifact(ctx context.Context, request *datacatalog.GetOrReserveArtifactRequest) (*datacatalog.GetOrReserveArtifactResponse, error) {
	tagKey := transformers.ToTagKey(request.DatasetId, request.TagName)
	tag, err := r.repo.TagRepo().Get(ctx, tagKey)
	if err != nil {
		if errors2.IsDoesNotExistError(err) {
			// Tag does not exist yet, let's reserve a spot to work on
			// generating the artifact.
			status, err := r.makeReservation(ctx, request)
			if err != nil {
				r.systemMetrics.makeReservationFailure.Inc(ctx)
				return nil, err
			}

			return &datacatalog.GetOrReserveArtifactResponse{
				Value: &datacatalog.GetOrReserveArtifactResponse_ReservationStatus{
					ReservationStatus: &status,
				},
			}, nil
		}
		logger.Errorf(ctx, "Failed retrieve tag: %+v, err: %v", tagKey, err)
		r.systemMetrics.getTagFailure.Inc(ctx)
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
				ExpireAt:       r.now().Add(r.reservationTimeout),
			})

			if err != nil {
				logger.Errorf(ctx, "Failed to create reservation: %+v, err %v", reservationKey, err)

				return datacatalog.ReservationStatus{}, err
			}

			r.systemMetrics.reservationAcquiredViaCreate.Inc(ctx)

			return datacatalog.ReservationStatus{
				State:   datacatalog.ReservationStatus_ACQUIRED,
				OwnerId: request.OwnerId,
			}, nil
		}
		return datacatalog.ReservationStatus{}, err
	}

	now := r.now()
	// Reservation already exists so there is a task already working on it
	// Let's check if the reservation is expired.
	if rsv.ExpireAt.Before(now) {
		// The reservation is expired, let's try to grab the reservation
		rowsAffected, err := repo.Update(ctx, reservationKey,
			rsv.ExpireAt,
			now.Add(r.reservationTimeout), request.OwnerId)
		if err != nil {
			return datacatalog.ReservationStatus{}, err
		}

		if rowsAffected > 0 {
			r.systemMetrics.reservationAcquiredViaUpdate.Inc(ctx)
			return datacatalog.ReservationStatus{
				State:   datacatalog.ReservationStatus_ACQUIRED,
				OwnerId: request.OwnerId,
			}, nil
		}
	}

	logger.Debugf(ctx, "Reservation: %+v is hold by %s", reservationKey, rsv.OwnerID)

	r.systemMetrics.reservationAlreadyInProgress.Inc(ctx)
	return datacatalog.ReservationStatus{
		State:   datacatalog.ReservationStatus_ALREADY_IN_PROGRESS,
		OwnerId: rsv.OwnerID,
	}, nil
}

func (r *reservationManager) ExtendReservation(context.Context, *datacatalog.ExtendReservationRequest) (*datacatalog.ExtendReservationResponse, error) {
	return nil, errors.New("not implemented")
}

func (r *reservationManager) ReleaseReservation(context.Context, *datacatalog.ReleaseReservationRequest) (*datacatalog.ReleaseReservationResponse, error) {
	return nil, errors.New("not implemented")
}
