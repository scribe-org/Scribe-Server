// SPDX-License-Identifier: GPL-3.0-or-later

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapSQLiteTypeToMariaDB(t *testing.T) {
	tests := []struct {
		sqliteType string
		expected   string
	}{
		{"TEXT", "TEXT"},
		{"INTEGER", "BIGINT"},
		{"REAL", "DOUBLE"},
		{"BLOB", "BLOB"},
		{"DATETIME", "TIMESTAMP"},
		{"TIMESTAMP", "TIMESTAMP"},
		// Types are matched regardless of case.
		{"integer", "BIGINT"},
		{"Real", "DOUBLE"},
		// Other types are stored as text.
		{"VARCHAR(10)", "TEXT"},
		{"NUMERIC", "TEXT"},
		{"", "TEXT"},
	}

	for _, tt := range tests {
		t.Run(tt.sqliteType, func(t *testing.T) {
			assert.Equal(t, tt.expected, MapSQLiteTypeToMariaDB(tt.sqliteType))
		})
	}
}

func TestMapColumnTypeToMariaDB(t *testing.T) {
	tests := []struct {
		name       string
		columnName string
		sqliteType string
		expected   string
	}{
		{"last modified column is a timestamp", "lastModified", "TEXT", "TIMESTAMP"},
		{"last modified column regardless of case", "LASTMODIFIED", "TEXT", "TIMESTAMP"},
		{"other columns use their type", "count", "INTEGER", "BIGINT"},
		{"other text columns", "noun", "TEXT", "TEXT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, MapColumnTypeToMariaDB(tt.columnName, tt.sqliteType))
		})
	}
}
