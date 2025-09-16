package templates

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"sententiawebapi/handlers/apis/images"
	"sententiawebapi/handlers/apis/tenantManagement"
	"sententiawebapi/handlers/models"
	"sententiawebapi/utilities"

	"github.com/gin-gonic/gin"
)

func CreateProjectTemplate(c *gin.Context) {
	// Validate the user's identity
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	// Retrieve the projectID from the query parameters
	projectID := c.Query("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID must be provided"})
		return
	}

	// Declaring a project template object to store the new project template data
	var projectTemplate models.ProjectTemplate

	// Declaring a project object to store the fetched project data
	var project models.Project

	// Bind the request body to the project template object.
	if err := c.ShouldBindJSON(&projectTemplate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Start a new database transaction
	tx, err := tenantManagement.DB.Begin()
	if err != nil {
		if isDevelopmentEnvironment() {
			log.Printf("Error starting transaction: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database transaction error"})
		return
	}
	defer tx.Rollback() // Ensure rollback in case of error

	if isDevelopmentEnvironment() {
		log.Printf("Received tenantID: %v", tenantID)   // Log the userID for debugging
		log.Printf("Received projectID: %v", projectID) // Log the userID for debugging
	}

	// Error handling for missing fields
	if projectTemplate.Privacy == nil {
		defaultPrivacy := true
		projectTemplate.Privacy = &defaultPrivacy
	}

	if projectTemplate.UserID == nil {
		projectTemplate.UserID = &userID
	}

	if projectTemplate.TenantID == nil {
		projectTemplate.TenantID = &tenantID
	}

	// Set project defaults
	project.ID = &projectID
	project.UserID = &userID
	project.TenantID = &tenantID

	if isDevelopmentEnvironment() {
		log.Printf("Set all the variables")
	}

	// Fetch the project data from the database
	err = tx.QueryRow(`
    SELECT
		id, user_id, tenant_id, title, category, complexity
    FROM
		st_schema.projects
    WHERE
		id = $1
	AND
		tenant_id = $2`,
		&project.ID,
		&project.TenantID,
	).Scan(
		&project.ID,
		&project.UserID,
		&project.TenantID,
		&project.Title,
		&project.Category,
		&project.Complexity,
	)

	if project.Status == nil {
		defaultStatus := "Not Started"
		project.Status = &defaultStatus
	}

	if err != nil {
		if isDevelopmentEnvironment() {
			log.Printf("Err: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching project data"})
		return
	}

	err = tx.QueryRow(`
    WITH user_info AS (
        SELECT first_name, last_name, user_picture
        FROM st_schema.users
        WHERE id = $1
    )
    INSERT INTO
        st_schema.project_templates (user_id, tenant_id, title, description, category, complexity, privacy)
    VALUES
        ($1, $2, $3, $4, $5, $6, $7)
    RETURNING
        id, user_id, tenant_id, title, description, category, complexity, created_at, updated_at, privacy,
        (SELECT first_name FROM user_info),
        (SELECT last_name FROM user_info),
        (SELECT user_picture FROM user_info)`,
		*project.UserID,
		*project.TenantID,
		*projectTemplate.Title,
		*projectTemplate.Description,
		*project.Category,
		*project.Complexity,
		*projectTemplate.Privacy,
	).Scan(
		&projectTemplate.ID,
		&projectTemplate.UserID,
		&projectTemplate.TenantID,
		&projectTemplate.Title,
		&projectTemplate.Description,
		&projectTemplate.Category,
		&projectTemplate.Complexity,
		&projectTemplate.CreatedAt,
		&projectTemplate.UpdatedAt,
		&projectTemplate.Privacy,
		&projectTemplate.FirstName,
		&projectTemplate.LastName,
		&projectTemplate.UserPicture,
	)

	if err != nil {
		if isDevelopmentEnvironment() {
			log.Printf("Err: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting project template"})
		return
	}

	// Declare a cursor for documents associated with the original project
	_, err = tx.Exec(`
		DECLARE
			doc_cursor
		CURSOR FOR SELECT
			id, user_id, tenant_id, title, content_json, p_raw_content, complexity
		FROM
			st_schema.project_documents
		WHERE
			project_id = $1
		AND
			tenant_id = $2`,
		projectID,         // This should be the original project ID
		*project.TenantID, // This should be the tenant ID from the context
	)

	if err != nil {
		// Log the detailed error for debugging
		log.Printf("Error declaring document cursor: %v", err)
		// Send generic message to client
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred while processing your request"})
		return
	}

	if projectTemplate.ID == nil {
		log.Printf("Not received properly: %v", projectTemplate.ID)
	}

	// Iterate over the cursor, fetching documents one by one
	docIDs := make([]string, 0)
	docTemplateIDs := make([]string, 0)
	for {
		var docTemplate models.DocumentTemplate
		var doc models.Document

		// Use sql.NullString to handle potential NULL values
		var complexity = "low"

		docTemplate.Complexity = &complexity

		// Fetch the next document from the cursor and store it in the document object
		err = tx.QueryRow(`FETCH NEXT FROM doc_cursor`).Scan(
			&doc.ID,
			&doc.UserID,
			&doc.TenantId,
			&doc.Title,
			&doc.Content,
			&doc.RawContent,
			&doc.Complexity)

		if err != nil {
			if err == sql.ErrNoRows {
				break // No more documents, exit the loop
			}
			if isDevelopmentEnvironment() {
				log.Printf("Err: %v", err)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching next document from cursor"})
			return
		}

		if projectTemplate.ID == nil {
			log.Printf("Not received properly: %v", projectTemplate.ID)
		}
		// Prepare the document template for insertion
		docTemplate.UserID = doc.UserID // Set the UserID from the project template
		docTemplate.TenantID = doc.TenantId
		docTemplate.ProjectTemplateID = projectTemplate.ID // Associate with the new project template

		if docTemplate.ProjectTemplateID == nil {
			docTemplate.ProjectTemplateID = projectTemplate.ID
		}

		docTemplate.Title = doc.Title
		docTemplate.Complexity = doc.Complexity
		docTemplate.Content = doc.Content
		docTemplate.RawContent = doc.RawContent

		// Insert the document template into the database and retrieve its ID
		err = tx.QueryRow(`
            INSERT INTO
				st_schema.document_templates (
					user_id, tenant_id, title,
					content_json, raw_content, p_content_json, p_raw_content,
					project_template_id, complexity, content
				)
            VALUES
				(
					$1, $2, $3,
					$4, $5, $6, $7,
					$8, $9, ''
				)
			RETURNING id`,
			*docTemplate.UserID,
			*docTemplate.TenantID,
			*docTemplate.Title,
			docTemplate.Content,
			docTemplate.RawContent,
			docTemplate.Content,
			docTemplate.RawContent,
			*docTemplate.ProjectTemplateID,
			*docTemplate.Complexity,
		).Scan(&docTemplate.ID)
		if err != nil {
			log.Printf("Err: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting document template"})
			return // Transaction will be rolled back due to deferred rollback call
		}
		newDocTemplateID := *docTemplate.ID

		docIDs = append(docIDs, *doc.ID)
		docTemplateIDs = append(docTemplateIDs, newDocTemplateID)
	}

	projectTemplate.DocumentTemplates = &docTemplateIDs
	// Close the cursor after use
	_, err = tx.Exec("CLOSE doc_cursor")

	if err != nil {
		log.Printf("Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error closing cursor"})
		return // Transaction will be rolled back due to deferred rollback call
	}

	err = images.CopyProjectFiles(context.Background(), tx, images.CopyProjectParams{
		SourceDocuments: images.DocumentsRef{
			Type: models.ResourceGroupProject,
			IDs:  docIDs,
		},
		DestinationDocuments: images.DocumentsRef{
			Type: models.ResourceGroupTemplate,
			IDs:  docTemplateIDs,
		},
		TenantID: tenantID,
	})
	if err != nil {
		log.Printf("Error while copying files: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error copying files"})
		return // Transaction will be rolled back due to deferred rollback call
	}

	// After document templates cursor is closed and before tx.Commit()
	// Add diagram templates handling
	_, err = tx.Exec(`
		DECLARE
			diagram_cursor
		CURSOR FOR SELECT
			id, user_id, tenant_id, title, diagram_type, diagram_status, category, design, raw_design
		FROM
			st_schema.diagrams
		WHERE
			project_id = $1
		AND
			tenant_id = $2`,
		projectID,
		*project.TenantID,
	)

	if err != nil {
		if isDevelopmentEnvironment() {
			log.Printf("Err: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error declaring diagram cursor"})
		return
	}

	// Iterate over the diagram cursor
	diagramTemplateIDs := make([]string, 0)
	for {
		var diagramTemplate models.DiagramTemplate
		var diagram struct {
			ID            *string          `json:"id"`
			UserID        *string          `json:"user_id"`
			TenantID      *string          `json:"tenant_id"`
			Title         *string          `json:"title"`
			DiagramType   *string          `json:"diagram_type"`
			DiagramStatus *string          `json:"diagram_status"`
			Category      *string          `json:"category"`
			Design        *json.RawMessage `json:"design"`
			RawDesign     []byte           `json:"raw_design"`
		}

		// Set default values
		defaultStatus := "Not Started"
		defaultType := "flowchart"
		defaultCategory := "general"

		err = tx.QueryRow(`FETCH NEXT FROM diagram_cursor`).Scan(
			&diagram.ID,
			&diagram.UserID,
			&diagram.TenantID,
			&diagram.Title,
			&diagram.DiagramType,
			&diagram.DiagramStatus,
			&diagram.Category,
			&diagram.Design,
			&diagram.RawDesign)

		if err != nil {
			if err == sql.ErrNoRows {
				break // No more diagrams
			}
			log.Printf("Err: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching next diagram from cursor"})
			return
		}

		// Handle null values with defaults
		if diagram.DiagramType == nil {
			diagram.DiagramType = &defaultType
		}
		if diagram.DiagramStatus == nil {
			diagram.DiagramStatus = &defaultStatus
		}
		if diagram.Category == nil {
			diagram.Category = &defaultCategory
		}

		// Prepare diagram template for insertion
		diagramTemplate.UserID = diagram.UserID
		diagramTemplate.TenantID = diagram.TenantID
		diagramTemplate.ProjectTemplateID = projectTemplate.ID
		diagramTemplate.Title = diagram.Title
		diagramTemplate.DiagramType = diagram.DiagramType
		diagramTemplate.DiagramStatus = diagram.DiagramStatus
		diagramTemplate.Category = diagram.Category
		diagramTemplate.RawDesign = diagram.RawDesign

		// Ensure Design is a valid JSON string or set a default
		if diagram.Design != nil && json.Valid([]byte(*diagram.Design)) {
			diagramTemplate.Design = diagram.Design
		} else {
			defaultDesign := json.RawMessage(`{}`) // Default to an empty JSON object
			diagramTemplate.Design = &defaultDesign
		}

		// Insert the diagram template
		err = tx.QueryRow(`
			INSERT INTO
				st_schema.diagram_templates (
					user_id,
					tenant_id,
					project_template_id,
					title,
					diagram_type,
					diagram_status,
					category,
					design,
					raw_design,
					p_design,
					p_raw_design
				)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			RETURNING id`,
			*diagramTemplate.UserID,
			*diagramTemplate.TenantID,
			*diagramTemplate.ProjectTemplateID,
			*diagramTemplate.Title,
			*diagramTemplate.DiagramType,
			*diagramTemplate.DiagramStatus,
			*diagramTemplate.Category,
			*diagramTemplate.Design,
			diagramTemplate.RawDesign,
			*diagramTemplate.Design,
			diagramTemplate.RawDesign,
		).Scan(&diagramTemplate.ID)

		if err != nil {
			if isDevelopmentEnvironment() {
				log.Printf("Err: %v", err)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting diagram template"})
			return
		}

		diagramTemplateIDs = append(diagramTemplateIDs, *diagramTemplate.ID)
	}

	// Close the diagram cursor
	_, err = tx.Exec("CLOSE diagram_cursor")
	if err != nil {
		if isDevelopmentEnvironment() {
			log.Printf("Err: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error closing diagram cursor"})
		return
	}

	projectTemplate.DiagramTemplates = &diagramTemplateIDs

	// Commit the transaction after successfully processing all documents

	if err := tx.Commit(); err != nil {
		log.Printf("Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit error"})
		return
	}

	responseStructure := gin.H{
		"message": "Project template created successfully",
		"data": gin.H{
			"id":                 *projectTemplate.ID,
			"user_id":            *projectTemplate.UserID,
			"tenant_id":          *projectTemplate.TenantID,
			"title":              *projectTemplate.Title,
			"description":        *projectTemplate.Description,
			"category":           *projectTemplate.Category,
			"document_templates": *projectTemplate.DocumentTemplates,
			"diagram_templates":  *projectTemplate.DiagramTemplates,
			"complexity":         *projectTemplate.Complexity,
			"created_at":         *projectTemplate.CreatedAt,
			"updated_at":         *projectTemplate.UpdatedAt,
			"privacy":            *projectTemplate.Privacy,
			"first_name":         projectTemplate.FirstName,
			"last_name":          projectTemplate.LastName,
			"user_picture":       projectTemplate.UserPicture,
		},
	}
	// Respond with the ID of the newly created project template
	c.JSON(http.StatusCreated, responseStructure)
}

