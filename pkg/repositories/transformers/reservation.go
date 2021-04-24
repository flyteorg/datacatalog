package transformers

import (
	"github.com/flyteorg/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"
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
