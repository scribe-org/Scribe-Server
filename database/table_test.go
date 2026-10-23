// SPDX-License-Identifier: GPL-3.0-or-later

package database

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MARK: Table Exists

func TestTableExists(t *testing.T) {
	query := regexp.QuoteMeta("FROM information_schema.TABLES")

	t.Run("true when the table exists", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectQuery(query).
			WithArgs(testDatabaseName, "ENLanguageDataNounsScribe").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		exists, err := TableExists("ENLanguageDataNounsScribe")

		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("false when the table doesn't exist", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectQuery(query).
			WithArgs(testDatabaseName, "FRLanguageDataNounsScribe").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		exists, err := TableExists("FRLanguageDataNounsScribe")

		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("error when the query fails", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectQuery(query).WillReturnError(errors.New("connection lost"))

		exists, err := TableExists("ENLanguageDataNounsScribe")

		assert.ErrorContains(t, err, "error checking table existence: connection lost")
		assert.False(t, exists)
	})
}

// MARK: Table Schema

func TestGetTableSchema(t *testing.T) {
	query := regexp.QuoteMeta("FROM information_schema.COLUMNS")

	t.Run("returns column names and types", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectQuery(query).
			WithArgs(testDatabaseName, "ENLanguageDataNounsScribe").
			WillReturnRows(sqlmock.NewRows([]string{"COLUMN_NAME", "COLUMN_TYPE"}).
				AddRow("wdLexemeId", "text").
				AddRow("lastModified", "timestamp"))

		schema, err := GetTableSchema("ENLanguageDataNounsScribe")

		require.NoError(t, err)
		assert.Equal(t, map[string]string{"wdLexemeId": "text", "lastModified": "timestamp"}, schema)
	})

	t.Run("rejects an invalid table name without querying", func(t *testing.T) {
		// Note: No queries are expected, so the test fails if one is sent.
		newMockDB(t)

		schema, err := GetTableSchema("ENLanguageDataNounsScribe; DROP TABLE users")

		assert.EqualError(t, err, "invalid table name")
		assert.Nil(t, schema)
	})

	t.Run("error when the query fails", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectQuery(query).WillReturnError(errors.New("connection lost"))

		_, err := GetTableSchema("ENLanguageDataNounsScribe")

		assert.ErrorContains(t, err, "error querying table schema: connection lost")
	})

	t.Run("error when a column can't be read", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectQuery(query).
			WillReturnRows(sqlmock.NewRows([]string{"COLUMN_NAME", "COLUMN_TYPE"}).AddRow(nil, "text"))

		_, err := GetTableSchema("ENLanguageDataNounsScribe")

		assert.ErrorContains(t, err, "error scanning column info")
	})

	t.Run("error when reading the rows fails", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectQuery(query).
			WillReturnRows(sqlmock.NewRows([]string{"COLUMN_NAME", "COLUMN_TYPE"}).
				AddRow("wdLexemeId", "text").
				AddRow("lastModified", "timestamp").
				RowError(1, errors.New("connection lost")))

		schema, err := GetTableSchema("ENLanguageDataNounsScribe")

		// Partial results shouldn't be returned as if they were complete.
		assert.ErrorContains(t, err, "connection lost")
		assert.Nil(t, schema)
	})
}

// MARK: Table Data

func TestGetTableData(t *testing.T) {
	query := regexp.QuoteMeta("SELECT * FROM ENLanguageDataNounsScribe")

	t.Run("returns rows with text, numbers, dates and nulls", func(t *testing.T) {
		modified := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
		mock := newMockDB(t)
		mock.ExpectQuery(query).
			WillReturnRows(sqlmock.NewRows([]string{"wdLexemeId", "count", "lastModified", "gender"}).
				AddRow([]byte("L1"), int64(3), modified, nil).
				AddRow([]byte("L2"), int64(5), modified, []byte("feminine")))

		data, err := GetTableData("ENLanguageDataNounsScribe")

		require.NoError(t, err)
		assert.Equal(t, []map[string]any{
			{"wdLexemeId": "L1", "count": int64(3), "lastModified": modified, "gender": nil},
			{"wdLexemeId": "L2", "count": int64(5), "lastModified": modified, "gender": "feminine"},
		}, data)
	})

	t.Run("no rows for an empty table", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectQuery(query).WillReturnRows(sqlmock.NewRows([]string{"wdLexemeId"}))

		data, err := GetTableData("ENLanguageDataNounsScribe")

		require.NoError(t, err)
		assert.Empty(t, data)
	})

	t.Run("rejects an invalid table name without querying", func(t *testing.T) {
		newMockDB(t)

		data, err := GetTableData("users")

		assert.EqualError(t, err, "invalid table name")
		assert.Nil(t, data)
	})

	t.Run("error when the query fails", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectQuery(query).WillReturnError(errors.New("connection lost"))

		_, err := GetTableData("ENLanguageDataNounsScribe")

		assert.ErrorContains(t, err, "error querying table data: connection lost")
	})

	t.Run("error when reading the rows fails", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectQuery(query).
			WillReturnRows(sqlmock.NewRows([]string{"wdLexemeId"}).
				AddRow([]byte("L1")).
				AddRow([]byte("L2")).
				RowError(1, errors.New("connection lost")))

		data, err := GetTableData("ENLanguageDataNounsScribe")

		// Partial results shouldn't be returned as if they were complete.
		assert.ErrorContains(t, err, "connection lost")
		assert.Nil(t, data)
	})
}
