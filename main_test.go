// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupConfigDir changes to a new directory with the given config.yaml (if any) and resets the configuration.
func setupConfigDir(t *testing.T, config string) {
	t.Helper()

	dir := t.TempDir()
	if config != "" {
		require.NoError(t, os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(config), 0o600))
	}
	t.Chdir(dir)

	// Note: Environment variables are cleared so the developer's environment doesn't affect the tests.
	t.Setenv("HOSTPORT", "")
	t.Setenv("FILESYSTEM", "")

	viper.Reset()
	t.Cleanup(viper.Reset)
}

// MARK: Environment Logging

func TestLogEnvironment(t *testing.T) {
	for name, tc := range map[string]struct {
		env, ginMode         string
		wantEnv, wantGinMode string
	}{
		"defaults when unset": {env: "", ginMode: "", wantEnv: "dev", wantGinMode: "debug"},
		"values when set":     {env: "prod", ginMode: "release", wantEnv: "prod", wantGinMode: "release"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Setenv("ENV", tc.env)
			t.Setenv("GIN_MODE", tc.ginMode)

			var logs bytes.Buffer
			original := log.Writer()
			log.SetOutput(&logs)
			t.Cleanup(func() { log.SetOutput(original) })

			logEnvironment()

			assert.Contains(t, logs.String(), "Environment: "+tc.wantEnv)
			assert.Contains(t, logs.String(), "GIN_MODE: "+tc.wantGinMode)
		})
	}
}

// MARK: Config Setup

func TestLoadConfig(t *testing.T) {
	t.Run("reads config.yaml", func(t *testing.T) {
		setupConfigDir(t, "hostPort: 8080\nfileSystem: ./packs\n")

		require.NoError(t, loadConfig())
		assert.Equal(t, "8080", viper.GetString("hostPort"))
		assert.Equal(t, "./contracts", viper.GetString("contractsDir"))
	})

	t.Run("keeps contractsDir from config.yaml", func(t *testing.T) {
		setupConfigDir(t, "hostPort: 8080\ncontractsDir: /data/contracts\n")

		require.NoError(t, loadConfig())
		assert.Equal(t, "/data/contracts", viper.GetString("contractsDir"))
	})

	t.Run("falls back to environment variables", func(t *testing.T) {
		setupConfigDir(t, "")
		t.Setenv("HOSTPORT", "8000")
		t.Setenv("FILESYSTEM", "./packs")

		require.NoError(t, loadConfig())
		assert.Equal(t, "8000", viper.GetString("hostPort"))
		assert.Equal(t, "./contracts", viper.GetString("contractsDir"))
	})

	t.Run("error without config.yaml or environment variables", func(t *testing.T) {
		setupConfigDir(t, "")
		t.Setenv("HOSTPORT", "8000")

		err := loadConfig()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "fatal error config environment")
	})

	t.Run("error for an invalid config.yaml", func(t *testing.T) {
		setupConfigDir(t, "hostPort: [unclosed\n")

		err := loadConfig()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "fatal error config file")
	})
}

// MARK: Main Entry Point

func TestMainEntryPoint(t *testing.T) {
	// startedServer replaces starting the server and returns how many times it was started.
	startedServer := func(t *testing.T) *int {
		t.Helper()

		started := 0
		original := handleRequests
		handleRequests = func() { started++ }
		t.Cleanup(func() { handleRequests = original })

		return &started
	}

	t.Run("starts the server", func(t *testing.T) {
		setupConfigDir(t, "hostPort: 8080\nfileSystem: ./packs\n")
		started := startedServer(t)

		main()

		assert.Equal(t, 1, *started)
	})

	t.Run("panics without configuration", func(t *testing.T) {
		setupConfigDir(t, "")
		started := startedServer(t)

		defer func() {
			recovered := recover()
			require.NotNil(t, recovered, "main should panic")
			assert.Contains(t, fmt.Sprint(recovered), "fatal error config environment")
			assert.Zero(t, *started, "the server shouldn't be started")
		}()

		main()
	})
}
