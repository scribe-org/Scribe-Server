// SPDX-License-Identifier: GPL-3.0-or-later
package models

// LanguageDataResponse represents the response for GET /data/:lang
type LanguageDataResponse struct {
	Language string                 `json:"language"`
	Contract Contract              `json:"contract"`
	Data     map[string]interface{} `json:"data"`
}

// Contract represents the schema contract for language data
type Contract struct {
	Version   string                       `json:"version"`
	UpdatedAt string                       `json:"updated_at"`
	Fields    map[string]map[string]string `json:"fields"`
}

// [WIP] LanguageVersionResponse represents the response for GET /data-version/:lang
type LanguageVersionResponse struct {
	Language                 string `json:"language"`
	NounsLastModified       string `json:"nouns_last_modified"`
	VerbsLastModified       string `json:"verbs_last_modified"`
	PrepositionsLastModified string `json:"prepositions_last_modified"`
}

// ErrorResponse represents API error responses
type ErrorResponse struct {
	Error string `json:"error"`
}