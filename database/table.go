// SPDX-License-Identifier: GPL-3.0-or-later

package database

import (
	"fmt"

	"github.com/spf13/viper"
)

func GetTableSchema(tableName string) (map[string]string, error) {
	if !isValidTableName(tableName) {
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

func GetTableData(tableName string) ([]map[string]interface{}, error) {
	if !isValidTableName(tableName) {
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

	var results []map[string]interface{}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		rowMap := make(map[string]interface{})
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
