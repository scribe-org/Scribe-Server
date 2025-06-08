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
	case "DATETIME", "TIMESTAMP":
		return "TIMESTAMP"
	default:
		return "TEXT"
	}
}

// MapColumnTypeToMariaDB converts SQLite types to MariaDB types, with special handling for known columns
func MapColumnTypeToMariaDB(columnName, sqliteType string) string {
	// Special handling for known timestamp columns
	if strings.EqualFold(columnName, "Lastmodified") {
		return "TIMESTAMP"
	}
	// Default type mapping
	return MapSQLiteTypeToMariaDB(sqliteType)
}
