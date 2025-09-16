package projects

// This package contains the handlers for managing all diagrams associated with projects.

// The handlers are:
// 1. NewDiagram - Creates a new diagram for a project
// 2. GetDiagram - Retrieves a diagram for a project
// 3. GetDiagrams - Retrieves all diagrams for a project
// 4. UpdateDiagram - Updates a diagram for a project
// 5. DeleteDiagram - Deletes a diagram for a project
// 6. CloneDiagram - Clones an existing diagram with optional overrides

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"

	"sententiawebapi/handlers/apis/tenantManagement"
	"sententiawebapi/handlers/models"
	"sententiawebapi/utilities"

	"github.com/gin-gonic/gin"
)

func NewDiagram(c *gin.Context) {
	var diagram models.Diagram

	// Get the user ID and tenant ID from the context
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	// Extracting project ID from the URL parameter
	projectID := c.Query("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	if err := c.ShouldBindJSON(&diagram); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	diagram.UserID = &userID
	diagram.TenantID = &tenantID
	diagram.ProjectID = &projectID

	// Start transaction
	tx, err := tenantManagement.DB.Begin()
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process diagram creation"})
		return
	}
	defer tx.Rollback() // Will be no-op if transaction is committed

	err = tx.QueryRow(`
		INSERT INTO st_schema.diagrams (
			user_id,
			tenant_id,
			project_id,
			document_id,
			title,
			diagram_type,
			diagram_status,
			category,
			design,
			short_description
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, user_id, tenant_id, project_id, document_id, title, diagram_type,
				  diagram_status, category, design, created_at, updated_at, short_description
	`,
		diagram.UserID,
		diagram.TenantID,
		diagram.ProjectID,
		diagram.DocumentID,
		diagram.Title,
		diagram.DiagramType,
		diagram.DiagramStatus,
		diagram.Category,
		diagram.Design,
		diagram.ShortDescription,
	).Scan(
		&diagram.ID,
		&diagram.UserID,
		&diagram.TenantID,
		&diagram.ProjectID,
		&diagram.DocumentID,
		&diagram.Title,
		&diagram.DiagramType,
		&diagram.DiagramStatus,
		&diagram.Category,
		&diagram.Design,
		&diagram.CreatedAt,
		&diagram.UpdatedAt,
		&diagram.ShortDescription,
	)

	if err != nil {
		log.Printf("Failed to create diagram: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create the diagram"})
		return
	}

	// Update project timestamp
	_, err = tx.Exec(`
		UPDATE st_schema.projects
		SET updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
	`, projectID, tenantID)
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

	// Return the new diagram data
	c.JSON(http.StatusOK, gin.H{
		"data":    diagram,
		"message": "New project diagram created successfully!",
	})
}

func GetDiagram(c *gin.Context) {
	// Get the user ID and tenant ID from the context
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	// Extract project ID and diagram ID from query parameters
	projectID := c.Query("project_id")
	diagramID := c.Query("diagram_id")

	if projectID == "" || diagramID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID and Diagram ID are required"})
		return
	}

	// Note: Corrected table name from "diagrams" to "diagrams" if that's the correct table name
	query := `
		SELECT id, user_id, tenant_id, project_id, document_id, title, diagram_type,
			   diagram_status, category, design, created_at, updated_at, short_description
		FROM st_schema.diagrams
		WHERE id = $1 AND project_id = $2 AND tenant_id = $3
	`

	var diagram models.Diagram
	err := tenantManagement.DB.QueryRow(query, diagramID, projectID, tenantID).Scan(
		&diagram.ID,
		&diagram.UserID,
		&diagram.TenantID,
		&diagram.ProjectID,
		&diagram.DocumentID,
		&diagram.Title,
		&diagram.DiagramType,
		&diagram.DiagramStatus,
		&diagram.Category,
		&diagram.Design,
		&diagram.CreatedAt,
		&diagram.UpdatedAt,
		&diagram.ShortDescription,
	)

	if err != nil {
		log.Printf("Failed to retrieve diagram: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Diagram not found or access denied"})
		return
	}

	// Return the diagram data
	c.JSON(http.StatusOK, gin.H{
		"data":    diagram,
		"message": "Project diagram retrieved successfully!",
	})
}

