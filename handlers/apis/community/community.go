package community

// This package contains the handlers for rerieving Solution Pilot published AI Prompt Configs ( So the AI templates that we expose to the public)
// Configurations stored in the 'sp_prompt_config_templates' table. This handler is used by the website

// The handlers are:
// 1. GetSpAiTemplates - Returns all Solution Pilot published AI Prompt Configurations.
// 2. GetSpAiTemplate - Returns a single Solution Pilot published AI Prompt Configuration by ID.

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"sententiawebapi/handlers/apis/tenantManagement"
	"sententiawebapi/handlers/models"

	"github.com/gin-gonic/gin"
)

var PDB *sql.DB // This will be initialized from main.go

func GetSpAiTemplatesPub(c *gin.Context) {
	templateData := `
        SELECT
            id, title, description, category, ai_vendor, ai_model, configuration,
            published_by, created_at, updated_at
        FROM
            st_schema.sp_prompt_config_templates;
    `

	stmt, err := PDB.Prepare(templateData)
	if err != nil {
		log.Printf("Database Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error"})
		return
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		log.Printf("Database Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error"})
		return
	}
	defer rows.Close()

	var templates []models.SpAiTemplate
	for rows.Next() {
		var template models.SpAiTemplate
		var rawConfig []byte

		err := rows.Scan(
			&template.ID, &template.Title, &template.Description, &template.Category,
			&template.AiVendor, &template.AiModel, &rawConfig,
			&template.PublishedBy, &template.CreatedAt, &template.UpdatedAt,
		)
		if err != nil {
			log.Printf("Database Error: %v", err)
			continue
		}

		var config models.AiConfiguration
		if err := json.Unmarshal(rawConfig, &config); err != nil {
			log.Printf("JSON Unmarshal Err: %v", err)
			continue
		}
		template.Configuration = &config

		templates = append(templates, template)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Database Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error after row iteration"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    templates,
		"message": models.StatusSuccess,
	})
}

func GetSpAiTemplatePub(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "No ID provided"})
		return
	}

	templateData := `
        SELECT
            id, title, description, category, ai_vendor, ai_model, configuration,
            published_by, created_at, updated_at
        FROM
            st_schema.sp_prompt_config_templates
        WHERE
            id = $1;
    `

	stmt, err := PDB.Prepare(templateData)
	if err != nil {
		log.Printf("Database Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error"})
		return
	}
	defer stmt.Close()

	row := stmt.QueryRow(id)

	var template models.SpAiTemplate
	var rawConfig []byte

	err = row.Scan(
		&template.ID, &template.Title, &template.Description, &template.Category,
		&template.AiVendor, &template.AiModel, &rawConfig,
		&template.PublishedBy, &template.CreatedAt, &template.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"message": "Template not found"})
		} else {
			log.Printf("Database Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error"})
		}
		return
	}

	var config models.AiConfiguration
	if err := json.Unmarshal(rawConfig, &config); err != nil {
		log.Printf("JSON Unmarshal Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "JSON unmarshalling error"})
		return
	}
	template.Configuration = &config

	c.JSON(http.StatusOK, gin.H{
		"data":    template,
		"message": models.StatusSuccess,
	})
}

func GetPublicUserTemplatesPub(c *gin.Context) {
	query := `
	SELECT
		id, 
		user_id, 
		title, 
		description, 
		category, 
		ai_vendor, 
		ai_model, 
		configuration, 
		published_by, 
		created_at, 
		updated_at
	FROM
		st_schema.community_prompt_config_templates
	`

	category := c.Query("category")

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

func GetPublicUserTemplatePub(c *gin.Context) {
	templateID := c.Query("id")
	if templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template ID required"})
		return
	}

	stmt, err := PDB.Prepare(`
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

// GetSpAiTemplates retrieves all Solution Pilot published AI Prompt Configurations
// @Summary Retrieve all Solution Pilot AI Prompt Configurations
// @Description Retrieves all AI Prompt Configurations published by Solution Pilot stored in the 'sp_prompt_config_templates' table.
// @Tags Solution Pilot AI Templates
// @Accept json
// @Produce json
// @Success 200 {array} SpAiTemplate "Array of Solution Pilot AI Prompt Configurations"
// @Failure 500 {object} string "Internal server error"
// @Router /api/spAiTemplates [get]
func GetSpAiTemplates(c *gin.Context) {
	templateData := `
        SELECT
            id, title, description, category, ai_vendor, ai_model, configuration,
            published_by, created_at, updated_at
        FROM
            st_schema.sp_prompt_config_templates;
    `

	stmt, err := tenantManagement.DB.Prepare(templateData)
	if err != nil {
		log.Printf("Database Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error"})
		return
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		log.Printf("Database Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error"})
		return
	}
	defer rows.Close()

	var templates []models.SpAiTemplate
	for rows.Next() {
		var template models.SpAiTemplate
		var rawConfig []byte

		err := rows.Scan(
			&template.ID, &template.Title, &template.Description, &template.Category,
			&template.AiVendor, &template.AiModel, &rawConfig,
			&template.PublishedBy, &template.CreatedAt, &template.UpdatedAt,
		)
		if err != nil {
			log.Printf("Database Error: %v", err)
			continue
		}

		var config models.AiConfiguration
		if err := json.Unmarshal(rawConfig, &config); err != nil {
			log.Printf("JSON Unmarshal Err: %v", err)
			continue
		}
		template.Configuration = &config

		templates = append(templates, template)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Database Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error after row iteration"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    templates,
		"message": models.StatusSuccess,
	})
}

// GetSpAiTemplate retrieves a single Solution Pilot Published Template by ID
// @Summary Retrieve a Solution Pilot AI Prompt Configuration by ID
// @Description Retrieves a single AI Prompt Configuration published by Solution Pilot using its ID.
// @Tags Solution Pilot AI Templates
// @Accept json
// @Produce json
// @Param id query string true "Template ID"
// @Success 200 {object} SpAiTemplate "Solution Pilot AI Prompt Configuration"
// @Failure 400 {object} string "Bad request, no ID provided"
// @Failure 404 {object} string "Template not found"
// @Failure 500 {object} string "Internal server error"
// @Router /api/spAiTemplate [get]
func GetSpAiTemplate(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "No ID provided"})
		return
	}

	templateData := `
        SELECT
            id, title, description, category, ai_vendor, ai_model, configuration,
            published_by, created_at, updated_at
        FROM
            st_schema.sp_prompt_config_templates
        WHERE
            id = $1;
    `

	stmt, err := tenantManagement.DB.Prepare(templateData)
	if err != nil {
		log.Printf("Database Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error"})
		return
	}
	defer stmt.Close()

	row := stmt.QueryRow(id)

	var template models.SpAiTemplate
	var rawConfig []byte

	err = row.Scan(
		&template.ID, &template.Title, &template.Description, &template.Category,
		&template.AiVendor, &template.AiModel, &rawConfig,
		&template.PublishedBy, &template.CreatedAt, &template.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"message": "Template not found"})
		} else {
			log.Printf("Database Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error"})
		}
		return
	}

	var config models.AiConfiguration
	if err := json.Unmarshal(rawConfig, &config); err != nil {
		log.Printf("JSON Unmarshal Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "JSON unmarshalling error"})
		return
	}
	template.Configuration = &config

	c.JSON(http.StatusOK, gin.H{
		"data":    template,
		"message": models.StatusSuccess,
	})
}