func GetProjectTemplate(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	// Retrieve the templateID from the query parameters
	templateID := c.Query("template_id")
	if templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template ID must be provided"})
		return
	}

	// Query for the project template
	var projectTemplate models.ProjectTemplate
	err := tenantManagement.DB.QueryRow(`
    SELECT
        pt.id,
        pt.user_id,
        pt.tenant_id,
        pt.title,
        pt.description,
        pt.complexity,
        pt.category,
        pt.created_at,
        pt.privacy,
        pt.updated_at,
        pt.public_template_ref,
        u.first_name,
        u.last_name,
        u.user_picture
    FROM
        st_schema.project_templates pt
    LEFT JOIN
        st_schema.users u ON pt.user_id = u.id
    WHERE
        pt.id = $1
    AND
        pt.tenant_id = $2`,
		templateID, tenantID).Scan(
		&projectTemplate.ID,
		&projectTemplate.UserID,
		&projectTemplate.TenantID,
		&projectTemplate.Title,
		&projectTemplate.Description,
		&projectTemplate.Complexity,
		&projectTemplate.Category,
		&projectTemplate.CreatedAt,
		&projectTemplate.Privacy,
		&projectTemplate.UpdatedAt,
		&projectTemplate.PublicTemplateRef,
		&projectTemplate.FirstName,
		&projectTemplate.LastName,
		&projectTemplate.UserPicture)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project template not found"})
		} else {
			log.Printf("Database error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching project template"})
		}
		return
	}

	// Query for the associated document templates
	docRows, err := tenantManagement.DB.Query(`
		SELECT
			id, user_id, tenant_id, project_template_id, title, p_content_json, complexity, created_at, updated_at, privacy
		FROM
			st_schema.document_templates
		WHERE
			project_template_id = $1
		AND
			tenant_id = $2
	`, templateID, tenantID)

	if err != nil {
		if isDevelopmentEnvironment() {
			log.Printf("Database error: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching document templates"})
		return
	}
	defer docRows.Close()

	documentTemplates := make([]models.DocumentTemplate, 0)
	for docRows.Next() {
		var docTemplate models.DocumentTemplate
		if err := docRows.Scan(
			&docTemplate.ID,
			&docTemplate.UserID,
			&docTemplate.TenantID,
			&docTemplate.ProjectTemplateID,
			&docTemplate.Title,
			&docTemplate.Content,
			&docTemplate.Complexity,
			&docTemplate.CreatedAt,
			&docTemplate.UpdatedAt,
			&docTemplate.Privacy); err != nil {
			log.Printf("Row parsing error: %v", err)
			continue
		}
		documentTemplates = append(documentTemplates, docTemplate)
	}

	// Query for the associated diagram templates
	diagramRows, err := tenantManagement.DB.Query(`
		SELECT
			id, user_id, tenant_id, project_template_id, title, diagram_type, diagram_status, category, p_design, created_at, updated_at
		FROM
			st_schema.diagram_templates
		WHERE
			project_template_id = $1
		AND
			tenant_id = $2
	`, templateID, tenantID)

	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching diagram templates"})
		return
	}
	defer diagramRows.Close()

	diagramTemplates := make([]models.DiagramTemplate, 0)
	for diagramRows.Next() {
		var diagTemplate models.DiagramTemplate
		if err := diagramRows.Scan(
			&diagTemplate.ID,
			&diagTemplate.UserID,
			&diagTemplate.TenantID,
			&diagTemplate.ProjectTemplateID,
			&diagTemplate.Title,
			&diagTemplate.DiagramType,
			&diagTemplate.DiagramStatus,
			&diagTemplate.Category,
			&diagTemplate.Design,
			&diagTemplate.CreatedAt,
			&diagTemplate.UpdatedAt); err != nil {
			if isDevelopmentEnvironment() {
				log.Printf("Row parsing error: %v", err)
			}
			continue
		}
		diagramTemplates = append(diagramTemplates, diagTemplate)
	}

	// Convert document and diagram template IDs to string slices for the response
	docTemplateIDs := make([]string, len(documentTemplates))
	for i, doc := range documentTemplates {
		docTemplateIDs[i] = *doc.ID
	}
	projectTemplate.DocumentTemplates = &docTemplateIDs

	diagTemplateIDs := make([]string, len(diagramTemplates))
	for i, diag := range diagramTemplates {
		diagTemplateIDs[i] = *diag.ID
	}
	projectTemplate.DiagramTemplates = &diagTemplateIDs

	// Compile and send the response
	c.JSON(http.StatusOK, gin.H{
		"message": "Project template retrieved successfully",
		"data": gin.H{
			"project_template":   projectTemplate,
			"document_templates": documentTemplates,
			"diagram_templates":  diagramTemplates,
		},
	})
}

