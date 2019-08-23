package interfaces

import (
	"context"

	datacatalog "github.com/lyft/flyteidl/gen/pb-go/flyteidl/datacatalog"
)

type ArtifactManager interface {
	CreateArtifact(ctx context.Context, request datacatalog.CreateArtifactRequest) (*datacatalog.CreateArtifactResponse, error)
	GetArtifact(ctx context.Context, request datacatalog.GetArtifactRequest) (*datacatalog.GetArtifactResponse, error)
}
