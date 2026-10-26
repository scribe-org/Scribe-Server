// SPDX-License-Identifier: GPL-3.0-or-later

package schema

import (
	"database/sql"
	"errors"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MARK: Helpers

// newSQLiteDB creates a SQLite database file for a test and runs the given statements on it.
func newSQLiteDB(t *testing.T, statements ...string) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "test.sqlite"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	for _, statement := range statements {
		_, err := db.Exec(statement)
		require.NoError(t, err)
	}

	return db
}

// newMockDB creates a sqlmock database for errors that can't be caused with a real SQLite database.
func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() {
		// Every expected query must have been sent.
		assert.NoError(t, mock.ExpectationsWereMet())
		_ = db.Close()
	})

	return db, mock
}

// MARK: Get Tables

func TestGetTables(t *testing.T) {
	query := regexp.QuoteMeta("FROM sqlite_master")

	t.Run("returns the data tables", func(t *testing.T) {
		db := newSQLiteDB(t,
			"CREATE TABLE nouns (id INTEGER PRIMARY KEY AUTOINCREMENT, noun TEXT)",
			"CREATE TABLE verbs (wdLexemeId TEXT, infinitive TEXT)",
		)

		tables, err := GetTables(db)

		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"nouns", "verbs"}, tables)
	})

	t.Run("returns no tables for an empty database", func(t *testing.T) {
		db := newSQLiteDB(t)

		tables, err := GetTables(db)

		require.NoError(t, err)
		assert.Empty(t, tables)
	})

	t.Run("error when the query fails", func(t *testing.T) {
		db, mock := newMockDB(t)
		mock.ExpectQuery(query).WillReturnError(errors.New("database is locked"))

		tables, err := GetTables(db)

		assert.EqualError(t, err, "database is locked")
		assert.Nil(t, tables)
	})

	t.Run("error when a table name can't be read", func(t *testing.T) {
		db, mock := newMockDB(t)
		mock.ExpectQuery(query).WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow(nil))

		tables, err := GetTables(db)

		assert.ErrorContains(t, err, "converting NULL to string")
		assert.Nil(t, tables)
	})

	t.Run("error when reading the rows fails", func(t *testing.T) {
		db, mock := newMockDB(t)
		mock.ExpectQuery(query).
			WillReturnRows(sqlmock.NewRows([]string{"name"}).
				AddRow("nouns").
				AddRow("verbs").
				RowError(1, errors.New("disk I/O error")))

		tables, err := GetTables(db)

		// Partial results shouldn't be returned as if they were complete.
		assert.EqualError(t, err, "disk I/O error")
		assert.Nil(t, tables)
	})
}

// MARK: Get Table Schema

func TestGetTableSchema(t *testing.T) {
	pragmaColumns := []string{"cid", "name", "type", "notnull", "dflt_value", "pk"}

	t.Run("returns the columns in order with MariaDB types", func(t *testing.T) {
		db := newSQLiteDB(t, `CREATE TABLE nouns (
			wdLexemeId TEXT,
			count INTEGER,
			score REAL,
			image BLOB,
			created DATETIME,
			lastModified TEXT,
			gender VARCHAR(10)
		)`)

		schema, err := GetTableSchema(db, "nouns")

		require.NoError(t, err)
		assert.Equal(t,
			[]string{"wdLexemeId", "count", "score", "image", "created", "lastModified", "gender"},
			schema.ColumnNames,
		)
		assert.Equal(t,
			// Note: lastModified is a TIMESTAMP although it's stored as TEXT in SQLite.
			[]string{"TEXT", "BIGINT", "DOUBLE", "BLOB", "TIMESTAMP", "TIMESTAMP", "TEXT"},
			schema.ColumnTypes,
		)
	})

	t.Run("reads tables with names that need quoting", func(t *testing.T) {
		db := newSQLiteDB(t, "CREATE TABLE `emoji-keywords` (word TEXT)")

		schema, err := GetTableSchema(db, "emoji-keywords")

		require.NoError(t, err)
		assert.Equal(t, []string{"word"}, schema.ColumnNames)
		assert.Equal(t, []string{"TEXT"}, schema.ColumnTypes)
	})

	t.Run("returns no columns for a table that doesn't exist", func(t *testing.T) {
		db := newSQLiteDB(t)

		schema, err := GetTableSchema(db, "nouns")

		require.NoError(t, err)
		assert.Empty(t, schema.ColumnNames)
		assert.Empty(t, schema.ColumnTypes)
	})

	t.Run("error when the query fails", func(t *testing.T) {
		db, mock := newMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta("PRAGMA table_info(`nouns`)")).
			WillReturnError(errors.New("database is locked"))

		schema, err := GetTableSchema(db, "nouns")

		assert.EqualError(t, err, "database is locked")
		assert.Nil(t, schema)
	})

	t.Run("error when a column can't be read", func(t *testing.T) {
		db, mock := newMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta("PRAGMA table_info(`nouns`)")).
			WillReturnRows(sqlmock.NewRows(pragmaColumns).AddRow("not a number", "noun", "TEXT", false, nil, false))

		schema, err := GetTableSchema(db, "nouns")

		assert.ErrorContains(t, err, "converting driver.Value type string")
		assert.Nil(t, schema)
	})

	t.Run("error when reading the rows fails", func(t *testing.T) {
		db, mock := newMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta("PRAGMA table_info(`nouns`)")).
			WillReturnRows(sqlmock.NewRows(pragmaColumns).
				AddRow(0, "wdLexemeId", "TEXT", false, nil, false).
				AddRow(1, "noun", "TEXT", false, nil, false).
				RowError(1, errors.New("disk I/O error")))

		schema, err := GetTableSchema(db, "nouns")

		// Partial results shouldn't be returned as if they were complete.
		assert.EqualError(t, err, "disk I/O error")
		assert.Nil(t, schema)
	})
}
