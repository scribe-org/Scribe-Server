// SPDX-License-Identifier: GPL-3.0-or-later

package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServeHome(t *testing.T) {
	previousMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() { gin.SetMode(previousMode) })

	router := gin.New()
	router.GET("/", ServeHome)

	t.Run("serves static/index.html as HTML", func(t *testing.T) {
		root := t.TempDir()
		require.NoError(t, os.Mkdir(filepath.Join(root, "static"), 0o700))
		require.NoError(t, os.WriteFile(filepath.Join(root, "static", "index.html"), []byte("<h1>Scribe-Server</h1>"), 0o600))
		t.Chdir(root)

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "text/html; charset=utf-8", rec.Header().Get("Content-Type"))
		assert.Equal(t, "<h1>Scribe-Server</h1>", rec.Body.String())
	})

	t.Run("404 when static/index.html is missing", func(t *testing.T) {
		t.Chdir(t.TempDir())

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}
