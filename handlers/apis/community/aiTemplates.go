package community

// This package contains handlers for retrieving AI community templates. Right now it only works with
// Azure Open AI Templates. This package requires authentication becuase its used within the
// application.

// The handlers are:
// 1. GetPublicUserTemplates - Retrieves all the templates that have been published by a user
// 2. GetPublicUserTemplate - Retrieves a single template data that has been published by a user

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"sententiawebapi/handlers/apis/tenantManagement"
	"sententiawebapi/handlers/models"
	"sententiawebapi/utilities"

	"github.com/gin-gonic/gin"
)

func GetPublicPromptTemplates(c *gin.Context) {
	// Validate user identity
	_, _, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	category := c.Query("category")

	query := `
		SELECT
			t.id,
			t.user_id,
			t.title,
			t.description,
			t.category,
			t.ai_vendor,
			t.ai_model,
			t.configuration,
			t.published_by,
			t.created_at,
			t.updated_at,
			u.first_name,
			u.last_name,
			u.user_picture
		FROM
			st_schema.community_prompt_config_templates t
		JOIN
			st_schema.users u ON t.user_id = u.id
	`

	if category != "" {
		query += " WHERE category = $1"
	}

	var rows *sql.Rows
	var err error

	if category != "" {
		rows, err = tenantManagement.DB.Query(query, category)
	} else {
		rows, err = tenantManagement.DB.Query(query)
	}

	if err != nil {
		log.Printf("Database Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database preparation error"})
		return
	}

	defer rows.Close()

	var publicTemplates []models.TenantAiTemplate
	for rows.Next() {
		var template models.TenantAiTemplate
		var rawConfig []byte // To store raw JSON configuration

		err := rows.Scan(
			&template.ID, &template.UserID, &template.Title, &template.Description,
			&template.Category, &template.AiVendor, &template.AiModel, &rawConfig,
			&template.PublishedBy, &template.CreatedAt, &template.UpdatedAt,
			&template.FirstName, &template.LastName, &template.UserPicture,
		)
		if err != nil {
			log.Printf("Database Err: %v", err)
			continue
		}

		var config models.AiConfiguration
		if err := json.Unmarshal(rawConfig, &config); err != nil {
			log.Printf("JSON Unmarshal Err: %v", err)
			continue
		}
		template.Configuration = &config

		publicTemplates = append(publicTemplates, template)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Database Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error during row iteration"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    publicTemplates,
		"message": models.StatusSuccess,
	})
}

// @Summary Retrieve a single public user template
// @Description Retrieves a specific template published by a user and available publicly.
// @Tags AI Community User Templates
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer [Token]"
// @Param id query string true "Public Template ID"
// @Success 200 {object} TenantAiTemplate "Public template retrieved successfully"
// @Failure 401 {object} map[string]string "Authorization required"
// @Failure 404 {object} map[string]string "Template not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/ppublicTemplate [get]
func GetPublicPromptTemplate(c *gin.Context) {
	templateID := c.Query("id")
	if templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template ID required"})
		return
	}

	stmt, err := tenantManagement.DB.Prepare(`
        SELECT
            id, user_id, title, description, category, ai_vendor, ai_model, configuration,
            published_by, created_at, updated_at
        FROM
            st_schema.community_prompt_config_templates
        WHERE
            id = $1
    `)
	if err != nil {
		log.Printf("Database Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database preparation error"})
		return
	}
	defer stmt.Close()

	var template models.TenantAiTemplate
	var rawConfig []byte // To store raw JSON configuration

	err = stmt.QueryRow(templateID).Scan(
		&template.ID, &template.UserID, &template.Title, &template.Description,
		&template.Category, &template.AiVendor, &template.AiModel, &rawConfig,
		&template.PublishedBy, &template.CreatedAt, &template.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		} else {
			log.Printf("Database Err: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query error"})
		}
		return
	}

	var config models.AiConfiguration
	if err := json.Unmarshal(rawConfig, &config); err != nil {
		log.Printf("JSON Unmarshal Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "JSON unmarshalling error"})
		return
	}
	template.Configuration = &config

	c.JSON(http.StatusOK, gin.H{
		"data":    template,
		"message": models.StatusSuccess,
	})
}

