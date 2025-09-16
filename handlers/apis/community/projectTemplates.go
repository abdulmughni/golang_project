package community

// This package contains the handlers for the community project templates.

// The handlers are:
// 1. PublishProjectTemplate - Publishes a user-created project template to the community including all of the associated documents.
// 2. UnpublishProjectTemplate - Removes a published project template and its associated documents from the community.
//    It also updates the privacy status of the project template to true. (When value is set to false the project template is public)
// 3. GetPublicProjectTemplate - Retrieves a single public project template by ID.
// 4. GetPublicProjectTemplates - Retrieves all public project templates.

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"sententiawebapi/handlers/apis/images"
	"sententiawebapi/handlers/apis/tenantManagement"
	"sententiawebapi/handlers/models"
	"sententiawebapi/utilities"

	"github.com/gin-gonic/gin"
)

// THis file contains all the handlers for the community design templates

// PublishProjectTemplate publishes a user-created project template to the community.
func PublishProjectTemplate(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	objectId := c.Query("template_id")
	if objectId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Object ID is required"})
		return
	}

	tx, err := tenantManagement.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database transaction start error"})
		return
	}
	defer tx.Rollback()

	// Fetch the project template data
	var projectTemplate models.ProjectTemplate
	var publicProjectTemplate models.PublicProjectTemplate

	if err := c.ShouldBindJSON(&publicProjectTemplate); err != nil {
		log.Printf("JSON binding error in PublishProjectTemplate: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})

		return
	}

	// Fetch the project template data
	publicProjectTemplate.UserID = &userID
	publicProjectTemplate.TenantID = &tenantID
	publicProjectTemplate.ProjectTemplateID = &objectId

	err = tx.QueryRow(`
    SELECT
        id, user_id, tenant_id, title, description, complexity, category, document_templates
    FROM
        st_schema.project_templates
    WHERE
        id = $1
	AND
		tenant_id = $2`, objectId, tenantID).Scan(
		&projectTemplate.ID,
		&projectTemplate.UserID,
		&projectTemplate.TenantID,
		&projectTemplate.Title,
		&projectTemplate.Description,
		&projectTemplate.Complexity,
		&projectTemplate.Category,
		&projectTemplate.DocumentTemplates)

	if err != nil {
		log.Printf("Error retrieving project template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve project template: " + err.Error()})
		return
	}

	// Publish the project template into the public templates table
	err = tx.QueryRow(`
        INSERT INTO
            st_schema.cm_project_templates
            (project_template_id, user_id, tenant_id, version, title, category, description, complexity)
        VALUES
            ($1, $2, $3, '1.0.0', $4, $5, $6, $7)
        RETURNING id, published_at, last_update_at`,
		publicProjectTemplate.ProjectTemplateID,
		publicProjectTemplate.UserID,
		publicProjectTemplate.TenantID,
		publicProjectTemplate.Title,
		publicProjectTemplate.Category,
		publicProjectTemplate.Description,
		publicProjectTemplate.Complexity).Scan(&publicProjectTemplate.ID, &publicProjectTemplate.PublishedAt, &publicProjectTemplate.LastUpdateAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish project template: " + err.Error()})
		return
	}

	// Declare and open a cursor for related documents
	_, err = tx.Exec(`
        DECLARE doc_cursor CURSOR FOR
        SELECT
            id, title, p_content_json, p_raw_content, complexity, description, category
        FROM
            st_schema.document_templates
        WHERE
            project_template_id = $1
        AND
            tenant_id = $2`, objectId, tenantID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error declaring cursor: " + err.Error()})
		return
	}

	// Iterate through documents using cursor
	prevDocsIDs := make([]string, 0)
	docTemplateIDs := make([]string, 0)
	for {
		var docID, title string
		var content *json.RawMessage
		var rawContent []byte
		var complexity, description, category sql.NullString

		err = tx.QueryRow("FETCH NEXT FROM doc_cursor").Scan(
			&docID,
			&title,
			&content,
			&rawContent,
			&complexity,
			&description,
			&category)

		if err != nil {
			if err == sql.ErrNoRows {
				break // No more rows, break the loop
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching document from cursor: " + err.Error()})
			return
		}

		// Correct insertion into cm_document_templates table
		var newDocTemplateID string
		err = tx.QueryRow(`
        INSERT INTO
            st_schema.cm_document_templates
            (
				community_project_template_id, user_id, tenant_id, title,
				content_json, raw_content, p_content_json, p_raw_content,
				complexity, category, description
			)
        VALUES
            ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        RETURNING id`,
			publicProjectTemplate.ID, // This should be the ID of the community project template
			userID,
			tenantID,
			title,
			content,
			rawContent,
			content,
			rawContent,
			complexity.String,
			category,
			description).Scan(&newDocTemplateID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert document into public templates: " + err.Error()})
			return
		}

		prevDocsIDs = append(prevDocsIDs, docID)
		docTemplateIDs = append(docTemplateIDs, newDocTemplateID)
	}

	publicProjectTemplate.DocumentTemplates = &docTemplateIDs
	// Close cursor after iteration
	_, err = tx.Exec("CLOSE doc_cursor")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error closing cursor: " + err.Error()})
		return
	}

	err = images.CopyProjectFiles(context.Background(), tx, images.CopyProjectParams{
		SourceDocuments: images.DocumentsRef{
			Type: models.ResourceGroupTemplate,
			IDs:  prevDocsIDs,
		},
		DestinationDocuments: images.DocumentsRef{
			Type: models.ResourceGroupCommunity,
			IDs:  docTemplateIDs,
		},
		TenantID: tenantID,
	})
	if err != nil {
		log.Printf("Error while copying files: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error copying files"})
		return
	}

	// Handle diagram templates
	_, err = tx.Exec(`
        DECLARE diagram_cursor CURSOR FOR
        SELECT
            id, title, diagram_type, diagram_status, category, p_design, p_raw_design
        FROM
            st_schema.diagram_templates
        WHERE
            project_template_id = $1
        AND
            tenant_id = $2`, objectId, tenantID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error declaring diagram cursor: " + err.Error()})
		return
	}

	// Iterate through diagrams using cursor
	diagTemplateIDs := make([]string, 0)
	for {
		var diagID, title string
		var design *json.RawMessage
		var rawDesign []byte
		var diagramType, diagramStatus, category sql.NullString

		err = tx.QueryRow("FETCH NEXT FROM diagram_cursor").Scan(
			&diagID,
			&title,
			&diagramType,
			&diagramStatus,
			&category,
			&design,
			&rawDesign)

		if err != nil {
			if err == sql.ErrNoRows {
				break // No more diagrams
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching diagram from cursor: " + err.Error()})
			return
		}

		// Insert into cm_diagram_templates
		var newDiagTemplateID string
		err = tx.QueryRow(`
        INSERT INTO
            st_schema.cm_diagram_templates
            (
				community_project_template_id, user_id, tenant_id, title, diagram_type, diagram_status, category,
				design, raw_design, p_design, p_raw_design
			)
        VALUES
            ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        RETURNING id`,
			publicProjectTemplate.ID,
			userID,
			tenantID,
			title,
			diagramType.String,
			diagramStatus.String,
			category.String,
			design,
			rawDesign,
			design,
			rawDesign,
		).Scan(&newDiagTemplateID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert diagram into public templates: " + err.Error()})
			return
		}

		diagTemplateIDs = append(diagTemplateIDs, newDiagTemplateID)
	}

	// Close diagram cursor
	_, err = tx.Exec("CLOSE diagram_cursor")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error closing diagram cursor: " + err.Error()})
		return
	}

	publicProjectTemplate.DiagramTemplates = &diagTemplateIDs

	// Update the project template with the privacy status set to false

	result, err := tx.Exec(`
    UPDATE
        st_schema.project_templates
    SET
        privacy = false
    WHERE
        id = $1
    AND
        tenant_id = $2`, projectTemplate.ID, projectTemplate.TenantID)

	if err != nil {
		log.Printf("Error updating private project template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update private project template: " + err.Error()})
		return
	}

	// Check if the row was actually updated
	affectedRows, err := result.RowsAffected()
	if err != nil || affectedRows == 0 {
		log.Printf("No rows were updated, privacy might not have been set to false as expected")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update privacy of the project template"})
		return
	}

	// Inject the generated public project template ID into the project template
	pubIdresult, err := tx.Exec(`
    UPDATE
        st_schema.project_templates
    SET
        public_template_ref = $1
    WHERE
        id = $2
    AND
        tenant_id = $3`, *publicProjectTemplate.ID, projectTemplate.ID, projectTemplate.TenantID)

	if err != nil {
		log.Printf("Error updating private project template with public ID: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update private project template: " + err.Error()})
		return
	}

	// Check if the row was actually updated
	pubAffectedRows, err := pubIdresult.RowsAffected()
	if err != nil || pubAffectedRows == 0 {
		log.Printf("No rows were updated, privacy might not have been set to false as expected")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update with public ID project template"})
		return
	}

	err = tx.QueryRow(`
        SELECT
            privacy
        FROM
            st_schema.project_templates
        WHERE id = $1
        AND
            tenant_id = $2`,
		projectTemplate.ID,
		projectTemplate.TenantID).Scan(&projectTemplate.Privacy)

	if err != nil {
		log.Printf("Error retrieving updated privacy value: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve updated privacy value"})
		return
	}
	// Commit the transaction
	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit failed: " + err.Error()})
		return
	}

	// Send successful response
	responseStructure := gin.H{
		"message": "Project template published successfully",
		"data": gin.H{
			"id":                  *publicProjectTemplate.ID,
			"project_template_id": *publicProjectTemplate.ProjectTemplateID,
			"user_id":             *publicProjectTemplate.UserID,
			"tenant_id":           *publicProjectTemplate.TenantID,
			"title":               *publicProjectTemplate.Title,
			"description":         *publicProjectTemplate.Description,
			"complexity":          *publicProjectTemplate.Complexity,
			"category":            *publicProjectTemplate.Category,
			"privacy":             *projectTemplate.Privacy,
			"created_at":          *publicProjectTemplate.PublishedAt,
			"updated_at":          *publicProjectTemplate.LastUpdateAt,
			"document_templates":  *publicProjectTemplate.DocumentTemplates,
			"diagram_templates":   *publicProjectTemplate.DiagramTemplates,
		},
	}
	c.JSON(http.StatusCreated, responseStructure)
}

