// SPDX-License-Identifier: GPL-3.0-or-later

package dbqueries

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/scribe-org/scribe-server/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetTranslationTableData tests retrieving and nesting translation data under various database scenarios.
func TestGetTranslationTableData(t *testing.T) {
	tableExistsQuery := regexp.QuoteMeta("FROM information_schema.TABLES")
	selectQuery := regexp.QuoteMeta("SELECT word, wordType, wordOrder, description, translation FROM `TranslationDataBNFromDE`")

	t.Run("returns nested translations for a valid translation table", func(t *testing.T) {
		mock := newMockDB(t)

		mock.ExpectQuery(tableExistsQuery).
			WithArgs(testDatabaseName, "TranslationDataBNFromDE").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		mock.ExpectQuery(selectQuery).
			WillReturnRows(sqlmock.NewRows([]string{"word", "wordType", "wordOrder", "description", "translation"}).
				AddRow("buch", "noun", "1", "ein gebundenes Schriftwerk", "book").
				AddRow("buch", "noun", "2", "ein Notizheft", "notebook").
				AddRow("laufen", "verb", "1", "sich schnell zu Fuß fortbewegen", "run"))

		result, err := GetTranslationTableData("bn", "de")

		require.NoError(t, err)
		expected := map[string]map[string]map[string]models.TranslationEntry{
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
		assert.Equal(t, expected, result)
	})

	t.Run("returns empty map when table has no rows", func(t *testing.T) {
		mock := newMockDB(t)

		mock.ExpectQuery(tableExistsQuery).
			WithArgs(testDatabaseName, "TranslationDataBNFromDE").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		mock.ExpectQuery(selectQuery).
			WillReturnRows(sqlmock.NewRows([]string{"word", "wordType", "wordOrder", "description", "translation"}))

		result, err := GetTranslationTableData("bn", "de")

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("rejects an invalid table name format without querying", func(t *testing.T) {
		newMockDB(t)

		result, err := GetTranslationTableData("b", "de")

		assert.EqualError(t, err, "invalid translation table name: TranslationDataBFromDE")
		assert.Nil(t, result)
	})

	t.Run("rejects invalid characters without querying", func(t *testing.T) {
		newMockDB(t)

		result, err := GetTranslationTableData("bn", "de; DROP TABLE users")

		assert.EqualError(t, err, "invalid translation table name: TranslationDataBNFromDE; DROP TABLE USERS")
		assert.Nil(t, result)
	})

	t.Run("error when checking table existence fails", func(t *testing.T) {
		mock := newMockDB(t)

		mock.ExpectQuery(tableExistsQuery).
			WithArgs(testDatabaseName, "TranslationDataBNFromDE").
			WillReturnError(errors.New("connection lost"))

		result, err := GetTranslationTableData("bn", "de")

		assert.ErrorContains(t, err, "error checking table existence for TranslationDataBNFromDE: error checking table existence: connection lost")
		assert.Nil(t, result)
	})

	t.Run("error when translation table does not exist", func(t *testing.T) {
		mock := newMockDB(t)

		mock.ExpectQuery(tableExistsQuery).
			WithArgs(testDatabaseName, "TranslationDataBNFromDE").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		result, err := GetTranslationTableData("bn", "de")

		assert.EqualError(t, err, "translation table TranslationDataBNFromDE does not exist")
		assert.Nil(t, result)
	})

	t.Run("error when query fails", func(t *testing.T) {
		mock := newMockDB(t)

		mock.ExpectQuery(tableExistsQuery).
			WithArgs(testDatabaseName, "TranslationDataBNFromDE").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		mock.ExpectQuery(selectQuery).
			WillReturnError(errors.New("query failed"))

		result, err := GetTranslationTableData("bn", "de")

		assert.ErrorContains(t, err, "error querying TranslationDataBNFromDE: query failed")
		assert.Nil(t, result)
	})

	t.Run("error when scanning row fails", func(t *testing.T) {
		mock := newMockDB(t)

		mock.ExpectQuery(tableExistsQuery).
			WithArgs(testDatabaseName, "TranslationDataBNFromDE").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		mock.ExpectQuery(selectQuery).
			WillReturnRows(sqlmock.NewRows([]string{"word", "wordType", "wordOrder", "description", "translation"}).
				AddRow("buch", "noun", "1", nil, "book"))

		result, err := GetTranslationTableData("bn", "de")

		assert.ErrorContains(t, err, "error scanning row")
		assert.Nil(t, result)
	})

	t.Run("error when iterating rows fails", func(t *testing.T) {
		mock := newMockDB(t)

		mock.ExpectQuery(tableExistsQuery).
			WithArgs(testDatabaseName, "TranslationDataBNFromDE").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		mock.ExpectQuery(selectQuery).
			WillReturnRows(sqlmock.NewRows([]string{"word", "wordType", "wordOrder", "description", "translation"}).
				AddRow("buch", "noun", "1", "description", "book").
				RowError(0, errors.New("connection broken")))

		result, err := GetTranslationTableData("bn", "de")

		assert.ErrorContains(t, err, "error iterating rows: connection broken")
		assert.Nil(t, result)
	})
}
