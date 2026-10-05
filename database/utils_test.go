// SPDX-License-Identifier: GPL-3.0-or-later

package database

import "testing"

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