// UnpublishProjectTemplate removes a published project template and its associated documents from the community.
func UnpublishProjectTemplate(c *gin.Context) {

	// Extract userID from the authenticated user's context
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	templateID := c.Query("template_id")
	if templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template ID is required"})
		return
	}

	// Start the database transaction
	tx, err := tenantManagement.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database transaction start error"})
		return
	}
	defer tx.Rollback()

	log.Printf("Attempting to unpublish template: %s for tenant: %s", templateID, tenantID)

	// Retrieve data about the project template to be deleted
	var publicProjectTemplate models.PublicProjectTemplate
	err = tx.QueryRow(`
        SELECT
			id, project_template_id, user_id, tenant_id, version, category, description, complexity, published_at, last_update_at, title
        FROM
			st_schema.cm_project_templates
        WHERE
			id = $1
		AND
			tenant_id = $2`, templateID, tenantID).Scan(
		&publicProjectTemplate.ID,
		&publicProjectTemplate.ProjectTemplateID,
		&publicProjectTemplate.UserID,
		&publicProjectTemplate.TenantID,
		&publicProjectTemplate.Version,
		&publicProjectTemplate.Category,
		&publicProjectTemplate.Description,
		&publicProjectTemplate.Complexity,
		&publicProjectTemplate.PublishedAt,
		&publicProjectTemplate.LastUpdateAt,
		&publicProjectTemplate.Title)

	if err != nil {
		log.Printf("Error retrieving project template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve project template"})
		return
	}

	// Retrieve document template IDs
	rows, err := tx.Query(`
        SELECT
			id
        FROM
			st_schema.cm_document_templates
        WHERE
			community_project_template_id = $1`, templateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve document templates"})
		return
	}
	defer rows.Close()

	documentIDs := make([]string, 0)
	for rows.Next() {
		var docID string
		if err := rows.Scan(&docID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning document IDs"})
			return
		}
		documentIDs = append(documentIDs, docID)
	}

	// Delete document templates
	_, err = tx.Exec(`
        DELETE FROM
			st_schema.cm_document_templates
        WHERE
			community_project_template_id = $1 AND tenant_id = $2`, templateID, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete document templates"})
		return
	}

	// Delete the project template
	_, err = tx.Exec(`
        DELETE FROM
			st_schema.cm_project_templates
        WHERE
			id = $1
		AND
			tenant_id = $2`, templateID, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project template"})
		return
	}

	stmt := `
		UPDATE st_schema.project_templates
		SET privacy = true, public_template_ref = NULL
		WHERE id = $1 AND tenant_id = $2`

	result, err := tx.Exec(stmt, publicProjectTemplate.ProjectTemplateID, tenantID)

	if err != nil {
		log.Printf("Error updating private project template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update private project template: " + err.Error()})
		return
	}

	// Check if the row was actually updated
	rowUpdate, err := result.RowsAffected()
	if err != nil || rowUpdate == 0 {
		log.Printf("Failed to update public_template_ref to null")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update with public ID project template"})
		return
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit failed"})
		return
	}

	// Construct response with the data of the deleted project template and its document templates
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"category":           publicProjectTemplate.Category,
			"complexity":         publicProjectTemplate.Complexity,
			"created_at":         publicProjectTemplate.CreatedAt,
			"description":        publicProjectTemplate.Description,
			"id":                 publicProjectTemplate.ID,
			"title":              publicProjectTemplate.Title,
			"updated_at":         publicProjectTemplate.LastUpdateAt,
			"user_id":            publicProjectTemplate.UserID,
			"tenant_id":          publicProjectTemplate.TenantID,
			"document_templates": documentIDs,
		},
		"message": "Public project template successfully deleted",
	})
}

