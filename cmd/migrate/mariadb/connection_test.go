// SPDX-License-Identifier: GPL-3.0-or-later

package mariadb

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/scribe-org/scribe-server/cmd/migrate/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupMariaDB(t *testing.T) {
	t.Run("returns a MySQL connection without connecting", func(t *testing.T) {
		// Note: Nothing is listening on port 1, so the test fails if it tries to connect.
		db, err := SetupMariaDB(types.DatabaseConfig{
			User:     "scribe",
			Password: "password",
			Host:     "127.0.0.1",
			Port:     "1",
			Name:     "scribe_server",
		})

		require.NoError(t, err)
		t.Cleanup(func() { _ = db.Close() })
		assert.IsType(t, &mysql.MySQLDriver{}, db.Driver())
	})

	t.Run("error when the connection can't be opened", func(t *testing.T) {
		original := openDatabase
		openDatabase = func(string, string) (*sql.DB, error) { return nil, errors.New("unknown driver") }
		t.Cleanup(func() { openDatabase = original })

		db, err := SetupMariaDB(types.DatabaseConfig{Name: "scribe_server"})

		assert.EqualError(t, err, "failed to connect to scribe_server database: unknown driver")
		assert.Nil(t, db)
	})
}
