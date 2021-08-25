package impl

import (
	"context"
	"time"

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"

	"github.com/flyteorg/datacatalog/pkg/errors"
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
	reservationReleased          labeled.Counter
	reservationAlreadyInProgress labeled.Counter
	acquireReservationFailure    labeled.Counter
	getTagFailure                labeled.Counter
	releaseReservationFailure    labeled.Counter
}

type NowFunc func() time.Time

type reservationManager struct {
	repo                           repositories.RepositoryInterface
	heartbeatGracePeriodMultiplier time.Duration
	heartbeatInterval              time.Duration
	now                            NowFunc
	systemMetrics                  reservationMetrics
}

func NewReservationManager(
	repo repositories.RepositoryInterface,
	heartbeatGracePeriodMultiplier time.Duration,
	heartbeatInterval time.Duration,
	nowFunc NowFunc, // Easier to mock time.Time for testing
	reservationScope promutils.Scope,
) interfaces.ReservationManager {
	systemMetrics := reservationMetrics{
		scope: reservationScope,
		reservationAcquired: labeled.NewCounter(
			"reservation_acquired",
			"Number of times a reservation was acquired",
			reservationScope),
		reservationReleased: labeled.NewCounter(
			"reservation_released",
			"Number of times a reservation was released",
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
		releaseReservationFailure: labeled.NewCounter(
			"release_reservation_failure",
			"Number of times we failed to release a reservation",
			reservationScope,
		),
	}

	return &reservationManager{
		repo:                           repo,
		heartbeatGracePeriodMultiplier: heartbeatGracePeriodMultiplier,
		heartbeatInterval:              heartbeatInterval,
		now:                            nowFunc,
		systemMetrics:                  systemMetrics,
	}
}

func (r *reservationManager) GetOrReserveArtifact(ctx context.Context, request *datacatalog.GetOrReserveArtifactRequest) (*datacatalog.GetOrReserveArtifactResponse, error) {
	reservationId := request.ReservationId
	tagKey := transformers.ToTagKey(reservationId.DatasetId, reservationId.TagName)
	tag, err := r.repo.TagRepo().Get(ctx, tagKey)
	if err != nil {
		if errors.IsDoesNotExistError(err) {
			// Tag does not exist yet, let's acquire the reservation to work on
			// generating the artifact.
			status, err := r.tryAcquireReservation(ctx, reservationId, request.OwnerId)
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
func (r *reservationManager) tryAcquireReservation(ctx context.Context, reservationId *datacatalog.ReservationID, ownerId string) (datacatalog.ReservationStatus, error) {
	repo := r.repo.ReservationRepo()
	reservationKey := transformers.FromReservationID(reservationId)
	reservation, err := repo.Get(ctx, reservationKey)

	reservationExists := true
	if err != nil {
		if errors.IsDoesNotExistError(err) {
			// Reservation does not exist yet so let's create one
			reservationExists = false
		} else {
			return datacatalog.ReservationStatus{}, err
		}
	}

	now := r.now()
	newReservation := models.Reservation{
		ReservationKey: reservationKey,
		OwnerID:        ownerId,
		ExpiresAt:      now.Add(r.heartbeatInterval * r.heartbeatGracePeriodMultiplier),
	}

	// Conditional upsert on reservation. Race conditions are handled
	// within the reservation repository Create and Update function calls.
	var repoErr error
	if !reservationExists {
		repoErr = repo.Create(ctx, newReservation, now)
	} else if reservation.ExpiresAt.Before(now) || reservation.OwnerID == ownerId {
		repoErr = repo.Update(ctx, newReservation, now)
	} else {
		logger.Debugf(ctx, "Reservation: %+v is held by %s", reservationKey, reservation.OwnerID)

		reservationStatus, err := transformers.CreateReservationStatus(
			reservation, r.heartbeatInterval,
			datacatalog.ReservationStatus_ALREADY_IN_PROGRESS)
		if err != nil {
			return reservationStatus, err
		}

		r.systemMetrics.reservationAlreadyInProgress.Inc(ctx)
		return reservationStatus, nil
	}

	if repoErr != nil {
		if repoErr.Error() == repo_errors.ReservationAlreadyInProgress {
			// Looks like someone else tried to obtain the reservation
			// at the same time and they won. Let's find out who won.
			rsv1, err := repo.Get(ctx, reservationKey)
			if err != nil {
				return datacatalog.ReservationStatus{}, err
			}

			reservationStatus, err := transformers.CreateReservationStatus(
				rsv1, r.heartbeatInterval,
				datacatalog.ReservationStatus_ALREADY_IN_PROGRESS)
			if err != nil {
				return reservationStatus, err
			}

			r.systemMetrics.reservationAlreadyInProgress.Inc(ctx)
			return reservationStatus, nil
		}

		return datacatalog.ReservationStatus{}, repoErr
	}

	// Reservation has been acquired or extended without error
	reservationStatus, err := transformers.CreateReservationStatus(newReservation,
		r.heartbeatInterval, datacatalog.ReservationStatus_ACQUIRED)
	if err != nil {
		return reservationStatus, err
	}

	r.systemMetrics.reservationAlreadyInProgress.Inc(ctx)
	return reservationStatus, nil
}

func (r *reservationManager) ReleaseReservation(ctx context.Context, request *datacatalog.ReleaseReservationRequest) (*datacatalog.ReleaseReservationResponse, error) {
	repo := r.repo.ReservationRepo()
	reservationKey := transformers.FromReservationID(request.ReservationId)

	err := repo.Delete(ctx, reservationKey)
	if err != nil {
		logger.Errorf(ctx, "Failed to release reservation: %+v, err: %v", reservationKey, err)

		r.systemMetrics.releaseReservationFailure.Inc(ctx)
		return nil, err
	}

	r.systemMetrics.reservationReleased.Inc(ctx)
	return &datacatalog.ReleaseReservationResponse{}, nil
}
