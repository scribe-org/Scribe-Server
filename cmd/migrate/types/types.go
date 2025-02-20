// SPDX-License-Identifier: GPL-3.0-or-later
package types

import "database/sql"

// TableSchema represents the structure of a database table.
type TableSchema struct {
	ColumnNames []string
	ColumnTypes []string
}

// DBConnections holds database connections
type DBConnections struct {
	MariaDB *sql.DB
	SQLite  *sql.DB
} 

// Config represents the overall configuration structure.
type Config struct {
	HostPort   int            `yaml:"hostPort"`
	FileSystem string         `yaml:"fileSystem"`
	Database   DatabaseConfig `yaml:"database"`
}

// DatabaseConfig holds the database connection parameters.
type DatabaseConfig struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Name     string `yaml:"name"`
}
