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

// stubLanguageDatabase temporarily swaps package-level database function pointers for isolated unit testing.
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

// MARK: Translation Tests

type translationTableDataFunc func(targetLang, sourceLang string) (map[string]map[string]map[string]models.TranslationEntry, error)

func setupTranslationTestRouter(t *testing.T) *gin.Engine {
	t.Helper()

	previousMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() { gin.SetMode(previousMode) })

	router := gin.New()
	router.GET("/api/v1/translations", GetTranslationData)
	return router
}

func setTranslationTableDataFunc(t *testing.T, fn translationTableDataFunc) {
	t.Helper()

	original := getTranslationTableData
	getTranslationTableData = fn
	t.Cleanup(func() {
		getTranslationTableData = original
	})
}

// TestGetTranslationData_SuccessResponseShape tests the successful retrieval of translations
func TestGetTranslationData_SuccessResponseShape(t *testing.T) {
	router := setupTranslationTestRouter(t)

	t.Run("nested response structure with multiple words, types, and orders", func(t *testing.T) {
		mockData := map[string]map[string]map[string]models.TranslationEntry{
			"buch": {
				"noun": {
					"1": {
						Description: "ein gebundenes Schriftwerk",
						Translation: "book",
					},
					"2": {
						Description: "ein Notizheft",
						Translation: "notebook",
					},
				},
			},
			"laufen": {
				"verb": {
					"1": {
						Description: "sich schnell zu Fuß fortbewegen",
						Translation: "run",
					},
				},
			},
		}

		var capturedTarget, capturedSource string
		setTranslationTableDataFunc(t, func(targetLang, sourceLang string) (map[string]map[string]map[string]models.TranslationEntry, error) {
			capturedTarget = targetLang
			capturedSource = sourceLang
			return mockData, nil
		})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/translations?source_lang=de&target_lang=en", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

		var resp models.TranslationDataResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, "en", capturedTarget)
		assert.Equal(t, "de", capturedSource)
		assert.Equal(t, "en", resp.TargetLang)
		assert.Equal(t, "de", resp.SourceLang)
		assert.Equal(t, mockData, resp.Data)

		// Verification of every level of the nested structure:
		// Level 1: word
		require.Contains(t, resp.Data, "buch")
		require.Contains(t, resp.Data, "laufen")

		// Level 2: wordType
		buchWordTypes := resp.Data["buch"]
		require.Contains(t, buchWordTypes, "noun")
		laufenWordTypes := resp.Data["laufen"]
		require.Contains(t, laufenWordTypes, "verb")

		// Level 3: wordOrder
		buchNounOrders := buchWordTypes["noun"]
		require.Contains(t, buchNounOrders, "1")
		require.Contains(t, buchNounOrders, "2")
		laufenVerbOrders := laufenWordTypes["verb"]
		require.Contains(t, laufenVerbOrders, "1")

		// Leaf: TranslationEntry {Description, Translation}
		assert.Equal(t, "ein gebundenes Schriftwerk", buchNounOrders["1"].Description)
		assert.Equal(t, "book", buchNounOrders["1"].Translation)
		assert.Equal(t, "ein Notizheft", buchNounOrders["2"].Description)
		assert.Equal(t, "notebook", buchNounOrders["2"].Translation)

		assert.Equal(t, "sich schnell zu Fuß fortbewegen", laufenVerbOrders["1"].Description)
		assert.Equal(t, "run", laufenVerbOrders["1"].Translation)
	})

	t.Run("empty translation data map", func(t *testing.T) {
		emptyData := make(map[string]map[string]map[string]models.TranslationEntry)

		setTranslationTableDataFunc(t, func(targetLang, sourceLang string) (map[string]map[string]map[string]models.TranslationEntry, error) {
			return emptyData, nil
		})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/translations?source_lang=es&target_lang=en", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

		var resp models.TranslationDataResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, "en", resp.TargetLang)
		assert.Equal(t, "es", resp.SourceLang)
		assert.Empty(t, resp.Data)
	})
}

// Empty or missing source_lang or target_lang query parameters.
func TestGetTranslationData_EmptyTranslationCodeError(t *testing.T) {
	router := setupTranslationTestRouter(t)

	// Ensure db query is never reached on invalid input.
	setTranslationTableDataFunc(t, func(targetLang, sourceLang string) (map[string]map[string]map[string]models.TranslationEntry, error) {
		t.Fatal("getTranslationTableData must not be called when language code is empty or missing")
		return nil, nil
	})

	tests := []struct {
		name string
		url  string
	}{
		{
			name: "missing both parameters",
			url:  "/api/v1/translations",
		},
		{
			name: "missing source_lang",
			url:  "/api/v1/translations?target_lang=en",
		},
		{
			name: "missing target_lang",
			url:  "/api/v1/translations?source_lang=es",
		},
		{
			name: "empty source_lang",
			url:  "/api/v1/translations?source_lang=&target_lang=en",
		},
		{
			name: "empty target_lang",
			url:  "/api/v1/translations?source_lang=es&target_lang=",
		},
		{
			name: "both parameters empty",
			url:  "/api/v1/translations?source_lang=&target_lang=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			require.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

			var resp models.ErrorResponse
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Equal(t, constants.EmptyTranslationCodeError, resp.Error)
		})
	}
}

