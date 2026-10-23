// SPDX-License-Identifier: GPL-3.0-or-later

package database

import (
	"errors"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/scribe-org/scribe-server/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildLanguageStatQueryAllDataTypes(t *testing.T) {
	t.Parallel()
	query, dataTypes, err := buildLanguageStatQuery("de", []string{"nouns", "prepositions", "profanity", "verbs"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"nouns", "verbs", "prepositions", "profanity"}
	if !reflect.DeepEqual(dataTypes, want) {
		t.Fatalf("data types %v, want %v", dataTypes, want)
	}

	for _, part := range []string{
		"(SELECT COUNT(*) FROM DELanguageDataNounsScribe) AS nouns",
		"(SELECT COUNT(*) FROM DELanguageDataVerbsScribe) AS verbs",
		"(SELECT COUNT(*) FROM DELanguageDataPrepositionsScribe) AS prepositions",
		"(SELECT COUNT(*) FROM DELanguageDataProfanityScribe) AS profanity",
	} {
		if !strings.Contains(query, part) {
			t.Fatalf("query %q missing %q", query, part)
		}
	}
}

func TestBuildLanguageStatQueryOptionalDataTypesMissing(t *testing.T) {
	t.Parallel()
	// English has no prepositions table, which shouldn't exclude the language.
	query, dataTypes, err := buildLanguageStatQuery("en", []string{"nouns", "profanity", "verbs"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"nouns", "verbs", "profanity"}
	if !reflect.DeepEqual(dataTypes, want) {
		t.Fatalf("data types %v, want %v", dataTypes, want)
	}
	if strings.Contains(query, "Prepositions") {
		t.Fatalf("query %q should not count prepositions", query)
	}
}

func TestBuildLanguageStatQueryRequiredDataTypeMissing(t *testing.T) {
	t.Parallel()
	for _, available := range [][]string{
		{"verbs", "profanity"},
		{"nouns", "prepositions"},
		{},
	} {
		query, dataTypes, err := buildLanguageStatQuery("fr", available)
		if err != nil {
			t.Fatalf("unexpected error for %v: %v", available, err)
		}
		if query != "" || len(dataTypes) != 0 {
			t.Fatalf("expected no query for %v, got %q %v", available, query, dataTypes)
		}
	}
}

func TestBuildLanguageStatQueryInvalidLanguage(t *testing.T) {
	t.Parallel()
	// Table names for invalid language codes are rejected before a query is built.
	for _, lang := range []string{"e1", "eng", "e;", ""} {
		query, dataTypes, err := buildLanguageStatQuery(lang, []string{"nouns", "verbs"})
		if err == nil {
			t.Fatalf("expected an error for %q, got %q %v", lang, query, dataTypes)
		}
		if query != "" || dataTypes != nil {
			t.Fatalf("expected no query for %q, got %q %v", lang, query, dataTypes)
		}
	}
}

func TestBuildLanguageStatQueryIgnoresOtherDataTypes(t *testing.T) {
	t.Parallel()
	_, dataTypes, err := buildLanguageStatQuery("sv", []string{"emojikeywords", "nouns", "verbs"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"nouns", "verbs"}
	if !reflect.DeepEqual(dataTypes, want) {
		t.Fatalf("data types %v, want %v", dataTypes, want)
	}
}

// MARK: Database Queries

var (
	availableLanguagesQuery = regexp.QuoteMeta("SELECT DISTINCT SUBSTRING(TABLE_NAME, 1, 2) as language_code")
	dataTypesQuery          = regexp.QuoteMeta("SELECT TABLE_NAME")
)

// tableRows returns query rows with a single column of table names.
func tableRows(tableNames ...string) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{"TABLE_NAME"})
	for _, tableName := range tableNames {
		rows.AddRow(tableName)
	}

	return rows
}

func TestGetAvailableLanguages(t *testing.T) {
	t.Run("returns lowercase language codes", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectQuery(availableLanguagesQuery).
			WithArgs(testDatabaseName).
			WillReturnRows(sqlmock.NewRows([]string{"language_code"}).AddRow("DE").AddRow("EN"))

		languages, err := GetAvailableLanguages()

		require.NoError(t, err)
		assert.Equal(t, []string{"de", "en"}, languages)
	})

	t.Run("error when the query fails", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectQuery(availableLanguagesQuery).WillReturnError(errors.New("connection lost"))

		_, err := GetAvailableLanguages()

		assert.ErrorContains(t, err, "error querying available languages: connection lost")
	})

	t.Run("error when a language can't be read", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectQuery(availableLanguagesQuery).
			WillReturnRows(sqlmock.NewRows([]string{"language_code"}).AddRow(nil))

		_, err := GetAvailableLanguages()

		assert.ErrorContains(t, err, "error scanning language")
	})

	t.Run("error when reading the rows fails", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectQuery(availableLanguagesQuery).
			WillReturnRows(sqlmock.NewRows([]string{"language_code"}).
				AddRow("DE").AddRow("EN").RowError(1, errors.New("connection lost")))

		languages, err := GetAvailableLanguages()

		assert.ErrorContains(t, err, "error reading languages: connection lost")
		assert.Nil(t, languages)
	})
}

