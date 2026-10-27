// SPDX-License-Identifier: GPL-3.0-or-later

package sqlite

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MARK: Helpers

// setupPacksDir changes to a temporary directory with an empty packs/sqlite directory for a test.
func setupPacksDir(t *testing.T) {
	t.Helper()

	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "packs", "sqlite"), 0o755))
	t.Chdir(dir)
}

// createSQLiteFile creates a SQLite file in packs/sqlite and runs the given statements on it.
func createSQLiteFile(t *testing.T, name string, statements ...string) string {
	t.Helper()

	path := filepath.Join("packs", "sqlite", name)
	db, err := sql.Open("sqlite", path)
	require.NoError(t, err)
	defer func() { require.NoError(t, db.Close()) }()

	for _, statement := range statements {
		_, err := db.Exec(statement)
		require.NoError(t, err)
	}

	return path
}

// newMockDB creates a sqlmock database in place of MariaDB.
func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() {
		// Every expected query must have been sent.
		assert.NoError(t, mock.ExpectationsWereMet())
		// Note: Closing isn't an expected call for sqlmock, so its error isn't checked.
		_ = db.Close()
	})

	return db, mock
}

// expectMigration expects a new MariaDB table to be created and filled with a single row.
func expectMigration(mock sqlmock.Sqlmock, tableName string, columns string, values ...driver.Value) {
	mock.ExpectQuery(regexp.QuoteMeta("FROM information_schema.tables")).
		WithArgs(tableName).
		WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(0))
	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS `" + tableName + "`")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectBegin()
	mock.ExpectPrepare(regexp.QuoteMeta("INSERT IGNORE INTO `" + tableName + "` (" + columns + ")")).
		ExpectExec().WithArgs(values...).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
}

// MARK: Process Files

func TestProcessSQLiteFiles(t *testing.T) {
	t.Run("migrates the SQLite files in packs/sqlite", func(t *testing.T) {
		setupPacksDir(t)
		createSQLiteFile(t, "ENLanguageData.sqlite",
			"CREATE TABLE nouns (wdLexemeId TEXT, noun TEXT)",
			"INSERT INTO nouns VALUES ('L1', 'house')",
		)
		// Note: Only .sqlite files are migrated.
		require.NoError(t, os.WriteFile(filepath.Join("packs", "sqlite", "README.md"), []byte("not data"), 0o644))
		mariaDB, mock := newMockDB(t)
		expectMigration(mock, "ENLanguageDataNounsScribe", "`wdLexemeId`, `noun`", "L1", "house")

		err := ProcessSQLiteFiles(mariaDB)

		require.NoError(t, err)
	})

	t.Run("error when the directory can't be searched", func(t *testing.T) {
		original := sqliteDir
		// An unclosed bracket is an invalid pattern for finding the files.
		sqliteDir = "packs/[sqlite"
		t.Cleanup(func() { sqliteDir = original })
		mariaDB, _ := newMockDB(t)

		err := ProcessSQLiteFiles(mariaDB)

		assert.EqualError(t, err, "failed to read directory: syntax error in pattern")
	})

	t.Run("does nothing without SQLite files", func(t *testing.T) {
		setupPacksDir(t)
		// Note: No MariaDB queries are expected, so the test fails if one is sent.
		mariaDB, _ := newMockDB(t)

		err := ProcessSQLiteFiles(mariaDB)

		require.NoError(t, err)
	})

	t.Run("returns the errors of every file that fails", func(t *testing.T) {
		setupPacksDir(t)
		createSQLiteFile(t, "DELanguageData.sqlite", "CREATE TABLE nouns (noun TEXT)")
		createSQLiteFile(t, "ENLanguageData.sqlite", "CREATE TABLE nouns (noun TEXT)")
		mariaDB, mock := newMockDB(t)
		// Note: Files are migrated concurrently, so the order of the queries isn't known.
		mock.MatchExpectationsInOrder(false)
		for _, tableName := range []string{"DELanguageDataNounsScribe", "ENLanguageDataNounsScribe"} {
			mock.ExpectQuery(regexp.QuoteMeta("FROM information_schema.tables")).
				WithArgs(tableName).
				WillReturnError(errors.New("connection lost"))
		}

		err := ProcessSQLiteFiles(mariaDB)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "error processing packs/sqlite/DELanguageData.sqlite: failed to migrate table nouns")
		assert.Contains(t, err.Error(), "error processing packs/sqlite/ENLanguageData.sqlite: failed to migrate table nouns")
	})
}

// MARK: Process File

func TestProcessSQLiteFile(t *testing.T) {
	t.Run("migrates translation tables", func(t *testing.T) {
		setupPacksDir(t)
		path := createSQLiteFile(t, "TranslationData.sqlite",
			"CREATE TABLE bn_translations_from_de (word TEXT, translation TEXT)",
			"INSERT INTO bn_translations_from_de VALUES ('Haus', 'বাড়ি')",
		)
		mariaDB, mock := newMockDB(t)
		expectMigration(mock, "TranslationDataBNFromDE", "`word`, `translation`", "Haus", "বাড়ি")

		err := processSQLiteFile(path, mariaDB)

		require.NoError(t, err)
	})

	t.Run("migrates the other tables when one fails", func(t *testing.T) {
		setupPacksDir(t)
		path := createSQLiteFile(t, "ENLanguageData.sqlite",
			"CREATE TABLE nouns (noun TEXT)",
			"CREATE TABLE verbs (infinitive TEXT)",
			"INSERT INTO verbs VALUES ('run')",
		)
		mariaDB, mock := newMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta("FROM information_schema.tables")).
			WithArgs("ENLanguageDataNounsScribe").
			WillReturnError(errors.New("connection lost"))
		expectMigration(mock, "ENLanguageDataVerbsScribe", "`infinitive`", "run")

		err := processSQLiteFile(path, mariaDB)

		assert.EqualError(t, err, "failed to migrate table nouns: failed to check table existence: connection lost")
	})

	t.Run("error when the file isn't a SQLite database", func(t *testing.T) {
		setupPacksDir(t)
		path := filepath.Join("packs", "sqlite", "ENLanguageData.sqlite")
		require.NoError(t, os.WriteFile(path, []byte("this isn't a SQLite database, it's just some text"), 0o644))
		// Note: No MariaDB queries are expected, so the test fails if one is sent.
		mariaDB, _ := newMockDB(t)

		err := processSQLiteFile(path, mariaDB)

		assert.ErrorContains(t, err, "failed to get tables")
	})

	t.Run("error when the file can't be opened", func(t *testing.T) {
		original := openDatabase
		openDatabase = func(string, string) (*sql.DB, error) { return nil, errors.New("unknown driver") }
		t.Cleanup(func() { openDatabase = original })
		mariaDB, _ := newMockDB(t)

		err := processSQLiteFile("packs/sqlite/ENLanguageData.sqlite", mariaDB)

		assert.EqualError(t, err, "failed to open SQLite file: unknown driver")
	})
}
