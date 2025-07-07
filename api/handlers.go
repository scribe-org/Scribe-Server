// SPDX-License-Identifier: GPL-3.0-or-later
package api

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/scribe-org/scribe-server/database"
	"github.com/scribe-org/scribe-server/internal/constants"
	"github.com/scribe-org/scribe-server/models"
)

func hello(c *gin.Context) {
	c.String(http.StatusOK, "Hello, I'm Scribe!")
}

// isValidLanguageCode validates ISO 639-1 language codes and we support it
// that is, we know that data exists in our db
// for now, we know that EN(English) and FR(French) exists
func isValidLanguageCode(lang string) bool {
	allowedLanguages := map[string]bool{
		"en": true,
		"fr": true,
	}

	// Basic validation: exactly 2 lowercase letters
	if len(lang) != 2 {
		return false
	}

	return allowedLanguages[lang]
}

// sanitizeLanguageCode ensures language code is safe for table name construction
func sanitizeLanguageCode(lang string) string {
	if !isValidLanguageCode(lang) {
		return ""
	}
	// Convert to uppercase for table name prefix
	return strings.ToUpper(lang)
}

// getAvailableLanguages handles GET /languages
func getAvailableLanguages(c *gin.Context) {
	languages, err := database.GetAvailableLanguages()
	if err != nil {
		log.Printf("Error fetching available languages: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: constants.ErrorFetchingLanguages,
		})
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

	c.JSON(http.StatusOK, models.AvailableLanguagesResponse{
		Languages: languageInfos,
	})
}

// getLanguageData handles GET /data/:lang
func getLanguageData(c *gin.Context) {
	lang := c.Param("lang")

	// Validate language code
	if !isValidLanguageCode(lang) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: constants.InvalidLanguageCodeError,
		})
		return
	}

	// Check if language exists in database
	availableLanguages, err := database.GetAvailableLanguages()
	if err != nil {
		log.Printf("Error checking available languages: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to check language availability",
		})
		return
	}

	// can we improve this?
	languageExists := false
	for _, availableLang := range availableLanguages {
		if availableLang == lang {
			languageExists = true
			break
		}
	}

	if !languageExists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: fmt.Sprintf("Language '%s' not supported", lang),
		})
		return
	}

	// Get data types for the language
	dataTypes, err := database.GetLanguageDataTypes(lang)
	if err != nil {
		log.Printf("Error fetching data types for %s: %v", lang, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to fetch language data types",
		})
		return
	}

	// Build the response
	response := models.LanguageDataResponse{
		Language: lang,
		Contract: models.Contract{
			Version:   constants.APIVersion,
			UpdatedAt: time.Now().Format(constants.DateFormat),
			Fields:    make(map[string]map[string]string),
		},
		Data: make(map[string]interface{}),
	}

	// For each data type, get schema and data
	for _, dataType := range dataTypes {
		// Sanitize inputs before constructing table name
		sanitizedLang := sanitizeLanguageCode(lang)
		if sanitizedLang == "" {
			log.Printf("Failed to sanitize language code: %s", lang)
			continue
		}

		tableName := fmt.Sprintf("%sLanguageData_%s", strings.ToUpper(lang), dataType)

		// Get schema
		schema, err := database.GetTableSchema(tableName)
		if err != nil {
			log.Printf("Error fetching schema for %s: %v", tableName, err)
			continue
		}

		// Get data
		data, err := database.GetTableData(tableName)
		if err != nil {
			log.Printf("Error fetching data for %s: %v", tableName, err)
			continue
		}

		// Add to response
		response.Contract.Fields[dataType] = schema
		response.Data[dataType] = data
	}

	c.JSON(http.StatusOK, response)
}

// getLanguageVersion handles GET /data-version/:lang
func getLanguageVersion(c *gin.Context) {
	lang := c.Param("lang")

	// Validate language code
	if !isValidLanguageCode(lang) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: constants.InvalidLanguageCodeError,
		})
		return
	}

	// Check if language exists in database
	availableLanguages, err := database.GetAvailableLanguages()
	if err != nil {
		log.Printf("Error checking available languages: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to check language availability",
		})
		return
	}

	languageExists := false
	for _, availableLang := range availableLanguages {
		if availableLang == lang {
			languageExists = true
			break
		}
	}

	if !languageExists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: fmt.Sprintf("Language '%s' not supported", lang),
		})
		return
	}

	// Get version information
	versions, err := database.GetLanguageVersions(lang)
	if err != nil {
		log.Printf("Error fetching versions for %s: %v", lang, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: constants.ErrorFetchingLanguageVersions,
		})
		return
	}

	response := models.LanguageVersionResponse{
		Language: lang,
		Versions: versions,
	}

	c.JSON(http.StatusOK, response)
}