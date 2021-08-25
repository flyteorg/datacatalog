package transformers

import (
	"time"

	"github.com/flyteorg/datacatalog/pkg/errors"
	"github.com/flyteorg/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"

	"github.com/golang/protobuf/ptypes"

	"google.golang.org/grpc/codes"
)

func ToReservationKey(datasetID datacatalog.DatasetID, tagName string) models.ReservationKey {
	return models.ReservationKey{
		DatasetProject: datasetID.Project,
		DatasetName:    datasetID.Name,
		DatasetDomain:  datasetID.Domain,
		DatasetVersion: datasetID.Version,
		TagName:        tagName,
	}
}

func CreateReservationStatus(reservation models.Reservation, heartbeatInterval time.Duration, state datacatalog.ReservationStatus_State) (datacatalog.ReservationStatus, error) {
	expiresAtPb, err := ptypes.TimestampProto(reservation.CreatedAt)
	if err != nil {
		return datacatalog.ReservationStatus{}, errors.NewDataCatalogErrorf(codes.Internal, "failed to serialize expires at time")
	}

	heartbeatIntervalPb := ptypes.DurationProto(heartbeatInterval)
	return datacatalog.ReservationStatus{
		State:             state,
		ExpiresAt:         expiresAtPb,
		HeartbeatInterval: heartbeatIntervalPb,
		OwnerId:           reservation.OwnerID,
	}, nil
}
