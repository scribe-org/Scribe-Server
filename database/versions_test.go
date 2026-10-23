// SPDX-License-Identifier: GPL-3.0-or-later

package database

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLatestVersionDate(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		versions map[string]string
		want     string
	}{
		{
			name: "most recent date across data types",
			versions: map[string]string{
				"nouns_last_modified":     "2026-09-30 03:21:30",
				"verbs_last_modified":     "2026-09-28 21:24:20",
				"profanity_last_modified": "2026-10-01 00:52:27",
			},
			want: "2026-10-01",
		},
		{
			name: "RFC3339 values",
			versions: map[string]string{
				"nouns_last_modified": "2026-06-07T08:35:17+01:00",
				"verbs_last_modified": "2026-08-24T18:37:32Z",
			},
			want: "2026-08-24",
		},
		{
			name: "default dates for missing values are ignored",
			versions: map[string]string{
				"nouns_last_modified":        "2026-09-29 15:47:49",
				"prepositions_last_modified": "1970-01-01",
			},
			want: "2026-09-29",
		},
		{
			name:     "only default dates",
			versions: map[string]string{"nouns_last_modified": "1970-01-01"},
			want:     "",
		},
		{
			name:     "values without dates",
			versions: map[string]string{"nouns_last_modified": "", "verbs_last_modified": "unknown"},
			want:     "",
		},
		{
			name:     "no data types",
			versions: map[string]string{},
			want:     "",
		},
	}

	for _, tc := range cases {
		if got := latestVersionDate(tc.versions); got != tc.want {
			t.Fatalf("%s: got %q, want %q", tc.name, got, tc.want)
		}
	}
}

// MARK: Database Queries

var (
	createVersionsTableQuery = regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS language_data_versions")
	updateVersionQuery       = regexp.QuoteMeta("INSERT INTO language_data_versions")
	lastModifiedColumnQuery  = regexp.QuoteMeta("AND COLUMN_NAME = 'lastModified'")
)

// lastModifiedQuery returns a pattern for the query of the most recent last modified date of a table.
func lastModifiedQuery(tableName string) string {
	return "(?s)" + regexp.QuoteMeta("COALESCE(MAX(lastModified), '1970-01-01')") + ".*" + regexp.QuoteMeta("FROM "+tableName)
}

func TestCreateLanguageDataVersionsTable(t *testing.T) {
	t.Run("creates the table", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectExec(createVersionsTableQuery).WillReturnResult(sqlmock.NewResult(0, 0))

		assert.NoError(t, CreateLanguageDataVersionsTable())
	})

	t.Run("error when the table can't be created", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectExec(createVersionsTableQuery).WillReturnError(errors.New("permission denied"))

		err := CreateLanguageDataVersionsTable()

		assert.ErrorContains(t, err, "error creating language_data_versions table: permission denied")
	})
}

func TestUpdateLanguageVersion(t *testing.T) {
	t.Run("inserts or updates the language", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectExec(updateVersionQuery).WithArgs("en").WillReturnResult(sqlmock.NewResult(1, 1))

		assert.NoError(t, UpdateLanguageVersion("en"))
	})

	t.Run("error when the language can't be updated", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectExec(updateVersionQuery).WithArgs("en").WillReturnError(errors.New("connection lost"))

		err := UpdateLanguageVersion("en")

		assert.ErrorContains(t, err, "error updating language version: connection lost")
	})
}

// expectLanguageVersions expects the queries of GetLanguageVersions for English with each fallback case.
func expectLanguageVersions(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(dataTypesQuery).
		WithArgs(testDatabaseName, "ENLanguageData%Scribe").
		WillReturnRows(tableRows(
			"ENLanguageDataNounsScribe",
			"ENLanguageDataPrepositionsScribe",
			"ENLanguageDataProfanityScribe",
			"ENLanguageDataVerbsScribe",
		))

	// Nouns have a last modified date.
	mock.ExpectQuery(lastModifiedColumnQuery).
		WithArgs(testDatabaseName, "ENLanguageDataNounsScribe").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(lastModifiedQuery("ENLanguageDataNounsScribe")).
		WillReturnRows(sqlmock.NewRows([]string{"max_last_modified"}).AddRow("2026-09-30 03:21:30"))

	// Checking the prepositions column fails.
	mock.ExpectQuery(lastModifiedColumnQuery).
		WithArgs(testDatabaseName, "ENLanguageDataPrepositionsScribe").
		WillReturnError(errors.New("connection lost"))

	// Getting the most recent profanity date fails.
	mock.ExpectQuery(lastModifiedColumnQuery).
		WithArgs(testDatabaseName, "ENLanguageDataProfanityScribe").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(lastModifiedQuery("ENLanguageDataProfanityScribe")).
		WillReturnError(errors.New("connection lost"))

	// Verbs don't have a lastModified column, so there's no query for their most recent date.
	mock.ExpectQuery(lastModifiedColumnQuery).
		WithArgs(testDatabaseName, "ENLanguageDataVerbsScribe").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
}

func TestGetLanguageVersions(t *testing.T) {
	t.Run("returns last modified dates with 1970-01-01 when they can't be found", func(t *testing.T) {
		mock := newMockDB(t)
		expectLanguageVersions(mock)

		versions, err := GetLanguageVersions("en")

		require.NoError(t, err)
		assert.Equal(t, map[string]string{
			"nouns_last_modified":        "2026-09-30 03:21:30",
			"prepositions_last_modified": "1970-01-01",
			"profanity_last_modified":    "1970-01-01",
			"verbs_last_modified":        "1970-01-01",
		}, versions)
	})

	t.Run("error when the data types can't be fetched", func(t *testing.T) {
		mock := newMockDB(t)
		mock.ExpectQuery(dataTypesQuery).WillReturnError(errors.New("connection lost"))

		_, err := GetLanguageVersions("en")

		assert.ErrorContains(t, err, "error querying data types: connection lost")
	})
}

func TestGetLanguageUpdatedAt(t *testing.T) {
	t.Run("returns the most recent last modified date", func(t *testing.T) {
		mock := newMockDB(t)
		expectLanguageVersions(mock)

		updatedAt, err := GetLanguageUpdatedAt("en")

		require.NoError(t, err)
		assert.Equal(t, "2026-09-30", updatedAt)
	})

	t.Run("error when the versions can't be fetched", func(t *testing.T) {
		newMockDB(t)

		_, err := GetLanguageUpdatedAt("e")

		assert.ErrorContains(t, err, "error fetching versions for e: invalid language code")
	})
}
