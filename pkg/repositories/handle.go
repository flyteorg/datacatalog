package repositories

import (
	"context"

	"fmt"

	"github.com/flyteorg/datacatalog/pkg/repositories/config"
	"github.com/flyteorg/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"gorm.io/gorm"
)

type DBHandle struct {
	db *gorm.DB
}

func NewDBHandle(dbConfigValues config.DbConfig, catalogScope promutils.Scope) (*DBHandle, error) {
	dbConfig := config.DbConfig{
		Host:         dbConfigValues.Host,
		Port:         dbConfigValues.Port,
		DbName:       dbConfigValues.DbName,
		User:         dbConfigValues.User,
		Password:     dbConfigValues.Password,
		ExtraOptions: dbConfigValues.ExtraOptions,
	}

	//TODO: abstract away the type of db we are connecting to
	db, err := config.OpenDbConnection(config.NewPostgresConfigProvider(dbConfig, catalogScope.NewSubScope("postgres")))
	if err != nil {
		return nil, err
	}

	out := &DBHandle{
		db: db,
	}

	return out, nil
}

func (h *DBHandle) CreateDB(dbName string) error {
	type DatabaseResult struct {
		Exists bool
	}
	var checkExists DatabaseResult
	result := h.db.Raw("SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = ?)", dbName).Scan(&checkExists)
	if result.Error != nil {
		return result.Error
	}

	// create db if it does not exist
	if !checkExists.Exists {
		logger.Infof(context.TODO(), "Creating Database %v since it does not exist", dbName)

		// NOTE: golang sql drivers do not support parameter injection for CREATE calls
		createDBStatement := fmt.Sprintf("CREATE DATABASE %s", dbName)
		result = h.db.Exec(createDBStatement)

		if result.Error != nil {
			return result.Error
		}
	}

	return nil
}

func (h *DBHandle) Migrate(ctx context.Context) error {
	if h.db.Config.Dialector.Name() == config.Postgres {
		logger.Infof(context.TODO(), "Creating postgres extension uuid-ossp if it does not exist")
		h.db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
	}
	if err := h.db.AutoMigrate(&models.Dataset{}); err != nil {
		logger.Errorf(ctx, "Failed to migrate Dataset. err: %v", err)
		return err
	}

	if err := h.db.AutoMigrate(&models.Artifact{}); err != nil {
		logger.Errorf(ctx, "Failed to migrate Artifact. err: %v", err)
		return err
	}

	if err := h.db.AutoMigrate(&models.ArtifactData{}); err != nil {
		logger.Errorf(ctx, "Failed to migrate ArtifactData. err: %v", err)
		return err
	}

	if err := h.db.AutoMigrate(&models.Tag{}); err != nil {
		logger.Errorf(ctx, "Failed to migrate Tag. err: %v", err)
		return err
	}

	if err := h.db.AutoMigrate(&models.PartitionKey{}); err != nil {
		logger.Errorf(ctx, "Failed to migrate PartitionKey. err: %v", err)
		return err
	}

	if err := h.db.AutoMigrate(&models.Partition{}); err != nil {
		logger.Errorf(ctx, "Failed to migrate Partition. err: %v", err)
		return err
	}

	return nil
}
