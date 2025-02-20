package db

import (
	"fmt"
	"io/ioutil"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/scribe-org/scribe-server/cmd/migrate/types"
	"gopkg.in/yaml.v2"
)

// LoadConfig reads and unmarshals the YAML config file.
func LoadConfig(path string) (*types.Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}
	var config types.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}
	return &config, nil
}

// SetupMariaDB initializes a MariaDB connection with the given configuration
func SetupMariaDB(dbConfig types.DatabaseConfig) (*sql.DB, error) {
	// Build connection string from config
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
	)

	// Connect without database name first
	mariaDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MariaDB: %v", err)
	}

	// Create database if not exists
	if _, err := mariaDB.Exec("CREATE DATABASE IF NOT EXISTS " + dbConfig.Name); err != nil {
		return nil, fmt.Errorf("failed to create database: %v", err)
	}

	// Connect to the specific database
	dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Name,
	)
	mariaDB, err = sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s database: %v", dbConfig.Name, err)
	}

	return mariaDB, nil
}
