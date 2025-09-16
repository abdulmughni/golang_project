package projects

// Project Feature Handlers
// This file contains the following functions:
// 1. NewProject: Creates a new project and default document for the user.
// 2. GetProject: Retrieves details of a specific project for the user.
// 3. GetProjects: Retrieves a list of projects for the user.
// 4. UpdateProject: Updates an existing project in the 'projects' table and retrieves document ID from 'project_documents' table.
// 5. DeleteProject: Deletes an existing project from the 'projects' table and any associated documents from 'project_documents' table.

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"sententiawebapi/handlers/apis/images"
	"sententiawebapi/handlers/apis/tenantManagement"
	"sententiawebapi/handlers/models"
	"sententiawebapi/utilities"
	"time"

	"github.com/gin-gonic/gin"
)

type ProjectResource struct {
	models.Project
	Requirements []models.Requirement `json:"requirements"`
}

func NewProject(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	var input ProjectResource
	if err := c.ShouldBindJSON(&input); err != nil {
		utilities.Response(c, http.StatusBadRequest, "Invalid request data", gin.H{"error": err.Error()})
		return
	}

	tx, err := tenantManagement.DB.Begin()
	if err != nil {
		log.Printf(models.DatabaseError, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to the database..."})
		return
	}
	defer tx.Rollback()

	err = tx.QueryRow(`
		INSERT INTO
			st_schema.projects (
				user_id,
				tenant_id,
				title,
				status,
				category,
				description,
				complexity,
				short_description
			)
		VALUES
			($1, $2, $3, $4, $5, COALESCE($6, ''), $7, COALESCE($8, ''))
		RETURNING
			id,
			user_id,
			tenant_id,
			title,
			status,
			category,
			created_at,
			updated_at,
			description,
			complexity,
			short_description`,
		userID,
		tenantID,
		input.Title,
		input.Status,
		input.Category,
		input.Description,
		input.Complexity,
		input.ShortDescription,
	).Scan(
		&input.ID,
		&input.UserID,
		&input.TenantID,
		&input.Title,
		&input.Status,
		&input.Category,
		&input.CreatedAt,
		&input.UpdatedAt,
		&input.Description,
		&input.Complexity,
		&input.ShortDescription,
	)
	if err != nil {
		log.Printf("Failed to insert project with error: %v", err)
		utilities.Response(c, http.StatusInternalServerError, "Failed to create the project", nil)
		return
	}

	requirementsApi := NewRequirementsApi(tenantID, userID, *input.ID)
	if _, err := requirementsApi.AddMany(tx, input.Requirements); err != nil {
		log.Printf("Failed to insert requirements: %v", err)
		utilities.Response(c, http.StatusInternalServerError, "Failed to create project requirements", nil)
		return
	}

	if err = tx.Commit(); err != nil {
		log.Printf(models.DatabaseError, err)
		utilities.Response(c, http.StatusInternalServerError, "Failed to commit database transaction", nil)
		return
	}

	projectData := map[string]interface{}{
		"id":         *input.ID,
		"user_id":    userID,
		"tenant_id":  tenantID,
		"title":      *input.Title,
		"status":     *input.Status,
		"complexity": *input.Complexity,
		"category":   *input.Category,
		"created_at": *input.CreatedAt,
		"updated_at": *input.UpdatedAt,
	}

	if input.Description != nil {
		projectData["description"] = *input.Description
	}
	if input.ShortDescription != nil {
		projectData["short_description"] = *input.ShortDescription
	}

	utilities.Response(c, http.StatusOK, models.ResponseSuccess, projectData)
}

func GetProject(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	projectId := c.Query("id")
	if projectId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Resource ID parameter cannot be empty..."})
		return
	}

	query := `
		SELECT
			id,
			user_id,
			tenant_id,
			title,
			status,
			category,
			created_at,
			updated_at,
			description,
			complexity,
			short_description
		FROM
			st_schema.projects
		WHERE
			id = $1
		AND
			tenant_id = $2
	`

	row := tenantManagement.DB.QueryRow(query, projectId, tenantID)

	var ProjectResource models.Project
	err := row.Scan(
		&ProjectResource.ID,
		&ProjectResource.UserID,
		&ProjectResource.TenantID,
		&ProjectResource.Title,
		&ProjectResource.Status,
		&ProjectResource.Category,
		&ProjectResource.CreatedAt,
		&ProjectResource.UpdatedAt,
		&ProjectResource.Description,
		&ProjectResource.Complexity,
		&ProjectResource.ShortDescription,
	)

	if err != nil {
		log.Printf(models.DatabaseError, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve the project resource data"})
		return
	}

	// Construct response data
	responseData := gin.H{
		"id":                ProjectResource.ID,
		"user_id":           ProjectResource.UserID,
		"tenant_id":         ProjectResource.TenantID,
		"title":             ProjectResource.Title,
		"status":            ProjectResource.Status,
		"category":          ProjectResource.Category,
		"complexity":        ProjectResource.Complexity,
		"created_at":        ProjectResource.CreatedAt,
		"updated_at":        ProjectResource.UpdatedAt,
		"description":       ProjectResource.Description,
		"short_description": ProjectResource.ShortDescription,
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    responseData,
		"message": models.ResponseSuccess,
	})
}

