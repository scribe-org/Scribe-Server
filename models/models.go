// SPDX-License-Identifier: GPL-3.0-or-later

// Package models defines data structures used across the Scribe Server.
package models

import "time"

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

// Contract represents the data schema contract.
type Contract struct {
	Version   string                       `json:"version"`
	UpdatedAt string                       `json:"updated_at"`
	Fields    map[string]map[string]string `json:"fields"`
}

// LanguageDataResponse represents the full language data response.
type LanguageDataResponse struct {
	Language string                 `json:"language"`
	Contract Contract               `json:"contract"`
	Data     map[string]interface{} `json:"data"`
}

// LanguageVersionResponse represents version information for a language.
type LanguageVersionResponse struct {
	Language string            `json:"language"`
	Versions map[string]string `json:"versions"`
}

// LanguageDataVersion represents a row in the language_data_versions table.
type LanguageDataVersion struct {
	LanguageISO string    `json:"language_iso"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// LanguageInfo represents basic language information.
type LanguageInfo struct {
	Code      string   `json:"code"`
	DataTypes []string `json:"data_types"`
}

// AvailableLanguagesResponse represents the response for available languages.
type AvailableLanguagesResponse struct {
	Languages []LanguageInfo `json:"languages"`
}
