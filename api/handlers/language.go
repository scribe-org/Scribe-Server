// SPDX-License-Identifier: GPL-3.0-or-later

// Package handlers contains HTTP handlers for various API endpoints.
package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/scribe-org/scribe-server/api/dbqueries"
	"github.com/scribe-org/scribe-server/api/validators"
	"github.com/scribe-org/scribe-server/database"
	"github.com/scribe-org/scribe-server/internal/constants"
	"github.com/scribe-org/scribe-server/models"
	"github.com/spf13/viper"
)

// GetAvailableLanguages handles GET /languages.
// @Summary Get available languages.
// @Description Returns a list of all supported languages with their available data types.
// @Tags languages
// @Accept json
// @Produce json
// @Success 200 {object} models.AvailableLanguagesResponse "List of supported languages".
// @Failure 500 {object} models.ErrorResponse "Internal server error".
// @Router /api/v1/languages [get]
func GetAvailableLanguages(c *gin.Context) {
	languages, err := database.GetAvailableLanguages()
	if err != nil {
		log.Printf("Error fetching available languages: %v", err)
		HandleError(c, http.StatusInternalServerError, constants.ErrorFetchingLanguages)
		return
	}

	var languageInfos []models.LanguageInfo
	for _, lang := range languages {
		dataTypes, err := database.GetLanguageDataTypes(lang)
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

// GetLanguageData handles GET /data/:lang.
// @Summary Get language data.
// @Description Returns all available language data and schema contract for a specific language.
// @Tags language-data
// @Accept json
// @Produce json
// @Param lang path string true "Language code (ISO 639-1)" example(en).
// @Success 200 {object} models.LanguageDataResponse "Complete language data with schema".
// @Failure 400 {object} models.ErrorResponse "Invalid language code".
// @Failure 404 {object} models.ErrorResponse "Language not supported".
// @Failure 500 {object} models.ErrorResponse "Internal server error".
// @Router /api/v1/data/{lang} [get]
func GetLanguageData(c *gin.Context) {
	lang := c.Param("lang")

	// Validate language code.
	if !validators.IsValidLanguageCode(lang) {
		HandleError(c, http.StatusBadRequest, constants.InvalidLanguageCodeError)
		return
	}

	// Check if language exists in database.
	availableLanguages, err := database.GetAvailableLanguages()
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
	dataTypes, err := database.GetLanguageDataTypes(lang)
	if err != nil {
		log.Printf("Error fetching data types for %s: %v", lang, err)
		HandleError(c, http.StatusInternalServerError, "Failed to fetch language data types")
		return
	}

	// Build the response.
	response := models.LanguageDataResponse{
		Language: lang,
		Contract: models.Contract{
			Version:   constants.APIVersion,
			UpdatedAt: time.Now().Format(constants.DateFormat),
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

		tableData, err := dbqueries.GetLanguageTableData(lang, dataType)
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

// GetLanguageVersion handles GET /data-version/:lang.
// @Summary Get language data versions.
// @Description Returns last modified dates for each data type of a specific language.
// @Tags language-data
// @Accept json
// @Produce json
// @Param lang path string true "Language code (ISO 639-1)" example(en).
// @Success 200 {object} models.LanguageVersionResponse "Language version information".
// @Failure 400 {object} models.ErrorResponse "Invalid language code"
// @Failure 404 {object} models.ErrorResponse "Language not supported"
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
	availableLanguages, err := database.GetAvailableLanguages()
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
	versions, err := database.GetLanguageVersions(lang)
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

// GetContracts handles GET /contracts.
// @Summary Get contracts.
// @Description Returns schema contracts for all languages or a specific language if 'lang' query parameter is provided.
// @Tags contracts
// @Accept json
// @Produce json
// @Param lang query string false "Language code (ISO 639-1)" example(en).
// @Success 200 {object} models.ContractsResponse "Schema contracts".
// @Failure 400 {object} models.ErrorResponse "Invalid language code".
// @Failure 404 {object} models.ErrorResponse "Language not supported".
// @Failure 500 {object} models.ErrorResponse "Internal server error".
// @Router /api/v1/contracts [get]
func GetContracts(c *gin.Context) {
	lang := c.Query("lang")
	contractsDir := viper.GetString("contractsDir")

	if lang != "" {
		if !validators.IsValidLanguageCode(lang) {
			HandleError(c, http.StatusBadRequest, constants.InvalidLanguageCodeError)
			return
		}

		availableLanguages, err := database.GetAvailableLanguages()
		if err != nil {
			log.Printf("Error checking available languages: %v", err)
			HandleError(c, http.StatusInternalServerError, "Failed to check language availability")
			return
		}

		if !validators.IsLanguageSupported(lang, availableLanguages) {
			HandleError(c, http.StatusNotFound, fmt.Sprintf("Language '%s' not supported", lang))
			return
		}
	}

	var contracts map[string]any
	var err error

	if lang != "" {
		// Load a single language
		contracts, err = loadSingleContract(contractsDir, lang)
	} else {
		// Load all languages
		contracts, err = loadAllContracts(contractsDir)
	}

	if err != nil {
		log.Printf("Error loading contracts: %v", err)
		HandleError(c, http.StatusInternalServerError, "Failed to load contracts")
		return
	}

	HandleSuccess(c, models.ContractsResponse{
		Contracts: contracts,
	})
}

// loadSingleContract reads and unmarshals a single contract file.
func loadSingleContract(contractsDir, lang string) (map[string]any, error) {
	filePath := filepath.Join(contractsDir, lang+".json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not read contract file for %s: %w", lang, err)
	}

	var contract any
	if err := json.Unmarshal(data, &contract); err != nil {
		return nil, fmt.Errorf("could not unmarshal contract for %s: %w", lang, err)
	}

	return map[string]any{lang: contract}, nil
}

// loadAllContracts reads and unmarshals all .json files in a directory.
func loadAllContracts(contractsDir string) (map[string]any, error) {
	contracts := make(map[string]any)

	files, err := os.ReadDir(contractsDir)
	if err != nil {
		return nil, fmt.Errorf("could not read contracts directory: %w", err)
	}

	for _, file := range files {
		// Skip directories and non-json files
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		langCode := strings.TrimSuffix(file.Name(), ".json")
		filePath := filepath.Join(contractsDir, file.Name())

		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Warning: could not read contract file %s: %v", file.Name(), err)
			continue
		}

		var contract any
		if err := json.Unmarshal(data, &contract); err != nil {
			log.Printf("Warning: could not unmarshal contract file %s: %v", file.Name(), err)
			continue
		}

		contracts[langCode] = contract
	}

	return contracts, nil
}
