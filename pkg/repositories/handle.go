package repositories

import (
	"context"

	"gorm.io/driver/sqlite"

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
	var gormDb *gorm.DB
	var err error

	switch {
	case dbConfigValues.SQLiteConfig != nil:
		if dbConfigValues.SQLiteConfig.File == "" {
			return nil, fmt.Errorf("illegal sqlite database configuration. `file` is a required parameter and should be a path")
		}
		gormDb, err = gorm.Open(sqlite.Open(dbConfigValues.SQLiteConfig.File))
	case dbConfigValues.PostgresConfig != nil && (len(dbConfigValues.PostgresConfig.Host) > 0 || len(dbConfigValues.PostgresConfig.User) > 0 || len(dbConfigValues.PostgresConfig.DbName) > 0):
		dbConfig := config.DbConfig{
			Host:         dbConfigValues.PostgresConfig.Host,
			Port:         dbConfigValues.PostgresConfig.Port,
			DbName:       dbConfigValues.PostgresConfig.DbName,
			User:         dbConfigValues.PostgresConfig.User,
			Password:     dbConfigValues.PostgresConfig.Password,
			ExtraOptions: dbConfigValues.PostgresConfig.ExtraOptions,
			BaseConfig: config.BaseConfig{
				DisableForeignKeyConstraintWhenMigrating: true,
			},
		}
		gormDb, err = config.OpenDbConnection(config.NewPostgresConfigProvider(dbConfig, catalogScope.NewSubScope(config.Postgres)))
	case len(dbConfigValues.Host) > 0 || len(dbConfigValues.User) > 0 || len(dbConfigValues.DbName) > 0:
		dbConfig := config.DbConfig{
			Host:         dbConfigValues.Host,
			Port:         dbConfigValues.Port,
			DbName:       dbConfigValues.DbName,
			User:         dbConfigValues.User,
			Password:     dbConfigValues.Password,
			ExtraOptions: dbConfigValues.ExtraOptions,
			BaseConfig: config.BaseConfig{
				DisableForeignKeyConstraintWhenMigrating: true,
			},
		}
		gormDb, err = config.OpenDbConnection(config.NewPostgresConfigProvider(dbConfig, catalogScope.NewSubScope(config.Postgres)))
	default:
		return nil, fmt.Errorf("unrecognized database config, %v. Supported only postgres and sqlite", dbConfigValues)
	}

	if err != nil {
		return nil, err
	}

	out := &DBHandle{
		db: gormDb,
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
	if err := h.db.AutoMigrate(&models.Dataset{}); err != nil {
		return err
	}

	if err := h.db.Debug().AutoMigrate(&models.Artifact{}); err != nil {
		return err
	}

	if err := h.db.AutoMigrate(&models.ArtifactData{}); err != nil {
		return err
	}

	if err := h.db.AutoMigrate(&models.Tag{}); err != nil {
		return err
	}

	if err := h.db.AutoMigrate(&models.PartitionKey{}); err != nil {
		return err
	}

	if err := h.db.AutoMigrate(&models.Partition{}); err != nil {
		return err
	}

	if err := h.db.AutoMigrate(&models.Reservation{}); err != nil {
		return err
	}

	return nil
}