// This is the in application get function. It returns details about a single template.
// It returns the data only to authenticated users. If the user is not
// Authenticated, it returns an error message.
func GetPublicProjectTemplate(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	projectTemplateID := c.Query("public_template_id")
	if projectTemplateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project template ID is required"})
		return
	}

	log.Printf("Fetching public template with ID: %s", projectTemplateID)

	var template models.PublicProjectTemplate

	// Prepare the SELECT statement for fetching the project template
	stmt, err := tenantManagement.DB.Prepare(`
        SELECT
            pt.id,
            pt.project_template_id,
            pt.user_id,
			pt.tenant_id,
            pt.version,
            pt.category,
            pt.description,
            pt.complexity,
            pt.published_at,
            pt.last_update_at,
            pt.title,
            u.first_name,
            u.last_name,
            u.user_picture
        FROM
            st_schema.cm_project_templates pt
        LEFT JOIN
            st_schema.users u ON pt.user_id = u.id
        WHERE
            pt.id = $1`)
	if err != nil {
		log.Printf("Prepare Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer stmt.Close()

	// Execute the query and check for no rows
	err = stmt.QueryRow(projectTemplateID).Scan(
		&template.ID,
		&template.ProjectTemplateID,
		&template.UserID,
		&template.TenantID,
		&template.Version,
		&template.Category,
		&template.Description,
		&template.Complexity,
		&template.PublishedAt,
		&template.LastUpdateAt,
		&template.Title,
		&template.FirstName,
		&template.LastName,
		&template.UserPicture)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No template found with ID: %s", projectTemplateID)
			c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
			return
		}
		log.Printf("Query Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve public project template"})
		return
	}

	log.Printf("Found template: %+v", template)

	// Prepare the SELECT statement for fetching associated document templates
	docStmt, err := tenantManagement.DB.Prepare(`
        SELECT
			id, user_id, tenant_id, community_project_template_id, title, p_content_json, complexity, published_at
        FROM
			st_schema.cm_document_templates
        WHERE
			community_project_template_id = $1`)
	if err != nil {
		log.Printf("Prepare Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer docStmt.Close()

	// Execute the query to retrieve document templates
	rows, err := docStmt.Query(projectTemplateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve document templates: " + err.Error()})
		return
	}
	defer rows.Close()

	publicDocumentTemplates := make([]models.PublicDocumentTemplate, 0)

	for rows.Next() {
		var documentTemplate models.PublicDocumentTemplate

		if err := rows.Scan(
			&documentTemplate.ID,
			&documentTemplate.UserID,
			&documentTemplate.TenantID,
			&documentTemplate.ProjectTemplateID,
			&documentTemplate.Title,
			&documentTemplate.Content,
			&documentTemplate.Complexity,
			&documentTemplate.PublishedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning document IDs: " + err.Error()})
			return
		}
		publicDocumentTemplates = append(publicDocumentTemplates, documentTemplate)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating document templates"})
		return
	}

	// Prepare the SELECT statement for fetching associated diagram templates
	diagStmt, err := tenantManagement.DB.Prepare(`
        SELECT
            id, 
			user_id,
			tenant_id,
            community_project_template_id,
            title, 
            diagram_type, 
            category, 
            diagram_status, 
            p_design, 
            published_at, 
            last_update_at
        FROM
            st_schema.cm_diagram_templates
        WHERE
            community_project_template_id = $1
    `)
	if err != nil {
		log.Printf("Prepare Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer diagStmt.Close()

	// Execute the query to retrieve diagram templates
	diagRows, err := diagStmt.Query(projectTemplateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve diagram templates"})
		return
	}
	defer diagRows.Close()

	publicDiagramTemplates := make([]models.PublicDiagramTemplate, 0)

	for diagRows.Next() {
		var diagTemplate models.PublicDiagramTemplate

		if err := diagRows.Scan(
			&diagTemplate.ID,
			&diagTemplate.UserID,
			&diagTemplate.TenantID,
			&diagTemplate.CommunityProjectTemplateID,
			&diagTemplate.Title,
			&diagTemplate.DiagramType,
			&diagTemplate.Category,
			&diagTemplate.DiagramStatus,
			&diagTemplate.Design,
			&diagTemplate.CreatedAt,
			&diagTemplate.UpdatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning diagram templates"})
			return
		}
		publicDiagramTemplates = append(publicDiagramTemplates, diagTemplate)
	}

	if err := diagRows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating diagram templates"})
		return
	}

	// Update the response to include diagram templates
	if *template.TenantID == tenantID {
		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"document_templates": publicDocumentTemplates,
				"diagram_templates":  publicDiagramTemplates,
				"project_template": gin.H{
					"id":                  template.ID,
					"project_template_id": template.ProjectTemplateID,
					"user_id":             template.UserID,
					"tenant_id":           template.TenantID,
					"title":               template.Title,
					"description":         template.Description,
					"category":            template.Category,
					"complexity":          template.Complexity,
					"created_at":          template.PublishedAt,
					"last_update_at":      template.LastUpdateAt,
					"version":             template.Version,
					"updated_at":          template.PublishedAt,
					"first_name":          template.FirstName,
					"last_name":           template.LastName,
					"user_picture":        template.UserPicture,
				},
			},
			"message": "Public project template retrieved successfully",
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"document_templates": publicDocumentTemplates,
				"diagram_templates":  publicDiagramTemplates,
				"project_template": gin.H{
					"id":             template.ID,
					"user_id":        template.UserID,
					"tenant_id":      template.TenantID,
					"title":          template.Title,
					"description":    template.Description,
					"category":       template.Category,
					"complexity":     template.Complexity,
					"created_at":     template.PublishedAt,
					"last_update_at": template.LastUpdateAt,
					"version":        template.Version,
					"updated_at":     template.PublishedAt,
					"first_name":     template.FirstName,
					"last_name":      template.LastName,
					"user_picture":   template.UserPicture,
				},
			},
			"message": "Public project template retrieved successfully",
		})
	}
}

