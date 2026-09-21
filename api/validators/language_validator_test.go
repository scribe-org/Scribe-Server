// SPDX-License-Identifier: GPL-3.0-or-later

package validators

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// setSupportedLanguages configures the package-level language map for a test + restores an empty map so tests stay independent
func setSupportedLanguages(t *testing.T, langs []string) {
	t.Helper()

	InitLanguageValidator(langs)
	t.Cleanup(func() { InitLanguageValidator(nil) })
}

func TestInitLanguageValidator(t *testing.T) {
	t.Cleanup(func() { InitLanguageValidator(nil) })

	t.Run("supports the given languages", func(t *testing.T) {
		InitLanguageValidator([]string{"en", "de"})

		assert.True(t, IsValidLanguageCode("en"))
		assert.True(t, IsValidLanguageCode("de"))
		assert.False(t, IsValidLanguageCode("fr"))
	})

	t.Run("replaces previous languages", func(t *testing.T) {
		InitLanguageValidator([]string{"en", "de"})
		InitLanguageValidator([]string{"fr"})

		assert.True(t, IsValidLanguageCode("fr"))
		assert.False(t, IsValidLanguageCode("en"))
		assert.False(t, IsValidLanguageCode("de"))
	})

	t.Run("nil list clears all languages", func(t *testing.T) {
		InitLanguageValidator([]string{"en"})
		InitLanguageValidator(nil)

		assert.False(t, IsValidLanguageCode("en"))
	})

	t.Run("empty list clears all languages", func(t *testing.T) {
		InitLanguageValidator([]string{"en"})
		InitLanguageValidator([]string{})

		assert.False(t, IsValidLanguageCode("en"))
	})

	t.Run("duplicate entries are accepted", func(t *testing.T) {
		InitLanguageValidator([]string{"en", "en"})

		assert.True(t, IsValidLanguageCode("en"))
	})
}

func TestIsValidLanguageCode(t *testing.T) {
	setSupportedLanguages(t, []string{"en", "de", "fr"})

	tests := []struct {
		name string
		code string
		want bool
	}{
		{name: "supported code", code: "en", want: true},
		{name: "another supported code", code: "de", want: true},
		{name: "valid format but unsupported", code: "es", want: false},
		{name: "uppercase code", code: "EN", want: false},
		{name: "mixed case code", code: "En", want: false},
		{name: "empty code", code: "", want: false},
		{name: "single character", code: "e", want: false},
		{name: "three characters", code: "eng", want: false},
		{name: "leading whitespace", code: " e", want: false},
		{name: "trailing whitespace", code: "en ", want: false},
		{name: "digits", code: "12", want: false},
		{name: "SQL injection attempt", code: "en; DROP TABLE", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidLanguageCode(tt.code))
		})
	}
}

func TestIsValidLanguageCode_NoLanguagesConfigured(t *testing.T) {
	setSupportedLanguages(t, nil)

	assert.False(t, IsValidLanguageCode("en"))
}

func TestSanitizeLanguageCode(t *testing.T) {
	setSupportedLanguages(t, []string{"en", "de"})

	tests := []struct {
		name string
		code string
		want string
	}{
		{name: "supported code is uppercased", code: "en", want: "EN"},
		{name: "another supported code", code: "de", want: "DE"},
		{name: "unsupported code", code: "fr", want: ""},
		{name: "uppercase input is rejected", code: "EN", want: ""},
		{name: "empty code", code: "", want: ""},
		{name: "too long", code: "eng", want: ""},
		{name: "injection attempt", code: "en; DROP TABLE", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, SanitizeLanguageCode(tt.code))
		})
	}
}

func TestIsLanguageSupported(t *testing.T) {
	tests := []struct {
		name      string
		lang      string
		available []string
		want      bool
	}{
		{name: "language in list", lang: "en", available: []string{"en", "de"}, want: true},
		{name: "language not in list", lang: "fr", available: []string{"en", "de"}, want: false},
		{name: "case sensitive", lang: "EN", available: []string{"en", "de"}, want: false},
		{name: "empty list", lang: "en", available: []string{}, want: false},
		{name: "nil list", lang: "en", available: nil, want: false},
		{name: "empty language not in list", lang: "", available: []string{"en"}, want: false},
		{name: "empty language in list", lang: "", available: []string{""}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsLanguageSupported(tt.lang, tt.available))
		})
	}
}

func TestIsValidTranslationLangCode(t *testing.T) {
	tests := []struct {
		name string
		code string
		want bool
	}{
		{name: "two-letter code", code: "de", want: true},
		{name: "three-letter code", code: "pnb", want: true},
		{name: "four-letter code", code: "abcd", want: true},
		{name: "uppercase code", code: "DE", want: false},
		{name: "mixed case code", code: "dE", want: false},
		{name: "empty code", code: "", want: false},
		{name: "single letter", code: "d", want: false},
		{name: "five letters", code: "abcde", want: false},
		{name: "digits", code: "d1", want: false},
		{name: "hyphen", code: "de-", want: false},
		{name: "leading whitespace", code: " de", want: false},
		{name: "trailing newline", code: "de\n", want: false},
		{name: "non-ASCII letters", code: "dé", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidTranslationLangCode(tt.code))
		})
	}
}
