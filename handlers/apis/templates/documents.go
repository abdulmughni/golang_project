package templates

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"sententiawebapi/handlers/apis/tenantManagement"
	"sententiawebapi/handlers/models"
	"sententiawebapi/utilities"

	"github.com/gin-gonic/gin"
)

// This function creates a new internal document template
func NewInternalDocumentTemplate(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	var Document models.DocumentTemplate

	// Extracting project ID from the URL parameter
	projectTemplateID := c.Query("project_template_id")
	if projectTemplateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project Template ID is required..."})
		return
	}

	if err := c.ShouldBindJSON(&Document); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	Document.UserID = &userID
	Document.TenantID = &tenantID
	Document.ProjectTemplateID = &projectTemplateID

	err := tenantManagement.DB.QueryRow(`
		INSERT INTO st_schema.document_templates (
            user_id, tenant_id, project_template_id,
            content, content_json, raw_content, p_content_json, p_raw_content,
			title, complexity
        ) VALUES (
            $1, $2, $3, $4,
			$5, $6, $7, $8,
			$9, $10
        ) RETURNING id, user_id, tenant_id, project_template_id, title, p_content_json, created_at, updated_at, complexity
		`,
		Document.UserID, Document.TenantID, Document.ProjectTemplateID,
		"", Document.Content, Document.RawContent, Document.Content, Document.RawContent,
		Document.Title, Document.Complexity,
	).Scan(
		&Document.ID,
		&Document.UserID,
		&Document.TenantID,
		&Document.ProjectTemplateID,
		&Document.Title,
		&Document.Content,
		&Document.CreatedAt,
		&Document.UpdatedAt,
		&Document.Complexity,
	)

	if err != nil {
		log.Printf("Database Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": models.DatabaseError})
		return
	}

	// Return the new document data
	c.JSON(200, gin.H{
		"data":    Document,
		"message": "New document template created successfully!",
	})
}

// This function retrieves the details of a specific internal document template
func GetInternalDocumentTemplate(c *gin.Context) {
	// Get the user ID from the context
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	// Extracting document ID from the URL parameters
	documentTemplateId := c.Query("id")

	if documentTemplateId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document Template ID is required"})
		return
	}

	var documentTemplate models.DocumentTemplate
	err := tenantManagement.DB.QueryRow(`
        SELECT
			id, user_id, tenant_id, project_template_id, title, complexity, p_content_json, created_at, updated_at, privacy
        FROM
			st_schema.document_templates
        WHERE
			id = $1
		AND
			tenant_id = $2`,
		documentTemplateId, tenantID,
	).Scan(
		&documentTemplate.ID,
		&documentTemplate.UserID,
		&documentTemplate.TenantID,
		&documentTemplate.ProjectTemplateID,
		&documentTemplate.Title,
		&documentTemplate.Complexity,
		&documentTemplate.Content,
		&documentTemplate.CreatedAt,
		&documentTemplate.UpdatedAt,
		&documentTemplate.Privacy,
	)

	if err != nil {
		if isDevelopmentEnvironment() {
			log.Printf("Error retrieving document template: %v", err)
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found or access denied"})
		return
	}

	// Return the document data
	c.JSON(http.StatusOK, gin.H{
		"data":    documentTemplate,
		"message": "Project document retrieved successfully!",
	})
}

// This function retrieves an array of all internal document templates from across all project templates.
func GetInternalDocumentTemplates(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	category := c.Query("category")
	projectTemplateID := c.Query("project_template_id")

	query := `
        SELECT
            id,
            user_id,
            tenant_id,
            project_template_id,
            title,
            complexity,
            p_content_json,
            created_at,
			updated_at,
			privacy,
            category,
			description
        FROM
            st_schema.document_templates
    `

	conditions := []string{}
	arguments := []interface{}{}
	argCounter := 0

	if category != "" {
		argCounter++
		conditions = append(conditions, fmt.Sprintf("category = $%d", argCounter))
		arguments = append(arguments, category)
	}

	if projectTemplateID != "" {
		argCounter++
		conditions = append(conditions, fmt.Sprintf("projectTemplateID = $%d", argCounter))
		arguments = append(arguments, projectTemplateID)
	}

	argCounter++
	conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argCounter))
	arguments = append(arguments, tenantID)

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	rows, err := tenantManagement.DB.Query(query, arguments...)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve document templates: " + err.Error()})
		return
	}
	defer rows.Close()

	var documentTemplates []models.DocumentTemplate
	for rows.Next() {
		var documentTemplate models.DocumentTemplate
		err := rows.Scan(
			&documentTemplate.ID,
			&documentTemplate.UserID,
			&documentTemplate.TenantID,
			&documentTemplate.ProjectTemplateID,
			&documentTemplate.Title,
			&documentTemplate.Complexity,
			&documentTemplate.Content,
			&documentTemplate.CreatedAt,
			&documentTemplate.UpdatedAt,
			&documentTemplate.Privacy,
			&documentTemplate.Category,
			&documentTemplate.Description,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning document templates: " + err.Error()})
			return
		}
		documentTemplates = append(documentTemplates, documentTemplate)
	}

	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating document templates: " + err.Error()})
		return
	}

	// Construct and send the response
	c.JSON(http.StatusOK, gin.H{
		"data":    documentTemplates,
		"message": "Document templates retrieved successfully",
	})
}

