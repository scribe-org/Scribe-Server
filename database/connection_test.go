// SPDX-License-Identifier: GPL-3.0-or-later

package database

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// openedDatabase records how InitDatabase opened the database.
type openedDatabase struct {
	driver string
	dsn    string
}

// setupInitDatabase sets the database config and replaces opening the database with the given connection or error.
// Tests using it must not call t.Parallel() as it changes the DB, openDatabase and Viper globals.
func setupInitDatabase(t *testing.T, db *sql.DB, openErr error) *openedDatabase {
	t.Helper()

	viper.Reset()
	t.Cleanup(viper.Reset)
	viper.Set("database.user", "scribe")
	viper.Set("database.password", "secret")
	viper.Set("database.host", "localhost")
	viper.Set("database.port", "3306")
	viper.Set("database.name", testDatabaseName)

	originalDB, originalOpen := DB, openDatabase
	t.Cleanup(func() { DB, openDatabase = originalDB, originalOpen })

	opened := &openedDatabase{}
	openDatabase = func(driver, dsn string) (*sql.DB, error) {
		opened.driver, opened.dsn = driver, dsn
		return db, openErr
	}

	return opened
}

// newPingMockDB returns a sqlmock connection that checks pings.
func newPingMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, mock.ExpectationsWereMet())
		_ = db.Close()
	})

	return db, mock
}

func TestInitDatabase(t *testing.T) {
	t.Run("connects with the configured database", func(t *testing.T) {
		db, mock := newPingMockDB(t)
		mock.ExpectPing()
		opened := setupInitDatabase(t, db, nil)

		require.NoError(t, InitDatabase())

		assert.Equal(t, "mysql", opened.driver)
		assert.Equal(t, "scribe:secret@tcp(localhost:3306)/scribe_test?charset=utf8mb4&parseTime=True&loc=Local", opened.dsn)
		assert.Same(t, db, DB)
	})

	t.Run("error when the database can't be opened", func(t *testing.T) {
		setupInitDatabase(t, nil, errors.New("unknown driver"))

		err := InitDatabase()

		assert.ErrorContains(t, err, "error opening database: unknown driver")
	})

	t.Run("error when the database can't be reached", func(t *testing.T) {
		db, mock := newPingMockDB(t)
		mock.ExpectPing().WillReturnError(errors.New("connection refused"))
		setupInitDatabase(t, db, nil)

		err := InitDatabase()

		assert.ErrorContains(t, err, "error pinging database: connection refused")
	})
}
