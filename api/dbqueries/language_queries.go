// SPDX-License-Identifier: GPL-3.0-or-later

// Package dbqueries provides helper functions that fetch and transform language-specific
// data from the underlying database.
package dbqueries

import (
	"fmt"
	"strings"

	"github.com/scribe-org/scribe-server/database"
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
	// Construct table name.
	tableName := fmt.Sprintf("%sLanguageData_%s", strings.ToUpper(lang), dataType)

	// Get table schema.
	schema, err := database.GetTableSchema(tableName)
	if err != nil {
		return nil, fmt.Errorf("error fetching schema for %s: %w", tableName, err)
	}

	// Get data from table.
	data, err := database.GetTableData(tableName)
	if err != nil {
		return nil, fmt.Errorf("error fetching data for %s: %w", tableName, err)
	}

	return map[string]interface{}{
		"schema": schema,
		"data":   data,
	}, nil
}
