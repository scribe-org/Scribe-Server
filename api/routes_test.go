// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupRoutes(t *testing.T) {
	previousMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() { gin.SetMode(previousMode) })

	router := gin.New()
	SetupRoutes(router)

	t.Run("registers the home page and API routes", func(t *testing.T) {
		routes := map[string]bool{}
		for _, route := range router.Routes() {
			routes[route.Method+" "+route.Path] = true
		}

		assert.Equal(t, map[string]bool{
			"GET /":                          true,
			"GET /api/v1/data/:lang":         true,
			"GET /api/v1/data-version/:lang": true,
			"GET /api/v1/languages":          true,
			"GET /api/v1/contracts":          true,
			"GET /api/v1/language-stats":     true,
			"GET /api/v1/translations":       true,
		}, routes)
	})

	t.Run("404 for unknown paths", func(t *testing.T) {
		for _, path := range []string{"/api/v1/unknown", "/api/v2/languages", "/api"} {
			assert.Equal(t, http.StatusNotFound, get(router, path).Code, path)
		}
	})

	t.Run("404 for unsupported methods", func(t *testing.T) {
		// Note: Gin returns a 404 rather than a 405 as HandleMethodNotAllowed isn't enabled.
		for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(method, "/api/v1/languages", nil))
			assert.Equal(t, http.StatusNotFound, rec.Code, method)
		}
	})

	t.Run("redirects paths with a trailing slash", func(t *testing.T) {
		rec := get(router, "/api/v1/languages/")

		require.Equal(t, http.StatusMovedPermanently, rec.Code)
		assert.Equal(t, "/api/v1/languages", rec.Header().Get("Location"))
	})

	t.Run("serves the home page", func(t *testing.T) {
		root := t.TempDir()
		writeTestFile(t, filepath.Join(root, "static", "index.html"), "<h1>Scribe-Server</h1>")
		t.Chdir(root)

		rec := get(router, "/")

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "text/html; charset=utf-8", rec.Header().Get("Content-Type"))
		assert.Equal(t, "<h1>Scribe-Server</h1>", rec.Body.String())
	})
}
