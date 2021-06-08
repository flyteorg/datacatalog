package transformers

import (
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/magiconair/properties/assert"
)

func TestToReservationKey(t *testing.T) {
	datasetID := datacatalog.DatasetID{
		Project: "p",
		Name:    "n",
		Domain:  "d",
		Version: "v",
	}

	reservationKey := ToReservationKey(datasetID, "t")
	assert.Equal(t, datasetID.Project, reservationKey.DatasetProject)
	assert.Equal(t, datasetID.Name, reservationKey.DatasetName)
	assert.Equal(t, datasetID.Domain, reservationKey.DatasetDomain)
	assert.Equal(t, datasetID.Version, reservationKey.DatasetVersion)
	assert.Equal(t, "t", reservationKey.TagName)
}