func GetProjects(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	query := `
		SELECT
			p.id,
			p.user_id,
			p.tenant_id,
			p.title,
			p.status,
			p.category,
			p.created_at,
			p.updated_at,
			p.description,
			p.complexity,
			p.short_description,
			u.first_name,
			u.last_name,
			u.user_picture,
			COALESCE((
				SELECT json_agg(r)
				FROM (
					SELECT
						id,
						title,
						TO_CHAR(target_date, 'YYYY-MM-DD') AS target_date
					FROM st_schema.project_requirements
					WHERE project_id = p.id AND tenant_id = p.tenant_id
				) r
			), '[]') AS requirements
		FROM
			st_schema.projects p
		LEFT JOIN
			st_schema.users u ON p.user_id = u.id
		WHERE
			p.tenant_id = $1
	`

	rows, err := tenantManagement.DB.Query(query, tenantID)
	if err != nil {
		log.Printf(models.DatabaseError, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve the project resources..."})
		return
	}
	defer rows.Close()

	var projects []gin.H
	for rows.Next() {
		var ProjectResource models.Project
		var firstName, lastName, userPicture sql.NullString
		var shortDescription sql.NullString
		var requirementsJSON json.RawMessage

		err := rows.Scan(
			&ProjectResource.ID,
			&ProjectResource.UserID,
			&ProjectResource.TenantID,
			&ProjectResource.Title,
			&ProjectResource.Status,
			&ProjectResource.Category,
			&ProjectResource.CreatedAt,
			&ProjectResource.UpdatedAt,
			&ProjectResource.Description,
			&ProjectResource.Complexity,
			&shortDescription,
			&firstName,
			&lastName,
			&userPicture,
			&requirementsJSON,
		)
		if err != nil {
			log.Printf(models.DatabaseError, err)
			continue
		}

		projectData := gin.H{
			"id":           *ProjectResource.ID,
			"user_id":      *ProjectResource.UserID,
			"tenant_id":    *ProjectResource.TenantID,
			"title":        *ProjectResource.Title,
			"description":  *ProjectResource.Description,
			"complexity":   *ProjectResource.Complexity,
			"status":       *ProjectResource.Status,
			"category":     *ProjectResource.Category,
			"created_at":   *ProjectResource.CreatedAt,
			"updated_at":   *ProjectResource.UpdatedAt,
			"requirements": requirementsJSON,
		}

		if shortDescription.Valid {
			projectData["short_description"] = shortDescription.String
		} else {
			projectData["short_description"] = ""
		}

		if firstName.Valid {
			projectData["first_name"] = firstName.String
		}
		if lastName.Valid {
			projectData["last_name"] = lastName.String
		}
		if userPicture.Valid {
			projectData["user_picture"] = userPicture.String
		}

		projects = append(projects, projectData)
	}

	if err = rows.Err(); err != nil {
		log.Printf(models.DatabaseError, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process the project resources..."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    projects,
		"message": "Project resources retrieved successfully!",
	})
}

func UpdateProject(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	projectId := c.Query("id")
	if projectId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required..."})
		return
	}

	var updateData models.Project
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := tenantManagement.DB.Begin()
	if err != nil {
		log.Printf(models.DatabaseError, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start a new transaction"})
		return
	}

	updateQuery := `
		UPDATE
			st_schema.projects
		SET
			title = $1,
			status = $2,
			category = $3,
			updated_at = NOW(),
			description = CASE 
				WHEN $4::text IS NULL AND $4 IS NOT NULL THEN NULL  -- when explicitly set to null
				WHEN $4 IS NULL THEN description  -- when field is not in request
				ELSE $4  -- when value is provided
			END,
			complexity = $5,
			short_description = CASE 
				WHEN $6::text IS NULL AND $6 IS NOT NULL THEN NULL  -- when explicitly set to null
				WHEN $6 IS NULL THEN short_description  -- when field is not in request
				ELSE $6  -- when value is provided
			END
		WHERE
			id = $7
		AND
			tenant_id = $8
		RETURNING
			id, user_id, tenant_id, title, status, category, created_at, updated_at, description, complexity, short_description
	`

	err = tx.QueryRow(
		updateQuery,
		updateData.Title,
		updateData.Status,
		updateData.Category,
		updateData.Description,
		updateData.Complexity,
		updateData.ShortDescription,
		projectId,
		tenantID).Scan(
		&updateData.ID,
		&updateData.UserID,
		&updateData.TenantID,
		&updateData.Title,
		&updateData.Status,
		&updateData.Category,
		&updateData.CreatedAt,
		&updateData.UpdatedAt,
		&updateData.Description,
		&updateData.Complexity,
		&updateData.ShortDescription,
	)

	if err != nil {
		tx.Rollback()
		log.Printf(models.DatabaseError, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update the project"})
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf(models.DatabaseError, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit the transaction"})
		return
	}

	// Modify the response data construction
	responseData := gin.H{
		"id":         *updateData.ID,
		"user_id":    *updateData.UserID,
		"tenant_id":  *updateData.TenantID,
		"title":      *updateData.Title,
		"complexity": *updateData.Complexity,
		"status":     *updateData.Status,
		"category":   *updateData.Category,
		"created_at": *updateData.CreatedAt,
		"updated_at": *updateData.UpdatedAt,
	}

	// Add optional fields only if they're not nil
	if updateData.Description != nil {
		responseData["description"] = *updateData.Description
	}
	if updateData.ShortDescription != nil {
		responseData["short_description"] = *updateData.ShortDescription
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    responseData,
		"message": models.ResponseSuccess,
	})
}

func DeleteProject(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	projectId := c.Query("id")
	if projectId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required..."})
		return
	}

	tx, err := tenantManagement.DB.Begin()
	if err != nil {
		log.Printf(models.DatabaseError, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start a new transaction"})
		return
	}
	defer tx.Rollback()

	// Retrieve project data before deletion
	var deletedProject models.Project
	err = tx.QueryRow(`
		SELECT
			id, user_id, tenant_id, title, status, category, created_at, updated_at, description, complexity, short_description
		FROM
			st_schema.projects
		WHERE
			id = $1
		AND
			tenant_id = $2`,
		projectId, tenantID).Scan(
		&deletedProject.ID,
		&deletedProject.UserID,
		&deletedProject.TenantID,
		&deletedProject.Title,
		&deletedProject.Status,
		&deletedProject.Category,
		&deletedProject.CreatedAt,
		&deletedProject.UpdatedAt,
		&deletedProject.Description,
		&deletedProject.Complexity,
		&deletedProject.ShortDescription,
	)

	if err != nil {
		log.Printf(models.DatabaseError, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve project details"})
		return
	}

	// Now delete the project
	_, err = tx.Exec(`
		DELETE FROM
			st_schema.projects
		WHERE
			id = $1
		AND
			tenant_id = $2`, projectId, tenantID)
	if err != nil {
		log.Printf(models.DatabaseError, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete the project"})
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf(models.DatabaseError, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit the transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Project deleted successfully!",
		"data":    deletedProject,
	})
}

func NewProjectFromTemplate(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	templateID := c.Query("template_id")
	if templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template ID must be provided"})
		return
	}

	// Start a new database transaction
	tx, err := tenantManagement.DB.Begin()
	if err != nil {
		log.Printf("Transaction start error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start a new transaction"})
		return
	}
	defer tx.Rollback()

	// Fetch the project template
	var projectTemplate models.ProjectTemplate
	err = tx.QueryRow(`
        SELECT
            user_id, tenant_id, title, category, complexity, short_description
        FROM
            st_schema.project_templates
        WHERE
            id = $1
        AND
            tenant_id = $2`, templateID, tenantID).Scan(
		&projectTemplate.UserID,
		&projectTemplate.TenantID,
		&projectTemplate.Title,
		&projectTemplate.Category,
		&projectTemplate.Complexity,
		&projectTemplate.ShortDescription)

	if err := c.ShouldBindJSON(&projectTemplate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err != nil {
		log.Printf("Failed to fetch project template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch project template"})
		return
	}

	// Create the new project based on the template
	var newProjectID string
	createdAt := time.Now()
	updatedAt := createdAt // Initially the same as created_at
	defaultStatus := "Not Started"
	projectTemplate.Status = &defaultStatus

	err = tx.QueryRow(`
        INSERT INTO
            st_schema.projects (user_id, tenant_id, title, status, description, category, complexity, short_description)
        VALUES
            ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, created_at, updated_at`,
		userID,
		tenantID,
		projectTemplate.Title,
		projectTemplate.Status,
		projectTemplate.Description,
		projectTemplate.Category,
		projectTemplate.Complexity,
		projectTemplate.ShortDescription,
	).Scan(
		&newProjectID,
		&projectTemplate.CreatedAt,
		&projectTemplate.UpdatedAt,
	)

	if err != nil {
		log.Printf("Failed to insert new project: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project from template"})
		return
	}

	// Declare a cursor for documents associated with the template
	_, err = tx.Exec(`
        DECLARE
            doc_cursor
        CURSOR FOR SELECT
            id, title, p_content_json, p_raw_content, complexity
        FROM
            st_schema.document_templates
        WHERE
            project_template_id = $1`, templateID)

	if err != nil {
		log.Printf("Failed to declare cursor: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to declare cursor"})
		return
	}

	var documentIDs = []string{}
	var documentTemplateIDs = []string{}
	var lastDocumentID string
	// Iterate over the cursor, fetching documents one by one
	for {
		var docTemplateID, title string
		var content *json.RawMessage
		var rawContent []byte
		var complexity sql.NullString

		// Fetch the next document from the cursor
		err = tx.QueryRow(`FETCH NEXT FROM doc_cursor`).Scan(&docTemplateID, &title, &content, &rawContent, &complexity)
		if err != nil {
			if err == sql.ErrNoRows {
				break // No more documents, exit the loop
			}
			log.Printf("Failed to fetch document: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching document from cursor"})
			return
		}

		var documentID string
		err = tx.QueryRow(`
            INSERT INTO
                st_schema.project_documents (
					user_id, tenant_id, project_id, title,
					content_json, raw_content, p_raw_content,
					complexity, created_at, updated_at
				)
            VALUES
                ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
            RETURNING id`,
			userID,
			tenantID,
			newProjectID,
			title,
			content,
			rawContent,
			rawContent,
			complexity.String, // Handle NULL complexity appropriately
			createdAt,
			updatedAt).Scan(&documentID)

		if err != nil {
			log.Printf("Failed to insert document: %v", err)
			continue // Consider handling or logging as needed
		}

		documentIDs = append(documentIDs, documentID)
		documentTemplateIDs = append(documentTemplateIDs, docTemplateID)
		lastDocumentID = documentID
	}

	// Close the cursor after use
	_, err = tx.Exec("CLOSE doc_cursor")
	if err != nil {
		log.Printf("Failed to close cursor: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error closing cursor"})
		return
	}

	err = images.CopyProjectFiles(context.Background(), tx, images.CopyProjectParams{
		SourceDocuments: images.DocumentsRef{
			Type: models.ResourceGroupTemplate,
			IDs:  documentTemplateIDs,
		},
		DestinationDocuments: images.DocumentsRef{
			Type: models.ResourceGroupProject,
			IDs:  documentIDs,
		},
		TenantID: tenantID,
	})
	if err != nil {
		log.Printf("Error while copying files: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error copying files"})
		return
	}

	// Declare a cursor for diagrams associated with the template
	_, err = tx.Exec(`
        DECLARE
            diagram_cursor
        CURSOR FOR SELECT
            title, diagram_type, diagram_status, category, p_design, p_raw_design, short_description
        FROM
            st_schema.diagram_templates
        WHERE
            project_template_id = $1`, templateID)

	if err != nil {
		log.Printf("Failed to declare diagram cursor: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to declare diagram cursor"})
		return
	}

	var diagramID string
	// Iterate over the cursor, fetching diagrams one by one
	for {
		var title, diagramType string
		var diagramStatus, category, shortDescription sql.NullString
		var design *json.RawMessage
		var rawDesign []byte

		// Fetch the next diagram from the cursor
		err = tx.QueryRow(`FETCH NEXT FROM diagram_cursor`).Scan(
			&title,
			&diagramType,
			&diagramStatus,
			&category,
			&design,
			&rawDesign,
			&shortDescription)

		if err != nil {
			if err == sql.ErrNoRows {
				break // No more diagrams, exit the loop
			}
			log.Printf("Failed to fetch diagram: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching diagram from cursor"})
			return
		}

		err = tx.QueryRow(`
            INSERT INTO
                st_schema.diagrams (
					user_id, tenant_id, project_id, title, diagram_type, diagram_status, category,
					design, raw_design,
					created_at, updated_at,
					short_description
				)
            VALUES
                ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
            RETURNING id`,
			userID,
			tenantID,
			newProjectID,
			title,
			diagramType,
			diagramStatus.String,
			category.String,
			design,
			rawDesign,
			createdAt,
			updatedAt,
			shortDescription).Scan(&diagramID)

		if err != nil {
			log.Printf("Failed to insert diagram: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert diagram"})
			return
		}
	}

	// Close the cursor after use
	_, err = tx.Exec("CLOSE diagram_cursor")
	if err != nil {
		log.Printf("Failed to close diagram cursor: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error closing cursor"})
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data": gin.H{
			"id":                newProjectID,
			"user_id":           userID,
			"tenant_id":         tenantID,
			"title":             projectTemplate.Title,
			"description":       projectTemplate.Description,
			"complexity":        projectTemplate.Complexity,
			"category":          projectTemplate.Category,
			"status":            defaultStatus,
			"created_at":        projectTemplate.CreatedAt,
			"updated_at":        projectTemplate.UpdatedAt,
			"short_description": projectTemplate.ShortDescription,
			"components": gin.H{
				"document": gin.H{
					"id": lastDocumentID,
				},
				"diagram": gin.H{
					"id": diagramID,
				},
			},
		},
	})
}

