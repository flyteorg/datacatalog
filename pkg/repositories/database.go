package repositories

import (
	"github.com/flyteorg/flytestdlib/database"
	stdlibLogger "github.com/flyteorg/flytestdlib/logger"
	"gorm.io/gorm"

	"context"
)

const (
	Postgres = "postgres"
	Sqlite   = "sqlite"
)

// OpenDbConnection opens a connection to the database specified in the config.
// You must call CloseDbConnection at the end of your session!
func OpenDbConnection(ctx context.Context, dbConfig database.DbConfig) (*gorm.DB, error) {
	gormConfig := &gorm.Config{
		Logger:                                   database.GetGormLogger(ctx, stdlibLogger.GetConfig()),
		DisableForeignKeyConstraintWhenMigrating: !dbConfig.EnableForeignKeyConstraintWhenMigrating,
	}
	db, err := NewDBHandle(ctx, dbConfig, gormConfig)
	if err != nil {
		return nil, err
	}
	return db.db, nil
}
