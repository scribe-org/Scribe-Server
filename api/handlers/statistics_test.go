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
	"github.com/scribe-org/scribe-server/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Creates a minimal Gin engine configured in test mode for testing the language-stats endpoint.
func setupStatisticsTestRouter(t *testing.T) *gin.Engine {
	t.Helper()

	previousMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() { gin.SetMode(previousMode) })

	router := gin.New()
	router.GET("/api/v1/language-stats", GetLanguageStats)
	return router
}

// Temporarily swaps package-level database function pointers for isolated unit testing.
func stubStatisticsDatabase(
	t *testing.T,
	allStatsFn func() ([]models.LanguageStatisticsReponse, error),
	availableLangsFn func() ([]string, error),
	langStatFn func(lang string) (map[string]any, error),
) {
	t.Helper()

	origAllStats := getAllLanguageStats
	origAvailableLangs := getAvailableLanguages
	origLangStat := getLanguageStat

	if allStatsFn != nil {
		getAllLanguageStats = allStatsFn
	}
	if availableLangsFn != nil {
		getAvailableLanguages = availableLangsFn
	}
	if langStatFn != nil {
		getLanguageStat = langStatFn
	}

	t.Cleanup(func() {
		getAllLanguageStats = origAllStats
		getAvailableLanguages = origAvailableLangs
		getLanguageStat = origLangStat
	})
}

func intPtr(i int) *int {
	return &i
}

func strPtr(s string) *string {
	return &s
}

// Tests GET /api/v1/language-stats when no codes parameter is passed or when it is empty.
func TestGetLanguageStats_AllLanguages(t *testing.T) {
	router := setupStatisticsTestRouter(t)

	mockAllStats := []models.LanguageStatisticsReponse{
		{
			Code:         "en",
			LanguageName: strPtr("English"),
			Nouns:        intPtr(100),
			Verbs:        intPtr(50),
			Prepositions: intPtr(10),
			Profanity:    intPtr(5),
		},
		{
			Code:         "de",
			LanguageName: strPtr("German"),
			Nouns:        intPtr(200),
			Verbs:        intPtr(80),
			Prepositions: intPtr(15),
			Profanity:    intPtr(2),
		},
	}

	t.Run("200 returns all stats when codes query parameter is omitted", func(t *testing.T) {
		stubStatisticsDatabase(
			t,
			func() ([]models.LanguageStatisticsReponse, error) {
				return mockAllStats, nil
			},
			nil,
			nil,
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/language-stats", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

		var response []models.LanguageStatisticsReponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
		assert.Equal(t, mockAllStats, response)
	})

	t.Run("200 returns all stats when codes query parameter is empty string", func(t *testing.T) {
		stubStatisticsDatabase(
			t,
			func() ([]models.LanguageStatisticsReponse, error) {
				return mockAllStats, nil
			},
			nil,
			nil,
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/language-stats?codes=", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

		var response []models.LanguageStatisticsReponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
		assert.Equal(t, mockAllStats, response)
	})

	t.Run("500 when GetAllLanguageStats fails", func(t *testing.T) {
		stubStatisticsDatabase(
			t,
			func() ([]models.LanguageStatisticsReponse, error) {
				return nil, errors.New("database failure while querying all stats")
			},
			nil,
			nil,
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/language-stats", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

		var response models.ErrorResponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
		assert.Equal(t, "Failed to fetch all language statistics", response.Error)
	})
}

// Tests filtering with valid comma-separated codes.
func TestGetLanguageStats_CommaSeparatedCodes(t *testing.T) {
	router := setupStatisticsTestRouter(t)

	availableLangs := []string{"en", "de", "fr"}
	statsData := map[string]map[string]any{
		"en": {"nouns": 100, "verbs": 50, "prepositions": 10, "profanity": 5},
		"de": {"nouns": 200, "verbs": 80, "prepositions": 15, "profanity": 2},
		"fr": {"nouns": 150, "verbs": 60, "prepositions": 8, "profanity": 1},
	}

	stubStatisticsDatabase(
		t,
		nil,
		func() ([]string, error) {
			return availableLangs, nil
		},
		func(lang string) (map[string]any, error) {
			if stat, ok := statsData[lang]; ok {
				return stat, nil
			}
			return nil, nil
		},
	)

	t.Run("200 for single code", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/language-stats?codes=en", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

		var response []models.LanguageStatisticsReponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
		require.Len(t, response, 1)

		assert.Equal(t, "en", response[0].Code)
		assert.Equal(t, "English", *response[0].LanguageName)
		assert.Equal(t, 100, *response[0].Nouns)
		assert.Equal(t, 50, *response[0].Verbs)
		assert.Equal(t, 10, *response[0].Prepositions)
		assert.Equal(t, 5, *response[0].Profanity)
	})

	t.Run("200 for multiple comma-separated codes preserving order", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/language-stats?codes=de,en", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

		var response []models.LanguageStatisticsReponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
		require.Len(t, response, 2)

		assert.Equal(t, "de", response[0].Code)
		assert.Equal(t, "German", *response[0].LanguageName)
		assert.Equal(t, 200, *response[0].Nouns)

		assert.Equal(t, "en", response[1].Code)
		assert.Equal(t, "English", *response[1].LanguageName)
		assert.Equal(t, 100, *response[1].Nouns)
	})
}

// Verifies that leading, trailing, and internal whitespace around codes is trimmed.
func TestGetLanguageStats_WhitespaceTrimming(t *testing.T) {
	router := setupStatisticsTestRouter(t)

	availableLangs := []string{"en", "fr"}
	statsData := map[string]map[string]any{
		"en": {"nouns": 100, "verbs": 50, "prepositions": 10, "profanity": 5},
		"fr": {"nouns": 150, "verbs": 60, "prepositions": 8, "profanity": 1},
	}

	stubStatisticsDatabase(
		t,
		nil,
		func() ([]string, error) {
			return availableLangs, nil
		},
		func(lang string) (map[string]any, error) {
			if stat, ok := statsData[lang]; ok {
				return stat, nil
			}
			return nil, nil
		},
	)

	testCases := []struct {
		name        string
		queryString string
	}{
		{
			name:        "spaces around codes and commas",
			queryString: "?codes=%20en%20,%20fr%20",
		},
		{
			name:        "leading whitespace",
			queryString: "?codes=%20%20en,fr",
		},
		{
			name:        "trailing whitespace",
			queryString: "?codes=en,fr%20%20",
		},
		{
			name:        "tabs and mixed whitespace",
			queryString: "?codes=%09en%09,%09fr%09",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/language-stats"+tc.queryString, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

			var response []models.LanguageStatisticsReponse
			require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
			require.Len(t, response, 2)
			assert.Equal(t, "en", response[0].Code)
			assert.Equal(t, "fr", response[1].Code)
		})
	}
}