// This is the in application get function.
// It returns the data only to authenticated users. If the user is not
// Authenticated, it returns an error message.
func GetPublicProjectTemplates(c *gin.Context) {
	category := c.Query("category")

	query := `
        SELECT
            pt.id,
            pt.project_template_id,
            pt.user_id,
			pt.tenant_id,
            pt.title,
            pt.description,
            pt.category,
            pt.complexity,
            pt.published_at,
            u.first_name,
            u.last_name,
            u.user_picture
        FROM
            st_schema.cm_project_templates pt
        LEFT JOIN
            st_schema.users u ON pt.user_id = u.id
    `

	// Modify the query based on the presence of the category parameter
	if category != "" {
		query += " WHERE pt.category = $1"
	}

	var rows *sql.Rows
	var err error

	if category != "" {
		rows, err = tenantManagement.DB.Query(query, category)
	} else {
		rows, err = tenantManagement.DB.Query(query)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve public project templates: " + err.Error()})
		return
	}
	defer rows.Close()

	var templates []gin.H
	for rows.Next() {
		var template models.PublicProjectTemplate
		if err := rows.Scan(
			&template.ID,
			&template.ProjectTemplateID,
			&template.UserID,
			&template.TenantID,
			&template.Title,
			&template.Description,
			&template.Category,
			&template.Complexity,
			&template.PublishedAt,
			&template.FirstName,
			&template.LastName,
			&template.UserPicture,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning public project templates: " + err.Error()})
			return
		}

		// Fetching related document templates for each project template
		docQuery := `
            SELECT id
            FROM st_schema.cm_document_templates
            WHERE community_project_template_id = $1
        `
		docRows, docErr := tenantManagement.DB.Query(docQuery, template.ID)
		if docErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve document templates: " + docErr.Error()})
			return
		}
		var docIDs []string
		for docRows.Next() {
			var docID string
			if err := docRows.Scan(&docID); err != nil {
				docRows.Close()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning document template IDs: " + err.Error()})
				return
			}
			docIDs = append(docIDs, docID)
		}
		docRows.Close()

		// Fetch diagram templates for this project
		diagQuery := `
            SELECT id
            FROM st_schema.cm_diagram_templates
            WHERE community_project_template_id = $1
        `
		diagRows, diagErr := tenantManagement.DB.Query(diagQuery, template.ID)
		if diagErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve diagram templates: " + diagErr.Error()})
			return
		}
		var diagIDs []string
		for diagRows.Next() {
			var diagID string
			if err := diagRows.Scan(&diagID); err != nil {
				diagRows.Close()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning diagram template IDs: " + err.Error()})
				return
			}
			diagIDs = append(diagIDs, diagID)
		}
		diagRows.Close()

		templates = append(templates, gin.H{
			"id":                 template.ID,
			"user_id":            template.UserID,
			"tenant_id":          template.TenantID,
			"title":              template.Title,
			"description":        template.Description,
			"category":           template.Category,
			"complexity":         template.Complexity,
			"created_at":         template.CreatedAt,
			"published_at":       template.PublishedAt,
			"document_templates": docIDs,
			"diagram_templates":  diagIDs,
			"first_name":         template.FirstName,
			"last_name":          template.LastName,
			"user_picture":       template.UserPicture,
		})
	}

	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching public project templates: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    templates,
		"message": "Project resources retrieved successfully!",
	})
}

