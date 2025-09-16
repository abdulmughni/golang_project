package community

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

// This function creates a new community document template
func NewPublicTemplateDocument(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	var Document models.PublicDocumentTemplate

	// Extracting community project ID from the URL parameter
	communityProjectTemplateID := c.Query("cm_template_id")
	if communityProjectTemplateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Community Project Template ID is required..."})
		return
	}

	if err := c.ShouldBindJSON(&Document); err != nil {
		log.Printf("Error binding JSON: %v", err)

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	Document.UserID = &userID
	Document.TenantID = &tenantID
	Document.CommunityProjectTemplateID = &communityProjectTemplateID

	err := tenantManagement.DB.QueryRow(`
		INSERT INTO st_schema.cm_document_templates (
            user_id, tenant_id, community_project_template_id,
			content_json, raw_content, p_content_json, p_raw_content,
			title, description, complexity, published_at
        )
		VALUES (
            $1, $2, $3,
			$4, $5, $6, $7,
			$8, $9, $10, NOW()
        )
		RETURNING id, user_id, tenant_id, community_project_template_id, title, description, p_content_json, complexity, published_at
		`,
		Document.UserID, Document.TenantID, Document.CommunityProjectTemplateID,
		Document.Content, Document.RawContent, Document.Content, Document.RawContent,
		Document.Title, Document.Description, Document.Complexity,
	).Scan(
		&Document.ID,
		&Document.UserID,
		&Document.TenantID,
		&Document.CommunityProjectTemplateID,
		&Document.Title,
		&Document.Description,
		&Document.Content,
		&Document.Complexity,
		&Document.PublishedAt,
	)

	if err != nil {
		log.Printf("Database Err: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{"error": models.DatabaseError})
		return
	}

	// Return the new document data
	c.JSON(200, gin.H{
		"data":    Document,
		"message": "New community document template created successfully!",
	})
}

// Below functions retrieve the document template details when users opens up the public project template
// it retrieves the document details for the specific project template
func GetPublicTemplateDocument(c *gin.Context) {
	// Extracting project ID and document ID from the URL parameters
	projectDocumentTemplateID := c.Query("cm_template_document_id")

	if projectDocumentTemplateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parameters are missing..."})
		return
	}

	// Query to fetch the document for the specific user
	query := `
        SELECT
			id, user_id, community_project_template_id, title, description, p_content_json, complexity, published_at
        FROM
			st_schema.cm_document_templates
        WHERE
			id = $1`

	var Document models.PublicDocumentTemplate
	err := tenantManagement.DB.QueryRow(query, projectDocumentTemplateID).Scan(
		&Document.ID,
		&Document.UserID,
		&Document.CommunityProjectTemplateID,
		&Document.Title,
		&Document.Description,
		&Document.Content,
		&Document.Complexity,
		&Document.PublishedAt,
	)

	if err != nil {
		log.Printf(models.DatabaseError, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found or access denied"})
		return
	}

	// Return the document data
	c.JSON(http.StatusOK, gin.H{
		"data":    Document,
		"message": "Project document retrieved successfully!",
	})
}

// This function retrieves all public document templates
// Optionally filters by category and/or community template ID
func GetPublicTemplateDocuments(c *gin.Context) {
	category := c.Query("category")
	communityProjectTemplateID := c.Query("cm_template_id")
	query := `
        SELECT
            id,
            community_project_template_id,
            user_id,
            tenant_id,
            title,
            description,
            p_content_json,
            complexity,
            published_at,
            category,
			last_update
        FROM
            st_schema.cm_document_templates
    `

	conditions := []string{}
	arguments := []interface{}{}
	argCounter := 0

	// Modify the query based on the presence of the category parameter
	if category != "" {
		argCounter++
		conditions = append(conditions, fmt.Sprintf("category = $%d", argCounter))
		arguments = append(arguments, category)
	}

	if communityProjectTemplateID != "" {
		argCounter++
		conditions = append(conditions, fmt.Sprintf("community_project_template_id = $%d", argCounter))
		arguments = append(arguments, communityProjectTemplateID)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	rows, err := tenantManagement.DB.Query(query, arguments...)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve document templates: " + err.Error()})
		return
	}
	defer rows.Close()

	var documentTemplates []models.PublicDocumentTemplate
	for rows.Next() {
		var documentTemplate models.PublicDocumentTemplate
		err := rows.Scan(
			&documentTemplate.ID,
			&documentTemplate.ProjectTemplateID,
			&documentTemplate.UserID,
			&documentTemplate.TenantID,
			&documentTemplate.Title,
			&documentTemplate.Description,
			&documentTemplate.Content,
			&documentTemplate.Complexity,
			&documentTemplate.PublishedAt,
			&documentTemplate.Category,
			&documentTemplate.LastUpdateAt,
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

// This function updates an existing community document template
func UpdatePublicTemplateDocument(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	// Extracting community_project_template_id and document_template_id from the URL parameters
	communityProjectTemplateID := c.Query("cm_template_id")
	documentTemplateID := c.Query("cm_template_document_id")

	if communityProjectTemplateID == "" || documentTemplateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Community Project Template ID and Document Template ID are required"})
		return
	}

	var updateData struct {
		Title       *string          `json:"title"`
		Description *string          `json:"description"`
		Content     *json.RawMessage `json:"content"`
		RawContent  []byte           `json:"raw_content"`
		Complexity  *string          `json:"complexity"`
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
	if updateData.Description != nil {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argCounter))
		args = append(args, *updateData.Description)
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
	if updateData.Complexity != nil {
		setParts = append(setParts, fmt.Sprintf("complexity = $%d", argCounter))
		args = append(args, *updateData.Complexity)
		argCounter++
	}

	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No updatable fields provided"})
		return
	}

	setClause := strings.Join(setParts, ", ")
	query := fmt.Sprintf(`
        UPDATE
			st_schema.cm_document_templates
        SET
			%s, published_at = NOW()
        WHERE
			id = $%d
		AND
			community_project_template_id = $%d
		AND
			tenant_id = $%d
        RETURNING
			id, user_id, community_project_template_id,
			title, description, p_content_json, complexity, published_at
    `, setClause, argCounter, argCounter+1, argCounter+2)

	args = append(args, documentTemplateID, communityProjectTemplateID, tenantID)

	stmt, err := tenantManagement.DB.Prepare(query)
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": models.InternalServerError})
		return
	}
	defer stmt.Close()

	var updatedTemplate models.PublicDocumentTemplate
	err = stmt.QueryRow(args...).Scan(
		&updatedTemplate.ID,
		&updatedTemplate.UserID,
		&updatedTemplate.CommunityProjectTemplateID,
		&updatedTemplate.Title,
		&updatedTemplate.Description,
		&updatedTemplate.Content,
		&updatedTemplate.Complexity,
		&updatedTemplate.PublishedAt,
	)

	if err != nil {
		log.Printf(models.DatabaseError, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update the community document template"})
		return
	}

	// Return the updated document template data
	c.JSON(http.StatusOK, gin.H{
		"data":    updatedTemplate,
		"message": "Community document template updated successfully",
	})
}

// This function deletes an existing community document template
func DeletePublicTemplateDocument(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	communityProjectTemplateID := c.Query("cm_template_id")
	documentTemplateID := c.Query("cm_template_document_id")
	if communityProjectTemplateID == "" || documentTemplateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Community Project Template ID and Document Template ID are required"})
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
	var document models.PublicDocumentTemplate
	err = tx.QueryRow(`
        DELETE FROM st_schema.cm_document_templates
        WHERE id = $1 AND community_project_template_id = $2 AND tenant_id = $3
        RETURNING id, user_id, community_project_template_id, title, description, p_content_json, complexity, published_at`,
		documentTemplateID, communityProjectTemplateID, tenantID,
	).Scan(
		&document.ID,
		&document.UserID,
		&document.CommunityProjectTemplateID,
		&document.Title,
		&document.Description,
		&document.Content,
		&document.Complexity,
		&document.PublishedAt,
	)

	if err != nil {
		log.Printf("Failed to delete document: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete the community document"})
		return
	}

	// Update project timestamp
	updateCommunityTemplateQuery := `
        UPDATE st_schema.cm_project_templates
        SET last_update_at = NOW()
        WHERE id = $1 AND tenant_id = $2
    `
	_, err = tx.Exec(updateCommunityTemplateQuery, communityProjectTemplateID, tenantID)
	if err != nil {
		log.Printf("Failed to update community project template timestamp: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": models.InternalServerError})
		return
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": models.InternalServerError})
		return
	}

	// Return a success response
	c.JSON(http.StatusOK, gin.H{
		"data":    document,
		"message": "Community document template deleted successfully",
	})
}

