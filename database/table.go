// SPDX-License-Identifier: GPL-3.0-or-later

package database

import (
	"fmt"

	"github.com/spf13/viper"
)

// TableExists checks if a table exists in the database.
func TableExists(tableName string) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM information_schema.TABLES 
		WHERE TABLE_SCHEMA = ? 
		AND TABLE_NAME = ?
	`

	var count int
	err := DB.QueryRow(query, viper.GetString("database.name"), tableName).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("error checking table existence: %w", err)
	}

	return count > 0, nil
}

// GetTableSchema returns the column names and types for a specific table
// in the connected MySQL/MariaDB database.
func GetTableSchema(tableName string) (map[string]string, error) {
	if !IsValidTableName(tableName) {
		return nil, fmt.Errorf("invalid table name")
	}

	query := `
		SELECT COLUMN_NAME, COLUMN_TYPE
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`

	rows, err := DB.Query(query, viper.GetString("database.name"), tableName)
	if err != nil {
		return nil, fmt.Errorf("error querying table schema: %w", err)
	}
	defer rows.Close()

	schema := make(map[string]string)
	for rows.Next() {
		var columnName, columnType string
		if err := rows.Scan(&columnName, &columnType); err != nil {
			return nil, fmt.Errorf("error scanning column info: %w", err)
		}
		schema[columnName] = columnType
	}

	return schema, nil
}

// GetTableData retrieves all rows and columns from a given table.
// The result is a slice of maps, where each map represents a row with column-value pairs.
func GetTableData(tableName string) ([]map[string]any, error) {
	if !IsValidTableName(tableName) {
		return nil, fmt.Errorf("invalid table name")
	}

	query := fmt.Sprintf("SELECT * FROM %s", tableName)

	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying table data: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("error getting columns: %w", err)
	}

	var results []map[string]any

	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		rowMap := make(map[string]any)
		for i, col := range columns {
			val := values[i]
			if val == nil {
				rowMap[col] = nil
			} else if b, ok := val.([]byte); ok {
				rowMap[col] = string(b)
			} else {
				rowMap[col] = val
			}
		}
		results = append(results, rowMap)
	}

	return results, nil
}
