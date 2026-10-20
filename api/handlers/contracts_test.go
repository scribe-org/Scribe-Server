// SPDX-License-Identifier: GPL-3.0-or-later

package handlers

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/scribe-org/scribe-server/api/validators"
	"github.com/scribe-org/scribe-server/internal/constants"
	"github.com/scribe-org/scribe-server/models"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	enContract = "nouns:\n  genders:\n    canonical: [gender]\n"
	deContract = "prepositions:\n  case: grammaticalCase\n"
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

// contractLangs returns the languages of loaded contracts.
func contractLangs(contracts []contractFile) []string {
	langs := make([]string, 0, len(contracts))
	for _, contract := range contracts {
		langs = append(langs, contract.lang)
	}

	return langs
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

// supportLanguages sets the languages that are valid and available in the database for the contracts endpoint.
func supportLanguages(t *testing.T, langs ...string) {
	t.Helper()

	stubLanguageDatabase(
		t,
		func() ([]string, error) { return langs, nil },
		func(_ string) ([]string, error) { return nil, nil },
	)

	// Note: The server initializes the supported languages on startup.
	validators.InitLanguageValidator(langs)
	t.Cleanup(func() { validators.InitLanguageValidator(nil) })
}

// getContracts sends a request to the contracts endpoint.
func getContracts(router *gin.Engine, query string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/contracts"+query, nil))
	return rec
}

// MARK: YAML to JSON

func TestContractToJSON(t *testing.T) {
	t.Run("keeps the order of keys including numbered keys", func(t *testing.T) {
		var contract strings.Builder
		contract.WriteString("verbs:\n  conjugations:\n")
		for i := 1; i <= 11; i++ {
			fmt.Fprintf(&contract, "    %d:\n      sectionTitle: Section %d\n", i, i)
		}
		contract.WriteString("nouns:\n  numbers: {}\n")

		out, err := contractToJSON([]byte(contract.String()))
		require.NoError(t, err)

		// Keys should be in the order of the YAML file, not sorted (e.g. 1, 10, 11, 2 or nouns before verbs).
		order := []string{`"verbs"`}
		for i := 1; i <= 11; i++ {
			order = append(order, fmt.Sprintf(`"Section %d"`, i))
		}
		order = append(order, `"nouns"`)

		last := -1
		for _, key := range order {
			position := bytes.Index(out, []byte(key))
			require.NotEqual(t, -1, position, key)
			assert.Greater(t, position, last, "%s should come after the previous key", key)
			last = position
		}
	})

	t.Run("converts values to their JSON types", func(t *testing.T) {
		contract := `
string: text
quoted: "1"
number: 1
float: 1.5
bool: true
null_value: null
empty: []
list: [a, b]
date: 2026-10-01
html: <b>&</b>
`
		out, err := contractToJSON([]byte(contract))
		require.NoError(t, err)

		var got map[string]any
		require.NoError(t, json.Unmarshal(out, &got))
		assert.Equal(t, map[string]any{
			"string":     "text",
			"quoted":     "1",
			"number":     float64(1),
			"float":      1.5,
			"bool":       true,
			"null_value": nil,
			"empty":      []any{},
			"list":       []any{"a", "b"},
			"date":       "2026-10-01",
			"html":       "<b>&</b>",
		}, got)
		assert.Contains(t, string(out), `"<b>&</b>"`, "HTML characters shouldn't be escaped")
		assert.Contains(t, string(out), "\n  \"string\": \"text\"", "JSON should be indented")
	})

	t.Run("resolves aliases", func(t *testing.T) {
		out, err := contractToJSON([]byte("base: &base [gender]\ncopy: *base\n"))
		require.NoError(t, err)
		assert.JSONEq(t, `{"base": ["gender"], "copy": ["gender"]}`, string(out))
	})

	t.Run("empty and comment only files", func(t *testing.T) {
		for _, contract := range []string{"", "# Only a comment.\n"} {
			out, err := contractToJSON([]byte(contract))
			require.NoError(t, err)
			assert.Equal(t, "null", string(out))
		}
	})

	t.Run("errors", func(t *testing.T) {
		for name, contract := range map[string]string{
			"invalid YAML":               "nouns: [unclosed\n",
			"key that isn't a string":    "? [a, b]\n: value\n",
			"invalid tagged number":      "nouns:\n  count: !!int abc\n",
			"number JSON can't encode":   "nouns:\n  - .nan\n",
			"invalid value in a mapping": "nouns:\n  count: .inf\n",
		} {
			_, err := contractToJSON([]byte(contract))
			assert.Error(t, err, name)
		}
	})
}

// MARK: Contract Loading

func TestLoadContract(t *testing.T) {
	t.Run("loads a yaml contract as JSON", func(t *testing.T) {
		dir := writeContractFiles(t, map[string]string{"en.yaml": enContract})

		contract, err := loadContract(dir, "en")
		require.NoError(t, err)
		assert.Equal(t, "en", contract.lang)
		assert.JSONEq(t, `{"nouns": {"genders": {"canonical": ["gender"]}}}`, string(contract.json))
		assert.False(t, contract.modified.IsZero())
	})

	t.Run("falls back to a yml contract", func(t *testing.T) {
		dir := writeContractFiles(t, map[string]string{"de.yml": deContract})

		contract, err := loadContract(dir, "de")
		require.NoError(t, err)
		assert.JSONEq(t, `{"prepositions": {"case": "grammaticalCase"}}`, string(contract.json))
	})

	t.Run("prefers a yaml contract over a yml contract", func(t *testing.T) {
		dir := writeContractFiles(t, map[string]string{"en.yaml": enContract, "en.yml": deContract})

		contract, err := loadContract(dir, "en")
		require.NoError(t, err)
		assert.Contains(t, string(contract.json), "nouns")
	})

	t.Run("missing contract", func(t *testing.T) {
		_, err := loadContract(t.TempDir(), "en")
		assert.ErrorIs(t, err, fs.ErrNotExist)
	})

	t.Run("contract that can't be read", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.Mkdir(filepath.Join(dir, "en.yaml"), 0o700))

		_, err := loadContract(dir, "en")
		require.Error(t, err)
		assert.NotErrorIs(t, err, fs.ErrNotExist)
		assert.Contains(t, err.Error(), "could not read contract file for en")
	})

	t.Run("contract that can't be converted", func(t *testing.T) {
		dir := writeContractFiles(t, map[string]string{"en.yaml": "nouns: [unclosed\n"})

		_, err := loadContract(dir, "en")
		require.Error(t, err)
		assert.NotErrorIs(t, err, fs.ErrNotExist)
		assert.Contains(t, err.Error(), "could not convert contract for en")
	})
}

