// SPDX-License-Identifier: GPL-3.0-or-later

// Package validators provides shared, reusable validators across the application.
package validators

import (
	"slices"
	"strings"
	"sync"
)

var (
	supportedLanguagesMap = map[string]bool{}
	mu                    sync.RWMutex
)

// InitLanguageValidator initializes the list of supported languages.
// This should be called once during server startup.
func InitLanguageValidator(langs []string) {
	mu.Lock()
	defer mu.Unlock()

	supportedLanguagesMap = make(map[string]bool)
	for _, lang := range langs {
		supportedLanguagesMap[lang] = true
	}
}

// IsValidLanguageCode checks if the language code is a valid ISO 639-1 code
// and that it is among the supported languages.
func IsValidLanguageCode(lang string) bool {
	if len(lang) != 2 || lang != strings.ToLower(lang) {
		return false
	}

	mu.RLock()
	defer mu.RUnlock()

	return supportedLanguagesMap[lang]
}

// SanitizeLanguageCode ensures language code is safe for table name construction.
func SanitizeLanguageCode(lang string) string {
	if !IsValidLanguageCode(lang) {
		return ""
	}
	// Convert to uppercase for table name prefix.
	return strings.ToUpper(lang)
}

// IsLanguageSupported checks if a language exists in the provided list.
func IsLanguageSupported(lang string, availableLanguages []string) bool {
	return slices.Contains(availableLanguages, lang)
}
