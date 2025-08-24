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

// GetLanguageTableData fetches data for a specific language table.
func GetLanguageTableData(lang, dataType string) (map[string]interface{}, error) {
	// Construct table name with the new format: ENLanguageDataNounsScribe

	caser := cases.Title(language.English)

	tableName := fmt.Sprintf("%sLanguageData%sScribe",
		strings.ToUpper(lang),
		caser.String(dataType),
	)

	// Validate table name format and existence
	if !database.IsValidTableName(tableName) {
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