func TestLoadAllContracts(t *testing.T) {
	t.Run("loads yaml and yml contracts sorted by language", func(t *testing.T) {
		dir := writeContractFiles(t, map[string]string{
			"en.yaml":   enContract,
			"de.yml":    deContract,
			"README.md": "# Contracts\n",
		})

		contracts, err := loadAllContracts(dir)
		require.NoError(t, err)
		assert.Equal(t, []string{"de", "en"}, contractLangs(contracts))
	})

	t.Run("prefers a yaml contract over a yml contract", func(t *testing.T) {
		dir := writeContractFiles(t, map[string]string{"en.yaml": enContract, "en.yml": deContract})

		contracts, err := loadAllContracts(dir)
		require.NoError(t, err)
		require.Len(t, contracts, 1)
		assert.Contains(t, string(contracts[0].json), "nouns")
	})

	t.Run("skips directories and contracts that can't be read or converted", func(t *testing.T) {
		dir := writeContractFiles(t, map[string]string{
			"en.yaml": enContract,
			"fr.yaml": "nouns: [unclosed\n",
		})
		require.NoError(t, os.Mkdir(filepath.Join(dir, "nested.yaml"), 0o700))

		// Note: A broken symlink can't be read regardless of the user running the tests.
		require.NoError(t, os.Symlink(filepath.Join(dir, "missing.yaml"), filepath.Join(dir, "de.yaml")))

		contracts, err := loadAllContracts(dir)
		require.NoError(t, err)
		assert.Equal(t, []string{"en"}, contractLangs(contracts))
	})

	t.Run("missing directory", func(t *testing.T) {
		_, err := loadAllContracts(filepath.Join(t.TempDir(), "missing"))
		assert.ErrorIs(t, err, fs.ErrNotExist)
	})

	t.Run("directory without contracts", func(t *testing.T) {
		dir := writeContractFiles(t, map[string]string{"README.md": "# Contracts\n"})

		_, err := loadAllContracts(dir)
		assert.ErrorIs(t, err, fs.ErrNotExist)
	})
}

