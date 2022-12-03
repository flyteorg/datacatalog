package repositories

import (
	"context"

	"gorm.io/driver/sqlite"

	"fmt"

	"github.com/flyteorg/flytestdlib/database"

	"github.com/flyteorg/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flytestdlib/logger"
	"gorm.io/gorm"
)

type DBHandle struct {
	db *gorm.DB
}

func NewDBHandle(ctx context.Context, dbConfigValues database.DbConfig, gormConfig *gorm.Config) (*DBHandle, error) {
	var gormDb *gorm.DB
	var err error
	logConfig := logger.GetConfig()
	if gormConfig == nil {
		gormConfig = &gorm.Config{
			Logger: database.GetGormLogger(ctx, logConfig),
		}
	}

	switch {
	case !dbConfigValues.SQLite.IsEmpty():
		gormDb, err = gorm.Open(sqlite.Open(dbConfigValues.SQLite.File))
	case !dbConfigValues.Postgres.IsEmpty():
		gormDb, err = database.CreatePostgresDbIfNotExists(ctx, gormConfig, dbConfigValues.Postgres)
	case !dbConfigValues.Mysql.IsEmpty():
		gormDb, err = database.CreateMysqlDbIfNotExists(ctx, gormConfig, dbConfigValues.Mysql)
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
