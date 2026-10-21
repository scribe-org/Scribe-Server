// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	// Register the generated Swagger docs as the server does in main.go.
	_ "github.com/scribe-org/scribe-server/docs"
)

func TestSetupSwagger(t *testing.T) {
	previousMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() { gin.SetMode(previousMode) })

	router := gin.New()
	setupSwagger(router)

	t.Run("serves the Swagger UI", func(t *testing.T) {
		for _, path := range []string{"/swagger/index.html", "/docs/index.html"} {
			rec := get(router, path)
			require.Equal(t, http.StatusOK, rec.Code, path)
			assert.Contains(t, rec.Body.String(), "swagger-ui", path)
		}
	})

	t.Run("serves the generated API docs", func(t *testing.T) {
		rec := get(router, "/swagger/doc.json")

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"/api/v1/contracts"`)
	})
}
