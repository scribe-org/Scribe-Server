package main

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"

	"github.com/scribe-org/scribe-server/cmd/migrate/db"
	"github.com/scribe-org/scribe-server/cmd/migrate/queries"
)

func processSQLiteFiles(mariaDB *sql.DB) error {
	sqliteDir := "./packs/sqlite"
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
				errChan <- fmt.Errorf("Error processing %s: %v", filepath, err)
			}
		}(file)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		log.Printf("%v", err)
	}

	return nil
}

func processSQLiteFile(filePath string, mariaDB *sql.DB) error {
	log.Printf("Processing file: %s", filePath)

	langCode := extractLanguageCode(filePath)

	sqlite, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return fmt.Errorf("failed to open SQLite file: %v", err)
	}
	defer sqlite.Close()

	tables, err := queries.GetTables(sqlite)
	if err != nil {
		return fmt.Errorf("failed to get tables: %v", err)
	}

	for _, table := range tables {
		if err := db.MigrateTable(sqlite, mariaDB, langCode, table); err != nil {
			log.Printf("Error migrating table %s: %v", table, err)
		}
	}

	return nil
}

func extractLanguageCode(filePath string) string {
	base := filepath.Base(filePath)
	langCode := strings.TrimSuffix(strings.TrimPrefix(base, "TranslationData.sqlite_"), ".sqlite")
	if langCode == "" {
		langCode = strings.Split(base, "_")[0]
	}
	return langCode
}

func main() {
	// Load configuration
	config, err := db.LoadConfig("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	// MariaDB connection setup with config
	mariaDB, err := db.SetupMariaDB(config.Database)
	if err != nil {
		log.Fatal(err)
	}
	defer mariaDB.Close()

	// Process SQLite files
	if err := processSQLiteFiles(mariaDB); err != nil {
		log.Fatal(err)
	}
}