// Invalid translation language code format (2-4 lowercase ASCII letters).
func TestGetTranslationData_InvalidTranslationLangCodeError(t *testing.T) {
	router := setupTranslationTestRouter(t)

	// Ensure db query is never reached on invalid input.
	setTranslationTableDataFunc(t, func(targetLang, sourceLang string) (map[string]map[string]map[string]models.TranslationEntry, error) {
		t.Fatal("getTranslationTableData must not be called when language code format is invalid")
		return nil, nil
	})

	tests := []struct {
		name string
		url  string
	}{
		// Invalid target_lang
		{
			name: "target uppercase",
			url:  "/api/v1/translations?source_lang=es&target_lang=EN",
		},
		{
			name: "target single letter (too short)",
			url:  "/api/v1/translations?source_lang=es&target_lang=e",
		},
		{
			name: "target more than 4 letters (too long)",
			url:  "/api/v1/translations?source_lang=es&target_lang=english",
		},
		{
			name: "target with numbers",
			url:  "/api/v1/translations?source_lang=es&target_lang=e1",
		},
		{
			name: "target with hyphen",
			url:  "/api/v1/translations?source_lang=es&target_lang=en-us",
		},
		{
			name: "target with symbols",
			url:  "/api/v1/translations?source_lang=es&target_lang=e$",
		},
		// Invalid source_lang
		{
			name: "source uppercase",
			url:  "/api/v1/translations?source_lang=ES&target_lang=en",
		},
		{
			name: "source single letter (too short)",
			url:  "/api/v1/translations?source_lang=s&target_lang=en",
		},
		{
			name: "source more than 4 letters (too long)",
			url:  "/api/v1/translations?source_lang=spanish&target_lang=en",
		},
		{
			name: "source with numbers",
			url:  "/api/v1/translations?source_lang=12&target_lang=en",
		},
		{
			name: "source with underscore",
			url:  "/api/v1/translations?source_lang=es_es&target_lang=en",
		},
		// Both invalid
		{
			name: "both uppercase",
			url:  "/api/v1/translations?source_lang=ES&target_lang=EN",
		},
		{
			name: "both too long",
			url:  "/api/v1/translations?source_lang=spanish&target_lang=english",
		},
		{
			name: "both numbers",
			url:  "/api/v1/translations?source_lang=12&target_lang=34",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			require.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

			var resp models.ErrorResponse
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Equal(t, constants.InvalidTranslationLangCodeError, resp.Error)
		})
	}
}

// Translation data table does not exist (404 Not Found).
func TestGetTranslationData_NotFound(t *testing.T) {
	router := setupTranslationTestRouter(t)

	tests := []struct {
		name         string
		sourceLang   string
		targetLang   string
		dbErr        error
		expectedCode int
		expectedErr  string
	}{
		{
			name:        "table TranslationDataENFromES does not exist",
			sourceLang:  "es",
			targetLang:  "en",
			dbErr:       errors.New("translation table TranslationDataENFromES does not exist"),
			expectedErr: "No translation data for 'en' from 'es'",
		},
		{
			name:        "table TranslationDataDEFromBN does not exist",
			sourceLang:  "bn",
			targetLang:  "de",
			dbErr:       errors.New("translation table TranslationDataDEFromBN does not exist"),
			expectedErr: "No translation data for 'de' from 'bn'",
		},
		{
			name:        "generic error containing does not exist",
			sourceLang:  "fr",
			targetLang:  "it",
			dbErr:       errors.New("table does not exist in schema"),
			expectedErr: "No translation data for 'it' from 'fr'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setTranslationTableDataFunc(t, func(targetLang, sourceLang string) (map[string]map[string]map[string]models.TranslationEntry, error) {
				assert.Equal(t, tt.targetLang, targetLang)
				assert.Equal(t, tt.sourceLang, sourceLang)
				return nil, tt.dbErr
			})

			url := fmt.Sprintf("/api/v1/translations?source_lang=%s&target_lang=%s", tt.sourceLang, tt.targetLang)
			req := httptest.NewRequest(http.MethodGet, url, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			require.Equal(t, http.StatusNotFound, rec.Code)
			assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

			var resp models.ErrorResponse
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedErr, resp.Error)
		})
	}
}

// Database query failure or unexpected error (500 Internal Server Error).
func TestGetTranslationData_InternalServerError(t *testing.T) {
	router := setupTranslationTestRouter(t)

	tests := []struct {
		name  string
		dbErr error
	}{
		{
			name:  "database connection refused",
			dbErr: errors.New("dial tcp 127.0.0.1:3306: connect: connection refused"),
		},
		{
			name:  "syntax error in query",
			dbErr: errors.New("error querying TranslationDataENFromES: syntax error"),
		},
		{
			name:  "table existence check failed",
			dbErr: errors.New("error checking table existence for TranslationDataENFromES: timeout"),
		},
		{
			name:  "invalid translation table name error",
			dbErr: errors.New("invalid translation table name: TranslationDataINVALIDFromES"),
		},
		{
			name:  "row scanning error",
			dbErr: errors.New("error scanning row: sql: expected 5 destination arguments in Scan, not 4"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setTranslationTableDataFunc(t, func(targetLang, sourceLang string) (map[string]map[string]map[string]models.TranslationEntry, error) {
				return nil, tt.dbErr
			})

			req := httptest.NewRequest(http.MethodGet, "/api/v1/translations?source_lang=es&target_lang=en", nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			require.Equal(t, http.StatusInternalServerError, rec.Code)
			assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

			var resp models.ErrorResponse
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Equal(t, constants.ErrorFetchingTranslationData, resp.Error)
		})
	}
}