// MARK: Contracts Archive

// failingWriter is a writer that always fails.
type failingWriter struct{}

func (failingWriter) Write(_ []byte) (int, error) { return 0, errors.New("write failed") }

func TestWriteContractsZip(t *testing.T) {
	t.Run("writes a data_contracts directory of JSON contracts", func(t *testing.T) {
		modified := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
		contracts := []contractFile{
			{lang: "de", json: []byte(`{"prepositions": {}}`), modified: modified},
			{lang: "en", json: []byte(`{"nouns": {}}`), modified: modified},
		}

		var archive bytes.Buffer
		require.NoError(t, writeContractsZip(&archive, contracts))

		reader, err := zip.NewReader(bytes.NewReader(archive.Bytes()), int64(archive.Len()))
		require.NoError(t, err)
		require.Len(t, reader.File, 2)

		for i, file := range reader.File {
			assert.Equal(t, "data_contracts/"+contracts[i].lang+".json", file.Name)
			assert.True(t, file.Modified.Equal(modified))

			content, err := file.Open()
			require.NoError(t, err)
			data, err := io.ReadAll(content)
			require.NoError(t, err)
			assert.Equal(t, contracts[i].json, data)
		}
	})

	t.Run("returns errors from writing the archive", func(t *testing.T) {
		// Note: A small contract fails when the archive is closed and a large one when it's written.
		small := []contractFile{{lang: "en", json: []byte("{}")}}
		assert.Error(t, writeContractsZip(failingWriter{}, small))

		// Note: Hashes can't be compressed, so writing them to the archive fails before it's closed.
		var incompressible []byte
		block := sha256.Sum256([]byte("contracts"))
		for len(incompressible) < 1<<20 {
			incompressible = append(incompressible, block[:]...)
			block = sha256.Sum256(block[:])
		}
		large := []contractFile{{lang: "en", json: incompressible}}
		assert.Error(t, writeContractsZip(failingWriter{}, large))
	})
}

// MARK: Endpoint

