// SPDX-License-Identifier: GPL-3.0-or-later

package mariadb

import (
	"database/sql"
	"errors"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/scribe-org/scribe-server/cmd/migrate/types"
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

// newNounsDB creates a SQLite database with a nouns table that has two rows.
func newNounsDB(t *testing.T) *sql.DB {
	t.Helper()

	return newSQLiteDB(t,
		"CREATE TABLE nouns (wdLexemeId TEXT, noun TEXT)",
		"INSERT INTO nouns VALUES ('L1', 'Haus'), ('L2', 'Baum')",
	)
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

// setBatchSize changes the number of rows that are inserted per batch for a test.
func setBatchSize(t *testing.T, size int) {
	t.Helper()

	original := batchSize
	batchSize = size
	t.Cleanup(func() { batchSize = original })
}

const (
	nounsTable  = "ENLanguageDataNounsScribe"
	backupTable = "ENLanguageDataNounsScribeOld"
)

var (
	tableExistsQuery = regexp.QuoteMeta("FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?")
	createQuery      = regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS `" + nounsTable + "`")
	insertQuery      = regexp.QuoteMeta("INSERT IGNORE INTO `" + nounsTable + "` (`wdLexemeId`, `noun`) VALUES (?,?)")
	dropNewQuery     = regexp.QuoteMeta("DROP TABLE IF EXISTS `" + nounsTable + "`")
	dropBackupQuery  = regexp.QuoteMeta("DROP TABLE IF EXISTS `" + backupTable + "`")
	backupQuery      = regexp.QuoteMeta("RENAME TABLE `" + nounsTable + "` TO `" + backupTable + "`")
	restoreQuery     = regexp.QuoteMeta("RENAME TABLE `" + backupTable + "` TO `" + nounsTable + "`")
)

// expectTableExists expects the check for whether the nouns table exists in MariaDB.
func expectTableExists(mock sqlmock.Sqlmock, exists bool) {
	count := 0
	if exists {
		count = 1
	}

	mock.ExpectQuery(tableExistsQuery).
		WithArgs(nounsTable).
		WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(count))
}

// expectBackup expects the existing nouns table to be renamed to the backup table.
func expectBackup(mock sqlmock.Sqlmock) {
	mock.ExpectExec(dropBackupQuery).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(backupQuery).WillReturnResult(sqlmock.NewResult(0, 0))
}

// expectCopy expects the two rows of the nouns table to be inserted in a transaction.
func expectCopy(mock sqlmock.Sqlmock) {
	mock.ExpectBegin()
	insert := mock.ExpectPrepare(insertQuery)
	insert.ExpectExec().WithArgs("L1", "Haus").WillReturnResult(sqlmock.NewResult(0, 1))
	insert.ExpectExec().WithArgs("L2", "Baum").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
}

// MARK: Table Naming

func TestGenerateMariaTableName(t *testing.T) {
	tests := []struct {
		name      string
		langCode  string
		tableName string
		expected  string
	}{
		{"language data table", "ENLanguageData", "nouns", "ENLanguageDataNounsScribe"},
		{"language data table with sqlite prefix", "DELanguageData", "sqlite_prepositions", "DELanguageDataPrepositionsScribe"},
		{"language data table with underscores", "ENLanguageData", "emoji_keywords", "ENLanguageDataEmojikeywordsScribe"},
		{"translation table", "TranslationData", "bn_translations_from_de", "TranslationDataBNFromDE"},
		{"translation table with sqlite prefix", "TranslationData", "sqlite_en_translations_from_fr", "TranslationDataENFromFR"},
		{"other translation table", "TranslationData", "en_translations", "TranslationDataEntranslations"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, generateMariaTableName(tt.langCode, tt.tableName))
		})
	}
}

// MARK: Table Check

