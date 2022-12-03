package repositories

import (
	"context"
	"github.com/flyteorg/datacatalog/pkg/runtime"
	"github.com/flyteorg/flytestdlib/logger"
)

// Migrate This command will run all the migrations for the database
func Migrate(ctx context.Context) error {
	configProvider := runtime.NewConfigurationProvider()
	dbConfigValues := *configProvider.ApplicationConfiguration().GetDbConfig()

	dbHandle, err := NewDBHandle(ctx, dbConfigValues, nil)
	if err != nil {
		return err
	}

	logger.Infof(ctx, "Created DB connection.")

	// TODO: checkpoints for migrations
	if err := dbHandle.Migrate(ctx); err != nil {
		logger.Errorf(ctx, "Failed to migrate. err: %v", err)
		panic(err)
	}
	logger.Infof(ctx, "Ran DB migration successfully.")
	return nil
}
