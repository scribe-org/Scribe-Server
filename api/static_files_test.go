// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: Files outside of the packs directory contain this so that tests can check they're never served.
const outsideContent = "SECRET"

// writeTestFile creates a file with the given content, including its parent directories.
func writeTestFile(t *testing.T, path, content string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
}

// setupPacksDir creates a directory with a packs directory and files outside of it, and returns its path.
//
//	root/
//	├── outside-dir/
//	├── secret.sqlite
//	├── secret.txt
//	└── packs/
//	    └── sqlite/
//	        ├── DELanguageData.sqlite
//	        ├── ENLanguageData.sqlite
//	        ├── Folder.sqlite/
//	        ├── index.html
//	        └── notes.txt
func setupPacksDir(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	sqlite := filepath.Join(root, "packs", "sqlite")

	writeTestFile(t, filepath.Join(sqlite, "DELanguageData.sqlite"), "de-data")
	writeTestFile(t, filepath.Join(sqlite, "ENLanguageData.sqlite"), "en-data")
	writeTestFile(t, filepath.Join(sqlite, "index.html"), "<h1>Scribe-Server packs</h1>")
	writeTestFile(t, filepath.Join(sqlite, "notes.txt"), "notes")
	require.NoError(t, os.Mkdir(filepath.Join(sqlite, "Folder.sqlite"), 0o700))

	require.NoError(t, os.Mkdir(filepath.Join(root, "outside-dir"), 0o700))
	writeTestFile(t, filepath.Join(root, "secret.txt"), outsideContent)
	writeTestFile(t, filepath.Join(root, "secret.sqlite"), outsideContent)

	return root
}

// setupStaticTestRouter creates a Gin engine with static file serving for the given packs directory.
func setupStaticTestRouter(t *testing.T, fileSystem string) *gin.Engine {
	t.Helper()

	previousMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() { gin.SetMode(previousMode) })

	previousFileSystem := viper.GetString("fileSystem")
	viper.Set("fileSystem", fileSystem)
	t.Cleanup(func() { viper.Set("fileSystem", previousFileSystem) })

	router := gin.New()
	setupStaticFiles(router)
	return router
}

// get sends a GET request for a path exactly as given, without cleaning it like most clients do.
func get(router *gin.Engine, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

// MARK: Startup

func TestSetupStaticFilesStartup(t *testing.T) {
	t.Run("fatal error when the packs path can't be resolved", func(t *testing.T) {
		replace(t, &resolveAbsPath, func(string) (string, error) { return "", errors.New("getwd failed") })
		var fatal string
		replace(t, &fatalf, func(format string, v ...any) { fatal = fmt.Sprintf(format, v...) })

		router := setupStaticTestRouter(t, "packs")

		assert.Equal(t, "Error resolving absolute path for packs: getwd failed", fatal)
		assert.Equal(t, http.StatusNotFound, get(router, "/packs/sqlite/list").Code)
	})

	t.Run("no packs routes when the packs directory is missing", func(t *testing.T) {
		router := setupStaticTestRouter(t, filepath.Join(t.TempDir(), "missing"))

		for _, path := range []string{"/packs/", "/packs/sqlite/list", "/packs/sqlite/ENLanguageData.sqlite"} {
			assert.Equal(t, http.StatusNotFound, get(router, path).Code, path)
		}
	})

	t.Run("uses ./packs by default", func(t *testing.T) {
		root := setupPacksDir(t)
		t.Chdir(root)
		router := setupStaticTestRouter(t, "")

		rec := get(router, "/packs/sqlite/ENLanguageData.sqlite")

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "en-data", rec.Body.String())
	})
}

// MARK: SQLite List

func TestSQLiteList(t *testing.T) {
	t.Run("lists only SQLite files", func(t *testing.T) {
		router := setupStaticTestRouter(t, filepath.Join(setupPacksDir(t), "packs"))

		rec := get(router, "/packs/sqlite/list")

		require.Equal(t, http.StatusOK, rec.Code)
		var files []string
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &files))
		assert.Equal(t, []string{"DELanguageData.sqlite", "ENLanguageData.sqlite"}, files)
	})

	t.Run("empty list when there are no SQLite files", func(t *testing.T) {
		packs := filepath.Join(t.TempDir(), "packs")
		require.NoError(t, os.MkdirAll(filepath.Join(packs, "sqlite"), 0o700))
		router := setupStaticTestRouter(t, packs)

		rec := get(router, "/packs/sqlite/list")

		require.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, "[]", rec.Body.String())
	})

	t.Run("500 when the SQLite directory is missing", func(t *testing.T) {
		packs := filepath.Join(t.TempDir(), "packs")
		require.NoError(t, os.Mkdir(packs, 0o700))
		router := setupStaticTestRouter(t, packs)

		rec := get(router, "/packs/sqlite/list")

		require.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.JSONEq(t, `{"error": "Could not read directory"}`, rec.Body.String())
	})
}

// MARK: SQLite Download

