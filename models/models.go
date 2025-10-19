// SPDX-License-Identifier: GPL-3.0-or-later

// Package models defines data structures used across the Scribe Server.
package models

import "time"

// # MARK: - Error Models

// ErrorResponse represents a generic error message returned by the API.
// swagger:model ErrorResponse
type ErrorResponse struct {
	// Description of the error
	Error string `json:"error"`
}

// # MARK: - Contract Models

// Contract represents the data schema contract that defines structure and metadata for language data.
// swagger:model Contract
type Contract struct {
	// Contract version identifier
	Version string `json:"version"`
	// Last update timestamp (RFC3339 format)
	UpdatedAt string `json:"updated_at"`
	// Field definitions grouped by section and name
	Fields map[string]map[string]string `json:"fields"`
}

// ContractsResponse represents all contract metadata available for supported languages.
// swagger:model ContractsResponse
type ContractsResponse struct {
	Contracts map[string]any `json:"contracts"`
}

// # MARK: - Language Data Models

// LanguageDataResponse represents the complete response when fetching a language’s data.
// swagger:model LanguageDataResponse
type LanguageDataResponse struct {
	// ISO code of the language
	Language string `json:"language"`
	// Contract details defining the schema
	Contract Contract `json:"contract"`
	// Actual data, structured according to the contract
	Data map[string]any `json:"data"`
}

// LanguageDataVersion represents a single record in the language_data_versions table.
// swagger:model LanguageDataVersion
type LanguageDataVersion struct {
	LanguageISO string    `json:"language_iso"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// # MARK: - Metadata Models

// LanguageVersionResponse represents version information for a language dataset.
// swagger:model LanguageVersionResponse
type LanguageVersionResponse struct {
	// ISO code of the language
	Language string `json:"language"`
	// Map of data types to version identifiers
	Versions map[string]string `json:"versions"`
}

// LanguageInfo represents basic information about a supported language.
// swagger:model LanguageInfo
type LanguageInfo struct {
	// ISO code of the language (e.g. "en", "fr")
	Code string `json:"code"`
	// List of supported data types for this language
	DataTypes []string `json:"data_types"`
}

// AvailableLanguagesResponse represents a list of languages that are currently available.
// swagger:model AvailableLanguagesResponse
type AvailableLanguagesResponse struct {
	Languages []LanguageInfo `json:"languages"`
}

// # MARK: - Statistics Models

// LanguageStatisticsReponse represents linguistic statistics for a language.
// swagger:model LanguageStatisticsReponse
type LanguageStatisticsReponse struct {
	// ISO code of the language
	Code string `json:"code"`
	// Human-readable language name (nullable)
	LanguageName *string `json:"language_name"`
	// Count of noun entries
	Nouns *int `json:"nouns"`
	// Count of verb entries
	Verbs *int `json:"verbs"`
}
