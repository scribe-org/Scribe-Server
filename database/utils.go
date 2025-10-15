// SPDX-License-Identifier: GPL-3.0-or-later

package database

import (
	"regexp"
	"strings"

	"github.com/scribe-org/scribe-server/internal/constants"
	"github.com/scribe-org/scribe-server/models"
)

// IsValidTableName validates table names to prevent SQL injection.
func IsValidTableName(tableName string) bool {
	// Pattern to match the new table structure: ENLanguageDataNounsScribe.
	pattern := `^[A-Z]{2}LanguageData[A-Za-z]+Scribe$`
	matched, err := regexp.MatchString(pattern, tableName)
	if err != nil {
		return false
	}

	// Additional length check.
	if len(tableName) > 100 || len(tableName) < 10 {
		return false
	}

	// Check for only alphanumeric characters (no special chars that could be used for injection).
	for _, char := range tableName {
		if !constants.IsAlphaNumeric(char) {
			return false
		}
	}

	return matched
}

// ToIntPtr converts various numeric types to a pointer to int.
func ToIntPtr(v any) *int {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case int:
		return &val
	case int64:
		i := int(val)
		return &i
	default:
		return nil
	}
}

// ToStringPtr converts a string to a pointer to string.
func ToStringPtr(v any) *string {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case string:
		return &val
	default:
		return nil
	}
}

var languageNames = map[string]string{
	"EN": "English",
	"FR": "French",
	"DE": "German",
	"ES": "Spanish",
	"IT": "Italian",
	"PT": "Portuguese",
	"RU": "Russian",
	"SV": "Swedish",
}

// GetLanguageDisplayName returns the display name for a given language code.
func GetLanguageDisplayName(code string) string {
	if name, ok := languageNames[strings.ToUpper(code)]; ok {
		return name
	}
	return strings.ToUpper(code)
}

// BuildLanguageStatResponse constructs a LanguageStatisticsReponse object from raw stat data.
func BuildLanguageStatResponse(code string, stat map[string]any) models.LanguageStatisticsReponse {
	langCode := strings.ToUpper(code)
	langName := GetLanguageDisplayName(langCode)

	return models.LanguageStatisticsReponse{
		Code:         strings.ToLower(langCode),
		LanguageName: &langName,
		Nouns:        ToIntPtr(stat["nouns"].(int)),
		Verbs:        ToIntPtr(stat["verbs"].(int)),
	}
}
