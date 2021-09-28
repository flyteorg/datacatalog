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
	repo_errors "github.com/flyteorg/datacatalog/pkg/repositories/errors"
	"github.com/flyteorg/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/datacatalog/pkg/repositories/transformers"

	"github.com/flyteorg/datacatalog/pkg/manager/interfaces"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"
)

type reservationMetrics struct {
	scope                        promutils.Scope
	reservationAcquired          labeled.Counter
	reservationAlreadyInProgress labeled.Counter
	acquireReservationFailure    labeled.Counter
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
		reservationAcquired: labeled.NewCounter(
			"reservation_acquired",
			"Number of times a reservation was acquired",
			reservationScope),
		reservationAlreadyInProgress: labeled.NewCounter(
			"reservation_already_in_progress",
			"Number of times we try of acquire a reservation but the reservation is in progress",
			reservationScope,
		),
		acquireReservationFailure: labeled.NewCounter(
			"acquire_reservation_failure",
			"Number of times we failed to acquire reservation",
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
			// Tag does not exist yet, let's acquire the reservation to work on
			// generating the artifact.
			status, err := r.tryAcquireReservation(ctx, request)
			if err != nil {
				r.systemMetrics.acquireReservationFailure.Inc(ctx)
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

// tryAcquireReservation will fetch the reservation first and only create/update
// the reservation if it does not exist or has expired.
// This is an optimization to reduce the number of writes to db. We always need
// to do a GET here because we want to know who owns the reservation
// and show it to users on the UI. However, the reservation is held by a single
// task most of the times and there is no need to do a write.
func (r *reservationManager) tryAcquireReservation(ctx context.Context, request *datacatalog.GetOrReserveArtifactRequest) (datacatalog.ReservationStatus, error) {
	repo := r.repo.ReservationRepo()
	reservationKey := transformers.ToReservationKey(*request.DatasetId, request.TagName)
	rsv, err := repo.Get(ctx, reservationKey)

	resvExists := true
	if err != nil {
		if errors2.IsDoesNotExistError(err) {
			// Reservation does not exist yet so let's create one
			resvExists = false
		} else {
			return datacatalog.ReservationStatus{}, err
		}
	}

	now := r.now()
	if !resvExists || rsv.ExpireAt.Before(now) {
		// If the reservation does not exist or it is expired,
		// we try to acquire the reservation
		err := repo.CreateOrUpdate(
			ctx,
			models.Reservation{
				ReservationKey: reservationKey,
				OwnerID:        request.OwnerId,
				ExpireAt:       now.Add(r.reservationTimeout),
			},
			now,
		)

		failedToAcquireReservation := false
		if err != nil {
			if err.Error() == repo_errors.ReservationAlreadyInProgress {

				failedToAcquireReservation = true
			} else {
				return datacatalog.ReservationStatus{}, err
			}
		}

		if failedToAcquireReservation {
			// Looks like someone else tried to obtain the reservation
			// at the same time and they won. Let's find out who won.
			rsv1, err := repo.Get(ctx, reservationKey)
			if err != nil {
				return datacatalog.ReservationStatus{}, err
			}

			r.systemMetrics.reservationAlreadyInProgress.Inc(ctx)
			return datacatalog.ReservationStatus{
				State:   datacatalog.ReservationStatus_ALREADY_IN_PROGRESS,
				OwnerId: rsv1.OwnerID,
			}, err
		}

		r.systemMetrics.reservationAcquired.Inc(ctx)
		return datacatalog.ReservationStatus{
			State:   datacatalog.ReservationStatus_ACQUIRED,
			OwnerId: request.OwnerId,
		}, nil
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