func GetWebPublicProjectTemplateDocument(c *gin.Context) {

	// Extracting project ID and document ID from the URL parameters
	projectTemplateID := c.Query("cm_template_id")
	projectDocumentTemplateID := c.Query("cm_template_document_id")

	if projectTemplateID == "" || projectDocumentTemplateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parameters are missing..."})
		return
	}

	// Query to fetch the document for the specific user
	query := `
        SELECT
			id, user_id, tenant_id, community_project_template_id, title, description, p_content_json, complexity, published_at
        FROM
			st_schema.cm_document_templates
        WHERE
			community_project_template_id = $1
		AND
			id = $2`

	var Document models.PublicDocumentTemplate
	err := tenantManagement.DB.QueryRow(query, projectTemplateID, projectDocumentTemplateID).Scan(
		&Document.ID,
		&Document.UserID,
		&Document.TenantID,
		&Document.CommunityProjectTemplateID,
		&Document.Title,
		&Document.Description,
		&Document.Content,
		&Document.Complexity,
		&Document.PublishedAt,
	)

	if err != nil {
		log.Printf(models.DatabaseError, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found or access denied"})
		return
	}

	// Return the document data
	c.JSON(http.StatusOK, gin.H{
		"data":    Document,
		"message": "Project document retrieved successfully!",
	})
}