// Verifies that empty segments between commas are skipped.
func TestGetLanguageStats_EmptySegments(t *testing.T) {
	router := setupStatisticsTestRouter(t)

	availableLangs := []string{"en", "de"}
	statsData := map[string]map[string]any{
		"en": {"nouns": 100, "verbs": 50},
		"de": {"nouns": 200, "verbs": 80},
	}

	stubStatisticsDatabase(
		t,
		nil,
		func() ([]string, error) {
			return availableLangs, nil
		},
		func(lang string) (map[string]any, error) {
			if stat, ok := statsData[lang]; ok {
				return stat, nil
			}
			return nil, nil
		},
	)

	t.Run("consecutive, leading, and trailing empty segments are ignored", func(t *testing.T) {
		urls := []string{
			"/api/v1/language-stats?codes=,en,,de,",
			"/api/v1/language-stats?codes=en,,de",
			"/api/v1/language-stats?codes=en,,,de",
			"/api/v1/language-stats?codes=,,en,de,,",
		}

		for _, url := range urls {
			req := httptest.NewRequest(http.MethodGet, url, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

			var response []models.LanguageStatisticsReponse
			require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
			require.Len(t, response, 2)
			assert.Equal(t, "en", response[0].Code)
			assert.Equal(t, "de", response[1].Code)
		}
	})

	t.Run("only commas or whitespace returns 404", func(t *testing.T) {
		emptyQueries := []string{
			"/api/v1/language-stats?codes=,",
			"/api/v1/language-stats?codes=,,",
			"/api/v1/language-stats?codes=,,,",
			"/api/v1/language-stats?codes=%20,%20",
			"/api/v1/language-stats?codes=%20%20",
		}

		for _, url := range emptyQueries {
			req := httptest.NewRequest(http.MethodGet, url, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			require.Equal(t, http.StatusNotFound, rec.Code)
			assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

			var response models.ErrorResponse
			require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
			assert.Equal(t, "No statistics found for requested languages", response.Error)
		}
	})
}

// Verifies that duplicate language codes are deduplicated.
func TestGetLanguageStats_Duplicates(t *testing.T) {
	router := setupStatisticsTestRouter(t)

	availableLangs := []string{"en", "de"}
	statsData := map[string]map[string]any{
		"en": {"nouns": 100, "verbs": 50, "prepositions": 10, "profanity": 5},
		"de": {"nouns": 200, "verbs": 80, "prepositions": 15, "profanity": 2},
	}

	var fetchCalls []string
	stubStatisticsDatabase(
		t,
		nil,
		func() ([]string, error) {
			return availableLangs, nil
		},
		func(lang string) (map[string]any, error) {
			fetchCalls = append(fetchCalls, lang)
			if stat, ok := statsData[lang]; ok {
				return stat, nil
			}
			return nil, nil
		},
	)

	t.Run("duplicate codes are queried and returned for each occurrence", func(t *testing.T) {
		fetchCalls = nil

		req := httptest.NewRequest(http.MethodGet, "/api/v1/language-stats?codes=en,de,en", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

		var response []models.LanguageStatisticsReponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
		require.Len(t, response, 3)

		assert.Equal(t, "en", response[0].Code)
		assert.Equal(t, "de", response[1].Code)
		assert.Equal(t, "en", response[2].Code)
		assert.Equal(t, []string{"en", "de", "en"}, fetchCalls)
	})
}

// Verifies unsupported codes are silently dropped.
func TestGetLanguageStats_SilentDroppingUnsupportedCodes(t *testing.T) {
	router := setupStatisticsTestRouter(t)

	availableLangs := []string{"en", "de"}
	statsData := map[string]map[string]any{
		"en": {"nouns": 100, "verbs": 50},
		"de": {"nouns": 200, "verbs": 80},
	}

	var fetchCalls []string
	stubStatisticsDatabase(
		t,
		nil,
		func() ([]string, error) {
			return availableLangs, nil
		},
		func(lang string) (map[string]any, error) {
			fetchCalls = append(fetchCalls, lang)
			if stat, ok := statsData[lang]; ok {
				return stat, nil
			}
			return nil, nil
		},
	)

	t.Run("unsupported codes are ignored and only supported ones returned", func(t *testing.T) {
		fetchCalls = nil

		req := httptest.NewRequest(http.MethodGet, "/api/v1/language-stats?codes=xx,en,yy,de,zz", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

		var response []models.LanguageStatisticsReponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
		require.Len(t, response, 2)

		assert.Equal(t, "en", response[0].Code)
		assert.Equal(t, "de", response[1].Code)
		assert.Equal(t, []string{"en", "de"}, fetchCalls)
	})
}

// Verifies that 404 fires only when all requested codes are dropped or have no stats.
func TestGetLanguageStats_NotFound(t *testing.T) {
	router := setupStatisticsTestRouter(t)

	availableLangs := []string{"en", "de"}

	t.Run("404 when all requested codes are unsupported", func(t *testing.T) {
		stubStatisticsDatabase(
			t,
			nil,
			func() ([]string, error) {
				return availableLangs, nil
			},
			func(_ string) (map[string]any, error) {
				return nil, nil
			},
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/language-stats?codes=xx,yy,zz", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

		var response models.ErrorResponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
		assert.Equal(t, "No statistics found for requested languages", response.Error)
	})

	t.Run("404 when single unsupported code is requested", func(t *testing.T) {
		stubStatisticsDatabase(
			t,
			nil,
			func() ([]string, error) {
				return availableLangs, nil
			},
			func(_ string) (map[string]any, error) {
				return nil, nil
			},
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/language-stats?codes=invalid", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNotFound, rec.Code)
		var response models.ErrorResponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
		assert.Equal(t, "No statistics found for requested languages", response.Error)
	})

	t.Run("404 when supported codes have nil stats in database", func(t *testing.T) {
		stubStatisticsDatabase(
			t,
			nil,
			func() ([]string, error) {
				return availableLangs, nil
			},
			func(_ string) (map[string]any, error) {
				return nil, nil
			},
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/language-stats?codes=en,de", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNotFound, rec.Code)
		var response models.ErrorResponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
		assert.Equal(t, "No statistics found for requested languages", response.Error)
	})

	t.Run("404 does not fire when at least one valid code produces statistics", func(t *testing.T) {
		stubStatisticsDatabase(
			t,
			nil,
			func() ([]string, error) {
				return availableLangs, nil
			},
			func(lang string) (map[string]any, error) {
				if lang == "en" {
					return map[string]any{"nouns": 100, "verbs": 50}, nil
				}
				return nil, nil
			},
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/language-stats?codes=unsupported,en", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var response []models.LanguageStatisticsReponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
		require.Len(t, response, 1)
		assert.Equal(t, "en", response[0].Code)
	})
}

// Verifies error handling when database operations fail.
func TestGetLanguageStats_DatabaseErrors(t *testing.T) {
	router := setupStatisticsTestRouter(t)

	t.Run("500 when checking available languages fails", func(t *testing.T) {
		stubStatisticsDatabase(
			t,
			nil,
			func() ([]string, error) {
				return nil, errors.New("failed to connect to database")
			},
			nil,
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/language-stats?codes=en,de", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

		var response models.ErrorResponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
		assert.Equal(t, "Failed to check available languages", response.Error)
	})

	t.Run("partial failure when GetLanguageStat fails for one language but succeeds for another", func(t *testing.T) {
		stubStatisticsDatabase(
			t,
			nil,
			func() ([]string, error) {
				return []string{"en", "de"}, nil
			},
			func(lang string) (map[string]any, error) {
				if lang == "en" {
					return map[string]any{"nouns": 100, "verbs": 50}, nil
				}
				return nil, fmt.Errorf("error reading stats for %s", lang)
			},
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/language-stats?codes=en,de", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

		var response []models.LanguageStatisticsReponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
		require.Len(t, response, 1)
		assert.Equal(t, "en", response[0].Code)
	})

	t.Run("404 when GetLanguageStat errors for all requested languages", func(t *testing.T) {
		stubStatisticsDatabase(
			t,
			nil,
			func() ([]string, error) {
				return []string{"en", "de"}, nil
			},
			func(lang string) (map[string]any, error) {
				return nil, fmt.Errorf("error reading stats for %s", lang)
			},
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/language-stats?codes=en,de", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

		var response models.ErrorResponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
		assert.Equal(t, "No statistics found for requested languages", response.Error)
	})
}
