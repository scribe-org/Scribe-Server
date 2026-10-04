// SPDX-License-Identifier: GPL-3.0-or-later

package database

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/scribe-org/scribe-server/models"
	"github.com/spf13/viper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// MARK: Get Available Languages

// GetAvailableLanguages retrieves all available languages in the database.
func GetAvailableLanguages() ([]string, error) {
	query := `
		SELECT DISTINCT SUBSTRING(TABLE_NAME, 1, 2) as language_code
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = ?
		AND TABLE_NAME LIKE '%LanguageData%Scribe'
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

// MARK: Data Types Retrieval

// GetLanguageDataTypes retrieves all available data types in a sample language table.
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

	rows, err := DB.Query(query, viper.GetString("database.name"), langPrefix+"LanguageData%Scribe")
	if err != nil {
		return nil, fmt.Errorf("error querying data types: %w", err)
	}
	defer rows.Close()

	var dataTypes []string
	// Regex to extract data type from table name like ENLanguageDataNounsScribe.
	re := regexp.MustCompile(`^[A-Z]{2}LanguageData([A-Za-z]+)Scribe$`)

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("error scanning table name: %w", err)
		}

		matches := re.FindStringSubmatch(tableName)
		if len(matches) > 1 {
			dataType := strings.ToLower(matches[1]) // convert to lowercase (nouns, verbs)
			dataTypes = append(dataTypes, dataType)
		}
	}

	return dataTypes, nil
}

// MARK: Language Statistics

// statDataTypes are the data types counted in language statistics in the order they're queried.
var statDataTypes = []string{"nouns", "verbs", "prepositions", "profanity"}

// requiredStatDataTypes are the data types a language needs to be included in statistics.
// Other data types such as prepositions are only available for some languages.
var requiredStatDataTypes = []string{"nouns", "verbs"}

// buildLanguageStatQuery builds a query that counts the entries of each statistics data type available for a language.
// It returns the query and the data types it counts in order, or no data types if a required data type is missing.
func buildLanguageStatQuery(lang string, availableDataTypes []string) (string, []string, error) {
	available := make(map[string]bool, len(availableDataTypes))
	for _, dataType := range availableDataTypes {
		available[strings.ToLower(dataType)] = true
	}

	for _, dataType := range requiredStatDataTypes {
		if !available[dataType] {
			return "", nil, nil
		}
	}

	caser := cases.Title(language.Und)
	counts := make([]string, 0, len(statDataTypes))
	countedDataTypes := make([]string, 0, len(statDataTypes))

	for _, dataType := range statDataTypes {
		if !available[dataType] {
			continue
		}

		// Construct the table name format: ENLanguageDataNounsScribe.
		tableName := strings.ToUpper(lang) + "LanguageData" + caser.String(dataType) + "Scribe"
		if !IsValidTableName(tableName) {
			return "", nil, fmt.Errorf("invalid table name for language: %s", lang)
		}

		counts = append(counts, fmt.Sprintf("(SELECT COUNT(*) FROM %s) AS %s", tableName, dataType))
		countedDataTypes = append(countedDataTypes, dataType)
	}

	return "SELECT " + strings.Join(counts, ", "), countedDataTypes, nil
}

// GetLanguageStat retrieves statistics for the data types of a specific language.
func GetLanguageStat(lan string) (map[string]any, error) {
	// Normalize and validate language code (e.g., "EN", "FR").
	lang := strings.ToUpper(strings.TrimSpace(lan))
	if !regexp.MustCompile(`^[A-Z]{2}$`).MatchString(lang) {
		return nil, fmt.Errorf("invalid language code: %s", lang)
	}

	dataTypes, err := GetLanguageDataTypes(lang)
	if err != nil {
		return nil, fmt.Errorf("error fetching data types for %s: %w", lang, err)
	}

	query, countedDataTypes, err := buildLanguageStatQuery(lang, dataTypes)
	if err != nil {
		return nil, err
	}

	if len(countedDataTypes) == 0 {
		log.Printf("⚠️ Skipping %s — missing nouns or verbs table", lang)
		return nil, nil
	}

	// Execute the query.
	counts := make([]int, len(countedDataTypes))
	scanTargets := make([]any, len(counts))
	for i := range counts {
		scanTargets[i] = &counts[i]
	}

	err = DB.QueryRow(query).Scan(scanTargets...)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1146 {
			log.Printf("⚠️ Skipping %s — missing table: %s", lang, mysqlErr.Message)
			return nil, nil
		}
		if strings.Contains(err.Error(), "doesn't exist") ||
			strings.Contains(err.Error(), "no such table") {
			log.Printf("⚠️ Skipping %s — missing table(s): %v", lang, err)
			return nil, nil
		}
		return nil, fmt.Errorf("error scanning stats for %s: %w", lang, err)
	}

	stat := map[string]any{"code": strings.ToLower(lang)}
	for i, dataType := range countedDataTypes {
		stat[dataType] = counts[i]
	}

	return stat, nil
}

// GetAllLanguageStats retrieves statistics for all available languages that have nouns and verbs.
func GetAllLanguageStats() ([]models.LanguageStatisticsReponse, error) {
	availableLanguages, err := GetAvailableLanguages()
	if err != nil {
		return nil, fmt.Errorf("failed to get available languages: %w", err)
	}

	allStats := []models.LanguageStatisticsReponse{}

	for _, lan := range availableLanguages {
		stat, err := GetLanguageStat(lan)
		if err != nil {
			log.Printf("⚠️ Error fetching stats for %s: %v", lan, err)
			continue
		}
		if stat == nil {
			continue
		}

		allStats = append(allStats, BuildLanguageStatResponse(lan, stat))
	}

	return allStats, nil
}
