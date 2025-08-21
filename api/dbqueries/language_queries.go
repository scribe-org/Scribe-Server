// SPDX-License-Identifier: GPL-3.0-or-later

// Package dbqueries provides helper functions that fetch and transform language-specific
// data from the underlying database.
package dbqueries

import (
	"fmt"
	"strings"

	"github.com/scribe-org/scribe-server/database"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// GetAvailableLanguages fetches all available languages from the database.
func GetAvailableLanguages() ([]string, error) {
	return database.GetAvailableLanguages()
}

// GetLanguageDataTypes fetches available data types for a specific language.
func GetLanguageDataTypes(lang string) ([]string, error) {
	return database.GetLanguageDataTypes(lang)
}

// GetLanguageVersions fetches version information for a specific language.
func GetLanguageVersions(lang string) (map[string]string, error) {
	return database.GetLanguageVersions(lang)
}

// GetLanguageTableData fetches data for a specific language table.
func GetLanguageTableData(lang, dataType string) (map[string]interface{}, error) {
	// Construct table name with the new format: ENLanguageDataNounsScribe

	caser := cases.Title(language.English)

	tableName := fmt.Sprintf("%sLanguageData%sScribe",
		strings.ToUpper(lang),
		caser.String(dataType),
	)

	// Validate table name format and existence
	if !isValidScribeTableName(tableName) {
		return nil, fmt.Errorf("invalid table name format: %s", tableName)
	}

	// Check if table exists
	exists, err := database.TableExists(tableName)
	if err != nil {
		return nil, fmt.Errorf("error checking table existence for %s: %w", tableName, err)
	}
	if !exists {
		return nil, fmt.Errorf("table %s does not exist", tableName)
	}

	// Get table schema
	schema, err := database.GetTableSchema(tableName)
	if err != nil {
		return nil, fmt.Errorf("error fetching schema for %s: %w", tableName, err)
	}

	// Get data from table
	data, err := database.GetTableData(tableName)
	if err != nil {
		return nil, fmt.Errorf("error fetching data for %s: %w", tableName, err)
	}

	return map[string]interface{}{
		"schema": schema,
		"data":   data,
	}, nil
}

// isValidScribeTableName validates the table name follows the expected pattern.
func isValidScribeTableName(tableName string) bool {
	// Expected pattern: ENLanguageDataNounsScribe, FRLanguageDataVerbsScribe
	if len(tableName) < 10 {
		return false
	}

	// Check if it starts with 2-letter language code
	if len(tableName) < 2 || !isAlpha(tableName[:2]) {
		return false
	}

	// Check if it contains LanguageData
	if !strings.Contains(tableName, "LanguageData") {
		return false
	}

	// Check if it ends with Scribe
	if !strings.HasSuffix(tableName, "Scribe") {
		return false
	}

	// Only allow alphanumeric characters
	for _, char := range tableName {
		if !isAlphaNumeric(char) {
			return false
		}
	}

	return true
}

// isAlpha checks if string contains only alphabetic characters.
func isAlpha(s string) bool {
	for _, char := range s {
		if !((char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z')) {
			return false
		}
	}
	return true
}

// isAlphaNumeric checks if character is alphanumeric.
func isAlphaNumeric(char rune) bool {
	return (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')
}
