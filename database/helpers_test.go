// SPDX-License-Identifier: GPL-3.0-or-later

package database

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testDatabaseName is the database name that queries are expected to use in tests.
const testDatabaseName = "scribe_test"

// newMockDB replaces the database connection with sqlmock and sets the database name for a test.
// Tests using it must not call t.Parallel() as it changes the DB and Viper globals.
func newMockDB(t *testing.T) sqlmock.Sqlmock {
	t.Helper()

	viper.Reset()
	viper.Set("database.name", testDatabaseName)
	t.Cleanup(viper.Reset)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	original := DB
	DB = db
	t.Cleanup(func() {
		// Every expected query must have been sent.
		assert.NoError(t, mock.ExpectationsWereMet())

		DB = original
		// Note: Closing isn't an expected call for sqlmock, so its error isn't checked.
		_ = db.Close()
	})

	return mock
}
