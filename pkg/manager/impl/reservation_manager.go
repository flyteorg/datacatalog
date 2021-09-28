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

func (r *reservationManager) GetOrExtendReservation(ctx context.Context, request *datacatalog.GetOrExtendReservationRequest) (*datacatalog.GetOrExtendReservationResponse, error) {
	reservationID := request.ReservationId
	reservation, err := r.tryAcquireReservation(ctx, reservationID, request.OwnerId)
	if err != nil {
		r.systemMetrics.acquireReservationFailure.Inc(ctx)
		return nil, err
	}

	return &datacatalog.GetOrExtendReservationResponse{
		Reservation: &reservation,
	}, nil
}

// tryAcquireReservation will fetch the reservation first and only create/update
// the reservation if it does not exist or has expired.
// This is an optimization to reduce the number of writes to db. We always need
// to do a GET here because we want to know who owns the reservation
// and show it to users on the UI. However, the reservation is held by a single
// task most of the times and there is no need to do a write.
func (r *reservationManager) tryAcquireReservation(ctx context.Context, reservationID *datacatalog.ReservationID, ownerID string) (datacatalog.Reservation, error) {
	repo := r.repo.ReservationRepo()
	reservationKey := transformers.FromReservationID(reservationID)
	repoReservation, err := repo.Get(ctx, reservationKey)

	reservationExists := true
	if err != nil {
		if errors.IsDoesNotExistError(err) {
			// Reservation does not exist yet so let's create one
			reservationExists = false
		} else {
			return datacatalog.Reservation{}, err
		}
	}

	now := r.now()
	newRepoReservation := models.Reservation{
		ReservationKey: reservationKey,
		OwnerID:        ownerID,
		ExpiresAt:      now.Add(r.heartbeatInterval * r.heartbeatGracePeriodMultiplier),
	}

	// Conditional upsert on reservation. Race conditions are handled
	// within the reservation repository Create and Update function calls.
	var repoErr error
	if !reservationExists {
		repoErr = repo.Create(ctx, newRepoReservation, now)
	} else if repoReservation.ExpiresAt.Before(now) || repoReservation.OwnerID == ownerID {
		repoErr = repo.Update(ctx, newRepoReservation, now)
	} else {
		logger.Debugf(ctx, "Reservation: %+v is held by %s", reservationKey, repoReservation.OwnerID)

		reservation, err := transformers.CreateReservation(&repoReservation, r.heartbeatInterval)
		if err != nil {
			return reservation, err
		}

		r.systemMetrics.reservationAlreadyInProgress.Inc(ctx)
		return reservation, nil
	}

	if repoErr != nil {
		if repoErr.Error() == repo_errors.AlreadyExists {
			// Looks like someone else tried to obtain the reservation
			// at the same time and they won. Let's find out who won.
			rsv1, err := repo.Get(ctx, reservationKey)
			if err != nil {
				return datacatalog.Reservation{}, err
			}

			reservation, err := transformers.CreateReservation(&rsv1, r.heartbeatInterval)
			if err != nil {
				return reservation, err
			}

			r.systemMetrics.reservationAlreadyInProgress.Inc(ctx)
			return reservation, nil
		}

		return datacatalog.Reservation{}, repoErr
	}

	// Reservation has been acquired or extended without error
	reservation, err := transformers.CreateReservation(&newRepoReservation, r.heartbeatInterval)
	if err != nil {
		return reservation, err
	}

	r.systemMetrics.reservationAlreadyInProgress.Inc(ctx)
	return reservation, nil
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
