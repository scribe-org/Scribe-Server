// SPDX-License-Identifier: GPL-3.0-or-later

package database

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

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

func GetLanguageDataTypes(lang string) ([]string, error) {
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
		parts := strings.Split(tableName, "_")
		if len(parts) >= 2 {
			dataType := strings.Join(parts[1:], "_")
			dataTypes = append(dataTypes, dataType)
		}
	}

	return dataTypes, nil
}