func TestSQLiteDownload(t *testing.T) {
	router := setupStaticTestRouter(t, filepath.Join(setupPacksDir(t), "packs"))

	t.Run("downloads a SQLite file", func(t *testing.T) {
		rec := get(router, "/packs/sqlite/ENLanguageData.sqlite")

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "en-data", rec.Body.String())
		assert.Equal(t, `attachment; filename="ENLanguageData.sqlite"`, rec.Header().Get("Content-Disposition"))
		assert.Equal(t, "application/x-sqlite3", rec.Header().Get("Content-Type"))
		assert.Equal(t, "binary", rec.Header().Get("Content-Transfer-Encoding"))
		assert.Equal(t, "File Transfer", rec.Header().Get("Content-Description"))
	})

	t.Run("404 for a missing SQLite file", func(t *testing.T) {
		rec := get(router, "/packs/sqlite/FRLanguageData.sqlite")

		require.Equal(t, http.StatusNotFound, rec.Code)
		assert.JSONEq(t, `{"error": "File not found"}`, rec.Body.String())
	})
}

// MARK: Directory Traversal

func TestDirectoryTraversal(t *testing.T) {
	router := setupStaticTestRouter(t, filepath.Join(setupPacksDir(t), "packs"))

	for _, tc := range []struct {
		name   string
		path   string
		status int
	}{
		// SQLite downloads.
		{"SQLite file outside of packs", "/packs/sqlite/../../secret.sqlite", http.StatusBadRequest},
		{"encoded slash in a SQLite file name", "/packs/sqlite/..%2f..%2fsecret.sqlite", http.StatusBadRequest},
		{"encoded dots and slashes in a SQLite file name", "/packs/sqlite/%2e%2e%2f%2e%2e%2fsecret.sqlite", http.StatusBadRequest},
		{"SQLite file in a subdirectory", "/packs/sqlite/Folder.sqlite%2fother.sqlite", http.StatusBadRequest},
		{"double encoded SQLite file name", "/packs/sqlite/%252e%252e%252f%252e%252e%252fsecret.sqlite", http.StatusNotFound},

		// Other files.
		{"file outside of packs", "/packs/../secret.txt", http.StatusBadRequest},
		{"directory outside of packs", "/packs/../outside-dir", http.StatusBadRequest},
		{"missing directory outside of packs", "/packs/../missing-dir", http.StatusBadRequest},
		{"file outside of packs from a subdirectory", "/packs/sqlite/../../secret.txt", http.StatusBadRequest},
		{"encoded file outside of packs", "/packs/%2e%2e%2fsecret.txt", http.StatusBadRequest},
		{"encoded directory outside of packs", "/packs/..%2foutside-dir", http.StatusBadRequest},
		{"double encoded file outside of packs", "/packs/%252e%252e%252fsecret.txt", http.StatusNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := get(router, tc.path)

			assert.Equal(t, tc.status, rec.Code)
			assert.NotContains(t, rec.Body.String(), outsideContent)

			// Note: Redirecting would show whether a directory outside of packs exists.
			assert.Empty(t, rec.Header().Get("Location"))
		})
	}
}

// MARK: Other Files

func TestOtherPacksFiles(t *testing.T) {
	router := setupStaticTestRouter(t, filepath.Join(setupPacksDir(t), "packs"))

	t.Run("serves a file", func(t *testing.T) {
		rec := get(router, "/packs/sqlite/notes.txt")

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "notes", rec.Body.String())
	})

	t.Run("serves the index of a directory", func(t *testing.T) {
		rec := get(router, "/packs/sqlite/")

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Scribe-Server packs")
	})

	t.Run("lists a directory without an index", func(t *testing.T) {
		rec := get(router, "/packs/")

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `href="sqlite/"`)
	})

	t.Run("redirects a directory to its path with a trailing slash", func(t *testing.T) {
		rec := get(router, "/packs/sqlite")

		require.Equal(t, http.StatusMovedPermanently, rec.Code)
		assert.Equal(t, "/packs/sqlite/", rec.Header().Get("Location"))
	})

	t.Run("404 for a missing file", func(t *testing.T) {
		assert.Equal(t, http.StatusNotFound, get(router, "/packs/sqlite/missing.txt").Code)
	})
}

// MARK: Static Assets

func TestStaticAssets(t *testing.T) {
	t.Run("serves the favicon and logo when they exist", func(t *testing.T) {
		root := setupPacksDir(t)
		writeTestFile(t, filepath.Join(root, "static", "favicon.ico"), "favicon")
		writeTestFile(t, filepath.Join(root, "static", "images", "ScribeServerLogoWhite.png"), "logo")
		t.Chdir(root)
		router := setupStaticTestRouter(t, "packs")

		for path, content := range map[string]string{
			"/favicon.ico":                      "favicon",
			"/images/ScribeServerLogoWhite.png": "logo",
		} {
			rec := get(router, path)
			require.Equal(t, http.StatusOK, rec.Code, path)
			assert.Equal(t, content, rec.Body.String(), path)
		}
	})

	t.Run("no routes for missing assets", func(t *testing.T) {
		root := setupPacksDir(t)
		t.Chdir(root)
		router := setupStaticTestRouter(t, "packs")

		for _, path := range []string{"/favicon.ico", "/images/ScribeServerLogoWhite.png"} {
			assert.Equal(t, http.StatusNotFound, get(router, path).Code, path)
		}
	})
}
