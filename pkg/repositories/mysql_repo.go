package repositories

import (
	"github.com/flyteorg/datacatalog/pkg/repositories/errors"
	"github.com/flyteorg/datacatalog/pkg/repositories/gormimpl"
	"github.com/flyteorg/datacatalog/pkg/repositories/interfaces"
	"github.com/flyteorg/flytestdlib/promutils"
	"gorm.io/gorm"
)

type MySqlRepo struct {
	datasetRepo     interfaces.DatasetRepo
	artifactRepo    interfaces.ArtifactRepo
	tagRepo         interfaces.TagRepo
	reservationRepo interfaces.ReservationRepo
}

func (dc *MySqlRepo) DatasetRepo() interfaces.DatasetRepo {
	return dc.datasetRepo
}

func (dc *MySqlRepo) ArtifactRepo() interfaces.ArtifactRepo {
	return dc.artifactRepo
}

func (dc *MySqlRepo) TagRepo() interfaces.TagRepo {
	return dc.tagRepo
}

func (dc *MySqlRepo) ReservationRepo() interfaces.ReservationRepo {
	return dc.reservationRepo
}

func NewMySqlRepo(db *gorm.DB, errorTransformer errors.ErrorTransformer, scope promutils.Scope) interfaces.DataCatalogRepo {
	return &MySqlRepo{
		datasetRepo:     gormimpl.NewDatasetRepo(db, errorTransformer, scope.NewSubScope("dataset")),
		artifactRepo:    gormimpl.NewArtifactRepo(db, errorTransformer, scope.NewSubScope("artifact")),
		tagRepo:         gormimpl.NewTagRepo(db, errorTransformer, scope.NewSubScope("tag")),
		reservationRepo: gormimpl.NewReservationRepo(db, errorTransformer, scope.NewSubScope("reservation")),
	}
}