func ClonePublicProjectTemplate(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	pubProjTemplateId := c.Query("id")
	if pubProjTemplateId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Public template ID is required"})
		return
	}

	var pubProjTemplate models.ProjectTemplate

	stmt, err := tenantManagement.DB.Prepare(
		`SELECT
			id, title, description, category, complexity
        FROM
            st_schema.cm_project_templates
        WHERE
			id = $1`)
	if err != nil {
		log.Printf("Prepare Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer stmt.Close()

	err = stmt.QueryRow(pubProjTemplateId).Scan(
		&pubProjTemplate.ID,
		&pubProjTemplate.Title,
		&pubProjTemplate.Description,
		&pubProjTemplate.Category,
		&pubProjTemplate.Complexity,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve public project template: " + err.Error()})
		return
	}

	docStmt, err := tenantManagement.DB.Prepare(
		`SELECT
			id, community_project_template_id, title, content, raw_content, complexity
        FROM
			st_schema.cm_document_templates
        WHERE
			community_project_template_id = $1`)
	if err != nil {
		log.Printf("Prepare Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer docStmt.Close()

	rows, err := docStmt.Query(pubProjTemplateId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve document templates: " + err.Error()})
		return
	}
	defer rows.Close()

	pubDocTemplates := make([]models.DocumentTemplate, 0)

	for rows.Next() {
		var pubDocTemplate models.DocumentTemplate

		if err := rows.Scan(
			&pubDocTemplate.ID,
			&pubDocTemplate.ProjectTemplateID,
			&pubDocTemplate.Title,
			&pubDocTemplate.Content,
			&pubDocTemplate.RawContent,
			&pubDocTemplate.Complexity,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning document templates: " + err.Error()})
			return
		}
		pubDocTemplates = append(pubDocTemplates, pubDocTemplate)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating document templates"})
		return
	}

	tx, err := tenantManagement.DB.Begin()
	if err != nil {
		log.Printf("Transaction Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer tx.Rollback()

	// First, create the new project
	var newProjectID string
	err = tx.QueryRow(
		`INSERT INTO st_schema.projects
		(
			user_id, tenant_id, title, description, category, complexity, created_at, updated_at, status
		)
		VALUES
		(
			$1, $2, $3, $4, $5, $6, NOW(), NOW(), 'Not Started'
		)
		RETURNING id
		`, userID, tenantID, pubProjTemplate.Title, pubProjTemplate.Description, pubProjTemplate.Category, pubProjTemplate.Complexity).Scan(&newProjectID)

	if err != nil {
		log.Printf("Insert Project Err: %v", err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clone project"})
		return
	}

	// Then fetch and insert diagrams
	diagStmt, err := tx.Prepare(
		`SELECT
			id, title, diagram_type, diagram_status, category, design
		FROM
			st_schema.cm_diagram_templates
		WHERE
			community_project_template_id = $1`)
	if err != nil {
		log.Printf("Prepare Err: %v", err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer diagStmt.Close()

	diagRows, err := diagStmt.Query(pubProjTemplateId)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve diagram templates"})
		return
	}
	defer diagRows.Close()

	pubDiagTemplates := make([]models.DiagramTemplate, 0)
	for diagRows.Next() {
		var diagTemplate models.DiagramTemplate
		if err := diagRows.Scan(
			&diagTemplate.ID,
			&diagTemplate.Title,
			&diagTemplate.DiagramType,
			&diagTemplate.DiagramStatus,
			&diagTemplate.Category,
			&diagTemplate.Design,
		); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning diagram templates"})
			return
		}
		pubDiagTemplates = append(pubDiagTemplates, diagTemplate)
	}

	// Create response for diagram templates
	responseDiagramTemplates := make([]gin.H, 0, len(pubDiagTemplates))
	for _, diagTemplate := range pubDiagTemplates {
		responseDiagramTemplates = append(responseDiagramTemplates, gin.H{
			"id":             diagTemplate.ID,
			"user_id":        userID,
			"tenant_id":      tenantID,
			"project_id":     newProjectID,
			"title":          diagTemplate.Title,
			"diagram_type":   diagTemplate.DiagramType,
			"diagram_status": diagTemplate.DiagramStatus,
			"category":       diagTemplate.Category,
			"design":         diagTemplate.Design,
			"created_at":     time.Now().UTC().Format(time.RFC3339),
			"updated_at":     time.Now().UTC().Format(time.RFC3339),
		})
	}

	// Continue with document templates and other operations...

	if err := tx.Commit(); err != nil {
		log.Printf("Transaction Commit Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	responseDocumentTemplates := make([]gin.H, 0, len(pubDocTemplates))
	for _, docTemplate := range pubDocTemplates {
		responseDocumentTemplates = append(responseDocumentTemplates, gin.H{
			"id":                  docTemplate.ID,
			"user_id":             userID,
			"tenant_id":           tenantID,
			"project_template_id": newProjectID,
			"title":               docTemplate.Title,
			"complexity":          docTemplate.Complexity,
			"content":             docTemplate.Content,
			"created_at":          time.Now().UTC().Format(time.RFC3339),
			"updated_at":          time.Now().UTC().Format(time.RFC3339),
			"privacy":             true,
			"category":            nil,
			"description":         nil,
		})
	}

	responseProjectTemplate := gin.H{
		"id":                  newProjectID,
		"user_id":             userID,
		"tenant_id":           tenantID,
		"title":               pubProjTemplate.Title,
		"description":         pubProjTemplate.Description,
		"category":            pubProjTemplate.Category,
		"document_templates":  nil,
		"complexity":          pubProjTemplate.Complexity,
		"created_at":          time.Now().UTC().Format(time.RFC3339),
		"updated_at":          time.Now().UTC().Format(time.RFC3339),
		"privacy":             true,
		"status":              "draft",
		"public_template_ref": pubProjTemplate.ID,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Public project template cloned successfully",
		"data": gin.H{
			"project_template":   responseProjectTemplate,
			"document_templates": responseDocumentTemplates,
			"diagram_templates":  responseDiagramTemplates,
		},
	})
}

func GetWebPublicProjectTemplate(c *gin.Context) {
	projectTemplateID := c.Query("public_template_id")
	if projectTemplateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project template ID is required"})
		return
	}

	var template models.PublicProjectTemplate

	query := `
        SELECT
            pt.id,
            pt.project_template_id,
            pt.user_id,
			pt.tenant_id,
            pt.version,
            pt.category,
            pt.description,
            pt.complexity,
            pt.published_at,
            pt.last_update_at,
            pt.title,
            u.first_name,
            u.last_name,
            u.user_picture
        FROM
            st_schema.cm_project_templates pt
        LEFT JOIN
            st_schema.users u ON pt.user_id = u.id
        WHERE
            pt.id = $1`

	err := tenantManagement.DB.QueryRow(query, projectTemplateID).Scan(
		&template.ID,
		&template.ProjectTemplateID,
		&template.UserID,
		&template.TenantID,
		&template.Version,
		&template.Category,
		&template.Description,
		&template.Complexity,
		&template.PublishedAt,
		&template.LastUpdateAt,
		&template.Title,
		&template.FirstName,
		&template.LastName,
		&template.UserPicture)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve public project template: " + err.Error()})
		return
	}

	// Retrieve associated document templates
	rows, err := tenantManagement.DB.Query(`
        SELECT
			id, user_id, tenant_id, community_project_template_id, title, content, complexity, published_at
        FROM
			st_schema.cm_document_templates
        WHERE
			community_project_template_id = $1
    `, projectTemplateID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve document templates: " + err.Error()})
		return
	}
	defer rows.Close()

	publicDocumentTemplates := make([]models.PublicDocumentTemplate, 0)

	for rows.Next() {
		var documentTemplate models.PublicDocumentTemplate

		if err := rows.Scan(
			&documentTemplate.ID,
			&documentTemplate.UserID,
			&documentTemplate.TenantID,
			&documentTemplate.ProjectTemplateID,
			&documentTemplate.Title,
			&documentTemplate.Content,
			&documentTemplate.Complexity,
			&documentTemplate.PublishedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning document IDs: " + err.Error()})
			return
		}
		publicDocumentTemplates = append(publicDocumentTemplates, documentTemplate)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating document templates"})
		return
	}

	// Add diagram templates query
	diagStmt, err := tenantManagement.DB.Query(`
        SELECT
            id, 
            community_project_template_id,
            title, 
            diagram_type, 
            category, 
            diagram_status, 
            design, 
            published_at, 
            last_update_at
        FROM
            st_schema.cm_diagram_templates
        WHERE
            community_project_template_id = $1
    `, projectTemplateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve diagram templates"})
		return
	}
	defer diagStmt.Close()

	publicDiagramTemplates := make([]models.PublicDiagramTemplate, 0)

	for diagStmt.Next() {
		var diagTemplate models.PublicDiagramTemplate

		if err := diagStmt.Scan(
			&diagTemplate.ID,
			&diagTemplate.CommunityProjectTemplateID,
			&diagTemplate.Title,
			&diagTemplate.DiagramType,
			&diagTemplate.Category,
			&diagTemplate.DiagramStatus,
			&diagTemplate.Design,
			&diagTemplate.CreatedAt,
			&diagTemplate.UpdatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning diagram templates"})
			return
		}
		publicDiagramTemplates = append(publicDiagramTemplates, diagTemplate)
	}

	if err := diagStmt.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating diagram templates"})
		return
	}

	// Update the response to include diagram templates
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"document_templates": publicDocumentTemplates,
			"diagram_templates":  publicDiagramTemplates,
			"project_template": gin.H{
				"id":             template.ID,
				"user_id":        template.UserID,
				"tenant_id":      template.TenantID,
				"title":          template.Title,
				"description":    template.Description,
				"category":       template.Category,
				"complexity":     template.Complexity,
				"created_at":     template.PublishedAt,
				"last_update_at": template.LastUpdateAt,
				"version":        template.Version,
				"updated_at":     template.PublishedAt,
				"first_name":     template.FirstName,
				"last_name":      template.LastName,
				"user_picture":   template.UserPicture,
			},
		},
		"message": "Public project template retrieved successfully",
	})
}

func GetWebPublicProjectTemplates(c *gin.Context) {
	query := `
        SELECT
            pt.id,
			pt.project_template_id,
			pt.user_id,
			pt.tenant_id,
			pt.title,
			pt.description,
			pt.category,
			pt.complexity,
			pt.published_at,
			u.first_name,
			u.last_name,
			u.user_picture
        FROM
            st_schema.cm_project_templates pt
        LEFT JOIN
            st_schema.users u ON pt.user_id = u.id
    `
	category := c.Query("category")

	if category != "" {
		query += " WHERE pt.category = $1"
	}

	var rows *sql.Rows
	var err error

	if category != "" {
		rows, err = tenantManagement.DB.Query(query, category)
	} else {
		rows, err = tenantManagement.DB.Query(query)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve public project templates: " + err.Error()})
		return
	}
	defer rows.Close()

	var templates []gin.H
	for rows.Next() {
		var template models.PublicProjectTemplate
		if err := rows.Scan(
			&template.ID,
			&template.ProjectTemplateID,
			&template.UserID,
			&template.TenantID,
			&template.Title,
			&template.Description,
			&template.Category,
			&template.Complexity,
			&template.PublishedAt,
			&template.FirstName,
			&template.LastName,
			&template.UserPicture,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning public project templates: " + err.Error()})
			return
		}

		// Fetching related document templates for each project template
		docQuery := `
            SELECT id
            FROM st_schema.cm_document_templates
            WHERE community_project_template_id = $1
        `
		docRows, docErr := tenantManagement.DB.Query(docQuery, template.ID)
		if docErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve document templates: " + docErr.Error()})
			return
		}
		var docIDs []string
		for docRows.Next() {
			var docID string
			if err := docRows.Scan(&docID); err != nil {
				docRows.Close()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning document template IDs: " + err.Error()})
				return
			}
			docIDs = append(docIDs, docID)
		}
		docRows.Close()

		// Fetch diagram templates for this project
		diagQuery := `
            SELECT id
            FROM st_schema.cm_diagram_templates
            WHERE community_project_template_id = $1
        `
		diagRows, diagErr := tenantManagement.DB.Query(diagQuery, template.ID)
		if diagErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve diagram templates: " + diagErr.Error()})
			return
		}
		var diagIDs []string
		for diagRows.Next() {
			var diagID string
			if err := diagRows.Scan(&diagID); err != nil {
				diagRows.Close()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning diagram template IDs: " + err.Error()})
				return
			}
			diagIDs = append(diagIDs, diagID)
		}
		diagRows.Close()

		templates = append(templates, gin.H{
			"id":                 template.ID,
			"user_id":            template.UserID,
			"tenant_id":          template.TenantID,
			"title":              template.Title,
			"description":        template.Description,
			"category":           template.Category,
			"complexity":         template.Complexity,
			"created_at":         template.CreatedAt,
			"published_at":       template.PublishedAt,
			"document_templates": docIDs,
			"diagram_templates":  diagIDs,
			"first_name":         template.FirstName,
			"last_name":          template.LastName,
			"user_picture":       template.UserPicture,
		})
	}

	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching public project templates: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    templates,
		"message": "Project resources retrieved successfully!",
	})
}

