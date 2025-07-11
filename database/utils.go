// SPDX-License-Identifier: GPL-3.0-or-later

package database

func isValidTableName(tableName string) bool {
	for _, char := range tableName {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_') {
			return false
		}
	}
	return len(tableName) > 0 && len(tableName) <= 64
}