func GetProjectTemplates(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	var templates []models.ProjectTemplate

	// Fetch all project templates for the user
	projectRows, err := tenantManagement.DB.Query(`
		SELECT
			pt.id,
			pt.user_id,
			pt.tenant_id,
			pt.title,
			pt.description,
			pt.complexity,
			pt.category,
			pt.created_at,
			pt.privacy,
			pt.updated_at,
			u.first_name,
			u.last_name,
			u.user_picture
		FROM
			st_schema.project_templates pt
		LEFT JOIN
			st_schema.users u ON pt.user_id = u.id
		WHERE
			pt.tenant_id = $1
	`, tenantID)
	if err != nil {
		if isDevelopmentEnvironment() {
			log.Printf("Database error: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching project templates"})
		return
	}
	defer projectRows.Close()

	for projectRows.Next() {
		var project models.ProjectTemplate
		err := projectRows.Scan(
			&project.ID,
			&project.UserID,
			&project.TenantID,
			&project.Title,
			&project.Description,
			&project.Complexity,
			&project.Category,
			&project.CreatedAt,
			&project.Privacy,
			&project.UpdatedAt,
			&project.FirstName,
			&project.LastName,
			&project.UserPicture)
		if err != nil {
			if isDevelopmentEnvironment() {
				log.Printf("Row parsing error: %v", err)
			}
			continue
		}

		// Initialize DocumentTemplates as an empty slice to avoid nil dereference
		project.DocumentTemplates = &[]string{}

		// For each project template, fetch associated document template IDs
		docIDs, docErr := tenantManagement.DB.Query(`
			SELECT
				id
			FROM
				st_schema.document_templates
			WHERE
				project_template_id = $1
		`, *project.ID)
		if docErr != nil {
			if isDevelopmentEnvironment() {
				log.Printf("Database error fetching document templates: %v", docErr)
			}
			continue // or handle as needed
		}

		var docID string
		for docIDs.Next() {
			if err := docIDs.Scan(&docID); err != nil {
				if isDevelopmentEnvironment() {
					log.Printf("Row parsing error: %v", err)
				}
				continue // or handle as needed
			}
			*project.DocumentTemplates = append(*project.DocumentTemplates, docID)
		}
		docIDs.Close()

		templates = append(templates, project)
	}

	if err := projectRows.Err(); err != nil {
		if isDevelopmentEnvironment() {
			log.Printf("Row iteration error: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating project templates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Project templates fetched successfully",
		"data":    templates,
	})
}

func DeleteProjectTemplate(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	// Extract templateID from path
	templateID := c.Query("template_id")
	if templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template ID must be provided"})
		return
	}

	// Fetch the project template to be deleted including its document templates
	var projectTemplate models.ProjectTemplate
	err := tenantManagement.DB.QueryRow(`
        SELECT
			id, user_id, tenant_id, title, description, complexity, category, created_at, updated_at, privacy
        FROM
			st_schema.project_templates
        WHERE
			id = $1
		AND
			tenant_id = $2`,
		templateID, tenantID).Scan(
		&projectTemplate.ID,
		&projectTemplate.UserID,
		&projectTemplate.TenantID,
		&projectTemplate.Title,
		&projectTemplate.Description,
		&projectTemplate.Complexity,
		&projectTemplate.Category,
		&projectTemplate.CreatedAt,
		&projectTemplate.UpdatedAt,
		&projectTemplate.Privacy,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project template not found"})
			return
		}
		if isDevelopmentEnvironment() {
			log.Printf("Err: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching project template data"})
		return
	}

	// Fetch associated document template IDs
	docRows, err := tenantManagement.DB.Query(`
        SELECT
			id
		FROM
			st_schema.document_templates
		WHERE
			project_template_id = $1
    `, templateID)
	if err != nil {
		if isDevelopmentEnvironment() {
			log.Printf("Database error: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching document templates"})
		return
	}
	defer docRows.Close()

	docTemplateIDs := make([]string, 0)
	for docRows.Next() {
		var docID string
		if err := docRows.Scan(&docID); err != nil {
			if isDevelopmentEnvironment() {
				log.Printf("Row parsing error: %v", err)
			}
			continue
		}
		docTemplateIDs = append(docTemplateIDs, docID)
	}
	projectTemplate.DocumentTemplates = &docTemplateIDs

	// Fetch associated diagram template IDs
	diagRows, err := tenantManagement.DB.Query(`
        SELECT
			id
		FROM
			st_schema.diagram_templates
		WHERE
			project_template_id = $1
    `, templateID)
	if err != nil {
		if isDevelopmentEnvironment() {
			log.Printf("Database error: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching diagram templates"})
		return
	}
	defer diagRows.Close()

	diagTemplateIDs := make([]string, 0)
	for diagRows.Next() {
		var diagID string
		if err := diagRows.Scan(&diagID); err != nil {
			if isDevelopmentEnvironment() {
				log.Printf("Row parsing error: %v", err)
			}
			continue
		}
		diagTemplateIDs = append(diagTemplateIDs, diagID)
	}
	projectTemplate.DiagramTemplates = &diagTemplateIDs

	// Perform the delete operation
	// The ON DELETE CASCADE will handle the associated templates
	_, err = tenantManagement.DB.Exec(`
        DELETE FROM
			st_schema.project_templates
		WHERE
			id = $1
		AND
			tenant_id = $2
    `, templateID, tenantID)
	if err != nil {
		if isDevelopmentEnvironment() {
			log.Printf("Database error: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting project template"})
		return
	}

	// If everything went fine, return the data of the deleted project template
	c.JSON(http.StatusOK, gin.H{
		"message": "Project template deleted successfully",
		"data": gin.H{
			"project_template": projectTemplate,
			"document_ids":     docTemplateIDs,
			"diagram_ids":      diagTemplateIDs,
		},
	})
}

