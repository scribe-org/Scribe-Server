// SPDX-License-Identifier: GPL-3.0-or-later
package main

import (
	"log"

	mariaDB "github.com/scribe-org/scribe-server/cmd/migrate/mariadb"
	"github.com/scribe-org/scribe-server/cmd/migrate/sqlite"
)

// Function variables allow mocking the config, databases and fatal errors in unit tests.
var (
	loadConfig         = LoadConfig
	setupMariaDB       = mariaDB.SetupMariaDB
	processSQLiteFiles = sqlite.ProcessSQLiteFiles
	fatal              = log.Fatal
)

func main() {
	// Load configuration.
	config, err := loadConfig("config.yaml")
	if err != nil {
		fatal(err)
		return
	}

	// MariaDB connection setup with config.
	mariaDB, err := setupMariaDB(config.Database)
	if err != nil {
		fatal(err)
		return
	}
	defer mariaDB.Close()

	// Process SQLite files.
	if err := processSQLiteFiles(mariaDB); err != nil {
		fatal(err)
	}
}