// This function updates an existing internal document template
func UpdateInternalDocumentTemplate(c *gin.Context) {
	// Get the tenant ID from the context
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	// Extracting project_template_id and document_template_id from the URL parameters
	projectTemplateID := c.Query("project_template_id")
	documentTemplateID := c.Query("document_template_id")

	if projectTemplateID == "" || documentTemplateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project Template ID and Document Template ID are required"})
		return
	}

	var updateData struct {
		Title         *string          `json:"title"`
		Complexity    *string          `json:"complexity"`
		Content       *json.RawMessage `json:"content"`
		RawContent    []byte           `json:"raw_content"`
		Privacy       *bool            `json:"privacy"`
		Category      *string          `json:"category"`
		Description   *string          `json:"description"`
		AiSuggestions *bool            `json:"ai_suggestions"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	setParts := []string{}
	args := []interface{}{}
	argCounter := 1

	if updateData.Title != nil {
		setParts = append(setParts, fmt.Sprintf("title = $%d", argCounter))
		args = append(args, *updateData.Title)
		argCounter++
	}
	if updateData.Complexity != nil {
		setParts = append(setParts, fmt.Sprintf("complexity = $%d", argCounter))
		args = append(args, *updateData.Complexity)
		argCounter++
	}
	if updateData.Content != nil {
		setParts = append(setParts, fmt.Sprintf("p_content_json = $%d", argCounter))
		args = append(args, *updateData.Content)
		argCounter++
	}
	if updateData.RawContent != nil {
		setParts = append(setParts, fmt.Sprintf("p_raw_content = $%d", argCounter))
		args = append(args, updateData.RawContent)
		argCounter++
	}
	if updateData.Privacy != nil {
		setParts = append(setParts, fmt.Sprintf("privacy = $%d", argCounter))
		args = append(args, *updateData.Privacy)
		argCounter++
	}
	if updateData.Category != nil {
		setParts = append(setParts, fmt.Sprintf("category = $%d", argCounter))
		args = append(args, *updateData.Category)
		argCounter++
	}
	if updateData.Description != nil {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argCounter))
		args = append(args, *updateData.Description)
		argCounter++
	}
	if updateData.AiSuggestions != nil {
		setParts = append(setParts, fmt.Sprintf("ai_suggestions = $%d", argCounter))
		args = append(args, *updateData.AiSuggestions)
		argCounter++
	}

	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No updatable fields provided"})
		return
	}

	setClause := strings.Join(setParts, ", ")
	query := fmt.Sprintf(`
        UPDATE
			st_schema.document_templates
        SET
			%s, updated_at = NOW()
        WHERE
			id = $%d
		AND
			project_template_id = $%d
		AND
			tenant_id = $%d
        RETURNING
			id, user_id, tenant_id, project_template_id,
			title, complexity, p_content_json, created_at,
			updated_at, privacy, category, description,
			ai_suggestions
    `, setClause, argCounter, argCounter+1, argCounter+2)

	args = append(args, documentTemplateID, projectTemplateID, tenantID)

	stmt, err := tenantManagement.DB.Prepare(query)
	if err != nil {
		if isDevelopmentEnvironment() {
			log.Printf("Error preparing statement: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	defer stmt.Close()

	var updatedTemplate models.DocumentTemplate
	err = stmt.QueryRow(args...).Scan(
		&updatedTemplate.ID,
		&updatedTemplate.UserID,
		&updatedTemplate.TenantID,
		&updatedTemplate.ProjectTemplateID,
		&updatedTemplate.Title,
		&updatedTemplate.Complexity,
		&updatedTemplate.Content,
		&updatedTemplate.CreatedAt,
		&updatedTemplate.UpdatedAt,
		&updatedTemplate.Privacy,
		&updatedTemplate.Category,
		&updatedTemplate.Description,
		&updatedTemplate.AiSuggestions,
	)

	if err != nil {
		log.Printf(models.DatabaseError, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update the document template"})
		return
	}

	// Return the updated document template data
	c.JSON(http.StatusOK, gin.H{
		"data":    updatedTemplate,
		"message": "Document template updated successfully",
	})
}

// This function deletes an existing internal document template
func DeleteInternalDocumentTemplate(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	projectTemplateID := c.Query("project_template_id")
	documentTemplateID := c.Query("document_template_id")
	if projectTemplateID == "" || documentTemplateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project Template ID and Document Template ID are required"})
		return
	}

	// Start transaction
	tx, err := tenantManagement.DB.Begin()
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process delete"})
		return
	}
	defer tx.Rollback() // Will be no-op if transaction is committed

	// Delete the document
	var document models.Document
	err = tx.QueryRow(`
        DELETE FROM st_schema.document_templates
        WHERE id = $1 AND project_template_id = $2 AND tenant_id = $3
        RETURNING id, user_id, tenant_id, project_template_id, title, p_content_json, created_at, updated_at,
                  complexity, ai_suggestions`,
		documentTemplateID, projectTemplateID, tenantID,
	).Scan(
		&document.ID,
		&document.UserID,
		&document.TenantId,
		&document.ProjectID,
		&document.Title,
		&document.Content,
		&document.CreatedAt,
		&document.UpdatedAt,
		&document.Complexity,
		&document.AiSuggestions,
	)

	if err != nil {
		log.Printf("Failed to delete document: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete the document"})
		return
	}

	// Update project timestamp
	updateProjectTemplateQuery := `
        UPDATE st_schema.project_templates
        SET updated_at = NOW()
        WHERE id = $1 AND tenant_id = $2
    `
	_, err = tx.Exec(updateProjectTemplateQuery, projectTemplateID, tenantID)
	if err != nil {
		log.Printf("Failed to update project template timestamp: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update the project template timestamp"})
		return
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete the delete"})
		return
	}

	// Return a success response
	c.JSON(http.StatusOK, gin.H{
		"data":    document,
		"message": "Document template deleted successfully",
	})
}
