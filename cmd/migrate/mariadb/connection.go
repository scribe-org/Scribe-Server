// SPDX-License-Identifier: GPL-3.0-or-later

// Package mariadb handles connecting to and interacting with the MariaDB database during migration.
package mariadb

import (
	"database/sql"
	"fmt"

	// Import MySQL and SQLite drivers for side effects (required by `database/sql`)
	_ "github.com/glebarez/sqlite"
	_ "github.com/go-sql-driver/mysql"
	"github.com/scribe-org/scribe-server/cmd/migrate/types"
)

// SetupMariaDB initializes a MariaDB connection with the given configuration.
func SetupMariaDB(dbConfig types.DatabaseConfig) (*sql.DB, error) {
	// Build connection string with database name directly.
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Name,
	)

	// Connect to the database.
	mariaDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s database: %v", dbConfig.Name, err)
	}

	return mariaDB, nil
}
