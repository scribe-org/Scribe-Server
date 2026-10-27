// SPDX-License-Identifier: GPL-3.0-or-later

package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/scribe-org/scribe-server/api/validators"
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
			func(_ string) ([]string, error) {
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

// MARK: Language Data Retrieval Tests

// setupLanguageDataTestRouter creates a minimal Gin engine configured in test mode for testing GetLanguageData.
func setupLanguageDataTestRouter(t *testing.T) *gin.Engine {
	t.Helper()

	previousMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() { gin.SetMode(previousMode) })

	router := gin.New()
	router.GET("/api/v1/data/:lang", GetLanguageData)
	return router
}

// stubLanguageData temporarily swaps package-level function pointers used by GetLanguageData.
func stubLanguageData(
	t *testing.T,
	languagesFn func() ([]string, error),
	dataTypesFn func(lang string) ([]string, error),
	updatedAtFn func(lang string) (string, error),
	tableDataFn func(lang, dataType string) (map[string]any, error),
) {
	t.Helper()

	origLanguages := getAvailableLanguages
	origDataTypes := getLanguageDataTypes
	origUpdatedAt := getLanguageUpdatedAt
	origTableData := getLanguageTableData

	if languagesFn != nil {
		getAvailableLanguages = languagesFn
	}
	if dataTypesFn != nil {
		getLanguageDataTypes = dataTypesFn
	}
	if updatedAtFn != nil {
		getLanguageUpdatedAt = updatedAtFn
	}
	if tableDataFn != nil {
		getLanguageTableData = tableDataFn
	}

	t.Cleanup(func() {
		getAvailableLanguages = origLanguages
		getLanguageDataTypes = origDataTypes
		getLanguageUpdatedAt = origUpdatedAt
		getLanguageTableData = origTableData
	})
}

// initLanguageValidator sets supported languages and registers cleanup to reset it.
func initLanguageValidator(t *testing.T, langs []string) {
	t.Helper()

	validators.InitLanguageValidator(langs)
	t.Cleanup(func() { validators.InitLanguageValidator(nil) })
}