func TestGetContracts(t *testing.T) {
	t.Run("200 with a single contract as JSON", func(t *testing.T) {
		dir := writeContractFiles(t, map[string]string{"en.yaml": enContract, "de.yaml": deContract})
		router := setupContractsTestRouter(t, dir)
		supportLanguages(t, "de", "en")

		rec := getContracts(router, "?lang=en")

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
		assert.Equal(t, `inline; filename="en.json"`, rec.Header().Get("Content-Disposition"))

		// The response is only the contract, without a wrapper such as contracts: en: ...
		assert.JSONEq(t, `{"nouns": {"genders": {"canonical": ["gender"]}}}`, rec.Body.String())
	})

	t.Run("200 with a zip archive of all contracts", func(t *testing.T) {
		dir := writeContractFiles(t, map[string]string{"en.yaml": enContract, "de.yaml": deContract})
		router := setupContractsTestRouter(t, dir)
		supportLanguages(t, "de", "en")

		rec := getContracts(router, "")

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/zip", rec.Header().Get("Content-Type"))
		assert.Equal(t, `attachment; filename="data_contracts.zip"`, rec.Header().Get("Content-Disposition"))

		reader, err := zip.NewReader(bytes.NewReader(rec.Body.Bytes()), int64(rec.Body.Len()))
		require.NoError(t, err)

		names := []string{}
		for _, file := range reader.File {
			names = append(names, file.Name)

			// Each contract in the archive is identical to the single contract response.
			content, err := file.Open()
			require.NoError(t, err)
			data, err := io.ReadAll(content)
			require.NoError(t, err)

			lang := strings.TrimSuffix(filepath.Base(file.Name), ".json")
			assert.Equal(t, getContracts(router, "?lang="+lang).Body.Bytes(), data)
		}
		assert.Equal(t, []string{"data_contracts/de.json", "data_contracts/en.json"}, names)
	})

	t.Run("400 for an invalid language code", func(t *testing.T) {
		router := setupContractsTestRouter(t, t.TempDir())
		supportLanguages(t, "en")

		for _, lang := range []string{"EN", "eng", "e1"} {
			rec := getContracts(router, "?lang="+lang)
			assert.Equal(t, http.StatusBadRequest, rec.Code, lang)
		}
	})

	t.Run("404 for a language without data", func(t *testing.T) {
		router := setupContractsTestRouter(t, writeContractFiles(t, map[string]string{"fr.yaml": enContract}))
		supportLanguages(t, "en")
		validators.InitLanguageValidator([]string{"en", "fr"})

		rec := getContracts(router, "?lang=fr")

		require.Equal(t, http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Body.String(), "Language 'fr' not supported")
	})

	t.Run("500 when the available languages can't be checked", func(t *testing.T) {
		router := setupContractsTestRouter(t, t.TempDir())
		supportLanguages(t, "en")
		stubLanguageDatabase(
			t,
			func() ([]string, error) { return nil, errors.New("database connection failed") },
			func(_ string) ([]string, error) { return nil, nil },
		)

		rec := getContracts(router, "?lang=en")

		require.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "Failed to check language availability")
	})

	t.Run("500 when a contract can't be converted", func(t *testing.T) {
		router := setupContractsTestRouter(t, writeContractFiles(t, map[string]string{"en.yaml": "nouns: [unclosed\n"}))
		supportLanguages(t, "en")

		rec := getContracts(router, "?lang=en")

		require.Equal(t, http.StatusInternalServerError, rec.Code)
		var response models.ErrorResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
		assert.Equal(t, "Failed to load contracts", response.Error)
	})

	t.Run("500 when the archive can't be created", func(t *testing.T) {
		router := setupContractsTestRouter(t, writeContractFiles(t, map[string]string{"en.yaml": enContract}))

		original := writeContractsZip
		writeContractsZip = func(_ io.Writer, _ []contractFile) error { return errors.New("zip failed") }
		t.Cleanup(func() { writeContractsZip = original })

		rec := getContracts(router, "")

		require.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "Failed to load contracts")
	})

	t.Run("503 when contracts aren't available", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "missing")

		for name, tc := range map[string]struct {
			dir   string
			query string
		}{
			"all contracts with a missing directory": {dir: missing, query: ""},
			"all contracts with an empty directory":  {dir: t.TempDir(), query: ""},
			"a single contract that's missing":       {dir: t.TempDir(), query: "?lang=en"},
		} {
			router := setupContractsTestRouter(t, tc.dir)
			supportLanguages(t, "en")

			rec := getContracts(router, tc.query)

			require.Equal(t, http.StatusServiceUnavailable, rec.Code, name)
			var response models.ErrorResponse
			require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
			assert.Equal(t, constants.ContractsNotAvailableError, response.Error, name)
		}
	})
}
