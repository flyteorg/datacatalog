package repositories

import (
	"context"
	"fmt"

	"github.com/flyteorg/datacatalog/pkg/repositories/config"
	"github.com/flyteorg/datacatalog/pkg/runtime"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/go-gormigrate/gormigrate/v2"
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

	m := gormigrate.New(dbHandle.db, gormigrate.DefaultOptions, config.ListMigrations(dbHandle.db))
	if err := m.Migrate(); err != nil {
		return fmt.Errorf("database migration failed: %v", err)
	}
	logger.Infof(ctx, "Migration ran successfully")

	return nil
}
