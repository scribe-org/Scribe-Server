// SPDX-License-Identifier: GPL-3.0-or-later

package dbqueries

import (
	"fmt"
	"strings"

	"github.com/scribe-org/scribe-server/database"
	"github.com/scribe-org/scribe-server/models"
)

// GetTranslationTableData fetches translation data for the given target and source language codes.
// It queries the TranslationData{TARGET}From{SOURCE} table and returns data nested as:
// word -> wordType -> wordOrder -> TranslationEntry.
func GetTranslationTableData(targetLang, sourceLang string) (map[string]map[string]map[string]models.TranslationEntry, error) {
	tableName := fmt.Sprintf("TranslationData%sFrom%s",
		strings.ToUpper(targetLang),
		strings.ToUpper(sourceLang),
	)

	if !database.IsValidTranslationTableName(tableName) {
		return nil, fmt.Errorf("invalid translation table name: %s", tableName)
	}

	exists, err := database.TableExists(tableName)
	if err != nil {
		return nil, fmt.Errorf("error checking table existence for %s: %w", tableName, err)
	}
	if !exists {
		return nil, fmt.Errorf("translation table %s does not exist", tableName)
	}

	rows, err := database.DB.Query(
		fmt.Sprintf("SELECT word, wordType, wordOrder, description, translation FROM `%s`", tableName),
	)
	if err != nil {
		return nil, fmt.Errorf("error querying %s: %w", tableName, err)
	}
	defer rows.Close()

	result := make(map[string]map[string]map[string]models.TranslationEntry)

	for rows.Next() {
		var word, wordType, wordOrder, description, translation string
		if err := rows.Scan(&word, &wordType, &wordOrder, &description, &translation); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		if result[word] == nil {
			result[word] = make(map[string]map[string]models.TranslationEntry)
		}
		if result[word][wordType] == nil {
			result[word][wordType] = make(map[string]models.TranslationEntry)
		}
		result[word][wordType][wordOrder] = models.TranslationEntry{
			Description: description,
			Translation: translation,
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return result, nil
}