func PublishTenantAiPromptTemplate(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": models.UserIdError})
		return
	}

	templateId := c.Query("id")
	if templateId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template ID is required"})
		return
	}

	// Begin a new transaction
	tx, err := tenantManagement.DB.Begin()
	if err != nil {
		log.Printf("Database Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer tx.Rollback()

	// Fetch template data
	var templateData models.TenantAiTemplate
	var rawConfig []byte
	err = tx.QueryRow(`
        SELECT id, user_id, tenant_id, title, description, category, ai_vendor, ai_model, configuration, privacy
        FROM st_schema.prompt_config_template
        WHERE id = $1 AND tenant_id = $2`, templateId, tenantID).Scan(
		&templateData.ID, &templateData.UserID, &templateData.TenantID, &templateData.Title, &templateData.Description,
		&templateData.Category, &templateData.AiVendor, &templateData.AiModel, &rawConfig, &templateData.Privacy,
	)
	if err != nil {
		log.Printf("Database Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch template data"})
		return
	}

	// Unmarshal and update configuration
	var config models.AiConfiguration
	if err := json.Unmarshal(rawConfig, &config); err != nil {
		log.Printf("JSON Unmarshal Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unmarshalling configuration"})
		return
	}
	templateData.Configuration = &config

	// Convert the updated configuration back to JSON
	updatedConfigJSON, err := json.Marshal(templateData.Configuration)
	if err != nil {
		log.Printf("JSON Marshal Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error marshaling configuration to JSON"})
		return
	}

	// Update original template's privacy to false in prompt_config_template
	_, err = tx.Exec(`
        UPDATE st_schema.prompt_config_template
        SET privacy = FALSE
        WHERE id = $1 AND tenant_id = $2`,
		templateId, tenantID)
	if err != nil {
		log.Printf("Database Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update original template"})
		tx.Rollback()
		return
	}

	// Update templateData's privacy to reflect the change
	*templateData.Privacy = false

	// Insert or update the template in community_prompt_config_templates
	_, err = tx.Exec(`
        INSERT INTO st_schema.community_prompt_config_templates
            (id, user_id, tenant_id, title, description, category, ai_vendor, ai_model, configuration, published_by)
        VALUES
            ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'communityTemplate')
        ON CONFLICT (id) DO UPDATE SET
            user_id = EXCLUDED.user_id,
            tenant_id = EXCLUDED.tenant_id,
            title = EXCLUDED.title,
            description = EXCLUDED.description,
            category = EXCLUDED.category,
            ai_vendor = EXCLUDED.ai_vendor,
            ai_model = EXCLUDED.ai_model,
            configuration = EXCLUDED.configuration,
            published_by = EXCLUDED.published_by
    `, templateData.ID, templateData.UserID, templateData.TenantID, templateData.Title, templateData.Description, templateData.Category,
		templateData.AiVendor, templateData.AiModel, updatedConfigJSON)
	if err != nil {
		log.Printf("Database Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process community template"})
		tx.Rollback()
		return
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Database Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Template published successfully",
		"data":    templateData,
	})
}

func UnpublishTenantAiPromptTemplate(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	templateId := c.Query("id")
	if templateId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template ID is required"})
		return
	}

	// Begin a new transaction
	tx, err := tenantManagement.DB.Begin()
	if err != nil {
		log.Printf("Database Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer tx.Rollback()

	// Remove the template from community_prompt_config_templates
	_, err = tx.Exec(`DELETE FROM st_schema.community_prompt_config_templates WHERE id = $1 AND tenant_id = $2`, templateId, tenantID)
	if err != nil {
		log.Printf("Database Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove from community templates"})
		return
	}

	// Update and fetch the private template data
	var unpublishedTemplate models.TenantAiTemplate
	var rawConfig []byte
	err = tx.QueryRow(`
		UPDATE st_schema.prompt_config_template
		SET published_by = 'userTemplate', privacy = TRUE
		WHERE id = $1 AND tenant_id = $2
		RETURNING id, title, description, category, ai_vendor, ai_model, configuration,
			privacy, original_publisher, published_by, created_at, updated_at`,
		templateId, tenantID).Scan(
		&unpublishedTemplate.ID, &unpublishedTemplate.Title, &unpublishedTemplate.Description,
		&unpublishedTemplate.Category, &unpublishedTemplate.AiVendor, &unpublishedTemplate.AiModel,
		&rawConfig, &unpublishedTemplate.Privacy,
		&unpublishedTemplate.OriginalPublisher, &unpublishedTemplate.PublishedBy,
		&unpublishedTemplate.CreatedAt, &unpublishedTemplate.UpdatedAt,
	)
	if err != nil {
		log.Printf("Database Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update and fetch template data"})
		return
	}

	// Unmarshal the raw JSON configuration into the AzOAConfiguration struct
	var config models.AiConfiguration
	if err := json.Unmarshal(rawConfig, &config); err != nil {
		log.Printf("JSON Unmarshal Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "JSON unmarshalling error"})
		return
	}
	unpublishedTemplate.Configuration = &config

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Database Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Template unpublished successfully",
		"data":    unpublishedTemplate,
	})
}