func TestGetLanguageDataTypes(t *testing.T) {
	t.Run("returns lowercase data types of language tables", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectQuery(dataTypesQuery).
			WithArgs(testDatabaseName, "ENLanguageData%Scribe").
			WillReturnRows(tableRows(
				"ENLanguageDataEmojiKeywordsScribe",
				"ENLanguageDataNounsScribe",
				"ENLanguageDataVerbsScribe",
				// Note: Tables that don't match the pattern, such as migration backups, are ignored.
				"ENLanguageDataNounsScribeOld",
			))

		dataTypes, err := GetLanguageDataTypes("en")

		require.NoError(t, err)
		assert.Equal(t, []string{"emojikeywords", "nouns", "verbs"}, dataTypes)
	})

	t.Run("rejects an invalid language code without querying", func(t *testing.T) {
		newMockDB(t)

		for _, lang := range []string{"", "e", "eng"} {
			_, err := GetLanguageDataTypes(lang)
			assert.EqualError(t, err, "invalid language code", lang)
		}
	})

	t.Run("error when the query fails", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectQuery(dataTypesQuery).WillReturnError(errors.New("connection lost"))

		_, err := GetLanguageDataTypes("en")

		assert.ErrorContains(t, err, "error querying data types: connection lost")
	})

	t.Run("error when a table name can't be read", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectQuery(dataTypesQuery).WillReturnRows(sqlmock.NewRows([]string{"TABLE_NAME"}).AddRow(nil))

		_, err := GetLanguageDataTypes("en")

		assert.ErrorContains(t, err, "error scanning table name")
	})

	t.Run("error when reading the rows fails", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectQuery(dataTypesQuery).
			WillReturnRows(tableRows("ENLanguageDataNounsScribe", "ENLanguageDataVerbsScribe").
				RowError(1, errors.New("connection lost")))

		dataTypes, err := GetLanguageDataTypes("en")

		assert.ErrorContains(t, err, "error reading data types: connection lost")
		assert.Nil(t, dataTypes)
	})
}

// MARK: Language Statistics

// expectStatQuery expects the query counting the entries of the given data types for a language.
func expectStatQuery(mock sqlmock.Sqlmock, lang string, dataTypes ...string) *sqlmock.ExpectedQuery {
	counts := make([]string, 0, len(dataTypes))
	for _, dataType := range dataTypes {
		tableName := strings.ToUpper(lang) + "LanguageData" + strings.ToUpper(dataType[:1]) + dataType[1:] + "Scribe"
		counts = append(counts, "(SELECT COUNT(*) FROM "+tableName+") AS "+dataType)
	}

	return mock.ExpectQuery("^" + regexp.QuoteMeta("SELECT "+strings.Join(counts, ", ")) + "$")
}

