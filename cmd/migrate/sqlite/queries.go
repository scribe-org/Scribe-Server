// SPDX-License-Identifier: GPL-3.0-or-later

// Package sqlite handles the reading and migration of SQLite language data into MariaDB.
package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"

	"github.com/scribe-org/scribe-server/cmd/migrate/mariadb"
	"github.com/scribe-org/scribe-server/cmd/migrate/schema"
)

var openDatabase = sql.Open

// sqliteDir is the directory of the SQLite files that are migrated.
// put as a variable so that tests can use other directories.
var sqliteDir = "./packs/sqlite"

// ProcessSQLiteFiles processes all SQLite files in the specified directory.
func ProcessSQLiteFiles(mariaDB *sql.DB) error {
	files, err := filepath.Glob(filepath.Join(sqliteDir, "*.sqlite"))
	if err != nil {
		return fmt.Errorf("failed to read directory: %v", err)
	}

	semaphore := make(chan struct{}, 4)
	errChan := make(chan error, len(files))
	var wg sync.WaitGroup

	for _, file := range files {
		wg.Add(1)
		go func(filepath string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if err := processSQLiteFile(filepath, mariaDB); err != nil {
				errChan <- fmt.Errorf("error processing %s: %v", filepath, err)
			}
		}(file)
	}

	wg.Wait()
	close(errChan)

	// Return all errors so that a failed migration doesn't exit successfully.
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// processSQLiteFile handles processing of a single SQLite file.
func processSQLiteFile(filePath string, mariaDB *sql.DB) error {
	log.Printf("Processing file: %s", filePath)

	// Extract language code.
	base := filepath.Base(filePath)
	langCode := strings.TrimSuffix(strings.TrimPrefix(base, "TranslationData.sqlite_"), ".sqlite")

	sqlite, err := openDatabase("sqlite", filePath)
	if err != nil {
		return fmt.Errorf("failed to open SQLite file: %v", err)
	}
	defer sqlite.Close()

	tables, err := schema.GetTables(sqlite)
	if err != nil {
		return fmt.Errorf("failed to get tables: %v", err)
	}

	// Remaining tables will be still migrated if one of them fails.
	var errs []error
	for _, table := range tables {
		if err := mariadb.MigrateTable(sqlite, mariaDB, langCode, table); err != nil {
			errs = append(errs, fmt.Errorf("failed to migrate table %s: %w", table, err))
		}
	}

	return errors.Join(errs...)
}
