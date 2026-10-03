// SPDX-License-Identifier: GPL-3.0-or-later

package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/scribe-org/scribe-server/internal/constants"
	"github.com/scribe-org/scribe-server/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRouter creates a minimal Gin engine configured in test mode for testing the endpoint.
func setupTestRouter(t *testing.T) *gin.Engine {
	t.Helper()

	previousMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() { gin.SetMode(previousMode) })

	router := gin.New()
	router.GET("/api/v1/languages", GetAvailableLanguages)
	return router
}

// stubLanguageDatabase temporarily swaps package-level database function pointers for isolated unit testing,
func stubLanguageDatabase(
	t *testing.T,
	languagesFn func() ([]string, error),
	dataTypesFn func(lang string) ([]string, error),
) {
	t.Helper()

	origLanguages := getAvailableLanguages
	origDataTypes := getLanguageDataTypes

	getAvailableLanguages = languagesFn
	getLanguageDataTypes = dataTypesFn

	t.Cleanup(func() {
		getAvailableLanguages = origLanguages
		getLanguageDataTypes = origDataTypes
	})
}

// TestGetAvailableLanguages tests the GET /api/v1/languages handler under various database scenarios.
func TestGetAvailableLanguages(t *testing.T) {
	router := setupTestRouter(t)

	t.Run("200 with a populated array", func(t *testing.T) {
		stubLanguageDatabase(
			t,
			func() ([]string, error) {
				return []string{"en", "de"}, nil
			},
			func(lang string) ([]string, error) {
				switch lang {
				case "en":
					return []string{"nouns", "verbs"}, nil
				case "de":
					return []string{"nouns", "verbs", "adjectives"}, nil
				default:
					return nil, fmt.Errorf("unexpected language: %s", lang)
				}
			},
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/languages", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

		var response models.AvailableLanguagesResponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))

		expected := models.AvailableLanguagesResponse{
			Languages: []models.LanguageInfo{
				{
					Code:      "en",
					DataTypes: []string{"nouns", "verbs"},
				},
				{
					Code:      "de",
					DataTypes: []string{"nouns", "verbs", "adjectives"},
				},
			},
		}
		assert.Equal(t, expected, response)
	})

	t.Run("500 when GetAvailableLanguages errors", func(t *testing.T) {
		stubLanguageDatabase(
			t,
			func() ([]string, error) {
				return nil, errors.New("database connection failed")
			},
			func(lang string) ([]string, error) {
				return nil, nil
			},
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/languages", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

		var response models.ErrorResponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
		assert.Equal(t, constants.ErrorFetchingLanguages, response.Error)
	})

	t.Run("partial-failure when GetLanguageDataTypes fails for a language", func(t *testing.T) {
		stubLanguageDatabase(
			t,
			func() ([]string, error) {
				return []string{"en", "fr"}, nil
			},
			func(lang string) ([]string, error) {
				if lang == "en" {
					return []string{"nouns", "verbs"}, nil
				}
				return nil, errors.New("data types retrieval failed for fr")
			},
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/languages", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

		var response models.AvailableLanguagesResponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))

		expected := models.AvailableLanguagesResponse{
			Languages: []models.LanguageInfo{
				{
					Code:      "en",
					DataTypes: []string{"nouns", "verbs"},
				},
			},
		}
		assert.Equal(t, expected, response)
	})
}