func TestTableExists(t *testing.T) {
	t.Run("true when the table exists", func(t *testing.T) {
		db, mock := newMockDB(t)
		expectTableExists(mock, true)

		exists, err := tableExists(db, nounsTable)

		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("false when the table doesn't exist", func(t *testing.T) {
		db, mock := newMockDB(t)
		expectTableExists(mock, false)

		exists, err := tableExists(db, nounsTable)

		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("error when the query fails", func(t *testing.T) {
		db, mock := newMockDB(t)
		mock.ExpectQuery(tableExistsQuery).WillReturnError(errors.New("connection lost"))

		exists, err := tableExists(db, nounsTable)

		assert.EqualError(t, err, "connection lost")
		assert.False(t, exists)
	})
}

// MARK: Migrate Table

func TestMigrateTable(t *testing.T) {
	t.Run("creates and fills a new table", func(t *testing.T) {
		mariaDB, mock := newMockDB(t)
		expectTableExists(mock, false)
		mock.ExpectExec(createQuery).WillReturnResult(sqlmock.NewResult(0, 0))
		expectCopy(mock)

		err := MigrateTable(newNounsDB(t), mariaDB, "ENLanguageData", "nouns")

		require.NoError(t, err)
	})

	t.Run("replaces an existing table and drops its backup", func(t *testing.T) {
		mariaDB, mock := newMockDB(t)
		expectTableExists(mock, true)
		expectBackup(mock)
		mock.ExpectExec(createQuery).WillReturnResult(sqlmock.NewResult(0, 0))
		expectCopy(mock)
		mock.ExpectExec(dropBackupQuery).WillReturnResult(sqlmock.NewResult(0, 0))

		err := MigrateTable(newNounsDB(t), mariaDB, "ENLanguageData", "nouns")

		require.NoError(t, err)
	})

	t.Run("succeeds when the backup can't be dropped", func(t *testing.T) {
		mariaDB, mock := newMockDB(t)
		expectTableExists(mock, true)
		expectBackup(mock)
		mock.ExpectExec(createQuery).WillReturnResult(sqlmock.NewResult(0, 0))
		expectCopy(mock)
		mock.ExpectExec(dropBackupQuery).WillReturnError(errors.New("connection lost"))

		err := MigrateTable(newNounsDB(t), mariaDB, "ENLanguageData", "nouns")

		// Note: The new data was migrated, so the leftover backup is only logged.
		require.NoError(t, err)
	})

	t.Run("error when the schema can't be read", func(t *testing.T) {
		// Note: No MariaDB queries are expected, so the test fails if one is sent.
		mariaDB, _ := newMockDB(t)
		sqlite := newNounsDB(t)
		require.NoError(t, sqlite.Close())

		err := MigrateTable(sqlite, mariaDB, "ENLanguageData", "nouns")

		assert.ErrorContains(t, err, "failed to get schema")
	})

	t.Run("error when checking if the table exists fails", func(t *testing.T) {
		mariaDB, mock := newMockDB(t)
		mock.ExpectQuery(tableExistsQuery).WillReturnError(errors.New("connection lost"))

		err := MigrateTable(newNounsDB(t), mariaDB, "ENLanguageData", "nouns")

		assert.EqualError(t, err, "failed to check table existence: connection lost")
	})

	t.Run("error when the existing table can't be backed up", func(t *testing.T) {
		mariaDB, mock := newMockDB(t)
		expectTableExists(mock, true)
		mock.ExpectExec(dropBackupQuery).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(backupQuery).WillReturnError(errors.New("permission denied"))

		err := MigrateTable(newNounsDB(t), mariaDB, "ENLanguageData", "nouns")

		assert.EqualError(t, err, "failed to rename existing table: permission denied")
	})

	t.Run("error when a new table can't be created", func(t *testing.T) {
		mariaDB, mock := newMockDB(t)
		expectTableExists(mock, false)
		mock.ExpectExec(createQuery).WillReturnError(errors.New("permission denied"))

		err := MigrateTable(newNounsDB(t), mariaDB, "ENLanguageData", "nouns")

		assert.EqualError(t, err, "failed to create table: permission denied")
	})

	t.Run("restores the backup when the table can't be created", func(t *testing.T) {
		mariaDB, mock := newMockDB(t)
		expectTableExists(mock, true)
		expectBackup(mock)
		mock.ExpectExec(createQuery).WillReturnError(errors.New("permission denied"))
		mock.ExpectExec(restoreQuery).WillReturnResult(sqlmock.NewResult(0, 0))

		err := MigrateTable(newNounsDB(t), mariaDB, "ENLanguageData", "nouns")

		assert.EqualError(t, err, "failed to create table: permission denied")
	})

	t.Run("error when the backup can't be restored after the table can't be created", func(t *testing.T) {
		mariaDB, mock := newMockDB(t)
		expectTableExists(mock, true)
		expectBackup(mock)
		mock.ExpectExec(createQuery).WillReturnError(errors.New("permission denied"))
		mock.ExpectExec(restoreQuery).WillReturnError(errors.New("connection lost"))

		err := MigrateTable(newNounsDB(t), mariaDB, "ENLanguageData", "nouns")

		assert.EqualError(t, err, "failed to create table and restore backup: "+
			"original error: permission denied, restore error: connection lost")
	})

	t.Run("error when the data of a new table can't be migrated", func(t *testing.T) {
		mariaDB, mock := newMockDB(t)
		expectTableExists(mock, false)
		mock.ExpectExec(createQuery).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectBegin().WillReturnError(errors.New("connection lost"))

		err := MigrateTable(newNounsDB(t), mariaDB, "ENLanguageData", "nouns")

		assert.EqualError(t, err, "failed to migrate data: failed to start transaction: connection lost")
	})

	t.Run("restores the backup when the data can't be migrated", func(t *testing.T) {
		mariaDB, mock := newMockDB(t)
		expectTableExists(mock, true)
		expectBackup(mock)
		mock.ExpectExec(createQuery).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectBegin().WillReturnError(errors.New("connection lost"))
		mock.ExpectExec(dropNewQuery).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(restoreQuery).WillReturnResult(sqlmock.NewResult(0, 0))

		err := MigrateTable(newNounsDB(t), mariaDB, "ENLanguageData", "nouns")

		assert.EqualError(t, err, "failed to migrate data: failed to start transaction: connection lost")
	})

	t.Run("error when the backup can't be restored after the data can't be migrated", func(t *testing.T) {
		mariaDB, mock := newMockDB(t)
		expectTableExists(mock, true)
		expectBackup(mock)
		mock.ExpectExec(createQuery).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectBegin().WillReturnError(errors.New("connection lost"))
		mock.ExpectExec(dropNewQuery).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(restoreQuery).WillReturnError(errors.New("table exists"))

		err := MigrateTable(newNounsDB(t), mariaDB, "ENLanguageData", "nouns")

		assert.EqualError(t, err, "failed to migrate data and restore backup: "+
			"original error: failed to start transaction: connection lost, restore error: table exists")
	})
}

// MARK: Data Migration

func TestPerformDataMigration(t *testing.T) {
	nounsSchema := &types.TableSchema{
		ColumnNames: []string{"wdLexemeId", "noun"},
		ColumnTypes: []string{"TEXT", "TEXT"},
	}

	t.Run("inserts all rows in a transaction", func(t *testing.T) {
		mariaDB, mock := newMockDB(t)
		expectCopy(mock)

		err := performDataMigration(newNounsDB(t), mariaDB, nounsSchema, "nouns", nounsTable)

		require.NoError(t, err)
	})

	t.Run("inserts full batches and the remaining rows", func(t *testing.T) {
		setBatchSize(t, 2)
		sqlite := newSQLiteDB(t,
			"CREATE TABLE nouns (wdLexemeId TEXT, noun TEXT)",
			"INSERT INTO nouns VALUES ('L1', 'Haus'), ('L2', 'Baum'), ('L3', 'Hund')",
		)
		mariaDB, mock := newMockDB(t)
		mock.ExpectBegin()
		insert := mock.ExpectPrepare(insertQuery)
		insert.ExpectExec().WithArgs("L1", "Haus").WillReturnResult(sqlmock.NewResult(0, 1))
		insert.ExpectExec().WithArgs("L2", "Baum").WillReturnResult(sqlmock.NewResult(0, 1))
		insert.ExpectExec().WithArgs("L3", "Hund").WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := performDataMigration(sqlite, mariaDB, nounsSchema, "nouns", nounsTable)

		require.NoError(t, err)
	})

	t.Run("keeps numbers, decimals and nulls", func(t *testing.T) {
		sqlite := newSQLiteDB(t,
			"CREATE TABLE stats (word TEXT, count INTEGER, score REAL)",
			"INSERT INTO stats VALUES ('Haus', 3, 0.5), (NULL, NULL, NULL)",
		)
		statsSchema := &types.TableSchema{
			ColumnNames: []string{"word", "count", "score"},
			ColumnTypes: []string{"TEXT", "BIGINT", "DOUBLE"},
		}
		mariaDB, mock := newMockDB(t)
		mock.ExpectBegin()
		insert := mock.ExpectPrepare(regexp.QuoteMeta("INSERT IGNORE INTO `StatsTable` (`word`, `count`, `score`) VALUES (?,?,?)"))
		insert.ExpectExec().WithArgs("Haus", int64(3), 0.5).WillReturnResult(sqlmock.NewResult(0, 1))
		insert.ExpectExec().WithArgs(nil, nil, nil).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := performDataMigration(sqlite, mariaDB, statsSchema, "stats", "StatsTable")

		require.NoError(t, err)
	})

	t.Run("error when the rows can't be selected", func(t *testing.T) {
		// Note: No MariaDB queries are expected, so the test fails if one is sent.
		mariaDB, _ := newMockDB(t)
		missingColumnSchema := &types.TableSchema{
			ColumnNames: []string{"wdLexemeId", "gender"},
			ColumnTypes: []string{"TEXT", "TEXT"},
		}

		err := performDataMigration(newNounsDB(t), mariaDB, missingColumnSchema, "nouns", nounsTable)

		assert.ErrorContains(t, err, "failed to select data")
	})

	t.Run("error when the insert can't be prepared", func(t *testing.T) {
		mariaDB, mock := newMockDB(t)
		mock.ExpectBegin()
		mock.ExpectPrepare(insertQuery).WillReturnError(errors.New("table doesn't exist"))
		mock.ExpectRollback()

		err := performDataMigration(newNounsDB(t), mariaDB, nounsSchema, "nouns", nounsTable)

		assert.EqualError(t, err, "failed to prepare statement: table doesn't exist")
	})

	t.Run("rolls back when a full batch can't be inserted", func(t *testing.T) {
		setBatchSize(t, 1)
		mariaDB, mock := newMockDB(t)
		mock.ExpectBegin()
		mock.ExpectPrepare(insertQuery).
			ExpectExec().WithArgs("L1", "Haus").WillReturnError(errors.New("data too long"))
		mock.ExpectRollback()

		err := performDataMigration(newNounsDB(t), mariaDB, nounsSchema, "nouns", nounsTable)

		assert.EqualError(t, err, "failed to execute batch: data too long")
	})

	t.Run("rolls back when the remaining rows can't be inserted", func(t *testing.T) {
		mariaDB, mock := newMockDB(t)
		mock.ExpectBegin()
		insert := mock.ExpectPrepare(insertQuery)
		insert.ExpectExec().WithArgs("L1", "Haus").WillReturnResult(sqlmock.NewResult(0, 1))
		insert.ExpectExec().WithArgs("L2", "Baum").WillReturnError(errors.New("data too long"))
		mock.ExpectRollback()

		err := performDataMigration(newNounsDB(t), mariaDB, nounsSchema, "nouns", nounsTable)

		assert.EqualError(t, err, "failed to execute final batch: data too long")
	})

	t.Run("error when the transaction can't be committed", func(t *testing.T) {
		mariaDB, mock := newMockDB(t)
		mock.ExpectBegin()
		insert := mock.ExpectPrepare(insertQuery)
		insert.ExpectExec().WithArgs("L1", "Haus").WillReturnResult(sqlmock.NewResult(0, 1))
		insert.ExpectExec().WithArgs("L2", "Baum").WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit().WillReturnError(errors.New("connection lost"))

		err := performDataMigration(newNounsDB(t), mariaDB, nounsSchema, "nouns", nounsTable)

		assert.EqualError(t, err, "failed to commit transaction: connection lost")
	})

	t.Run("error when reading the rows fails", func(t *testing.T) {
		// Note: The SQLite side is mocked here as a real database can't fail partway through.
		sqlite, sqliteMock := newMockDB(t)
		sqliteMock.ExpectQuery(regexp.QuoteMeta("SELECT `wdLexemeId`, `noun` FROM `nouns`")).
			WillReturnRows(sqlmock.NewRows([]string{"wdLexemeId", "noun"}).
				AddRow("L1", "Haus").
				AddRow("L2", "Baum").
				RowError(1, errors.New("disk I/O error")))
		mariaDB, mock := newMockDB(t)
		mock.ExpectBegin()
		mock.ExpectPrepare(insertQuery).
			ExpectExec().WithArgs("L1", "Haus").WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectRollback()

		err := performDataMigration(sqlite, mariaDB, nounsSchema, "nouns", nounsTable)

		// Partially migrated data shouldn't be committed.
		assert.EqualError(t, err, "error iterating rows: disk I/O error")
	})

	t.Run("rolls back when a row can't be read", func(t *testing.T) {
		// Note: The SQLite side is mocked to return more columns than were selected.
		sqlite, sqliteMock := newMockDB(t)
		sqliteMock.ExpectQuery(regexp.QuoteMeta("SELECT `wdLexemeId`, `noun` FROM `nouns`")).
			WillReturnRows(sqlmock.NewRows([]string{"wdLexemeId", "noun", "gender"}).AddRow("L1", "Haus", "neuter"))
		mariaDB, mock := newMockDB(t)
		mock.ExpectBegin()
		mock.ExpectPrepare(insertQuery)
		mock.ExpectRollback()

		err := performDataMigration(sqlite, mariaDB, nounsSchema, "nouns", nounsTable)

		assert.ErrorContains(t, err, "failed to scan row")
	})
}

// MARK: Batch Execution

func TestExecuteBatch(t *testing.T) {
	t.Run("stops at the first row that fails", func(t *testing.T) {
		db, mock := newMockDB(t)
		insert := mock.ExpectPrepare(insertQuery)
		insert.ExpectExec().WithArgs("L1", "Haus").WillReturnResult(sqlmock.NewResult(0, 1))
		insert.ExpectExec().WithArgs("L2", "Baum").WillReturnError(errors.New("data too long"))
		stmt, err := db.Prepare("INSERT IGNORE INTO `" + nounsTable + "` (`wdLexemeId`, `noun`) VALUES (?,?)")
		require.NoError(t, err)

		// Note: The third row isn't expected, so the test fails if it's inserted.
		err = ExecuteBatch(stmt, [][]interface{}{{"L1", "Haus"}, {"L2", "Baum"}, {"L3", "Hund"}})

		assert.EqualError(t, err, "data too long")
	})

	t.Run("does nothing for an empty batch", func(t *testing.T) {
		db, mock := newMockDB(t)
		mock.ExpectPrepare(insertQuery)
		stmt, err := db.Prepare("INSERT IGNORE INTO `" + nounsTable + "` (`wdLexemeId`, `noun`) VALUES (?,?)")
		require.NoError(t, err)

		require.NoError(t, ExecuteBatch(stmt, nil))
	})
}
