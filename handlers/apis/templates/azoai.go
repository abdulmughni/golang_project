package templates

// This package contains the handlers for creating tenant AI configurations using Azure Open AI.
// The handlers are:
// 1. NewTenantAiTemplate - Creates a new AI template for a tenant
// 2. GetTenantAiTemplate - Retrieves a specific AI template
// 3. GetTenantAiTemplates - Retrieves all AI templates for a tenant
// 4. UpdateTenantAiTemplate - Updates a specific AI template
// 5. DeleteTenantAiTemplate - Deletes a specific AI template

// Local functions:
// 1. isDevelopmentEnvironment - Checks if the current environment is development or not.

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"sententiawebapi/handlers/apis/tenantManagement"
	"sententiawebapi/handlers/models"
	"sententiawebapi/utilities"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func NewTenantAiTemplate(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	var template models.TenantAiTemplate

	// Load all of the request body data first
	if err := c.ShouldBindJSON(&template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Override/set critical fields after JSON binding
	template.UserID = &userID
	template.TenantID = &tenantID
	template.Privacy = new(bool)
	*template.Privacy = true // Setting default privacy to true

	// Convert the Configuration struct to JSON
	configJSON, err := json.Marshal(template.Configuration)
	if err != nil {
		if isDevelopmentEnvironment() {
			log.Printf("JSON Marshal Err: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error marshaling configuration to JSON"})
		return
	}

	var newID string
	err = tenantManagement.DB.QueryRow(`
        INSERT INTO st_schema.prompt_config_template
        (
            user_id, tenant_id, title, description, category, ai_vendor, ai_model,
            configuration, privacy, original_publisher, published_by,
            created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
        ) RETURNING id
    `,
		template.UserID,
		template.TenantID,
		template.Title,
		template.Description,
		template.Category,
		template.AiVendor,
		template.AiModel,
		configJSON,
		template.Privacy,
		template.OriginalPublisher,
		template.PublishedBy,
		time.Now(),
		time.Now(),
	).Scan(&newID)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		} else {
			if isDevelopmentEnvironment() {
				log.Printf("Database Err: %v", err)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	template.ID = &newID

	c.JSON(http.StatusCreated, gin.H{
		"data":    template,
		"message": models.StatusCreated,
	})
}

func GetTenantAiTemplate(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	templateID := c.Query("id") // Adjust based on your URL parameter name
	if templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template ID required"})
		return
	}

	var template models.TenantAiTemplate
	var rawConfig []byte

	err := tenantManagement.DB.QueryRow(`
        SELECT
            id, source_id, user_id, tenant_id, title, description, category,
            ai_vendor, ai_model, configuration, privacy, original_publisher,
            published_by, created_at, updated_at
        FROM
            st_schema.prompt_config_template
        WHERE
            id = $1 AND tenant_id = $2
    `, templateID, tenantID).Scan(
		&template.ID, &template.SourceID, &template.UserID, &template.TenantID,
		&template.Title, &template.Description, &template.Category,
		&template.AiVendor, &template.AiModel, &rawConfig,
		&template.Privacy, &template.OriginalPublisher,
		&template.PublishedBy, &template.CreatedAt, &template.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		} else {
			if isDevelopmentEnvironment() {
				log.Printf("Database Err: %v", err)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	// Unmarshal JSON configuration
	var config models.AiConfiguration
	if err := json.Unmarshal(rawConfig, &config); err != nil {
		if isDevelopmentEnvironment() {
			log.Printf("JSON Unmarshal Err: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unmarshalling configuration"})
		return
	}
	template.Configuration = &config // Assign the unmarshalled configuration

	c.JSON(http.StatusOK, gin.H{
		"data":    template,
		"message": models.StatusSuccess,
	})
}

// GetTenantAiTemplates retrieves all AI templates for a tenant
// @Summary Get all AI templates
// @Description Retrieves all the user's configuration templates in an array.
// @Tags Tenant AI Templates
// @Accept json
// @Produce json
// @Success 200 {array} TenantAiTemplate "Successfully retrieved all AI Templates"
// @Failure 500 {object} string "Internal Server Error"
// @Router /api/tenantAiTemplates [get]
func GetTenantAiTemplates(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	rows, err := tenantManagement.DB.Query(`
       SELECT
			p.id, p.source_id, p.user_id, p.tenant_id, p.title, p.description, p.category,
			p.ai_vendor, p.ai_model, p.configuration, p.privacy, p.original_publisher,
			p.published_by, p.created_at, p.updated_at,
			u.first_name, u.last_name, u.user_picture
		FROM
			st_schema.prompt_config_template p
		JOIN
			st_schema.users u ON p.user_id = u.id
		WHERE
			p.tenant_id = $1
    `, tenantID)

	if err != nil {
		if isDevelopmentEnvironment() {
			log.Printf("Database Err: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	defer rows.Close()

	var templates []models.TenantAiTemplate
	for rows.Next() {
		var template models.TenantAiTemplate
		var rawConfig []byte // To store raw JSON configuration

		err := rows.Scan(
			&template.ID, &template.SourceID, &template.UserID, &template.TenantID,
			&template.Title, &template.Description, &template.Category,
			&template.AiVendor, &template.AiModel, &rawConfig,
			&template.Privacy, &template.OriginalPublisher,
			&template.PublishedBy, &template.CreatedAt, &template.UpdatedAt,
			&template.FirstName, &template.LastName, &template.UserPicture,
		)

		if err != nil {
			if isDevelopmentEnvironment() {
				log.Printf("Database Err: %v", err)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			continue
		}

		// Unmarshal JSON configuration
		var config models.AiConfiguration
		if err := json.Unmarshal(rawConfig, &config); err != nil {
			if isDevelopmentEnvironment() {
				log.Printf("JSON Unmarshal Err: %v", err)
			}
			continue // Or handle the error as needed
		}
		template.Configuration = &config // Assign the unmarshalled configuration

		templates = append(templates, template)
	}

	if err := rows.Err(); err != nil {
		if isDevelopmentEnvironment() {
			log.Printf("Database Err: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    templates,
		"message": models.StatusSuccess,
	})
}

// UpdateTenantAiTemplate updates a specific AI template
// @Summary Update an AI template
// @Description Allows users to update an existing configuration template.
// @Tags Tenant AI Templates
// @Accept json
// @Produce json
// @Param id query string true "Template ID"
// @Param tenantAiTemplate body TenantAiTemplate true "AI Template Data"
// @Success 200 {object} TenantAiTemplate "Successfully updated AI Template"
// @Failure 400 {object} string "Bad Request"
// @Failure 500 {object} string "Internal Server Error"
// @Router /api/tenantAiTemplate [put]
func UpdateTenantAiTemplate(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	templateId := c.Query("id") // Retrieve the template ID from URL parameters
	if templateId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template ID is required"})
		return
	}

	var updatedTemplate models.TenantAiTemplate
	if err := c.ShouldBindJSON(&updatedTemplate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set update config
	setParts := []string{}
	args := []interface{}{}
	argCounter := 1

	// Check each field and add to setParts if it's not nil
	if updatedTemplate.Title != nil {
		setParts = append(setParts, fmt.Sprintf("title = $%d", argCounter))
		args = append(args, *updatedTemplate.Title)
		argCounter++
	}
	if updatedTemplate.Description != nil {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argCounter))
		args = append(args, *updatedTemplate.Description)
		argCounter++
	}
	if updatedTemplate.Category != nil {
		setParts = append(setParts, fmt.Sprintf("category = $%d", argCounter))
		args = append(args, *updatedTemplate.Category)
		argCounter++
	}
	if updatedTemplate.AiVendor != nil {
		setParts = append(setParts, fmt.Sprintf("ai_vendor = $%d", argCounter))
		args = append(args, *updatedTemplate.AiVendor)
		argCounter++
	}
	if updatedTemplate.AiModel != nil {
		setParts = append(setParts, fmt.Sprintf("ai_model = $%d", argCounter))
		args = append(args, *updatedTemplate.AiModel)
		argCounter++
	}

	// Convert the Configuration struct to JSON
	updatedConfig, err := json.Marshal(updatedTemplate.Configuration)
	if err != nil {
		if isDevelopmentEnvironment() {
			log.Printf("JSON Marshal Err: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing configuration"})
		return
	}

	setParts = append(setParts, fmt.Sprintf("configuration = $%d", argCounter))
	args = append(args, updatedConfig)
	argCounter++

	// Check if any fields were set for update
	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No updatable fields provided"})
		return
	}

	// Construct the SQL query
	setClause := strings.Join(setParts, ", ")
	query := fmt.Sprintf(`
        UPDATE st_schema.prompt_config_template
        SET %s, updated_at = NOW()
        WHERE id = $%d AND tenant_id = $%d
        RETURNING id, source_id, user_id, tenant_id, title, description, category,
                  ai_vendor, ai_model, configuration, privacy, original_publisher,
                  published_by, created_at, updated_at
    `, setClause, argCounter, argCounter+1)
	args = append(args, templateId, tenantID)

	var returnTemplate models.TenantAiTemplate
	var rawConfig []byte
	err = tenantManagement.DB.QueryRow(query, args...).Scan(
		&returnTemplate.ID,
		&returnTemplate.SourceID,
		&returnTemplate.UserID,
		&returnTemplate.TenantID,
		&returnTemplate.Title,
		&returnTemplate.Description,
		&returnTemplate.Category,
		&returnTemplate.AiVendor,
		&returnTemplate.AiModel,
		&rawConfig,
		&returnTemplate.Privacy,
		&returnTemplate.OriginalPublisher,
		&returnTemplate.PublishedBy,
		&returnTemplate.CreatedAt,
		&returnTemplate.UpdatedAt,
	)

	if err != nil {
		if isDevelopmentEnvironment() {
			log.Printf("Database Err: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Unmarshal the configuration JSON
	if err := json.Unmarshal(rawConfig, &returnTemplate.Configuration); err != nil {
		log.Printf("JSON Unmarshal Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing configuration JSON"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    returnTemplate,
		"message": models.StatusSuccess,
	})
}

func DeleteTenantAiTemplate(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	templateId := c.Query("id") // Retrieve the template ID from URL parameters
	if templateId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template ID is required"})
		return
	}

	var deletedTemplate models.TenantAiTemplate
	var rawConfig []byte

	err := tenantManagement.DB.QueryRow(`
        DELETE FROM st_schema.prompt_config_template
        WHERE id = $1 AND tenant_id = $2
        RETURNING id, source_id, user_id, tenant_id, title, description, category,
                  ai_vendor, ai_model, configuration, privacy, original_publisher,
                  published_by, created_at, updated_at
    `, templateId, tenantID).Scan(
		&deletedTemplate.ID, &deletedTemplate.SourceID,
		&deletedTemplate.UserID, &deletedTemplate.TenantID, &deletedTemplate.Title, &deletedTemplate.Description,
		&deletedTemplate.Category, &deletedTemplate.AiVendor, &deletedTemplate.AiModel,
		&rawConfig, // Scan the raw JSON data
		&deletedTemplate.Privacy, &deletedTemplate.OriginalPublisher,
		&deletedTemplate.PublishedBy, &deletedTemplate.CreatedAt, &deletedTemplate.UpdatedAt,
	)
	if err != nil {
		log.Printf("Database Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete template"})
		return
	}

	// Unmarshal JSON configuration
	var config models.AiConfiguration
	if err := json.Unmarshal(rawConfig, &config); err != nil {
		log.Printf("JSON Unmarshal Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing configuration"})
		return
	}
	deletedTemplate.Configuration = &config

	c.JSON(http.StatusOK, gin.H{
		"data":    deletedTemplate,
		"message": "Template successfully deleted",
	})
}

func isDevelopmentEnvironment() bool {
	env := os.Getenv("ENVIRONMENT")
	return env == "dev" || env == "local"
}
