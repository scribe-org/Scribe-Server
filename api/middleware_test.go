// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertCORSHeaders(t *testing.T, headers http.Header) {
	t.Helper()

	assert.Equal(t, "*", headers.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", headers.Get("Access-Control-Allow-Credentials"))
	assert.Equal(
		t,
		"Content-Type, Content-Length, Accept-Encoding, "+
			"X-CSRF-Token, Authorization, accept, origin, "+
			"Cache-Control, X-Requested-With",
		headers.Get("Access-Control-Allow-Headers"),
	)
	assert.Equal(t, "POST, OPTIONS, GET, PUT, DELETE", headers.Get("Access-Control-Allow-Methods"))
}

func setupTestRouter(t *testing.T) *gin.Engine {
	t.Helper()

	previousMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() { gin.SetMode(previousMode) })

	router := gin.New()
	router.Use(SetupCORS())
	return router
}

func TestSetupCORS_NormalGET(t *testing.T) {
	router := setupTestRouter(t)

	handlerReached := false
	router.GET("/test", func(c *gin.Context) {
		handlerReached = true
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, handlerReached, "handler should be reached for normal GET")
	assert.Equal(t, "ok", rec.Body.String())
	assertCORSHeaders(t, rec.Header())
}

func TestSetupCORS_OPTIONS(t *testing.T) {
	router := setupTestRouter(t)

	handlerReached := false
	router.OPTIONS("/test", func(c *gin.Context) {
		handlerReached = true
		c.String(http.StatusOK, "should not be reached")
	})

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	assert.False(t, handlerReached, "handler must not be reached on OPTIONS preflight")
	assert.Empty(t, rec.Body.String())
	assertCORSHeaders(t, rec.Header())
}

func TestSetupCORS_ErrorResponses(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		handler    gin.HandlerFunc
	}{
		{
			name:       "400 Bad Request with AbortWithStatusJSON",
			statusCode: http.StatusBadRequest,
			handler: func(c *gin.Context) {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "bad request"})
			},
		},
		{
			name:       "403 Forbidden with AbortWithStatus",
			statusCode: http.StatusForbidden,
			handler: func(c *gin.Context) {
				c.AbortWithStatus(http.StatusForbidden)
			},
		},
		{
			name:       "404 Not Found from handler",
			statusCode: http.StatusNotFound,
			handler: func(c *gin.Context) {
				c.String(http.StatusNotFound, "not found")
			},
		},
		{
			name:       "500 Internal Server Error with AbortWithStatus",
			statusCode: http.StatusInternalServerError,
			handler: func(c *gin.Context) {
				c.AbortWithStatus(http.StatusInternalServerError)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter(t)
			router.GET("/error", tt.handler)

			req := httptest.NewRequest(http.MethodGet, "/error", nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			require.Equal(t, tt.statusCode, rec.Code)
			assertCORSHeaders(t, rec.Header())
		})
	}
}

func TestSetupCORS_UnroutedRoute(t *testing.T) {
	router := setupTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	assertCORSHeaders(t, rec.Header())
}

func TestSetupCORS_OtherMethods(t *testing.T) {
	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			router := setupTestRouter(t)
			router.Handle(method, "/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(method, "/test", nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			assertCORSHeaders(t, rec.Header())
		})
	}
}
