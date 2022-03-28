package runtime

import (
	"context"
	"io/ioutil"
	"os"
	"strings"

	dbconfig "github.com/flyteorg/datacatalog/pkg/repositories/config"
	"github.com/flyteorg/datacatalog/pkg/runtime/configs"
	"github.com/flyteorg/flytestdlib/config"
	stdlibDb "github.com/flyteorg/flytestdlib/config/database"
	"github.com/flyteorg/flytestdlib/logger"
)

const database = "database"
const datacatalog = "datacatalog"

var datacatalogConfig = config.MustRegisterSection(datacatalog, &configs.DataCatalogConfig{})

// Defines the interface to return top-level config structs necessary to start up a datacatalog application.
type ApplicationConfiguration interface {
	GetDbConfig() dbconfig.DbConfig
	GetDataCatalogConfig() configs.DataCatalogConfig
}

type ApplicationConfigurationProvider struct{}

func (p *ApplicationConfigurationProvider) GetDbConfig() dbconfig.DbConfig {
	dbConfigSection := stdlibDb.GetConfig()
	password := dbConfigSection.DeprecatedPassword
	if len(dbConfigSection.DeprecatedPasswordPath) > 0 {
		if _, err := os.Stat(dbConfigSection.DeprecatedPasswordPath); os.IsNotExist(err) {
			logger.Fatalf(context.Background(),
				"missing database password at specified path [%s]", dbConfigSection.DeprecatedPasswordPath)
		}
		passwordVal, err := ioutil.ReadFile(dbConfigSection.DeprecatedPasswordPath)
		if err != nil {
			logger.Fatalf(context.Background(), "failed to read database password from path [%s] with err: %v",
				dbConfigSection.DeprecatedPasswordPath, err)
		}
		// Passwords can contain special characters as long as they are percent encoded
		// https://www.postgresql.org/docs/current/libpq-connect.html
		password = strings.TrimSpace(string(passwordVal))
	}

	var postgresConfig dbconfig.PostgresConfig
	var sqliteConfig dbconfig.SQLiteConfig

	if dbConfigSection.PostgresConfig != nil {
		postgresConfig = dbconfig.PostgresConfig(*dbConfigSection.PostgresConfig)
	}

	if dbConfigSection.SQLiteConfig != nil {
		sqliteConfig = dbconfig.SQLiteConfig(*dbConfigSection.SQLiteConfig)
	}

	return dbconfig.DbConfig{
		Host:           dbConfigSection.DeprecatedHost,
		Port:           dbConfigSection.DeprecatedPort,
		DbName:         dbConfigSection.DeprecatedDbName,
		User:           dbConfigSection.DeprecatedUser,
		Password:       password,
		ExtraOptions:   dbConfigSection.DeprecatedExtraOptions,
		BaseConfig:     dbconfig.BaseConfig{LogLevel: dbConfigSection.LogLevel},
		PostgresConfig: &postgresConfig,
		SQLiteConfig:   &sqliteConfig,
	}
}

func (p *ApplicationConfigurationProvider) GetDataCatalogConfig() configs.DataCatalogConfig {
	return *datacatalogConfig.GetConfig().(*configs.DataCatalogConfig)
}

func NewApplicationConfigurationProvider() ApplicationConfiguration {
	return &ApplicationConfigurationProvider{}
}
