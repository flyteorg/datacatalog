package runtime

import (
	"context"
	"io/ioutil"
	"os"
	"strings"

	"github.com/flyteorg/flytestdlib/config"

	dbconfig "github.com/flyteorg/datacatalog/pkg/repositories/config"
	"github.com/flyteorg/datacatalog/pkg/runtime/configs"
	"github.com/flyteorg/flytestdlib/database"
	"github.com/flyteorg/flytestdlib/logger"
)

const datacatalog = "datacatalog"

var datacatalogConfig = config.MustRegisterSection(datacatalog, &configs.DataCatalogConfig{})

// Defines the interface to return top-level config structs necessary to start up a datacatalog application.
type ApplicationConfiguration interface {
	GetDbConfig() dbconfig.DbConfig
	GetDataCatalogConfig() configs.DataCatalogConfig
}

type ApplicationConfigurationProvider struct{}

func (p *ApplicationConfigurationProvider) GetDbConfig() dbconfig.DbConfig {
	dbConfigSection := database.GetConfig()
	password := dbConfigSection.Postgres.Password
	if len(dbConfigSection.Postgres.PasswordPath) > 0 {
		if _, err := os.Stat(dbConfigSection.Postgres.PasswordPath); os.IsNotExist(err) {
			logger.Fatalf(context.Background(),
				"missing database password at specified path [%s]", dbConfigSection.Postgres.PasswordPath)
		}
		passwordVal, err := ioutil.ReadFile(dbConfigSection.Postgres.PasswordPath)
		if err != nil {
			logger.Fatalf(context.Background(), "failed to read database password from path [%s] with err: %v",
				dbConfigSection.Postgres.PasswordPath, err)
		}
		// Passwords can contain special characters as long as they are percent encoded
		// https://www.postgresql.org/docs/current/libpq-connect.html
		password = strings.TrimSpace(string(passwordVal))
	}

	return dbconfig.DbConfig{
		Host:         dbConfigSection.Postgres.Host,
		Port:         dbConfigSection.Postgres.Port,
		DbName:       dbConfigSection.Postgres.DbName,
		User:         dbConfigSection.Postgres.User,
		Password:     password,
		ExtraOptions: dbConfigSection.Postgres.ExtraOptions,
		BaseConfig:   dbconfig.BaseConfig{LogLevel: dbConfigSection.LogLevel},
		Postgres:     dbConfigSection.Postgres,
		SQLite:       dbConfigSection.SQLite,
	}
}

func (p *ApplicationConfigurationProvider) GetDataCatalogConfig() configs.DataCatalogConfig {
	return *datacatalogConfig.GetConfig().(*configs.DataCatalogConfig)
}

func NewApplicationConfigurationProvider() ApplicationConfiguration {
	return &ApplicationConfigurationProvider{}
}
