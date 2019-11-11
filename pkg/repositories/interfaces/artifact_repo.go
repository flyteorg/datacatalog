package interfaces

import (
	"context"

	"github.com/lyft/datacatalog/pkg/repositories/models"
)

type ArtifactRepo interface {
	Create(ctx context.Context, in models.Artifact) error
	Get(ctx context.Context, in models.ArtifactKey) (models.Artifact, error)
	List(ctx context.Context, datasetKey models.DatasetKey, in models.ListModelsInput) ([]models.Artifact, error)
}

// input:
// FilterExpression
// output:
// ArtifactFilter
// - this will be passed into the repo to construct the GORM expression
// -
// in GORM land we want to
// - construct a Tx and apply the
// - JOIN: based on the different entity filters
// - WHERE's: based on the different entity filters