// ClonePublicAiPromptTemplate clones a public AI template into the user's private template repository.
// It essentially creates a new template in the user's private repository with the same data as the public template by creating
// a copy with new ID.
func ClonePublicAiPromptTemplate(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	// Get public template ID from query parameters
	sourceTemplateId := c.Query("id")
	if sourceTemplateId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Public template ID is required"})
		return
	}

	// Fetch the public template and retrieve the configuration as []byte (JSONB)
	var template models.TenantAiTemplate
	var rawConfig []byte
	err := tenantManagement.DB.QueryRow(`
		SELECT id, user_id, title, description, category, ai_vendor, ai_model, configuration
		FROM st_schema.community_prompt_config_templates
		WHERE id = $1`,
		sourceTemplateId).Scan(
		&template.ID,
		&template.UserID,
		&template.Title,
		&template.Description,
		&template.Category,
		&template.AiVendor,
		&template.AiModel,
		&rawConfig,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Public template not found"})
		} else {
			log.Printf("Database Err: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch public template"})
		}
		return
	}

	// Unmarshal the raw JSON configuration into the AzOAConfiguration struct
	var config models.AiConfiguration
	if err := json.Unmarshal(rawConfig, &config); err != nil {
		log.Printf("JSON Unmarshal Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unmarshalling configuration"})
		return
	}

	// Insert and get the cloned template
	var newTemplateResource models.TenantAiTemplate
	var newTemplateResourceRawConfig []byte
	err = tenantManagement.DB.QueryRow(`
		INSERT INTO st_schema.prompt_config_template (
			source_id,
			user_id,
			tenant_id,
			title,
			description,
			category,
			ai_vendor,
			ai_model,
			configuration,
			privacy,
			original_publisher,
			published_by,
			created_at,
			updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12,
			NOW(), NOW()
		)
		RETURNING
			id,
			source_id,
			user_id,
			tenant_id,
			title,
			description,
			category,
			ai_vendor,
			ai_model,
			configuration,
			privacy,
			original_publisher,
			published_by,
			created_at,
			updated_at`,
		// Parameters
		sourceTemplateId,
		userID,
		tenantID,
		template.Title,
		template.Description,
		template.Category,
		template.AiVendor,
		template.AiModel,
		rawConfig,
		true,
		template.UserID,
		"Community",
	).Scan(
		// Scan results into struct fields
		&newTemplateResource.ID,
		&newTemplateResource.SourceID,
		&newTemplateResource.UserID,
		&newTemplateResource.TenantID,
		&newTemplateResource.Title,
		&newTemplateResource.Description,
		&newTemplateResource.Category,
		&newTemplateResource.AiVendor,
		&newTemplateResource.AiModel,
		&newTemplateResourceRawConfig,
		&newTemplateResource.Privacy,
		&newTemplateResource.OriginalPublisher,
		&newTemplateResource.PublishedBy,
		&newTemplateResource.CreatedAt,
		&newTemplateResource.UpdatedAt,
	)
	if err != nil {
		log.Printf("Database Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clone template"})
		return
	}

	// Unmarshal the configuration for the response
	var newTemplateConfig models.AiConfiguration
	if err := json.Unmarshal(newTemplateResourceRawConfig, &newTemplateConfig); err != nil {
		log.Printf("JSON Unmarshal Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unmarshalling cloned configuration"})
		return
	}
	newTemplateResource.Configuration = &newTemplateConfig

	c.JSON(http.StatusOK, gin.H{
		"message": "Template cloned successfully",
		"data":    newTemplateResource,
	})
}
