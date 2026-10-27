// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/scribe-org/scribe-server/cmd/migrate/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeConfig writes a config file with the given contents for a test and returns its path.
func writeConfig(t *testing.T, contents string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o644))

	return path
}

func TestLoadConfig(t *testing.T) {
	t.Run("reads the server and database config", func(t *testing.T) {
		path := writeConfig(t, `hostPort: 8080
fileSystem: ./packs
database:
  user: scribe
  password: password
  host: tools-db
  port: "3306"
  name: scribe_server
`)

		config, err := LoadConfig(path)

		require.NoError(t, err)
		assert.Equal(t, &types.Config{
			HostPort:   8080,
			FileSystem: "./packs",
			Database: types.DatabaseConfig{
				User:     "scribe",
				Password: "password",
				Host:     "tools-db",
				Port:     "3306",
				Name:     "scribe_server",
			},
		}, config)
	})

	t.Run("error when the file doesn't exist", func(t *testing.T) {
		config, err := LoadConfig(filepath.Join(t.TempDir(), "config.yaml"))

		assert.ErrorContains(t, err, "failed to read config file")
		assert.Nil(t, config)
	})

	t.Run("error for invalid YAML", func(t *testing.T) {
		path := writeConfig(t, "database: [user: scribe")

		config, err := LoadConfig(path)

		assert.ErrorContains(t, err, "failed to read config file")
		assert.Nil(t, config)
	})

	t.Run("error when a value has the wrong type", func(t *testing.T) {
		path := writeConfig(t, "hostPort: not a port\n")

		config, err := LoadConfig(path)

		assert.ErrorContains(t, err, "failed to unmarshal config")
		assert.Nil(t, config)
	})
}
