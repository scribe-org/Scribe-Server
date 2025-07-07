// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.Use(SetupCORS())
	SetupRoutes(r)
	return r
}

func TestHealthCheck(t *testing.T) {
	router := setupTestRouter()

	t.Run("Root Hello endpoint returns 200", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestCORS(t *testing.T) {
	router := setupTestRouter()

	t.Run("CORS headers present", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("OPTIONS", "/", nil)
		router.ServeHTTP(w, req)

		assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Origin"))
	})
}