func TestGetLanguageStat(t *testing.T) {
	t.Run("counts the available data types", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectQuery(dataTypesQuery).
			WithArgs(testDatabaseName, "DELanguageData%Scribe").
			WillReturnRows(tableRows(
				"DELanguageDataNounsScribe",
				"DELanguageDataPrepositionsScribe",
				"DELanguageDataProfanityScribe",
				"DELanguageDataVerbsScribe",
			))
		expectStatQuery(mock, "de", "nouns", "verbs", "prepositions", "profanity").
			WillReturnRows(sqlmock.NewRows([]string{"nouns", "verbs", "prepositions", "profanity"}).
				AddRow(178751, 23, 196, 113))

		// Note: Codes are trimmed and are case insensitive.
		stat, err := GetLanguageStat(" De ")

		require.NoError(t, err)
		assert.Equal(t, map[string]any{
			"code": "de", "nouns": 178751, "verbs": 23, "prepositions": 196, "profanity": 113,
		}, stat)
	})

	t.Run("doesn't count data types the language doesn't have", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectQuery(dataTypesQuery).
			WillReturnRows(tableRows("ENLanguageDataNounsScribe", "ENLanguageDataVerbsScribe"))
		expectStatQuery(mock, "en", "nouns", "verbs").
			WillReturnRows(sqlmock.NewRows([]string{"nouns", "verbs"}).AddRow(33498, 8787))

		stat, err := GetLanguageStat("en")

		require.NoError(t, err)
		assert.Equal(t, map[string]any{"code": "en", "nouns": 33498, "verbs": 8787}, stat)
	})

	t.Run("skips a language without nouns or verbs", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectQuery(dataTypesQuery).WillReturnRows(tableRows("FRLanguageDataNounsScribe"))

		stat, err := GetLanguageStat("fr")

		require.NoError(t, err)
		assert.Nil(t, stat)
	})

	t.Run("skips a language when a table is removed during the query", func(t *testing.T) {
		for name, queryErr := range map[string]error{
			"MariaDB missing table error":  &mysql.MySQLError{Number: 1146, Message: "Table 'ENLanguageDataVerbsScribe' doesn't exist"},
			"missing table message":        errors.New("Table 'ENLanguageDataVerbsScribe' doesn't exist"),
			"SQLite missing table message": errors.New("no such table: ENLanguageDataVerbsScribe"),
		} {
			mock := newMockDB(t)
			mock.ExpectQuery(dataTypesQuery).
				WillReturnRows(tableRows("ENLanguageDataNounsScribe", "ENLanguageDataVerbsScribe"))
			expectStatQuery(mock, "en", "nouns", "verbs").WillReturnError(queryErr)

			stat, err := GetLanguageStat("en")

			require.NoError(t, err, name)
			assert.Nil(t, stat, name)
		}
	})

	t.Run("rejects an invalid language code without querying", func(t *testing.T) {
		newMockDB(t)

		for _, lang := range []string{"", "e", "eng", "e1"} {
			_, err := GetLanguageStat(lang)
			assert.ErrorContains(t, err, "invalid language code", lang)
		}
	})

	t.Run("error when the data types can't be fetched", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectQuery(dataTypesQuery).WillReturnError(errors.New("connection lost"))

		_, err := GetLanguageStat("en")

		assert.ErrorContains(t, err, "error fetching data types for EN")
	})

	t.Run("error when the counts can't be fetched", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectQuery(dataTypesQuery).
			WillReturnRows(tableRows("ENLanguageDataNounsScribe", "ENLanguageDataVerbsScribe"))
		expectStatQuery(mock, "en", "nouns", "verbs").WillReturnError(errors.New("connection lost"))

		_, err := GetLanguageStat("en")

		assert.ErrorContains(t, err, "error scanning stats for EN: connection lost")
	})
}

func TestGetAllLanguageStats(t *testing.T) {
	t.Run("returns statistics for languages with nouns and verbs", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectQuery(availableLanguagesQuery).
			WillReturnRows(sqlmock.NewRows([]string{"language_code"}).AddRow("DE").AddRow("EN").AddRow("FR"))

		// German fails, so it's skipped.
		mock.ExpectQuery(dataTypesQuery).WithArgs(testDatabaseName, "DELanguageData%Scribe").
			WillReturnError(errors.New("connection lost"))

		// English is included.
		mock.ExpectQuery(dataTypesQuery).WithArgs(testDatabaseName, "ENLanguageData%Scribe").
			WillReturnRows(tableRows("ENLanguageDataNounsScribe", "ENLanguageDataVerbsScribe"))
		expectStatQuery(mock, "en", "nouns", "verbs").
			WillReturnRows(sqlmock.NewRows([]string{"nouns", "verbs"}).AddRow(33498, 8787))

		// French has no verbs, so it's skipped.
		mock.ExpectQuery(dataTypesQuery).WithArgs(testDatabaseName, "FRLanguageData%Scribe").
			WillReturnRows(tableRows("FRLanguageDataNounsScribe"))

		stats, err := GetAllLanguageStats()

		require.NoError(t, err)
		require.Len(t, stats, 1)
		assert.Equal(t, BuildLanguageStatResponse("en", map[string]any{"nouns": 33498, "verbs": 8787}), stats[0])
	})

	t.Run("empty list when there are no languages", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectQuery(availableLanguagesQuery).WillReturnRows(sqlmock.NewRows([]string{"language_code"}))

		stats, err := GetAllLanguageStats()

		require.NoError(t, err)
		assert.Equal(t, []models.LanguageStatisticsReponse{}, stats)
	})

	t.Run("error when the languages can't be fetched", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectQuery(availableLanguagesQuery).WillReturnError(errors.New("connection lost"))

		_, err := GetAllLanguageStats()

		assert.ErrorContains(t, err, "failed to get available languages")
	})
}
