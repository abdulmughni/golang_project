package projects

// This package contains the handlers for managing all documetns associated with projects.

// The handlers are:
// 1. NewDocument - Creates a new document for a project
// 2. GetDocument - Retrieves a document for a project
// 3. GetDocuments - Retrieves all documents for a project
// 4. UpdateDocument - Updates a document for a project
// 5. DeleteDocument - Deletes a document for a project

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"sententiawebapi/handlers/apis/images"
	"sententiawebapi/handlers/apis/tenantManagement"
	"sententiawebapi/handlers/models"
	"sententiawebapi/utilities"
	"strings"

	"github.com/gin-gonic/gin"
)

func NewDocument(c *gin.Context) {
	var Document models.Document

	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	// Extracting project ID from the URL parameter
	projectID := c.Query("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required..."})
		return
	}

	if err := c.ShouldBindJSON(&Document); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	Document.UserID = &userID
	Document.TenantId = &tenantID
	Document.ProjectID = &projectID

	row := tenantManagement.DB.QueryRow(`
		INSERT INTO st_schema.project_documents (
            user_id, tenant_id, project_id,
            content_json, raw_content, p_raw_content,
			title, complexity, document_type
        ) VALUES (
            $1, $2, $3,
			$4, $5, $6,
			$7, $8, $9
        ) RETURNING id, user_id, tenant_id, project_id, title, content_json, created_at, updated_at, complexity, document_type
		`,
		Document.UserID, Document.TenantId, Document.ProjectID,
		Document.Content, Document.RawContent, Document.PRawContent,
		Document.Title, Document.Complexity, Document.DocumentType,
	)

	err := row.Scan(
		&Document.ID,
		&Document.UserID,
		&Document.TenantId,
		&Document.ProjectID,
		&Document.Title,
		&Document.Content,
		&Document.CreatedAt,
		&Document.UpdatedAt,
		&Document.Complexity,
		&Document.DocumentType,
	)

	if err != nil {
		log.Printf("Database Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": models.InternalServerError})
		return
	}

	// Return the new document data
	c.JSON(200, gin.H{
		"data":    Document,
		"message": "New document created successfully!",
	})
}

func CloneDocument(c *gin.Context) {
	// Get the user ID and tenant ID from the context
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	projectID := c.Query("project_id")
	documentID := c.Query("document_id")

	if projectID == "" || documentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID and Document ID are required"})
		return
	}

	// Define struct for optional override values
	var overrideData struct {
		Title        string `json:"title"`
		Complexity   string `json:"complexity"`
		DocumentType string `json:"document_type"`
	}

	// Bind JSON body if provided
	if err := c.ShouldBindJSON(&overrideData); err != nil && err.Error() != "EOF" {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Start transaction
	tx, err := tenantManagement.DB.Begin()
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process clone"})
		return
	}
	defer tx.Rollback() // Will be no-op if transaction is committed

	// Fetch the existing document
	var existingDoc models.Document
	err = tx.QueryRow(`
        SELECT id, user_id, tenant_id, project_id, title, content_json, raw_content, p_raw_content, created_at, updated_at,
               complexity, ai_suggestions, document_type
        FROM st_schema.project_documents
        WHERE id = $1 AND project_id = $2 AND tenant_id = $3`,
		documentID, projectID, tenantID,
	).Scan(
		&existingDoc.ID,
		&existingDoc.UserID,
		&existingDoc.TenantId,
		&existingDoc.ProjectID,
		&existingDoc.Title,
		&existingDoc.Content,
		&existingDoc.RawContent,
		&existingDoc.PRawContent,
		&existingDoc.CreatedAt,
		&existingDoc.UpdatedAt,
		&existingDoc.Complexity,
		&existingDoc.AiSuggestions,
		&existingDoc.DocumentType,
	)

	if err != nil {
		log.Printf("Failed to fetch document: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch the document to clone"})
		return
	}

	// Apply overrides if provided
	title := existingDoc.Title
	if overrideData.Title != "" {
		title = &overrideData.Title
	}

	content := existingDoc.Content         // content_json
	rawContent := existingDoc.RawContent   // raw_content
	pRawContent := existingDoc.PRawContent // p_raw_content

	complexity := existingDoc.Complexity
	if overrideData.Complexity != "" {
		complexity = &overrideData.Complexity
	}

	documentType := existingDoc.DocumentType
	if overrideData.DocumentType != "" {
		documentType = &overrideData.DocumentType
	}

	// Create a new document with the same content but a new ID
	var newDocument models.Document
	err = tx.QueryRow(`
        INSERT INTO st_schema.project_documents (
            user_id, tenant_id, project_id,
            title, content_json, raw_content, p_raw_content,
            complexity, document_type
        ) VALUES (
            $1, $2, $3,
            $4, $5, $6, $7,
            $8, $9
        ) RETURNING id, user_id, tenant_id, project_id, title, content_json, raw_content, p_raw_content, created_at, updated_at,
                  complexity, ai_suggestions, document_type`,
		userID, tenantID, projectID,
		title, content, rawContent, pRawContent,
		complexity, documentType,
	).Scan(
		&newDocument.ID,
		&newDocument.UserID,
		&newDocument.TenantId,
		&newDocument.ProjectID,
		&newDocument.Title,
		&newDocument.Content,
		&newDocument.RawContent,
		&newDocument.PRawContent,
		&newDocument.CreatedAt,
		&newDocument.UpdatedAt,
		&newDocument.Complexity,
		&newDocument.AiSuggestions,
		&newDocument.DocumentType,
	)

	if err != nil {
		log.Printf("Failed to create cloned document: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create cloned document"})
		return
	}

	// Update project timestamp
	_, err = tx.Exec(`
        UPDATE st_schema.projects
        SET updated_at = NOW()
        WHERE id = $1 AND tenant_id = $2`,
		projectID, tenantID,
	)
	if err != nil {
		log.Printf("Failed to update project timestamp: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update the project timestamp"})
		return
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete the operation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    newDocument,
		"message": "Document cloned successfully!",
	})
}

