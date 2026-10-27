// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/scribe-org/scribe-server/cmd/migrate/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// migrationRun records what main did during a test.
type migrationRun struct {
	setupConfig types.DatabaseConfig
	migratedDB  *sql.DB
	fatal       string
}

// replace swaps a function variable for a test and restores it afterwards.
func replace[T any](t *testing.T, target *T, value T) {
	t.Helper()

	original := *target
	*target = value
	t.Cleanup(func() { *target = original })
}

// stubMigration replaces the config, databases and fatal errors of main with the given results.
func stubMigration(t *testing.T, configErr, setupErr, migrateErr error) (*migrationRun, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, mock.ExpectationsWereMet()) })

	run := &migrationRun{}
	replace(t, &loadConfig, func(path string) (*types.Config, error) {
		assert.Equal(t, "config.yaml", path)
		if configErr != nil {
			return nil, configErr
		}
		return &types.Config{Database: types.DatabaseConfig{Name: "scribe_server"}}, nil
	})
	replace(t, &setupMariaDB, func(config types.DatabaseConfig) (*sql.DB, error) {
		run.setupConfig = config
		if setupErr != nil {
			return nil, setupErr
		}
		return db, nil
	})
	replace(t, &processSQLiteFiles, func(mariaDB *sql.DB) error {
		run.migratedDB = mariaDB
		return migrateErr
	})
	replace(t, &fatal, func(v ...any) { run.fatal = fmt.Sprint(v...) })

	return run, mock
}

func TestMainEntryPoint(t *testing.T) {
	t.Run("migrates the SQLite files into MariaDB and closes the connection", func(t *testing.T) {
		run, mock := stubMigration(t, nil, nil, nil)
		mock.ExpectClose()

		main()

		assert.Equal(t, "scribe_server", run.setupConfig.Name)
		assert.NotNil(t, run.migratedDB)
		assert.Empty(t, run.fatal)
	})

	t.Run("fatal when the config can't be loaded", func(t *testing.T) {
		run, _ := stubMigration(t, errors.New("failed to read config file"), nil, nil)

		main()

		assert.Equal(t, "failed to read config file", run.fatal)
		assert.Empty(t, run.setupConfig.Name, "MariaDB shouldn't be set up")
	})

	t.Run("fatal when MariaDB can't be set up", func(t *testing.T) {
		run, _ := stubMigration(t, nil, errors.New("failed to connect"), nil)

		main()

		assert.Equal(t, "failed to connect", run.fatal)
		assert.Nil(t, run.migratedDB, "no files should be migrated")
	})

	t.Run("fatal when the migration fails", func(t *testing.T) {
		run, mock := stubMigration(t, nil, nil, errors.New("error processing packs/sqlite/ENLanguageData.sqlite"))
		mock.ExpectClose()

		main()

		assert.Equal(t, "error processing packs/sqlite/ENLanguageData.sqlite", run.fatal)
	})
}