func GetDiagrams(c *gin.Context) {
	// Get the user ID and tenant ID from the context
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	// Extract project ID from query parameters
	projectID := c.Query("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	query := `
        SELECT id, user_id, tenant_id, project_id, document_id, title, diagram_type,
               diagram_status, category, design, created_at, updated_at, short_description
        FROM st_schema.diagrams
        WHERE project_id = $1 AND tenant_id = $2
    `

	rows, err := tenantManagement.DB.Query(query, projectID, tenantID)
	if err != nil {
		log.Printf("Failed to query diagrams: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve diagrams"})
		return
	}
	defer rows.Close()

	var diagrams []models.Diagram
	for rows.Next() {
		var diagram models.Diagram
		err := rows.Scan(
			&diagram.ID,
			&diagram.UserID,
			&diagram.TenantID,
			&diagram.ProjectID,
			&diagram.DocumentID,
			&diagram.Title,
			&diagram.DiagramType,
			&diagram.DiagramStatus,
			&diagram.Category,
			&diagram.Design,
			&diagram.CreatedAt,
			&diagram.UpdatedAt,
			&diagram.ShortDescription,
		)
		if err != nil {
			log.Printf("Failed to scan diagram row: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read diagram data"})
			return
		}
		diagrams = append(diagrams, diagram)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating through rows: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed during retrieving diagrams"})
		return
	}

	// Return the list of diagrams
	c.JSON(http.StatusOK, gin.H{
		"data":    diagrams,
		"message": "Project diagrams retrieved successfully!",
	})
}

func UpdateDiagram(c *gin.Context) {
	var updateData models.Diagram

	// Get the user ID and tenant ID from the context
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	// Extract project ID and diagram ID from query parameters
	projectID := c.Query("project_id")
	diagramID := c.Query("diagram_id")

	if projectID == "" || diagramID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID and Diagram ID are required"})
		return
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build the set clause dynamically based on non-nil fields in the updateData
	setParts := []string{}
	args := []interface{}{}
	argCounter := 1

	if updateData.Title != nil {
		setParts = append(setParts, fmt.Sprintf("title = $%d", argCounter))
		args = append(args, *updateData.Title)
		argCounter++
	}
	if updateData.DiagramType != nil {
		setParts = append(setParts, fmt.Sprintf("diagram_type = $%d", argCounter))
		args = append(args, *updateData.DiagramType)
		argCounter++
	}
	if updateData.DiagramStatus != nil {
		setParts = append(setParts, fmt.Sprintf("diagram_status = $%d", argCounter))
		args = append(args, *updateData.DiagramStatus)
		argCounter++
	}
	if updateData.Category != nil {
		setParts = append(setParts, fmt.Sprintf("category = $%d", argCounter))
		args = append(args, *updateData.Category)
		argCounter++
	}
	if updateData.Design != nil {
		setParts = append(setParts, fmt.Sprintf("design = $%d", argCounter))
		args = append(args, *updateData.Design)
		argCounter++
	}
	if updateData.ShortDescription != nil {
		setParts = append(setParts, fmt.Sprintf("short_description = $%d", argCounter))
		args = append(args, *updateData.ShortDescription)
		argCounter++
	}

	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No updatable fields provided"})
		return
	}

	setClause := strings.Join(setParts, ", ")
	query := fmt.Sprintf(`
		UPDATE
			st_schema.diagrams
		SET
			%s, updated_at = NOW()
		WHERE
			id = $%d
		AND
			project_id = $%d
		AND
			tenant_id = $%d
		RETURNING
			id, user_id, tenant_id, project_id, document_id, title, diagram_type, diagram_status, category, design, short_description, created_at, updated_at
	`, setClause, argCounter, argCounter+1, argCounter+2)

	args = append(args, diagramID, projectID, tenantID)

	tx, err := tenantManagement.DB.Begin()
	if err != nil {
		log.Printf("Executing query: %s with args: %v", query, args)
		c.JSON(http.StatusInternalServerError, gin.H{"error": models.DatabaseError})
		return
	}

	stmt, err := tx.Prepare(query)
	if err != nil {
		tx.Rollback()
		log.Printf("Prepare statement error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": models.DatabaseError})
		return
	}
	defer stmt.Close()

	var updatedDiagram models.Diagram
	err = stmt.QueryRow(args...).Scan(
		&updatedDiagram.ID,
		&updatedDiagram.UserID,
		&updatedDiagram.TenantID,
		&updatedDiagram.ProjectID,
		&updatedDiagram.DocumentID,
		&updatedDiagram.Title,
		&updatedDiagram.DiagramType,
		&updatedDiagram.DiagramStatus,
		&updatedDiagram.Category,
		&updatedDiagram.Design,
		&updatedDiagram.ShortDescription,
		&updatedDiagram.CreatedAt,
		&updatedDiagram.UpdatedAt,
	)
	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			log.Printf("No diagram found with ID: %s, Project ID: %s, User ID: %s", diagramID, projectID, userID)
			c.JSON(http.StatusNotFound, gin.H{"error": "Diagram not found"})
		} else {
			log.Printf("Database Err: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update the diagram"})
		}
		return
	}

	// Update project timestamp
	_, err = tx.Exec(`
		UPDATE st_schema.projects
		SET updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
	`, projectID, tenantID)
	if err != nil {
		tx.Rollback()
		log.Printf("Failed to update project timestamp: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update the project timestamp"})
		return
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Database Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": models.DatabaseError})
		return
	}

	// Return the updated diagram data
	c.JSON(http.StatusOK, gin.H{
		"data":    updatedDiagram,
		"message": "Diagram updated successfully",
	})
}

func DeleteDiagram(c *gin.Context) {
	// Get the user ID and tenant ID from the context
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	// Extract project ID and diagram ID from query parameters
	projectID := c.Query("project_id")
	diagramID := c.Query("diagram_id")

	if projectID == "" || diagramID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID and Diagram ID are required"})
		return
	}

	// Start transaction
	tx, err := tenantManagement.DB.Begin()
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process diagram deletion"})
		return
	}
	defer tx.Rollback() // Will be no-op if transaction is committed

	query := `
		DELETE FROM
			st_schema.diagrams
		WHERE
			id = $1 AND project_id = $2 AND tenant_id = $3
		RETURNING id, user_id, tenant_id, project_id, document_id, title, diagram_type, diagram_status, category, design, short_description, created_at, updated_at
	`

	var diagram models.Diagram
	err = tx.QueryRow(query, diagramID, projectID, tenantID).Scan(
		&diagram.ID,
		&diagram.UserID,
		&diagram.TenantID,
		&diagram.ProjectID,
		&diagram.DocumentID,
		&diagram.Title,
		&diagram.DiagramType,
		&diagram.DiagramStatus,
		&diagram.Category,
		&diagram.Design,
		&diagram.ShortDescription,
		&diagram.CreatedAt,
		&diagram.UpdatedAt,
	)

	if err != nil {
		log.Printf("Failed to delete diagram: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Diagram not found or access denied"})
		return
	}

	// Update project timestamp
	_, err = tx.Exec(`
		UPDATE st_schema.projects
		SET updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
	`, projectID, tenantID)
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

	// Return a success response
	c.JSON(http.StatusOK, gin.H{
		"data":    diagram,
		"message": "Diagram deleted successfully",
	})
}

func CloneDiagram(c *gin.Context) {
	// Get the user ID and tenant ID from the context
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	projectID := c.Query("project_id")
	diagramID := c.Query("diagram_id")

	if projectID == "" || diagramID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID and Diagram ID are required"})
		return
	}

	// Define struct for optional override values
	var overrideData struct {
		Title            string `json:"title"`
		DiagramType      string `json:"diagram_type"`
		DiagramStatus    string `json:"diagram_status"`
		Category         string `json:"category"`
		ShortDescription string `json:"short_description"`
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

	// Fetch the existing diagram
	var existingDiagram models.Diagram
	err = tx.QueryRow(`
        SELECT id, user_id, tenant_id, project_id, document_id, title, diagram_type, diagram_status, category, design, raw_design, short_description, created_at, updated_at
        FROM st_schema.diagrams
        WHERE id = $1 AND project_id = $2 AND tenant_id = $3`,
		diagramID, projectID, tenantID,
	).Scan(
		&existingDiagram.ID,
		&existingDiagram.UserID,
		&existingDiagram.TenantID,
		&existingDiagram.ProjectID,
		&existingDiagram.DocumentID,
		&existingDiagram.Title,
		&existingDiagram.DiagramType,
		&existingDiagram.DiagramStatus,
		&existingDiagram.Category,
		&existingDiagram.Design,
		&existingDiagram.RawDesign,
		&existingDiagram.ShortDescription,
		&existingDiagram.CreatedAt,
		&existingDiagram.UpdatedAt,
	)

	if err != nil {
		log.Printf("Failed to fetch diagram: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch the diagram to clone"})
		return
	}

	// Apply overrides if provided
	title := existingDiagram.Title
	if overrideData.Title != "" {
		title = &overrideData.Title
	}
	diagramType := existingDiagram.DiagramType
	if overrideData.DiagramType != "" {
		diagramType = &overrideData.DiagramType
	}
	diagramStatus := existingDiagram.DiagramStatus
	if overrideData.DiagramStatus != "" {
		diagramStatus = &overrideData.DiagramStatus
	}
	category := existingDiagram.Category
	if overrideData.Category != "" {
		category = &overrideData.Category
	}
	shortDescription := existingDiagram.ShortDescription
	if overrideData.ShortDescription != "" {
		shortDescription = &overrideData.ShortDescription
	}

	design := existingDiagram.Design // Always use the original design
	documentID := existingDiagram.DocumentID

	// Create a new diagram with the same design but a new ID
	var newDiagram models.Diagram
	err = tx.QueryRow(`
        INSERT INTO st_schema.diagrams (
            user_id, tenant_id, project_id, document_id, title, diagram_type, diagram_status, category, design, raw_design, short_description
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
        ) RETURNING id, user_id, tenant_id, project_id, document_id, title, diagram_type, diagram_status, category, design, raw_design, created_at, updated_at, short_description`,
		userID, tenantID, projectID, documentID, title, diagramType, diagramStatus, category, design, existingDiagram.RawDesign, shortDescription,
	).Scan(
		&newDiagram.ID,
		&newDiagram.UserID,
		&newDiagram.TenantID,
		&newDiagram.ProjectID,
		&newDiagram.DocumentID,
		&newDiagram.Title,
		&newDiagram.DiagramType,
		&newDiagram.DiagramStatus,
		&newDiagram.Category,
		&newDiagram.Design,
		&newDiagram.RawDesign,
		&newDiagram.CreatedAt,
		&newDiagram.UpdatedAt,
		&newDiagram.ShortDescription,
	)

	if err != nil {
		log.Printf("Failed to create cloned diagram: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create cloned diagram"})
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
		"data":    newDiagram,
		"message": "Diagram cloned successfully!",
	})
}