func GetDocument(c *gin.Context) {
	// Get the user ID from the context
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	// Extracting project ID and document ID from the URL parameters
	projectID := c.Query("project_id")
	documentID := c.Query("document_id")

	if projectID == "" || documentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parameters are missing..."})
		return
	}

	// Query to fetch the document for the specific user
	query := `
        SELECT
			id, user_id, tenant_id, project_id, title, content, created_at, updated_at, complexity, ai_suggestions, document_type
        FROM
			st_schema.project_documents
        WHERE
			project_id = $1
			AND id = $2
			AND tenant_id = $3`

	var Document models.Document
	err := tenantManagement.DB.QueryRow(query, projectID, documentID, tenantID).Scan(
		&Document.ID,
		&Document.UserID,
		&Document.TenantId,
		&Document.ProjectID,
		&Document.Title,
		&Document.Content,
		&Document.CreatedAt,
		&Document.UpdatedAt,
		&Document.Complexity,
		&Document.AiSuggestions,
		&Document.DocumentType,
	)

	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found or access denied"})
		return
	}

	// Return the document data
	c.JSON(http.StatusOK, gin.H{
		"data":    Document,
		"message": "Project document retrieved successfully!",
	})
}

func GetDocuments(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	// Extracting project ID from the URL parameter
	projectID := c.Query("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	// Prepare the query
	stmt, err := tenantManagement.DB.Prepare(`
        SELECT
			id,
			user_id,
			tenant_id,
			project_id,
			title,
			created_at,
			updated_at,
			complexity,
			ai_suggestions,
			document_type
        FROM
			st_schema.project_documents
        WHERE
			tenant_id = $1
		AND
			project_id = $2
	`)

	if err != nil {
		log.Printf(models.DatabaseError, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare the query"})
		return
	}

	defer stmt.Close()

	// Execute the query
	rows, err := stmt.Query(tenantID, projectID)
	if err != nil {
		log.Printf("Failed to execute query: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute the query"})
		return
	}
	defer rows.Close()

	var documents []models.Document
	for rows.Next() {
		var doc models.Document
		err := rows.Scan(
			&doc.ID,
			&doc.UserID,
			&doc.TenantId,
			&doc.ProjectID,
			&doc.Title,
			&doc.CreatedAt,
			&doc.UpdatedAt,
			&doc.Complexity,
			&doc.AiSuggestions,
			&doc.DocumentType,
		)
		if err != nil {
			log.Printf(models.DatabaseError, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read document data"})
			return
		}
		documents = append(documents, doc)
	}

	// Check for errors from iterating over rows
	if err = rows.Err(); err != nil {
		log.Printf(models.DatabaseError, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed during retrieving documents"})
		return
	}

	// Return the list of documents
	c.JSON(http.StatusOK, gin.H{
		"data": documents,
	})
}

func UpdateDocument(c *gin.Context) {
	// Get the user ID from the context
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	projectID := c.Query("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	documentID := c.Query("document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document ID is required"})
		return
	}

	var updateData struct {
		Title         *string `json:"title"`
		Content       *string `json:"content"`
		Complexity    *string `json:"complexity"`
		AiSuggestions *bool   `json:"ai_suggestions"`
		DocumentType  *string `json:"document_type"`
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
	if updateData.Content != nil {
		setParts = append(setParts, fmt.Sprintf("content = $%d", argCounter))
		args = append(args, *updateData.Content)
		argCounter++
	}
	if updateData.Complexity != nil {
		setParts = append(setParts, fmt.Sprintf("complexity = $%d", argCounter))
		args = append(args, *updateData.Complexity)
		argCounter++
	}
	if updateData.AiSuggestions != nil {
		setParts = append(setParts, fmt.Sprintf("ai_suggestions = $%d", argCounter))
		args = append(args, *updateData.AiSuggestions)
		argCounter++
	}
	if updateData.DocumentType != nil {
		setParts = append(setParts, fmt.Sprintf("document_type = $%d", argCounter))
		args = append(args, *updateData.DocumentType)
		argCounter++
	}

	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No updatable fields provided"})
		return
	}

	setClause := strings.Join(setParts, ", ")

	// Start transaction
	tx, err := tenantManagement.DB.Begin()
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process update"})
		return
	}
	defer tx.Rollback() // Will be no-op if transaction is committed

	query := fmt.Sprintf(`
        UPDATE st_schema.project_documents
        SET %s, updated_at = NOW()
        WHERE tenant_id = $%d AND project_id = $%d AND id = $%d
        RETURNING id, user_id, tenant_id, project_id, title, content, created_at, updated_at,
                  complexity, ai_suggestions, document_type
    `, setClause, argCounter, argCounter+1, argCounter+2)

	args = append(args, tenantID, projectID, documentID)

	var updatedDocument models.Document
	err = tx.QueryRow(query, args...).Scan(
		&updatedDocument.ID,
		&updatedDocument.UserID,
		&updatedDocument.TenantId,
		&updatedDocument.ProjectID,
		&updatedDocument.Title,
		&updatedDocument.Content,
		&updatedDocument.CreatedAt,
		&updatedDocument.UpdatedAt,
		&updatedDocument.Complexity,
		&updatedDocument.AiSuggestions,
		&updatedDocument.DocumentType,
	)

	if err != nil {
		log.Printf("Failed to update document: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update the document"})
		return
	}

	// Update project timestamp within the same transaction
	updateProjectQuery := `
        UPDATE st_schema.projects
        SET updated_at = NOW()
        WHERE id = $1 AND tenant_id = $2
    `
	_, err = tx.Exec(updateProjectQuery, projectID, tenantID)
	if err != nil {
		log.Printf("Failed to update project timestamp: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update the project timestamp"})
		return
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete the update"})
		return
	}

	// Return the updated document data
	c.JSON(http.StatusOK, gin.H{
		"data":    updatedDocument,
		"message": "Document updated successfully",
	})
}