func GetWebPublicProjectTemplatesPagination(c *gin.Context) {
	// Add pagination parameters
	page := c.DefaultQuery("page", "1")
	pageSize := 4 // Fixed page size of 4 items

	pageNum, err := strconv.Atoi(page)
	if err != nil || pageNum < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
		return
	}

	offset := (pageNum - 1) * pageSize

	// Modified base query to include first diagram's data using LEFT JOIN and DISTINCT ON
	query := `
        SELECT
            pt.id,
            pt.project_template_id,
            pt.user_id,
            pt.tenant_id,
            pt.title,
            pt.description,
            pt.category,
            pt.complexity,
            pt.published_at,
            u.first_name,
            u.last_name,
            u.user_picture,
            COUNT(*) OVER() as total_count,
            dt.id as diagram_id,
            dt.design as diagram_design
        FROM
            st_schema.cm_project_templates pt
        LEFT JOIN
            st_schema.users u ON pt.user_id = u.id
        LEFT JOIN LATERAL (
            SELECT id, design
            FROM st_schema.cm_diagram_templates
            WHERE community_project_template_id = pt.id
            LIMIT 1
        ) dt ON true
    `

	category := c.Query("category")
	var rows *sql.Rows

	if category != "" {
		query += " WHERE pt.category = $1 LIMIT $2 OFFSET $3"
		rows, err = tenantManagement.DB.Query(query, category, pageSize, offset)
	} else {
		query += " LIMIT $1 OFFSET $2"
		rows, err = tenantManagement.DB.Query(query, pageSize, offset)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve public project templates: " + err.Error()})
		return
	}
	defer rows.Close()

	var templates []gin.H
	var totalCount int

	for rows.Next() {
		var template models.PublicProjectTemplate
		var diagramID sql.NullString
		var diagramDesign sql.NullString

		if err := rows.Scan(
			&template.ID,
			&template.ProjectTemplateID,
			&template.UserID,
			&template.TenantID,
			&template.Title,
			&template.Description,
			&template.Category,
			&template.Complexity,
			&template.PublishedAt,
			&template.FirstName,
			&template.LastName,
			&template.UserPicture,
			&totalCount,
			&diagramID,
			&diagramDesign,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning public project templates: " + err.Error()})
			return
		}

		// Fetching related document templates for each project template
		docQuery := `
            SELECT id
            FROM st_schema.cm_document_templates
            WHERE community_project_template_id = $1
        `
		docRows, docErr := tenantManagement.DB.Query(docQuery, template.ID)
		if docErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve document templates: " + docErr.Error()})
			return
		}
		var docIDs []string
		for docRows.Next() {
			var docID string
			if err := docRows.Scan(&docID); err != nil {
				docRows.Close()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning document template IDs: " + err.Error()})
				return
			}
			docIDs = append(docIDs, docID)
		}
		docRows.Close()

		// Fetch diagram templates for this project
		diagQuery := `
            SELECT id
            FROM st_schema.cm_diagram_templates
            WHERE community_project_template_id = $1
        `
		diagRows, diagErr := tenantManagement.DB.Query(diagQuery, template.ID)
		if diagErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve diagram templates: " + diagErr.Error()})
			return
		}
		var diagIDs []string
		for diagRows.Next() {
			var diagID string
			if err := diagRows.Scan(&diagID); err != nil {
				diagRows.Close()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning diagram template IDs: " + err.Error()})
				return
			}
			diagIDs = append(diagIDs, diagID)
		}
		diagRows.Close()

		templates = append(templates, gin.H{
			"id":                 template.ID,
			"user_id":            template.UserID,
			"tenant_id":          template.TenantID,
			"title":              template.Title,
			"description":        template.Description,
			"category":           template.Category,
			"complexity":         template.Complexity,
			"created_at":         template.CreatedAt,
			"published_at":       template.PublishedAt,
			"document_templates": docIDs,
			"diagram_templates":  diagIDs,
			"first_name":         template.FirstName,
			"last_name":          template.LastName,
			"user_picture":       template.UserPicture,
			"diagram": gin.H{
				"id":     diagramID.String,
				"design": diagramDesign.String,
			},
		})
	}

	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching public project templates: " + err.Error()})
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))

	c.JSON(http.StatusOK, gin.H{
		"data": templates,
		"pagination": gin.H{
			"current_page": pageNum,
			"total_pages":  totalPages,
			"page_size":    pageSize,
			"total_items":  totalCount,
		},
		"message": "Project resources retrieved successfully!",
	})
}
