// Package provisioner provides database provisioning functionality for Docker-DBM.
// It defines a common interface that allows different database implementations
// to be used interchangeably.
package provisioner

import (
	"errors"
	"fmt"
	"os"
)

// Config holds the configuration for database provisioning.
// It includes both admin credentials for the database server
// and the application-specific database details to be created.
type Config struct {
	// DBType specifies the database engine (postgres/mariadb)
	DBType string
	// DBHost is the hostname or IP of the database server
	DBHost string
	// DBPort is the port number of the database server
	DBPort string
	// AdminUser is the privileged user for database administration
	AdminUser string
	// AdminPass is the password for the admin user
	AdminPass string
	// AppDBName is the name of the database to create
	AppDBName string
	// AppDBUser is the application user to create
	AppDBUser string
	// AppDBPass is the password for the application user
	AppDBPass string
}

// DatabaseProvisioner defines the interface for database provisioning operations.
// This interface-driven design allows for easy extension to support
// additional database engines like MongoDB in the future.
type DatabaseProvisioner interface {
	// Provision creates the database and user if they don't exist.
	// It returns an error if the database or user already exists,
	// or if any provisioning step fails.
	Provision(config Config) error

	// Name returns the name of the database type (e.g., "postgres", "mariadb")
	Name() string
}

// ErrDatabaseExists is returned when the target database already exists.
var ErrDatabaseExists = errors.New("database already exists")

// ErrUserExists is returned when the target user already exists.
var ErrUserExists = errors.New("user already exists")

// LoadConfigFromEnv loads the provisioning configuration from environment variables.
func LoadConfigFromEnv() (Config, error) {
	config := Config{
		DBType:    os.Getenv("DB_TYPE"),
		DBHost:    os.Getenv("DB_HOST"),
		DBPort:    os.Getenv("DB_PORT"),
		AdminUser: os.Getenv("ADMIN_USER"),
		AdminPass: os.Getenv("ADMIN_PASS"),
		AppDBName: os.Getenv("APP_DB_NAME"),
		AppDBUser: os.Getenv("APP_DB_USER"),
		AppDBPass: os.Getenv("APP_DB_PASS"),
	}

	// Validate required fields
	var missingVars []string
	if config.DBType == "" {
		missingVars = append(missingVars, "DB_TYPE")
	}
	if config.DBHost == "" {
		missingVars = append(missingVars, "DB_HOST")
	}
	if config.DBPort == "" {
		missingVars = append(missingVars, "DB_PORT")
	}
	if config.AdminUser == "" {
		missingVars = append(missingVars, "ADMIN_USER")
	}
	if config.AdminPass == "" {
		missingVars = append(missingVars, "ADMIN_PASS")
	}
	if config.AppDBName == "" {
		missingVars = append(missingVars, "APP_DB_NAME")
	}
	if config.AppDBUser == "" {
		missingVars = append(missingVars, "APP_DB_USER")
	}
	if config.AppDBPass == "" {
		missingVars = append(missingVars, "APP_DB_PASS")
	}

	if len(missingVars) > 0 {
		return config, fmt.Errorf("missing required environment variables: %v", missingVars)
	}

	return config, nil
}

// GetProvisioner returns the appropriate DatabaseProvisioner based on the DB type.
func GetProvisioner(dbType string) (DatabaseProvisioner, error) {
	switch dbType {
	case "postgres":
		return &PostgresProvisioner{}, nil
	case "mariadb", "mysql":
		return &MariaDBProvisioner{}, nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s (supported: postgres, mariadb)", dbType)
	}
}
