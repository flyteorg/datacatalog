package repositories

import (
	"context"
	"github.com/flyteorg/flytestdlib/database"

	"github.com/flyteorg/datacatalog/pkg/repositories/errors"
	"github.com/flyteorg/datacatalog/pkg/repositories/interfaces"
	"github.com/flyteorg/flytestdlib/promutils"
)

// The RepositoryInterface indicates the methods that each Repository must support.
// A Repository indicates a Database which is collection of Tables/models.
// The goal is allow databases to be Plugged in easily.
type RepositoryInterface interface {
	DatasetRepo() interfaces.DatasetRepo
	ArtifactRepo() interfaces.ArtifactRepo
	TagRepo() interfaces.TagRepo
	ReservationRepo() interfaces.ReservationRepo
}

func GetRepository(ctx context.Context, dbConfig database.DbConfig, scope promutils.Scope) RepositoryInterface {
	db, err := OpenDbConnection(ctx, dbConfig)
	if err != nil {
		panic(err)
	}

	var errTransformer errors.ErrorTransformer
	if !dbConfig.Mysql.IsEmpty() {
		errTransformer = errors.NewMySqlErrorTransformer()
		return NewMySqlRepo(db, errTransformer, scope.NewSubScope("repositories"))
	} else if !dbConfig.Postgres.IsEmpty() {
		errTransformer = errors.NewPostgresErrorTransformer()
		return NewPostgresRepo(db, errTransformer, scope.NewSubScope("repositories"))
	} else {
		errTransformer = errors.NewGenericErrorTransformer()
	}
	panic("Unsupported database type")
}
