// SPDX-License-Identifier: GPL-3.0-or-later

package database

import (
	"regexp"
	"strings"

	"github.com/scribe-org/scribe-server/models"
)

// MARK: Validation Helpers

// Note: The patterns only allow letters, so table names can't include characters that could be used for SQL injection.
var (
	// tableNameRegex matches language data table names such as ENLanguageDataNounsScribe.
	tableNameRegex = regexp.MustCompile(`^[A-Z]{2}LanguageData[A-Za-z]+Scribe$`)

	// translationTableNameRegex matches translation table names such as TranslationDataBNFromDE.
	translationTableNameRegex = regexp.MustCompile(`^TranslationData[A-Z]{2,4}From[A-Z]{2,4}$`)
)

// maxTableNameLength is the maximum length of a language data table name.
const maxTableNameLength = 100

// IsValidTableName validates table names to prevent SQL injection.
func IsValidTableName(tableName string) bool {
	return len(tableName) <= maxTableNameLength && tableNameRegex.MatchString(tableName)
}

// IsValidTranslationTableName validates TranslationData table names.
// Pattern: TranslationData{TARGET}From{SOURCE} e.g. TranslationDataBNFromDE.
// Note: The pattern limits table names to 23-27 characters.
func IsValidTranslationTableName(tableName string) bool {
	return translationTableNameRegex.MatchString(tableName)
}

// MARK: Pointer Conversion

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

// MARK: Language Metadata

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

// MARK: Response Builders

// BuildLanguageStatResponse constructs a LanguageStatisticsResponse object from raw stat data.
func BuildLanguageStatResponse(code string, stat map[string]any) models.LanguageStatisticsReponse {
	langCode := strings.ToUpper(code)
	langName := GetLanguageDisplayName(langCode)

	return models.LanguageStatisticsReponse{
		Code:         strings.ToLower(langCode),
		LanguageName: &langName,
		Nouns:        ToIntPtr(stat["nouns"]),
		Verbs:        ToIntPtr(stat["verbs"]),
		Prepositions: ToIntPtr(stat["prepositions"]),
		Profanity:    ToIntPtr(stat["profanity"]),
	}
}
