package config

import (
	fixupmigrations "github.com/flyteorg/datacatalog/pkg/repositories/config/migrations/fixup"
	noopmigrations "github.com/flyteorg/datacatalog/pkg/repositories/config/migrations/noop"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func ListMigrations(db *gorm.DB) []*gormigrate.Migration {
	var Migrations []*gormigrate.Migration
	if db.Dialector.Name() == "postgres" {
		Migrations = append(Migrations, noopmigrations.Migrations...)
	}
	Migrations = append(Migrations, fixupmigrations.Migrations...)
	return Migrations
}
