// SPDX-License-Identifier: GPL-3.0-or-later

package database

import (
	"fmt"
	"log"
	"time"
)

// CreateLanguageDataVersionsTable creates the `language_data_versions` table if it does not already exist.
// This table tracks the last updated time for each language's dataset.
func CreateLanguageDataVersionsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS language_data_versions (
			language_iso VARCHAR(2) PRIMARY KEY,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`

	_, err := DB.Exec(query)
	if err != nil {
		return fmt.Errorf("error creating language_data_versions table: %w", err)
	}

	log.Println("✅ language_data_versions table ready")
	return nil
}

// UpdateLanguageVersion updates the `updated_at` timestamp for a specific language in the `language_data_versions` table.
// If the language does not exist, it inserts a new row.
func UpdateLanguageVersion(lang string) error {
	query := `
		INSERT INTO language_data_versions (language_iso, updated_at) 
		VALUES (?, NOW()) 
		ON DUPLICATE KEY UPDATE updated_at = NOW()
	`

	_, err := DB.Exec(query, lang)
	if err != nil {
		return fmt.Errorf("error updating language version: %w", err)
	}

	return nil
}

// GetLanguageVersions returns a map of data types to their last modified date for the given language.
// Currently, it uses the current date as a placeholder until version tracking is implemented per data type.
func GetLanguageVersions(lang string) (map[string]string, error) {
	dataTypes, err := GetLanguageDataTypes(lang)
	if err != nil {
		return nil, err
	}

	versions := make(map[string]string)
	currentDate := time.Now().Format("2006-01-02")

	for _, dataType := range dataTypes {
		versions[dataType+"_last_modified"] = currentDate
	}

	return versions, nil
}
