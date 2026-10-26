// SPDX-License-Identifier: GPL-3.0-or-later

// Package handlers contains HTTP handlers for various API endpoints.
package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/scribe-org/scribe-server/api/dbqueries"
	"github.com/scribe-org/scribe-server/api/validators"
	"github.com/scribe-org/scribe-server/database"
	"github.com/scribe-org/scribe-server/internal/constants"
	"github.com/scribe-org/scribe-server/models"
	"github.com/spf13/viper"
)

// MARK: Languages Endpoints

// Function variables allow mocking database queries in unit tests.
var (
	getAvailableLanguages = database.GetAvailableLanguages
	getLanguageDataTypes  = database.GetLanguageDataTypes
	getLanguageUpdatedAt  = database.GetLanguageUpdatedAt
	getLanguageTableData  = dbqueries.GetLanguageTableData
	getLanguageVersions   = database.GetLanguageVersions
)

// GetAvailableLanguages returns a list of all supported languages and their available data types.
//
// @Summary List all supported languages
// @Description Fetches all languages currently supported by Scribe, along with their associated data types.
// @Tags Languages
// @Accept  json
// @Produce  json
// @Success 200 {object} models.AvailableLanguagesResponse "Successfully retrieved available languages"
// @Failure 500 {object} models.ErrorResponse "Internal server error occurred while fetching languages"
// @Router /api/v1/languages [get]
func GetAvailableLanguages(c *gin.Context) {
	languages, err := getAvailableLanguages()
	if err != nil {
		log.Printf("Error fetching available languages: %v", err)
		HandleError(c, http.StatusInternalServerError, constants.ErrorFetchingLanguages)
		return
	}

	var languageInfos []models.LanguageInfo
	for _, lang := range languages {
		dataTypes, err := getLanguageDataTypes(lang)
		if err != nil {
			log.Printf("Error fetching data types for %s: %v", lang, err)
			continue
		}

		languageInfos = append(languageInfos, models.LanguageInfo{
			Code:      lang,
			DataTypes: dataTypes,
		})
	}

	HandleSuccess(c, models.AvailableLanguagesResponse{
		Languages: languageInfos,
	})
}

// MARK: Language Data Retrieval

// GetLanguageData returns the full dataset and schema contract for a specific language.
//
// @Summary Retrieve full language data
// @Description Returns all available language data and schema contract for the given ISO 639-1 language code.
// @Tags Language Data
// @Accept  json
// @Produce  json
// @Param lang path string true "Language code (ISO 639-1)" example(en)
// @Success 200 {object} models.LanguageDataResponse "Successfully retrieved language data"
// @Failure 400 {object} models.ErrorResponse "Invalid or malformed language code"
// @Failure 404 {object} models.ErrorResponse "Requested language not found or unsupported"
// @Failure 500 {object} models.ErrorResponse "Internal server error while fetching data"
// @Router /api/v1/data/{lang} [get]
func GetLanguageData(c *gin.Context) {
	lang := c.Param("lang")

	// Validate language code.
	if !validators.IsValidLanguageCode(lang) {
		HandleError(c, http.StatusBadRequest, constants.InvalidLanguageCodeError)
		return
	}

	// Check if language exists in database.
	availableLanguages, err := getAvailableLanguages()
	if err != nil {
		log.Printf("Error checking available languages: %v", err)
		HandleError(c, http.StatusInternalServerError, "Failed to check language availability")
		return
	}

	if !validators.IsLanguageSupported(lang, availableLanguages) {
		HandleError(c, http.StatusNotFound, fmt.Sprintf("Language '%s' not supported", lang))
		return
	}

	// Get data types for the language.
	dataTypes, err := getLanguageDataTypes(lang)
	if err != nil {
		log.Printf("Error fetching data types for %s: %v", lang, err)
		HandleError(c, http.StatusInternalServerError, "Failed to fetch language data types")
		return
	}

	updatedAt, err := getLanguageUpdatedAt(lang)
	if err != nil {
		log.Printf("Error fetching update date for %s: %v", lang, err)
	}

	// Build the response.
	response := models.LanguageDataResponse{
		Language: lang,
		Contract: models.Contract{
			Version:   constants.APIVersion,
			UpdatedAt: updatedAt,
			Fields:    make(map[string]map[string]string),
		},
		Data: make(map[string]any),
	}

	// For each data type, get schema and data.
	for _, dataType := range dataTypes {
		// Sanitize inputs before constructing table name.
		sanitizedLang := validators.SanitizeLanguageCode(lang)
		if sanitizedLang == "" {
			log.Printf("Failed to sanitize language code: %s", lang)
			continue
		}

		tableData, err := getLanguageTableData(lang, dataType)
		if err != nil {
			log.Printf("Error fetching table data for %s/%s: %v", lang, dataType, err)
			continue
		}

		// Add to response.
		if schema, ok := tableData["schema"].(map[string]string); ok {
			response.Contract.Fields[dataType] = schema
		}
		if data, ok := tableData["data"]; ok {
			response.Data[dataType] = data
		}
	}

	HandleSuccess(c, response)
}

