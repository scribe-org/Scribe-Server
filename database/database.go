// SPDX-License-Identifier: GPL-3.0-or-later

package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

var DB *sql.DB

// InitDatabase initializes the database connection
func InitDatabase() error {
	user := viper.GetString("database.user")
	password := viper.GetString("database.password")
	host := viper.GetString("database.host")
	port := viper.GetString("database.port")
	name := viper.GetString("database.name")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, name)

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}

	// Test the connection
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}

	log.Println("✅ Database connection established")
	return nil
}

// GetAvailableLanguages returns languages that have data in the database
func GetAvailableLanguages() ([]string, error) {
	query := `
		SELECT DISTINCT SUBSTRING(TABLE_NAME, 1, 2) as language_code
		FROM information_schema.TABLES 
		WHERE TABLE_SCHEMA = ? 
		AND TABLE_NAME LIKE '%LanguageData_%'
		ORDER BY language_code
	`

	rows, err := DB.Query(query, viper.GetString("database.name"))
	if err != nil {
		return nil, fmt.Errorf("error querying available languages: %w", err)
	}
	defer rows.Close()

	var languages []string
	for rows.Next() {
		var lang string
		if err := rows.Scan(&lang); err != nil {
			return nil, fmt.Errorf("error scanning language: %w", err)
		}
		languages = append(languages, strings.ToLower(lang))
	}

	return languages, nil
}

// GetLanguageDataTypes returns available data types for a language
func GetLanguageDataTypes(lang string) ([]string, error) {
	// Sanitize input - only allow 2-letter codes
	if len(lang) != 2 {
		return nil, fmt.Errorf("invalid language code")
	}

	langPrefix := strings.ToUpper(lang)

	query := `
		SELECT TABLE_NAME
		FROM information_schema.TABLES 
		WHERE TABLE_SCHEMA = ? 
		AND TABLE_NAME LIKE ?
		ORDER BY TABLE_NAME
	`

	rows, err := DB.Query(query, viper.GetString("database.name"), langPrefix+"LanguageData_%")
	if err != nil {
		return nil, fmt.Errorf("error querying data types: %w", err)
	}
	defer rows.Close()

	var dataTypes []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("error scanning table name: %w", err)
		}

		// Extract data type from table name (e.g., "ENLanguageData_nouns" -> "nouns")
		parts := strings.Split(tableName, "_")
		if len(parts) >= 2 {
			dataType := strings.Join(parts[1:], "_") // Handle multi-part names like "personal_pronouns"
			dataTypes = append(dataTypes, dataType)
		}
	}

	return dataTypes, nil
}

// GetTableSchema returns the schema for a specific table
func GetTableSchema(tableName string) (map[string]string, error) {
	// Sanitize table name - only allow alphanumeric and underscore
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

// GetTableData returns all data from a specific table
func GetTableData(tableName string) ([]map[string]interface{}, error) {
	// Sanitize table name
	if !isValidTableName(tableName) {
		return nil, fmt.Errorf("invalid table name")
	}


	query := fmt.Sprintf("SELECT * FROM %s", tableName)

	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying table data: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("error getting columns: %w", err)
	}

	var results []map[string]interface{}

	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		// Scan the row into the value pointers
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		// Create a map for this row
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]

			// Handle null values and convert byte slices to strings
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

// GetLanguageVersions returns the last modified dates for a language's data types
func GetLanguageVersions(lang string) (map[string]string, error) {
	// For now, we'll use a simple approach - check if `language_data_versions` table exists
	// If not, return current date for all data types

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

// isValidTableName checks if a table name is safe to use in queries
func isValidTableName(tableName string) bool {
	// Only allow alphanumeric characters and underscores
	for _, char := range tableName {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_') {
			return false
		}
	}
	return len(tableName) > 0 && len(tableName) <= 64 // MySQL table name limit
}

// CreateLanguageDataVersionsTable creates the versions tracking table if it doesn't exist
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

// UpdateLanguageVersion updates the last modified timestamp for a language
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
