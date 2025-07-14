// SPDX-License-Identifier: GPL-3.0-or-later

// Package constants provides shared, reusable constant values across the application.
package constants

const (
    // InvalidLanguageCodeError indicates that a language code is invalid or unsupported (expects ISO 639-1 format).
    InvalidLanguageCodeError = "Invalid language code or not supported. Use ISO 639-1 format (e.g., 'en', 'fr')"

    // ErrorFetchingLanguages indicates a failure when retrieving available languages.
    ErrorFetchingLanguages = "Failed to fetch available languages"

    // ErrorFetchingLanguageVersions indicates a failure when retrieving language version data.
    ErrorFetchingLanguageVersions = "Failed to fetch language versions"
)