func DeleteDocument(c *gin.Context) {
	// Get the user ID from the context
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	projectID := c.Query("project_id")
	documentID := c.Query("document_id")
	if projectID == "" || documentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID and Document ID are required"})
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
        DELETE FROM st_schema.project_documents
        WHERE id = $1 AND project_id = $2 AND tenant_id = $3
        RETURNING id, user_id, tenant_id, project_id, title, content_json, created_at, updated_at,
        complexity, ai_suggestions, document_type`,
		documentID, projectID, tenantID,
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
		&document.DocumentType,
	)

	if err != nil {
		log.Printf("Failed to delete document: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete the document"})
		return
	}

	// Update project timestamp
	updateProjectQuery := `
        UPDATE st_schema.projects
        SET updated_at = NOW()
        WHERE id = $1 AND tenant_id = $2
    `
	_, err = tx.Exec(updateProjectQuery, projectID, tenantID)
	if err != nil {
		log.Printf("Failed to update project timestamp: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update the project timestamp"})
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
		"message": "Document deleted successfully",
	})
}

// This function creates new document from private template
func NewDocumentFromTemplate(c *gin.Context) {
	// Get the user ID and tenant ID from the context
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	projectID := c.Query("project_id")
	templateID := c.Query("document_template_id")

	if projectID == "" || templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID and Template ID are required"})
		return
	}

	// Start transaction
	tx, err := tenantManagement.DB.Begin()
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process template"})
		return
	}
	defer tx.Rollback() // Will be no-op if transaction is committed

	// Fetch template
	var template models.DocumentTemplate
	err = tx.QueryRow(`
        SELECT user_id, title, complexity, p_content_json, p_raw_content, document_type
        FROM st_schema.document_templates
        WHERE id = $1 AND tenant_id = $2`,
		templateID, tenantID,
	).Scan(
		&template.UserID,
		&template.Title,
		&template.Complexity,
		&template.Content,
		&template.RawContent,
		&template.DocumentType,
	)
	if err != nil {
		log.Printf("Error fetching document template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch document template"})
		return
	}

	// Create the new document from the template
	var document models.Document
	err = tx.QueryRow(`
        INSERT INTO st_schema.project_documents (
            user_id,
            tenant_id,
            project_id,
            title,
            content_json,
			raw_content,
			p_raw_content,
            complexity,
			document_type
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        RETURNING id, user_id, tenant_id, project_id, title, content_json, created_at, updated_at,
                  complexity, ai_suggestions, document_type`,
		userID,
		tenantID,
		projectID,
		template.Title,
		template.Content,
		template.RawContent,
		template.RawContent,
		template.Complexity,
		template.DocumentType,
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
		&document.DocumentType,
	)
	if err != nil {
		log.Printf("Failed to create document from template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create document from template"})
		return
	}

	err = images.CopyFiles(context.Background(), tx, images.CopyParams{
		SourceDocument: images.DocumentRef{
			Type: models.ResourceGroupTemplate,
			ID:   templateID,
		},
		DestinationDocument: images.DocumentRef{
			Type: models.ResourceGroupProject,
			ID:   *document.ID,
		},
		TenantID: tenantID,
	})
	if err != nil {
		log.Printf("Error while copying files: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error copying files"})
		return
	}

	// Update project timestamp
	_, err = tx.Exec(`
        UPDATE st_schema.projects
        SET updated_at = NOW()
        WHERE id = $1 AND tenant_id = $2`,
		projectID, tenantID,
	)
	if err != nil {
		log.Printf("Failed to update project timestamp: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update the project timestamp"})
		return
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete the operation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    document,
		"message": "New project document created from template successfully!",
	})
}

// This handler creates a new document from a public document template in a private project
func NewDocumentFromPubTemplate(c *gin.Context) {
	// Get the user ID and tenant ID from the context
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	projectID := c.Query("project_id")
	templateID := c.Query("public_document_template_id")

	if projectID == "" || templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID and Template ID are required"})
		return
	}

	// Start transaction
	tx, err := tenantManagement.DB.Begin()
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process template"})
		return
	}
	defer tx.Rollback() // Will be no-op if transaction is committed

	// Fetch public template
	var template models.PublicDocumentTemplate
	err = tx.QueryRow(`
        SELECT user_id, title, complexity, content, raw_content, document_type
        FROM st_schema.cm_document_templates
        WHERE id = $1`,
		templateID,
	).Scan(
		&template.UserID,
		&template.Title,
		&template.Complexity,
		&template.Content,
		&template.RawContent,
		&template.DocumentType,
	)
	if err != nil {
		log.Printf("Error fetching public document template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch public document template"})
		return
	}

	// Create the new document from the public template
	var document models.Document
	err = tx.QueryRow(`
        INSERT INTO st_schema.project_documents (
            user_id,
            tenant_id,
            project_id,
            title,
            content,
			raw_content,
            complexity,
			document_type
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, user_id, tenant_id, project_id, title, content, created_at, updated_at,
                  complexity, ai_suggestions, document_type`,
		userID,
		tenantID,
		projectID,
		template.Title,
		template.Content,
		template.RawContent,
		template.Complexity,
		template.DocumentType,
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
		&document.DocumentType,
	)
	if err != nil {
		log.Printf("Failed to create document from public template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create document from public template"})
		return
	}

	// Update project timestamp
	_, err = tx.Exec(`
        UPDATE st_schema.projects
        SET updated_at = NOW()
        WHERE id = $1 AND tenant_id = $2`,
		projectID, tenantID,
	)
	if err != nil {
		log.Printf("Failed to update project timestamp: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update the project timestamp"})
		return
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete the operation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    document,
		"message": "New project document created from public template successfully!",
	})
}
