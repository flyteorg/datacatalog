package repositories

import (
	"fmt"

	"github.com/lyft/datacatalog/pkg/repositories/config"
	"github.com/lyft/datacatalog/pkg/repositories/errors"
	"github.com/lyft/datacatalog/pkg/repositories/interfaces"
	"github.com/lyft/flytestdlib/promutils"
)

type RepoConfig int32

const (
	POSTGRES RepoConfig = 0
)

var RepositoryConfigurationName = map[RepoConfig]string{
	POSTGRES: "POSTGRES",
}

// The RepositoryInterface indicates the methods that each Repository must support.
// A Repository indicates a Database which is collection of Tables/models.
// The goal is allow databases to be Plugged in easily.
type RepositoryInterface interface {
	DatasetRepo() interfaces.DatasetRepo
	ArtifactRepo() interfaces.ArtifactRepo
	TagRepo() interfaces.TagRepo
}

func GetRepository(repoType RepoConfig, dbConfig config.DbConfig, scope promutils.Scope) RepositoryInterface {
	switch repoType {
	case POSTGRES:
		db, err := config.OpenDbConnection(config.NewPostgresConfigProvider(dbConfig, scope.NewSubScope("postgres")))
		db.LogMode(true)
		if err != nil {
			panic(err)
		}
		return NewPostgresRepo(
			db,
			errors.NewPostgresErrorTransformer(),
			scope.NewSubScope("repositories"))
	default:
		panic(fmt.Sprintf("Invalid repoType %v", repoType))
	}
}
