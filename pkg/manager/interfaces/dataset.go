package interfaces

import (
	"context"

	datacatalog "github.com/lyft/flyteidl/gen/pb-go/flyteidl/datacatalog"
)

type DatasetManager interface {
	CreateDataset(ctx context.Context, request datacatalog.CreateDatasetRequest) (*datacatalog.CreateDatasetResponse, error)
	GetDataset(ctx context.Context, request datacatalog.GetDatasetRequest) (*datacatalog.GetDatasetResponse, error)
}
