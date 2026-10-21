// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/scribe-org/scribe-server/api/validators"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// replace sets a package-level variable for a test and restores it afterwards.
func replace[T any](t *testing.T, target *T, value T) {
	t.Helper()

	original := *target
	*target = value
	t.Cleanup(func() { *target = original })
}

// setViper sets a configuration value for a test and restores it afterwards.
func setViper(t *testing.T, key string, value any) {
	t.Helper()

	original := viper.Get(key)
	viper.Set(key, value)
	t.Cleanup(func() { viper.Set(key, original) })
}

// captureLog collects log output for a test.
func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()

	var buf bytes.Buffer
	original := log.Writer()
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(original) })

	return &buf
}

// serverRun records how the server was run.
type serverRun struct {
	router *gin.Engine
	addr   string
	fatal  string
}

// stubServer replaces the database, server and fatal errors so that HandleRequests can run in tests.
func stubServer(t *testing.T, languages []string, languagesErr, runErr error) *serverRun {
	t.Helper()

	previousMode := gin.Mode()
	t.Cleanup(func() { gin.SetMode(previousMode) })
	t.Cleanup(func() { validators.InitLanguageValidator(nil) })

	setViper(t, "GIN_MODE", "test")
	setViper(t, "hostPort", "8099")
	setViper(t, "fileSystem", filepath.Join(t.TempDir(), "missing"))

	run := &serverRun{}
	replace(t, &initDatabase, func() error { return nil })
	replace(t, &getAvailableLanguages, func() ([]string, error) { return languages, languagesErr })
	replace(t, &runServer, func(r *gin.Engine, addr string) error {
		run.router, run.addr = r, addr
		return runErr
	})
	replace(t, &fatalf, func(format string, v ...any) { run.fatal = fmt.Sprintf(format, v...) })

	return run
}

// MARK: Handle Requests

func TestHandleRequests(t *testing.T) {
	t.Run("starts the server with the router", func(t *testing.T) {
		run := stubServer(t, []string{"de", "en"}, nil, nil)

		HandleRequests()

		assert.Empty(t, run.fatal)
		assert.Equal(t, ":8099", run.addr)
		require.NotNil(t, run.router)
		assert.Equal(t, gin.TestMode, gin.Mode())

		routes := map[string]bool{}
		for _, route := range run.router.Routes() {
			routes[route.Path] = true
		}
		assert.True(t, routes["/api/v1/contracts"])
		assert.True(t, routes["/swagger/*any"])

		// The available languages are used to validate language codes.
		assert.True(t, validators.IsValidLanguageCode("de"))
		assert.False(t, validators.IsValidLanguageCode("fr"))
	})

	t.Run("logs unknown languages when they can't be fetched", func(t *testing.T) {
		stubServer(t, nil, errors.New("database connection failed"), nil)
		logs := captureLog(t)

		HandleRequests()

		assert.Contains(t, logs.String(), "Could not fetch available languages: database connection failed")
		assert.Contains(t, logs.String(), "Available languages: [unknown]")
	})

	t.Run("fatal error when the database can't be initialized", func(t *testing.T) {
		run := stubServer(t, nil, nil, nil)
		replace(t, &initDatabase, func() error { return errors.New("connection refused") })

		HandleRequests()

		assert.Equal(t, "Failed to initialize database: connection refused", run.fatal)
		assert.Nil(t, run.router, "the server shouldn't be run")
	})

	t.Run("fatal error when the router can't be set up", func(t *testing.T) {
		run := stubServer(t, nil, nil, nil)
		t.Setenv("ENV", "prod")
		replace(t, &prodTrustedProxies, []string{"not-a-network"})

		HandleRequests()

		assert.Contains(t, run.fatal, "Failed to set up the router")
		assert.Nil(t, run.router, "the server shouldn't be run")
	})

	t.Run("fatal error when the server stops", func(t *testing.T) {
		run := stubServer(t, []string{"en"}, nil, errors.New("address already in use"))

		HandleRequests()

		assert.Equal(t, "Failed to run the server: address already in use", run.fatal)
	})
}

// MARK: Router

func TestNewRouterGinMode(t *testing.T) {
	previousMode := gin.Mode()
	t.Cleanup(func() { gin.SetMode(previousMode) })
	setViper(t, "fileSystem", filepath.Join(t.TempDir(), "missing"))

	for ginMode, want := range map[string]string{
		"release": gin.ReleaseMode,
		"test":    gin.TestMode,
		"":        gin.DebugMode,
		"other":   gin.DebugMode,
	} {
		setViper(t, "GIN_MODE", ginMode)

		_, err := newRouter()
		require.NoError(t, err)
		assert.Equal(t, want, gin.Mode(), ginMode)
	}
}

func TestTrustedProxies(t *testing.T) {
	previousMode := gin.Mode()
	t.Cleanup(func() { gin.SetMode(previousMode) })
	setViper(t, "GIN_MODE", "test")
	setViper(t, "fileSystem", filepath.Join(t.TempDir(), "missing"))

	// clientIP returns the client IP that the router sees for a request forwarded by a private network proxy.
	clientIP := func(t *testing.T) string {
		t.Helper()

		router, err := newRouter()
		require.NoError(t, err)
		router.GET("/ip", func(c *gin.Context) { c.String(http.StatusOK, c.ClientIP()) })

		req := httptest.NewRequest(http.MethodGet, "/ip", nil)
		req.RemoteAddr = "10.1.2.3:1234"
		req.Header.Set("X-Forwarded-For", "203.0.113.7")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec.Body.String()
	}

	t.Run("trusts private network proxies in production", func(t *testing.T) {
		t.Setenv("ENV", "prod")
		assert.Equal(t, prodTrustedProxies, trustedProxies())
		assert.Equal(t, "203.0.113.7", clientIP(t))
	})

	t.Run("trusts no proxies in development", func(t *testing.T) {
		t.Setenv("ENV", "dev")
		assert.Empty(t, trustedProxies())
		assert.Equal(t, "10.1.2.3", clientIP(t))
	})
}
