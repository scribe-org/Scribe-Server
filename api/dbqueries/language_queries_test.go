// SPDX-License-Identifier: GPL-3.0-or-later

package dbqueries

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetLanguageTableData tests fetching schema and data for language tables under various database scenarios.
func TestGetLanguageTableData(t *testing.T) {
	tableExistsQuery := regexp.QuoteMeta("FROM information_schema.TABLES")
	schemaQuery := regexp.QuoteMeta("FROM information_schema.COLUMNS")
	dataQuery := regexp.QuoteMeta("SELECT * FROM ENLanguageDataNounsScribe")

	t.Run("returns schema and data for a valid language table", func(t *testing.T) {
		modified := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
		mock := newMockDB(t)

		mock.ExpectQuery(tableExistsQuery).
			WithArgs(testDatabaseName, "ENLanguageDataNounsScribe").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		mock.ExpectQuery(schemaQuery).
			WithArgs(testDatabaseName, "ENLanguageDataNounsScribe").
			WillReturnRows(sqlmock.NewRows([]string{"COLUMN_NAME", "COLUMN_TYPE"}).
				AddRow("wdLexemeId", "text").
				AddRow("noun", "text").
				AddRow("lastModified", "timestamp"))

		mock.ExpectQuery(dataQuery).
			WillReturnRows(sqlmock.NewRows([]string{"wdLexemeId", "noun", "lastModified"}).
				AddRow([]byte("L1"), []byte("Haus"), modified).
				AddRow([]byte("L2"), []byte("Baum"), modified))

		result, err := GetLanguageTableData("en", "nouns")

		require.NoError(t, err)
		assert.Equal(t, map[string]string{
			"wdLexemeId":   "text",
			"noun":         "text",
			"lastModified": "timestamp",
		}, result["schema"])
		assert.Equal(t, []map[string]any{
			{"wdLexemeId": "L1", "noun": "Haus", "lastModified": modified},
			{"wdLexemeId": "L2", "noun": "Baum", "lastModified": modified},
		}, result["data"])
	})

	t.Run("rejects an invalid table name format without querying", func(t *testing.T) {
		newMockDB(t)

		result, err := GetLanguageTableData("eng", "nouns")

		assert.EqualError(t, err, "invalid table name format: ENGLanguageDataNounsScribe")
		assert.Nil(t, result)
	})

	t.Run("rejects invalid characters in dataType without querying", func(t *testing.T) {
		newMockDB(t)

		result, err := GetLanguageTableData("en", "nouns; DROP TABLE users")

		assert.EqualError(t, err, "invalid table name format: ENLanguageDataNouns; Drop Table UsersScribe")
		assert.Nil(t, result)
	})

	t.Run("error when checking table existence fails", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectQuery(tableExistsQuery).
			WithArgs(testDatabaseName, "ENLanguageDataNounsScribe").
			WillReturnError(errors.New("connection lost"))

		result, err := GetLanguageTableData("en", "nouns")

		assert.ErrorContains(t, err, "error checking table existence for ENLanguageDataNounsScribe: error checking table existence: connection lost")
		assert.Nil(t, result)
	})

	t.Run("error when table does not exist", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectQuery(tableExistsQuery).
			WithArgs(testDatabaseName, "ENLanguageDataNounsScribe").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		result, err := GetLanguageTableData("en", "nouns")

		assert.EqualError(t, err, "table ENLanguageDataNounsScribe does not exist")
		assert.Nil(t, result)
	})

	t.Run("error when fetching table schema fails", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectQuery(tableExistsQuery).
			WithArgs(testDatabaseName, "ENLanguageDataNounsScribe").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		mock.ExpectQuery(schemaQuery).
			WithArgs(testDatabaseName, "ENLanguageDataNounsScribe").
			WillReturnError(errors.New("schema query failed"))

		result, err := GetLanguageTableData("en", "nouns")

		assert.ErrorContains(t, err, "error fetching schema for ENLanguageDataNounsScribe: error querying table schema: schema query failed")
		assert.Nil(t, result)
	})

	t.Run("error when fetching table data fails", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectQuery(tableExistsQuery).
			WithArgs(testDatabaseName, "ENLanguageDataNounsScribe").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		mock.ExpectQuery(schemaQuery).
			WithArgs(testDatabaseName, "ENLanguageDataNounsScribe").
			WillReturnRows(sqlmock.NewRows([]string{"COLUMN_NAME", "COLUMN_TYPE"}).
				AddRow("wdLexemeId", "text"))

		mock.ExpectQuery(dataQuery).
			WillReturnError(errors.New("data query failed"))

		result, err := GetLanguageTableData("en", "nouns")

		assert.ErrorContains(t, err, "error fetching data for ENLanguageDataNounsScribe: error querying table data: data query failed")
		assert.Nil(t, result)
	})
}
