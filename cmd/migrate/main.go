// SPDX-License-Identifier: GPL-3.0-or-later
package main

import (
	"log"

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