func NewProjectFromPublicTemplate(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	templateID := c.Query("public_template_id")
	if templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template ID must be provided"})
		return
	}

	// Start a new database transaction
	tx, err := tenantManagement.DB.Begin()
	if err != nil {
		log.Printf("Transaction start error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start a new transaction"})
		return
	}
	defer tx.Rollback()

	// Fetch the project template
	var publicProjectTemplate models.PublicProjectTemplate
	err = tx.QueryRow(`
        SELECT
            title, category, complexity, description, short_description
        FROM
            st_schema.cm_project_templates
        WHERE
            id = $1`, templateID).Scan(
		&publicProjectTemplate.Title,
		&publicProjectTemplate.Category,
		&publicProjectTemplate.Complexity,
		&publicProjectTemplate.Description,
		&publicProjectTemplate.ShortDescription,
	)

	if err := c.ShouldBindJSON(&publicProjectTemplate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err != nil {
		log.Printf("Failed to fetch project template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch project template"})
		return
	}

	// Create the new project based on the template
	var newProjectID string
	createdAt := time.Now()
	updatedAt := createdAt // Initially the same as created_at
	defaultStatus := "Not Started"

	err = tx.QueryRow(`
        INSERT INTO
            st_schema.projects (user_id, tenant_id, title, status, description, category, complexity, short_description)
        VALUES
            ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, created_at, updated_at`,
		userID,
		tenantID,
		publicProjectTemplate.Title,
		defaultStatus,
		publicProjectTemplate.Description,
		publicProjectTemplate.Category,
		publicProjectTemplate.Complexity,
		publicProjectTemplate.ShortDescription,
	).Scan(
		&newProjectID,
		&publicProjectTemplate.CreatedAt,
		&publicProjectTemplate.LastUpdateAt,
	)

	if err != nil {
		log.Printf("Failed to insert new project: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project from template"})
		return
	}

	// Declare a cursor for documents associated with the template
	_, err = tx.Exec(`
        DECLARE
            doc_cursor
        CURSOR FOR SELECT
            id, title, p_content_json, p_raw_content, complexity, category
        FROM
            st_schema.cm_document_templates
        WHERE
            community_project_template_id = $1`, templateID)

	if err != nil {
		log.Printf("Failed to declare cursor: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to declare cursor"})
		return
	}

	var documentID string
	var communityDocIDs = make([]string, 0)
	var docIDs = make([]string, 0)
	// Iterate over the cursor, fetching documents one by one
	for {
		var communityDocID, title string
		var content *json.RawMessage
		var rawContent []byte
		var complexity, category sql.NullString

		// Fetch the next document from the cursor
		err = tx.QueryRow(`FETCH NEXT FROM doc_cursor`).Scan(&communityDocID, &title, &content, &rawContent, &complexity, &category)
		if err != nil {
			if err == sql.ErrNoRows {
				break // No more documents, exit the loop
			}
			log.Printf("Failed to fetch document: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching document from cursor"})
			return
		}

		err = tx.QueryRow(`
            INSERT INTO
                st_schema.project_documents (
					user_id, tenant_id, project_id, title,
					content_json, raw_content, p_raw_content,
					complexity, created_at, updated_at
				)
            VALUES
                ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
            RETURNING id`,
			userID,
			tenantID,
			newProjectID,
			title,
			content,
			rawContent,
			rawContent,
			complexity.String, // Handle NULL complexity appropriately
			createdAt,
			updatedAt).Scan(&documentID)

		if err != nil {
			log.Printf("Failed to insert document: %v", err)
			continue // Consider handling or logging as needed
		}

		communityDocIDs = append(communityDocIDs, communityDocID)
		docIDs = append(docIDs, documentID)
	}

	// Close the cursor after use
	_, err = tx.Exec("CLOSE doc_cursor")
	if err != nil {
		log.Printf("Failed to close cursor: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error closing cursor"})
		return
	}

	err = images.CopyProjectFiles(context.Background(), tx, images.CopyProjectParams{
		SourceDocuments: images.DocumentsRef{
			Type: models.ResourceGroupCommunity,
			IDs:  communityDocIDs,
		},
		DestinationDocuments: images.DocumentsRef{
			Type: models.ResourceGroupProject,
			IDs:  docIDs,
		},
		TenantID: tenantID,
	})
	if err != nil {
		log.Printf("Error while copying files: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error copying files"})
		return
	}

	// Declare a cursor for diagrams associated with the template
	_, err = tx.Exec(`
        DECLARE
            diagram_cursor
        CURSOR FOR SELECT
            title, diagram_type, diagram_status, category, p_design, p_raw_design, published_at, last_update_at
        FROM
            st_schema.cm_diagram_templates
        WHERE
            community_project_template_id = $1`, templateID)

	if err != nil {
		log.Printf("Failed to declare diagram cursor: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to declare diagram cursor"})
		return
	}

	// Make sure we close the cursor before committing or rolling back
	defer func() {
		_, err := tx.Exec("CLOSE diagram_cursor")
		if err != nil {
			log.Printf("Failed to close diagram cursor: %v", err)
		}
	}()

	var diagramID string
	// Iterate over the cursor, fetching diagrams one by one
	for {
		var title, diagramType string
		var diagramStatus, category sql.NullString
		var design *json.RawMessage
		var rawDesign []byte
		var publishedAt, lastUpdateAt sql.NullTime

		// Fetch the next diagram from the cursor
		err = tx.QueryRow(`FETCH NEXT FROM diagram_cursor`).Scan(
			&title,
			&diagramType,
			&diagramStatus,
			&category,
			&design,
			&rawDesign,
			&publishedAt,
			&lastUpdateAt)

		if err != nil {
			if err == sql.ErrNoRows {
				break // No more diagrams, exit the loop
			}
			log.Printf("Failed to fetch diagram with error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching diagram from cursor"})
			return
		}

		// Insert into project_diagrams
		err = tx.QueryRow(`
            INSERT INTO
                st_schema.diagrams (
					user_id, tenant_id, project_id, title, diagram_type, diagram_status, category,
					design, raw_design,
					created_at, updated_at
				)
            VALUES
                ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
            RETURNING id`,
			userID,
			tenantID,
			newProjectID,
			title,
			diagramType,
			diagramStatus.String,
			category.String,
			design,
			rawDesign,
			createdAt,
			updatedAt,
		).Scan(&diagramID)

		if err != nil {
			log.Printf("Failed to insert diagram: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert diagram"})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data": gin.H{
			"id":                newProjectID,
			"user_id":           userID,
			"tenant_id":         tenantID,
			"title":             publicProjectTemplate.Title,
			"description":       publicProjectTemplate.Description,
			"complexity":        publicProjectTemplate.Complexity,
			"category":          publicProjectTemplate.Category,
			"status":            defaultStatus,
			"created_at":        publicProjectTemplate.CreatedAt,
			"updated_at":        publicProjectTemplate.LastUpdateAt,
			"short_description": publicProjectTemplate.ShortDescription,
			"components": gin.H{
				"document": gin.H{
					"id": documentID,
				},
				"diagram": gin.H{
					"id": diagramID,
				},
			},
		},
	})
}
