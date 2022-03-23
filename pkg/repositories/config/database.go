package config

import "gorm.io/gorm/logger"

//go:generate pflags DbConfigSection

// DbConfigSection corresponds to the  database section of in the config
type DbConfigSection struct {
	Host   string `json:"host"`
	Port   int    `json:"port"`
	DbName string `json:"dbname"`
	User   string `json:"username"`
	// Either Password or PasswordPath must be set.
	Password     string `json:"password"`
	PasswordPath string `json:"passwordPath"`
	// See http://gorm.io/docs/connecting_to_the_database.html for available options passed, in addition to the above.
	ExtraOptions string          `json:"options"`
	LogLevel     logger.LogLevel `json:"log_level" pflag:"-,"`
}

// DbConfig is database config. Contains values necessary to open a database connection.
type DbConfig struct {
	BaseConfig
	Host           string          `json:"host"`
	Port           int             `json:"port"`
	DbName         string          `json:"dbname"`
	User           string          `json:"user"`
	Password       string          `json:"password"`
	ExtraOptions   string          `json:"options"`
	PostgresConfig *PostgresConfig `json:"postgres,omitempty"`
	SQLiteConfig   *SQLiteConfig   `json:"sqlite,omitempty"`
}

type SQLiteConfig struct {
	File string `json:"file" pflag:",The path to the file (existing or new) where the DB should be created / stored. If existing, then this will be re-used, else a new will be created"`
}

type PostgresConfig struct {
	Host   string `json:"host" pflag:",The host name of the database server"`
	Port   int    `json:"port" pflag:",The port name of the database server"`
	DbName string `json:"dbname" pflag:",The database name"`
	User   string `json:"username" pflag:",The database user who is connecting to the server."`
	// Either Password or PasswordPath must be set.
	Password     string `json:"password" pflag:",The database password."`
	PasswordPath string `json:"passwordPath" pflag:",Points to the file containing the database password."`
	ExtraOptions string `json:"options" pflag:",See http://gorm.io/docs/connecting_to_the_database.html for available options passed, in addition to the above."`
	Debug        bool   `json:"debug" pflag:" Whether or not to start the database connection with debug mode enabled."`
}
