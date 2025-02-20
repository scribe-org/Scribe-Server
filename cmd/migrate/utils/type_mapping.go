// SPDX-License-Identifier: GPL-3.0-or-later
package utils

import "strings"

// MapSQLiteTypeToMariaDB converts SQLite types to MariaDB types.
func MapSQLiteTypeToMariaDB(sqliteType string) string {
	switch strings.ToUpper(sqliteType) {
	case "TEXT":
		return "TEXT"
	case "INTEGER":
		return "BIGINT"
	case "REAL":
		return "DOUBLE"
	case "BLOB":
		return "BLOB"
	default:
		return "TEXT"
	}
} 
