// SPDX-License-Identifier: GPL-3.0-or-later

package dbqueries

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/scribe-org/scribe-server/database"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testDatabaseName = "scribe_test"

// newMockDB replaces the database connection with sqlmock and sets the database name for a test.
func newMockDB(t *testing.T) sqlmock.Sqlmock {
	t.Helper()

	viper.Reset()
	viper.Set("database.name", testDatabaseName)
	t.Cleanup(viper.Reset)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	original := database.DB
	database.DB = db
	t.Cleanup(func() {
		assert.NoError(t, mock.ExpectationsWereMet())
		database.DB = original
		_ = db.Close()
	})

	return mock
}
