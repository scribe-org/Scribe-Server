// SPDX-License-Identifier: GPL-3.0-or-later

package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToIntPtr(t *testing.T) {
	t.Parallel()
	if ToIntPtr(nil) != nil {
		t.Fatal("nil -> nil")
	}
	v := ToIntPtr(42)
	if v == nil || *v != 42 {
		t.Fatalf("int: got %v", v)
	}
	v = ToIntPtr(int64(7))
	if v == nil || *v != 7 {
		t.Fatalf("int64: got %v", v)
	}
	if ToIntPtr("nope") != nil {
		t.Fatal("string -> nil")
	}
}

func TestToStringPtr(t *testing.T) {
	t.Parallel()
	if ToStringPtr(nil) != nil {
		t.Fatal("nil -> nil")
	}
	v := ToStringPtr("hi")
	if v == nil || *v != "hi" {
		t.Fatalf("got %v", v)
	}
	if ToStringPtr(1) != nil {
		t.Fatal("int -> nil")
	}
}

func TestGetLanguageDisplayName(t *testing.T) {
	t.Parallel()
	if got := GetLanguageDisplayName("en"); got != "English" {
		t.Fatalf("en: %q", got)
	}
	if got := GetLanguageDisplayName("zz"); got != "ZZ" {
		t.Fatalf("unknown: %q", got)
	}
}

func TestBuildLanguageStatResponse(t *testing.T) {
	t.Parallel()
	// Missing keys must not panic; ToIntPtr yields nil.
	resp := BuildLanguageStatResponse("en", map[string]any{})
	if resp.Code != "en" {
		t.Fatalf("code %q", resp.Code)
	}
	if resp.LanguageName == nil || *resp.LanguageName != "English" {
		t.Fatalf("name %+v", resp.LanguageName)
	}
	if resp.Nouns != nil || resp.Verbs != nil || resp.Prepositions != nil || resp.Profanity != nil {
		t.Fatalf(
			"expected nil counts, nouns=%v verbs=%v prepositions=%v profanity=%v",
			resp.Nouns, resp.Verbs, resp.Prepositions, resp.Profanity,
		)
	}

	resp = BuildLanguageStatResponse("de", map[string]any{
		"nouns": 3, "verbs": int64(9), "prepositions": 2, "profanity": 1,
	})
	if resp.Nouns == nil || *resp.Nouns != 3 {
		t.Fatalf("nouns %+v", resp.Nouns)
	}
	if resp.Verbs == nil || *resp.Verbs != 9 {
		t.Fatalf("verbs %+v", resp.Verbs)
	}
	if resp.Prepositions == nil || *resp.Prepositions != 2 {
		t.Fatalf("prepositions %+v", resp.Prepositions)
	}
	if resp.Profanity == nil || *resp.Profanity != 1 {
		t.Fatalf("profanity %+v", resp.Profanity)
	}

	// Languages without prepositions have a nil count.
	resp = BuildLanguageStatResponse("en", map[string]any{"nouns": 3, "verbs": 9, "profanity": 1})
	if resp.Prepositions != nil {
		t.Fatalf("expected nil prepositions, got %v", *resp.Prepositions)
	}
}

func TestIsValidTableName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		tableName string
		want      bool
	}{
		{name: "empty table name", tableName: "", want: false},
		{name: "valid noun table", tableName: "ENLanguageDataNounsScribe", want: true},
		{name: "valid verb table", tableName: "DELanguageDataVerbsScribe", want: true},
		{name: "valid long datatype", tableName: "FRLanguageDataPrepositionsScribe", want: true},
		{name: "invalid prefix lowercase", tableName: "enLanguageDataNounsScribe", want: false},
		{name: "invalid prefix three letters", tableName: "ENGLanguageDataNounsScribe", want: false},
		{name: "invalid missing suffix", tableName: "ENLanguageDataNouns", want: false},
		{name: "invalid lowercase suffix", tableName: "ENLanguageDataNounsscribe", want: false},
		{name: "invalid suffix with extra text", tableName: "ENLanguageDataNounsScribeExtra", want: false},
		{name: "invalid numbers in datatype", tableName: "ENLanguageData123Scribe", want: false},
		{name: "sql injection semicolon", tableName: "ENLanguageDataNounsScribe;", want: false},
		{name: "sql injection drop table", tableName: "ENLanguageDataNounsScribe; DROP TABLE scribe", want: false},
		{name: "sql injection union", tableName: "ENLanguageDataNounsScribe UNION SELECT 1", want: false},
		{name: "sql injection line comment", tableName: "ENLanguageDataNounsScribe--", want: false},
		{name: "sql injection block comment", tableName: "ENLanguageDataNounsScribe/*", want: false},
		{name: "sql injection quotes", tableName: "ENLanguageDataNounsScribe' OR '1'='1", want: false},
		{name: "sql injection space", tableName: "ENLanguageDataNounsScribe ", want: false},
		{name: "invalid special chars", tableName: "ENLanguageData-NounsScribe", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidTableName(tt.tableName))
		})
	}
}

func TestIsValidTranslationTableName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		tableName string
		want      bool
	}{
		{name: "empty table name", tableName: "", want: false},
		{name: "valid 2-letter codes", tableName: "TranslationDataENFromDE", want: true},
		{name: "valid 3-letter codes", tableName: "TranslationDataBNFromENG", want: true},
		{name: "valid 4-letter codes", tableName: "TranslationDataFRENFromGERM", want: true},
		{name: "invalid lowercase from", tableName: "TranslationDataENfromDE", want: false},
		{name: "invalid lowercase target code", tableName: "TranslationDataenFromDE", want: false},
		{name: "invalid lowercase source code", tableName: "TranslationDataENFromde", want: false},
		{name: "invalid missing target", tableName: "TranslationDataFromDE", want: false},
		{name: "invalid missing source", tableName: "TranslationDataENFrom", want: false},
		{name: "invalid suffix with extra text", tableName: "TranslationDataENFromDEExtra", want: false},
		{name: "sql injection semicolon", tableName: "TranslationDataENFromDE;", want: false},
		{name: "sql injection drop table", tableName: "TranslationDataENFromDE; DROP TABLE scribe", want: false},
		{name: "sql injection union", tableName: "TranslationDataENFromDE UNION SELECT 1", want: false},
		{name: "sql injection line comment", tableName: "TranslationDataENFromDE--", want: false},
		{name: "sql injection block comment", tableName: "TranslationDataENFromDE/*", want: false},
		{name: "sql injection quotes", tableName: "TranslationDataENFromDE' OR '1'='1", want: false},
		{name: "sql injection space", tableName: "TranslationDataENFromDE ", want: false},
		{name: "invalid special chars", tableName: "TranslationDataEN-FromDE", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidTranslationTableName(tt.tableName))
		})
	}
}