// TestGetLanguageData_Success tests the successful retrieval of full language data.
func TestGetLanguageData_Success(t *testing.T) {
	router := setupLanguageDataTestRouter(t)
	initLanguageValidator(t, []string{"en", "de"})

	nounsSchema := map[string]string{"singular": "string", "plural": "string"}
	nounsData := []any{map[string]any{"singular": "book", "plural": "books"}}
	verbsSchema := map[string]string{"infinitive": "string", "past": "string"}
	verbsData := []any{map[string]any{"infinitive": "run", "past": "ran"}}

	stubLanguageData(
		t,
		func() ([]string, error) { return []string{"en", "de"}, nil },
		func(lang string) ([]string, error) {
			if lang == "en" {
				return []string{"nouns", "verbs"}, nil
			}
			return nil, fmt.Errorf("unexpected language: %s", lang)
		},
		func(lang string) (string, error) {
			assert.Equal(t, "en", lang)
			return "2026-10-20", nil
		},
		func(lang, dataType string) (map[string]any, error) {
			assert.Equal(t, "en", lang)
			if dataType == "nouns" {
				return map[string]any{"schema": nounsSchema, "data": nounsData}, nil
			}
			return map[string]any{"schema": verbsSchema, "data": verbsData}, nil
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/data/en", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

	var resp models.LanguageDataResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "en", resp.Language)
	assert.Equal(t, constants.APIVersion, resp.Contract.Version)
	assert.Equal(t, "2026-10-20", resp.Contract.UpdatedAt)
	assert.Equal(t, nounsSchema, resp.Contract.Fields["nouns"])
	assert.Equal(t, verbsSchema, resp.Contract.Fields["verbs"])
	assert.NotEmpty(t, resp.Data["nouns"])
	assert.NotEmpty(t, resp.Data["verbs"])
}

// TestGetLanguageData_UpdatedAtError tests that a failure fetching updatedAt still returns 200 with empty updatedAt.
func TestGetLanguageData_UpdatedAtError(t *testing.T) {
	router := setupLanguageDataTestRouter(t)
	initLanguageValidator(t, []string{"en"})

	stubLanguageData(
		t,
		func() ([]string, error) { return []string{"en"}, nil },
		func(_ string) ([]string, error) { return []string{"nouns"}, nil },
		func(_ string) (string, error) { return "", errors.New("failed to fetch updated_at") },
		func(_, _ string) (map[string]any, error) {
			return map[string]any{
				"schema": map[string]string{"singular": "string"},
				"data":   []any{map[string]any{"singular": "cat"}},
			}, nil
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/data/en", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

	var resp models.LanguageDataResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "en", resp.Language)
	assert.Empty(t, resp.Contract.UpdatedAt)
	assert.Equal(t, map[string]string{"singular": "string"}, resp.Contract.Fields["nouns"])
	assert.NotEmpty(t, resp.Data["nouns"])
}

// TestGetLanguageData_PartialFailure tests that a failure in one data type's table fetch is skipped while others succeed.
func TestGetLanguageData_PartialFailure(t *testing.T) {
	router := setupLanguageDataTestRouter(t)
	initLanguageValidator(t, []string{"en"})

	stubLanguageData(
		t,
		func() ([]string, error) { return []string{"en"}, nil },
		func(_ string) ([]string, error) { return []string{"nouns", "verbs"}, nil },
		func(_ string) (string, error) { return "2026-10-20", nil },
		func(_, dataType string) (map[string]any, error) {
			if dataType == "verbs" {
				return nil, errors.New("failed to query verbs table")
			}
			return map[string]any{
				"schema": map[string]string{"singular": "string"},
				"data":   []any{map[string]any{"singular": "dog"}},
			}, nil
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/data/en", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

	var resp models.LanguageDataResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "en", resp.Language)
	assert.Contains(t, resp.Contract.Fields, "nouns")
	assert.NotContains(t, resp.Contract.Fields, "verbs")
	assert.Contains(t, resp.Data, "nouns")
	assert.NotContains(t, resp.Data, "verbs")
}

// TestGetLanguageData_IncompleteTableData tests that missing schema or data keys in tableData do not cause panics.
func TestGetLanguageData_IncompleteTableData(t *testing.T) {
	router := setupLanguageDataTestRouter(t)
	initLanguageValidator(t, []string{"en"})

	stubLanguageData(
		t,
		func() ([]string, error) { return []string{"en"}, nil },
		func(_ string) ([]string, error) { return []string{"nouns"}, nil },
		func(_ string) (string, error) { return "2026-10-20", nil },
		func(_, _ string) (map[string]any, error) {
			return map[string]any{"schema": "invalid-schema-type"}, nil
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/data/en", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp models.LanguageDataResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Empty(t, resp.Contract.Fields["nouns"])
	assert.NotContains(t, resp.Data, "nouns")
}

// TestGetLanguageData_SanitizeLanguageCodeFails tests the defensive branch where sanitization fails during the loop.
func TestGetLanguageData_SanitizeLanguageCodeFails(t *testing.T) {
	router := setupLanguageDataTestRouter(t)
	initLanguageValidator(t, []string{"en"})

	stubLanguageData(
		t,
		func() ([]string, error) { return []string{"en"}, nil },
		func(_ string) ([]string, error) {
			validators.InitLanguageValidator(nil)
			return []string{"nouns"}, nil
		},
		func(_ string) (string, error) { return "2026-10-20", nil },
		func(_, _ string) (map[string]any, error) {
			return map[string]any{"schema": map[string]string{"singular": "string"}}, nil
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/data/en", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp models.LanguageDataResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "en", resp.Language)
	assert.Empty(t, resp.Contract.Fields["nouns"])
	assert.Empty(t, resp.Data["nouns"])
}

// TestGetLanguageData_InvalidLanguageCode tests that invalid language code formats return 400 Bad Request.
func TestGetLanguageData_InvalidLanguageCode(t *testing.T) {
	router := setupLanguageDataTestRouter(t)
	initLanguageValidator(t, []string{"en", "de"})

	tests := []struct {
		name string
		lang string
	}{
		{name: "uppercase code", lang: "EN"},
		{name: "single character", lang: "e"},
		{name: "three characters", lang: "eng"},
		{name: "numeric code", lang: "12"},
		{name: "alphanumeric code", lang: "e1"},
		{name: "special characters", lang: "e!"},
		{name: "unsupported format code not in validator", lang: "zz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/data/"+tt.lang, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			require.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

			var resp models.ErrorResponse
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			assert.Equal(t, constants.InvalidLanguageCodeError, resp.Error)
		})
	}
}

// TestGetLanguageData_UnsupportedLanguage tests that valid code format unsupported in DB returns 404 Not Found.
func TestGetLanguageData_UnsupportedLanguage(t *testing.T) {
	router := setupLanguageDataTestRouter(t)
	initLanguageValidator(t, []string{"en", "de"})

	stubLanguageData(
		t,
		func() ([]string, error) { return []string{"en"}, nil },
		nil,
		nil,
		nil,
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/data/de", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

	var resp models.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "Language 'de' not supported", resp.Error)
}

// TestGetLanguageData_DatabaseErrors tests 500 responses when DB queries error.
func TestGetLanguageData_DatabaseErrors(t *testing.T) {
	router := setupLanguageDataTestRouter(t)
	initLanguageValidator(t, []string{"en"})

	t.Run("500 when checking available languages fails", func(t *testing.T) {
		stubLanguageData(
			t,
			func() ([]string, error) { return nil, errors.New("db connection failure") },
			nil,
			nil,
			nil,
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/data/en", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

		var resp models.ErrorResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, "Failed to check language availability", resp.Error)
	})

	t.Run("500 when fetching language data types fails", func(t *testing.T) {
		stubLanguageData(
			t,
			func() ([]string, error) { return []string{"en"}, nil },
			func(_ string) ([]string, error) { return nil, errors.New("query data types error") },
			nil,
			nil,
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/data/en", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

		var resp models.ErrorResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, "Failed to fetch language data types", resp.Error)
	})
}

// MARK: Language Version Tests

// setupLanguageVersionTestRouter creates a minimal Gin engine configured in test mode for testing GetLanguageVersion.
func setupLanguageVersionTestRouter(t *testing.T) *gin.Engine {
	t.Helper()

	previousMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() { gin.SetMode(previousMode) })

	router := gin.New()
	router.GET("/api/v1/data-version/:lang", GetLanguageVersion)
	return router
}

// stubLanguageVersion temporarily swaps package-level function pointers used by GetLanguageVersion.
func stubLanguageVersion(
	t *testing.T,
	languagesFn func() ([]string, error),
	versionsFn func(lang string) (map[string]string, error),
) {
	t.Helper()

	origLanguages := getAvailableLanguages
	origVersions := getLanguageVersions

	if languagesFn != nil {
		getAvailableLanguages = languagesFn
	}
	if versionsFn != nil {
		getLanguageVersions = versionsFn
	}

	t.Cleanup(func() {
		getAvailableLanguages = origLanguages
		getLanguageVersions = origVersions
	})
}

// TestGetLanguageVersion_SuccessMapShape tests that GetLanguageVersion returns 200 with the <datatype>_last_modified map shape.
func TestGetLanguageVersion_SuccessMapShape(t *testing.T) {
	router := setupLanguageVersionTestRouter(t)
	initLanguageValidator(t, []string{"en", "fr"})

	expectedVersions := map[string]string{
		"nouns_last_modified":        "2026-10-20 14:32:00",
		"verbs_last_modified":        "2026-10-18 09:15:30",
		"prepositions_last_modified": "2026-09-01 00:00:00",
		"profanity_last_modified":    "2026-10-22 18:45:12",
	}

	var capturedLang string
	stubLanguageVersion(
		t,
		func() ([]string, error) {
			return []string{"en", "fr"}, nil
		},
		func(lang string) (map[string]string, error) {
			capturedLang = lang
			return expectedVersions, nil
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/data-version/en", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

	var resp models.LanguageVersionResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	assert.Equal(t, "en", capturedLang)
	assert.Equal(t, "en", resp.Language)
	assert.Equal(t, expectedVersions, resp.Versions)

	// Verify the <datatype>_last_modified map shape.
	for key, value := range resp.Versions {
		assert.True(t, strings.HasSuffix(key, "_last_modified"), "expected key %q to end with _last_modified", key)
		dataType := strings.TrimSuffix(key, "_last_modified")
		assert.NotEmpty(t, dataType, "datatype prefix must not be empty")
		assert.Contains(t, []string{"nouns", "verbs", "prepositions", "profanity"}, dataType)
		assert.Equal(t, expectedVersions[key], value)
	}
}

// TestGetLanguageVersion_FallbackDate tests the 1970-01-01 fallback date when a table has no lastModified column.
func TestGetLanguageVersion_FallbackDate(t *testing.T) {
	router := setupLanguageVersionTestRouter(t)
	initLanguageValidator(t, []string{"en"})

	t.Run("table without lastModified column uses 1970-01-01 fallback while others retain dates", func(t *testing.T) {
		fallbackVersions := map[string]string{
			"nouns_last_modified":        "2026-10-20 14:32:00",
			"verbs_last_modified":        "1970-01-01",
			"prepositions_last_modified": "1970-01-01",
		}

		stubLanguageVersion(
			t,
			func() ([]string, error) { return []string{"en"}, nil },
			func(lang string) (map[string]string, error) {
				assert.Equal(t, "en", lang)
				return fallbackVersions, nil
			},
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/data-version/en", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

		var resp models.LanguageVersionResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

		assert.Equal(t, "en", resp.Language)
		assert.Equal(t, fallbackVersions, resp.Versions)
		assert.Equal(t, "2026-10-20 14:32:00", resp.Versions["nouns_last_modified"])
		assert.Equal(t, "1970-01-01", resp.Versions["verbs_last_modified"])
		assert.Equal(t, "1970-01-01", resp.Versions["prepositions_last_modified"])
	})

	t.Run("all tables missing lastModified column fall back to 1970-01-01", func(t *testing.T) {
		fallbackVersions := map[string]string{
			"nouns_last_modified": "1970-01-01",
		}

		stubLanguageVersion(
			t,
			func() ([]string, error) { return []string{"en"}, nil },
			func(_ string) (map[string]string, error) {
				return fallbackVersions, nil
			},
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/data-version/en", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

		var resp models.LanguageVersionResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

		assert.Equal(t, "en", resp.Language)
		assert.Equal(t, "1970-01-01", resp.Versions["nouns_last_modified"])
	})
}

// TestGetLanguageVersion_EmptyMap tests that a language with zero data types returns an empty map in versions.
func TestGetLanguageVersion_EmptyMap(t *testing.T) {
	router := setupLanguageVersionTestRouter(t)
	initLanguageValidator(t, []string{"de"})

	stubLanguageVersion(
		t,
		func() ([]string, error) { return []string{"de"}, nil },
		func(lang string) (map[string]string, error) {
			assert.Equal(t, "de", lang)
			return map[string]string{}, nil
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/data-version/de", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

	var resp models.LanguageVersionResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	assert.Equal(t, "de", resp.Language)
	assert.NotNil(t, resp.Versions)
	assert.Empty(t, resp.Versions)

	var raw map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &raw))
	assert.Equal(t, "de", raw["language"])
	assert.Equal(t, map[string]any{}, raw["versions"])
}

// TestGetLanguageVersion_InvalidLanguageCode tests 400 Bad Request responses for malformed or invalid language codes.
func TestGetLanguageVersion_InvalidLanguageCode(t *testing.T) {
	router := setupLanguageVersionTestRouter(t)
	initLanguageValidator(t, []string{"en", "de"})

	// Ensure database functions are never called when the language code is invalid.
	stubLanguageVersion(
		t,
		func() ([]string, error) {
			t.Fatal("getAvailableLanguages must not be called when language code is invalid")
			return nil, nil
		},
		func(_ string) (map[string]string, error) {
			t.Fatal("getLanguageVersions must not be called when language code is invalid")
			return nil, nil
		},
	)

	tests := []struct {
		name string
		url  string
	}{
		{
			name: "uppercase ISO code",
			url:  "/api/v1/data-version/EN",
		},
		{
			name: "mixed case ISO code",
			url:  "/api/v1/data-version/En",
		},
		{
			name: "single letter code (too short)",
			url:  "/api/v1/data-version/e",
		},
		{
			name: "three letter ISO code (too long for ISO 639-1)",
			url:  "/api/v1/data-version/eng",
		},
		{
			name: "full language name",
			url:  "/api/v1/data-version/english",
		},
		{
			name: "numeric code",
			url:  "/api/v1/data-version/12",
		},
		{
			name: "code with digit",
			url:  "/api/v1/data-version/e1",
		},
		{
			name: "code with hyphen",
			url:  "/api/v1/data-version/en-us",
		},
		{
			name: "code with special character",
			url:  "/api/v1/data-version/e$",
		},
		{
			name: "unsupported valid-length code not in supported languages map",
			url:  "/api/v1/data-version/xx",
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
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			assert.Equal(t, constants.InvalidLanguageCodeError, resp.Error)
		})
	}
}

// TestGetLanguageVersion_NotFound tests 404 Not Found responses when the language code is valid format but unsupported in the database.
func TestGetLanguageVersion_NotFound(t *testing.T) {
	router := setupLanguageVersionTestRouter(t)
	initLanguageValidator(t, []string{"en", "fr", "es"})

	stubLanguageVersion(
		t,
		func() ([]string, error) {
			// Available languages in database does not include "es".
			return []string{"en", "fr"}, nil
		},
		func(_ string) (map[string]string, error) {
			t.Fatal("getLanguageVersions must not be called when language is not supported")
			return nil, nil
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/data-version/es", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

	var resp models.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "Language 'es' not supported", resp.Error)
}

// TestGetLanguageVersion_DatabaseErrors tests 500 Internal Server Error responses when database operations fail.
func TestGetLanguageVersion_DatabaseErrors(t *testing.T) {
	router := setupLanguageVersionTestRouter(t)
	initLanguageValidator(t, []string{"en"})

	t.Run("500 when checking available languages fails", func(t *testing.T) {
		stubLanguageVersion(
			t,
			func() ([]string, error) {
				return nil, errors.New("database connection refused")
			},
			func(_ string) (map[string]string, error) {
				t.Fatal("getLanguageVersions must not be called when available languages check fails")
				return nil, nil
			},
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/data-version/en", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

		var resp models.ErrorResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, "Failed to check language availability", resp.Error)
	})

	t.Run("500 when fetching language versions fails", func(t *testing.T) {
		stubLanguageVersion(
			t,
			func() ([]string, error) {
				return []string{"en"}, nil
			},
			func(lang string) (map[string]string, error) {
				assert.Equal(t, "en", lang)
				return nil, errors.New("database error querying version timestamps")
			},
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/data-version/en", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

		var resp models.ErrorResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, constants.ErrorFetchingLanguageVersions, resp.Error)
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

		setTranslationTableDataFunc(t, func(_, _ string) (map[string]map[string]map[string]models.TranslationEntry, error) {
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
	setTranslationTableDataFunc(t, func(_, _ string) (map[string]map[string]map[string]models.TranslationEntry, error) {
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
	setTranslationTableDataFunc(t, func(_, _ string) (map[string]map[string]map[string]models.TranslationEntry, error) {
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
			setTranslationTableDataFunc(t, func(_, _ string) (map[string]map[string]map[string]models.TranslationEntry, error) {
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
