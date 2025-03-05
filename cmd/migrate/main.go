// SPDX-License-Identifier: GPL-3.0-or-later
package main

import (
	"log"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	mariaDB "github.com/scribe-org/scribe-server/cmd/migrate/mariadb"
	"github.com/scribe-org/scribe-server/cmd/migrate/sqlite"
)

func main() {
	// Load configuration
	config, err := LoadConfig("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	// MariaDB connection setup with config
	mariaDB, err := mariaDB.SetupMariaDB(config.Database)
	if err != nil {
		log.Fatal(err)
	}
	defer mariaDB.Close()

	// Process SQLite files
	if err := sqlite.ProcessSQLiteFiles(mariaDB); err != nil {
		log.Fatal(err)
	}
}
