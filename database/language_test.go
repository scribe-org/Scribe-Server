// SPDX-License-Identifier: GPL-3.0-or-later

package database

import (
	"reflect"
	"strings"
	"testing"
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