// MARK: Language Version Info

// GetLanguageVersion returns version information (last modified dates) for all data types of a given language.
//
// @Summary Get version information for a language
// @Description Provides last modified timestamps for each data type of the specified language.
// @Tags Language Data
// @Accept  json
// @Produce  json
// @Param lang path string true "Language code (ISO 639-1)" example(fr)
// @Success 200 {object} models.LanguageVersionResponse "Successfully retrieved language version details"
// @Failure 400 {object} models.ErrorResponse "Invalid language code"
// @Failure 404 {object} models.ErrorResponse "Language not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /api/v1/data-version/{lang} [get]
func GetLanguageVersion(c *gin.Context) {
	lang := c.Param("lang")

	// Validate language code.
	if !validators.IsValidLanguageCode(lang) {
		HandleError(c, http.StatusBadRequest, constants.InvalidLanguageCodeError)
		return
	}

	// Check if language exists in database.
	availableLanguages, err := getAvailableLanguages()
	if err != nil {
		log.Printf("Error checking available languages: %v", err)
		HandleError(c, http.StatusInternalServerError, "Failed to check language availability")
		return
	}

	if !validators.IsLanguageSupported(lang, availableLanguages) {
		HandleError(c, http.StatusNotFound, fmt.Sprintf("Language '%s' not supported", lang))
		return
	}

	// Get version information.
	versions, err := getLanguageVersions(lang)
	if err != nil {
		log.Printf("Error fetching versions for %s: %v", lang, err)
		HandleError(c, http.StatusInternalServerError, constants.ErrorFetchingLanguageVersions)
		return
	}

	response := models.LanguageVersionResponse{
		Language: lang,
		Versions: versions,
	}

	HandleSuccess(c, response)
}

// MARK: Contracts Retrieval

