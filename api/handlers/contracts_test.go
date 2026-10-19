// SPDX-License-Identifier: GPL-3.0-or-later

package handlers

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/scribe-org/scribe-server/api/validators"
	"github.com/scribe-org/scribe-server/internal/constants"
	"github.com/scribe-org/scribe-server/models"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// writeContractFiles creates a contracts directory with the given files and returns its path.
func writeContractFiles(t *testing.T, files map[string]string) string {
	t.Helper()

	dir := t.TempDir()
	for name, content := range files {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
	}

	return dir
}

// setupContractsTestRouter creates a Gin engine for the contracts endpoint using the given contracts directory.
func setupContractsTestRouter(t *testing.T, contractsDir string) *gin.Engine {
	t.Helper()

	previousMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() { gin.SetMode(previousMode) })

	previousDir := viper.GetString("contractsDir")
	viper.Set("contractsDir", contractsDir)
	t.Cleanup(func() { viper.Set("contractsDir", previousDir) })

	router := gin.New()
	router.GET("/api/v1/contracts", GetContracts)
	return router
}

func TestLoadAllContracts(t *testing.T) {
	t.Run("loads yaml and yml files and skips other files", func(t *testing.T) {
		dir := writeContractFiles(t, map[string]string{
			"en.yaml":   "nouns:\n  numbers:\n    1:\n      singular: singular\n",
			"de.yml":    "prepositions:\n  case: grammaticalCase\n",
			"README.md": "# Contracts\n",
		})

		contracts, err := loadAllContracts(dir)
		require.NoError(t, err)
		assert.Len(t, contracts, 2)
		assert.Contains(t, contracts, "en")
		assert.Contains(t, contracts, "de")
	})

	t.Run("missing directory", func(t *testing.T) {
		_, err := loadAllContracts(filepath.Join(t.TempDir(), "missing"))
		require.Error(t, err)
		assert.ErrorIs(t, err, fs.ErrNotExist)
	})

	t.Run("directory without contracts", func(t *testing.T) {
		dir := writeContractFiles(t, map[string]string{"README.md": "# Contracts\n"})

		_, err := loadAllContracts(dir)
		require.Error(t, err)
		assert.ErrorIs(t, err, fs.ErrNotExist)
	})
}

func TestLoadSingleContract(t *testing.T) {
	t.Run("loads a yaml contract", func(t *testing.T) {
		dir := writeContractFiles(t, map[string]string{"en.yaml": "nouns:\n  genders:\n    canonical: [gender]\n"})

		contracts, err := loadSingleContract(dir, "en")
		require.NoError(t, err)
		assert.Contains(t, contracts, "en")
	})

	t.Run("falls back to a yml contract", func(t *testing.T) {
		dir := writeContractFiles(t, map[string]string{"de.yml": "nouns:\n  genders:\n    canonical: [gender]\n"})

		contracts, err := loadSingleContract(dir, "de")
		require.NoError(t, err)
		assert.Contains(t, contracts, "de")
	})

	t.Run("missing contract file", func(t *testing.T) {
		_, err := loadSingleContract(t.TempDir(), "en")
		require.Error(t, err)
		assert.ErrorIs(t, err, fs.ErrNotExist)
	})
}

func TestGetContracts(t *testing.T) {
	t.Run("200 with all contracts", func(t *testing.T) {
		dir := writeContractFiles(t, map[string]string{
			"en.yaml": "nouns:\n  genders:\n    canonical: [gender]\n",
			"de.yaml": "prepositions:\n  case: grammaticalCase\n",
		})
		router := setupContractsTestRouter(t, dir)

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/contracts", nil))

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/x-yaml; charset=utf-8", rec.Header().Get("Content-Type"))
		assert.Equal(t, `attachment; filename="contracts.yaml"`, rec.Header().Get("Content-Disposition"))

		var response struct {
			Contracts map[string]any `yaml:"contracts"`
		}
		require.NoError(t, yaml.Unmarshal(rec.Body.Bytes(), &response))
		assert.Contains(t, response.Contracts, "en")
		assert.Contains(t, response.Contracts, "de")
	})

	t.Run("200 with a single contract named by language", func(t *testing.T) {
		dir := writeContractFiles(t, map[string]string{
			"en.yaml": "nouns:\n  genders:\n    canonical: [gender]\n",
			"de.yaml": "prepositions:\n  case: grammaticalCase\n",
		})
		router := setupContractsTestRouter(t, dir)
		stubLanguageDatabase(
			t,
			func() ([]string, error) { return []string{"de", "en"}, nil },
			func(_ string) ([]string, error) { return nil, nil },
		)

		// Note: The server initializes the supported languages on startup.
		validators.InitLanguageValidator([]string{"de", "en"})
		t.Cleanup(func() { validators.InitLanguageValidator(nil) })

		for _, lang := range []string{"en", "de"} {
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/contracts?lang="+lang, nil))

			require.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, `attachment; filename="contracts_`+lang+`.yaml"`, rec.Header().Get("Content-Disposition"))

			var response struct {
				Contracts map[string]any `yaml:"contracts"`
			}
			require.NoError(t, yaml.Unmarshal(rec.Body.Bytes(), &response))
			assert.Len(t, response.Contracts, 1)
			assert.Contains(t, response.Contracts, lang)
		}
	})

	t.Run("503 when the contracts directory is missing", func(t *testing.T) {
		router := setupContractsTestRouter(t, filepath.Join(t.TempDir(), "missing"))

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/contracts", nil))

		require.Equal(t, http.StatusServiceUnavailable, rec.Code)

		var response models.ErrorResponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
		assert.Equal(t, constants.ContractsNotAvailableError, response.Error)
	})

	t.Run("503 when the contracts directory is empty", func(t *testing.T) {
		router := setupContractsTestRouter(t, t.TempDir())

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/contracts", nil))

		require.Equal(t, http.StatusServiceUnavailable, rec.Code)

		var response models.ErrorResponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
		assert.Equal(t, constants.ContractsNotAvailableError, response.Error)
	})

	t.Run("503 when no contract can be parsed", func(t *testing.T) {
		dir := writeContractFiles(t, map[string]string{"en.yaml": "nouns: [unclosed\n"})
		router := setupContractsTestRouter(t, dir)

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/contracts", nil))

		// Note: Unparsable files are skipped, so a directory with only unparsable files has no contracts.
		require.Equal(t, http.StatusServiceUnavailable, rec.Code)
	})
}