// Expects the following query parameters:
// - template_id: The ID of the project template to update.
// Expects a JSON body with any of the following fields for partial updates:
// - title
// - description
// - category
// - complexity
// - privacy
func UpdateProjectTemplate(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	// Retrieve the templateID from the query parameters
	templateID := c.Query("template_id")
	if templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template ID is required"})
		return
	}

	// Bind the request body to a struct to handle partial updates
	var updateData struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		Category    *string `json:"category"`
		Complexity  *string `json:"complexity"`
		Privacy     *bool   `json:"privacy"`
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
	if updateData.Category != nil {
		setParts = append(setParts, fmt.Sprintf("category = $%d", argCounter))
		args = append(args, *updateData.Category)
		argCounter++
	}
	if updateData.Complexity != nil {
		setParts = append(setParts, fmt.Sprintf("complexity = $%d", argCounter))
		args = append(args, *updateData.Complexity)
		argCounter++
	}
	if updateData.Privacy != nil {
		setParts = append(setParts, fmt.Sprintf("privacy = $%d", argCounter))
		args = append(args, *updateData.Privacy)
		argCounter++
	}

	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No updatable fields provided"})
		return
	}

	setClause := strings.Join(setParts, ", ")
	query := fmt.Sprintf(`
        UPDATE
			st_schema.project_templates
        SET
			%s, updated_at = NOW()
        WHERE
			tenant_id = $%d
		AND
			id = $%d
        RETURNING
			id, user_id, tenant_id, title, description, category, complexity, privacy, created_at, updated_at
    `, setClause, argCounter, argCounter+1)

	args = append(args, tenantID, templateID)

	// Use sql.NullString for nullable fields
	var (
		id, dbUserID, dbTenantID, title, description, category, complexity sql.NullString
		privacy                                                            sql.NullBool
		createdAt, updatedAt                                               sql.NullTime
	)

	err := tenantManagement.DB.QueryRow(query, args...).Scan(
		&id, &dbUserID, &dbTenantID, &title, &description, &category, &complexity, &privacy, &createdAt, &updatedAt,
	)

	if err != nil {
		if isDevelopmentEnvironment() {
			log.Printf(models.DatabaseError, err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update the project template"})
		return
	}

	// Create a map to store non-null values
	updatedTemplate := map[string]interface{}{
		"id":         id.String,
		"user_id":    dbUserID.String,
		"tenant_id":  dbTenantID.String,
		"created_at": createdAt.Time,
		"updated_at": updatedAt.Time,
	}

	if title.Valid {
		updatedTemplate["title"] = title.String
	}
	if description.Valid {
		updatedTemplate["description"] = description.String
	}
	if category.Valid {
		updatedTemplate["category"] = category.String
	}
	if complexity.Valid {
		updatedTemplate["complexity"] = complexity.String
	}
	if privacy.Valid {
		updatedTemplate["privacy"] = privacy.Bool
	}

	// Return the updated project template data
	c.JSON(http.StatusOK, gin.H{
		"data":    updatedTemplate,
		"message": "Project template updated successfully",
	})
}

func UpdatePublicProjectTemplate(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	// Get template ID from query parameters
	templateID := c.Query("cm_template_id")
	if templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template ID is required"})
		return
	}

	// Define update data structure
	var updateData struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		Category    *string `json:"category"`
		Complexity  *string `json:"complexity"`
		Version     *string `json:"version"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build dynamic update query
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
	if updateData.Category != nil {
		setParts = append(setParts, fmt.Sprintf("category = $%d", argCounter))
		args = append(args, *updateData.Category)
		argCounter++
	}
	if updateData.Complexity != nil {
		setParts = append(setParts, fmt.Sprintf("complexity = $%d", argCounter))
		args = append(args, *updateData.Complexity)
		argCounter++
	}
	if updateData.Version != nil {
		setParts = append(setParts, fmt.Sprintf("version = $%d", argCounter))
		args = append(args, *updateData.Version)
		argCounter++
	}

	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No updatable fields provided"})
		return
	}

	setClause := strings.Join(setParts, ", ")
	query := fmt.Sprintf(`
        UPDATE st_schema.cm_project_templates
        SET %s
        WHERE id = $%d AND tenant_id = $%d
        RETURNING id, user_id, tenant_id, title, description, category, complexity, version
    `, setClause, argCounter, argCounter+1)

	args = append(args, templateID, tenantID)

	var (
		id, dbUserID, dbTenantID, title, description, category, complexity, version sql.NullString
	)

	err := tenantManagement.DB.QueryRow(query, args...).Scan(
		&id, &dbUserID, &dbTenantID, &title, &description, &category, &complexity, &version,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project template not found or you don't have permission to update it"})
		} else {
			log.Printf("Error updating public project template: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update the project template"})
		}
		return
	}

	updatedTemplate := map[string]interface{}{
		"id":        id.String,
		"user_id":   dbUserID.String,
		"tenant_id": dbTenantID.String,
	}

	if title.Valid {
		updatedTemplate["title"] = title.String
	}
	if description.Valid {
		updatedTemplate["description"] = description.String
	}
	if category.Valid {
		updatedTemplate["category"] = category.String
	}
	if complexity.Valid {
		updatedTemplate["complexity"] = complexity.String
	}
	if version.Valid {
		updatedTemplate["version"] = version.String
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    updatedTemplate,
		"message": "Public project template updated successfully",
	})
}