// GetContracts returns the data contract for a language as JSON, or a zip archive of all contracts.
//
// @Summary Retrieve data contracts
// @Description Returns a language's contract as JSON ({lang}.json), or a zip of all contracts (scribe_data_contracts.zip) without 'lang'.
// @Tags Contracts
// @Accept  json
// @Produce  json,application/zip
// @Param lang query string false "Language code (ISO 639-1)" example(es)
// @Success 200 {file} file "The language's contract as JSON, or a zip archive of all contracts"
// @Header 200 {string} Content-Disposition "inline; filename={lang}.json, or attachment; filename=scribe_data_contracts.zip for all languages"
// @Failure 400 {object} models.ErrorResponse "Invalid language code provided"
// @Failure 404 {object} models.ErrorResponse "Language not supported"
// @Failure 500 {object} models.ErrorResponse "Failed to load contract files"
// @Failure 503 {object} models.ErrorResponse "Data contracts are not available"
// @Router /api/v1/contracts [get]
func GetContracts(c *gin.Context) {
	lang := c.Query("lang")
	contractsDir := viper.GetString("contractsDir")

	if lang == "" {
		contracts, err := loadAllContracts(contractsDir)
		if err != nil {
			handleContractsError(c, err)
			return
		}

		var archive bytes.Buffer
		if err := writeContractsZip(&archive, contracts); err != nil {
			handleContractsError(c, fmt.Errorf("could not create contracts archive: %w", err))
			return
		}

		HandleFileDownload(c, contractsArchiveDir+".zip", "application/zip", archive.Bytes())
		return
	}

	if !validators.IsValidLanguageCode(lang) {
		HandleError(c, http.StatusBadRequest, constants.InvalidLanguageCodeError)
		return
	}

	availableLanguages, err := getAvailableLanguages()
	if err != nil {
		log.Printf("Error checking available languages: %v", err)
		HandleError(c, http.StatusInternalServerError, "Failed to check language availability")
		return
	}

	if !validators.IsLanguageSupported(lang, availableLanguages) {
		HandleError(c, http.StatusNotFound, fmt.Sprintf("Language '%s' not supported", lang))
		return
	}

	contract, err := loadContract(contractsDir, lang)
	if err != nil {
		handleContractsError(c, err)
		return
	}

	HandleFileInline(c, lang+".json", "application/json; charset=utf-8", contract.json)
}

// handleContractsError sends a 503 if the contracts are missing and a 500 for other errors.
func handleContractsError(c *gin.Context, err error) {
	log.Printf("Error loading contracts: %v", err)
	if errors.Is(err, fs.ErrNotExist) {
		HandleError(c, http.StatusServiceUnavailable, constants.ContractsNotAvailableError)
		return
	}

	HandleError(c, http.StatusInternalServerError, "Failed to load contracts")
}

// MARK: Translation Data Retrieval

// getTranslationTableData fetches translation data. It is a package-level variable to allow mocking in unit tests.
var getTranslationTableData = dbqueries.GetTranslationTableData

// GetTranslationData returns translation data from a source language into a target language.
//
// @Summary Retrieve translation data
// @Description Returns nested translation data for the given target and source language ISO codes.
// @Tags Translations
// @Accept  json
// @Produce  json
// @Param source_lang query string true "Source language code (ISO 639-1)" example(de)
// @Param target_lang query string true "Target language code (ISO 639-1)" example(bn)
// @Success 200 {object} models.TranslationDataResponse "Successfully retrieved translation data"
// @Failure 400 {object} models.ErrorResponse "Invalid language code"
// @Failure 404 {object} models.ErrorResponse "Translation data not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /api/v1/translations [get]
func GetTranslationData(c *gin.Context) {
	sourceLang := c.Query("source_lang")
	targetLang := c.Query("target_lang")

	if sourceLang == "" || targetLang == "" {
		HandleError(c, http.StatusBadRequest, constants.EmptyTranslationCodeError)
		return
	}

	if !validators.IsValidTranslationLangCode(targetLang) || !validators.IsValidTranslationLangCode(sourceLang) {
		HandleError(c, http.StatusBadRequest, constants.InvalidTranslationLangCodeError)
		return
	}

	data, err := getTranslationTableData(targetLang, sourceLang)
	if err != nil {
		log.Printf("Error fetching translation data for %s/%s: %v", targetLang, sourceLang, err)
		if strings.Contains(err.Error(), "does not exist") {
			HandleError(c, http.StatusNotFound, fmt.Sprintf("No translation data for '%s' from '%s'", targetLang, sourceLang))
			return
		}
		HandleError(c, http.StatusInternalServerError, constants.ErrorFetchingTranslationData)
		return
	}

	HandleSuccess(c, models.TranslationDataResponse{
		TargetLang: targetLang,
		SourceLang: sourceLang,
		Data:       data,
	})
}
