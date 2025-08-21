// SPDX-License-Identifier: GPL-3.0-or-later

package database

import "regexp"

// isValidTableName validates table names to prevent SQL injection.
func isValidTableName(tableName string) bool {
	// Pattern to match the new table structure: ENLanguageDataNounsScribe
	pattern := `^[A-Z]{2}LanguageData[A-Za-z]+Scribe$`
	matched, err := regexp.MatchString(pattern, tableName)
	if err != nil {
		return false
	}

	// Additional length check
	if len(tableName) > 100 || len(tableName) < 10 {
		return false
	}

	// Check for only alphanumeric characters (no special chars that could be used for injection)
	for _, char := range tableName {
		if !isAlphaNumeric(char) {
			return false
		}
	}

	return matched
}
