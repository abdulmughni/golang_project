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

// This function creates a new diagram template
func NewInternalDiagramTemplate(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)

	if !ok {
		return
	}

	var diagram models.DiagramTemplate

	// Extracting project ID from the URL parameter
	projectTemplateID := c.Query("project_template_id")

	if projectTemplateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project Template ID is required..."})

		return
	}

	if err := c.ShouldBindJSON(&diagram); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	diagram.UserID = &userID
	diagram.TenantID = &tenantID
	diagram.ProjectTemplateID = &projectTemplateID

	err := tenantManagement.DB.QueryRow(`
		INSERT INTO st_schema.diagram_templates (
            user_id, tenant_id, project_template_id,
            title, diagram_type, diagram_status, category,
            design, raw_design, p_design, p_raw_design,
            short_description
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
        ) RETURNING id, user_id, tenant_id, project_template_id, title,
            diagram_type, diagram_status, category, p_design,
            created_at, updated_at, short_description
		`,
		diagram.UserID, diagram.TenantID, diagram.ProjectTemplateID,
		diagram.Title, diagram.DiagramType, diagram.DiagramStatus, diagram.Category,
		diagram.Design, diagram.RawDesign, diagram.Design, diagram.RawDesign,
		diagram.ShortDescription,
	).Scan(
		&diagram.ID,
		&diagram.UserID,
		&diagram.TenantID,
		&diagram.ProjectTemplateID,
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
		log.Printf("Database Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": models.DatabaseError})

		return
	}

	// Return the new diagram data
	c.JSON(200, gin.H{
		"data":    diagram,
		"message": "New diagram template created successfully!",
	})
}

// This function retrieves the details of a specific diagram template
func GetInternalDiagramTemplate(c *gin.Context) {
	// Get the tenant ID from the context
	_, tenantID, ok := utilities.ProcessIdentity(c)

	if !ok {
		return
	}

	// Extracting diagram ID from the URL parameters
	diagramTemplateId := c.Query("id")

	if diagramTemplateId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Diagram Template ID is required"})

		return
	}

	var diagramTemplate models.DiagramTemplate
	err := tenantManagement.DB.QueryRow(`
        SELECT
            id, user_id, tenant_id, project_template_id,
            title, diagram_type, diagram_status, category,
            p_design, created_at, updated_at, short_description
        FROM
            st_schema.diagram_templates
        WHERE
            id = $1
        AND
            tenant_id = $2`,
		diagramTemplateId, tenantID,
	).Scan(
		&diagramTemplate.ID,
		&diagramTemplate.UserID,
		&diagramTemplate.TenantID,
		&diagramTemplate.ProjectTemplateID,
		&diagramTemplate.Title,
		&diagramTemplate.DiagramType,
		&diagramTemplate.DiagramStatus,
		&diagramTemplate.Category,
		&diagramTemplate.Design,
		&diagramTemplate.CreatedAt,
		&diagramTemplate.UpdatedAt,
		&diagramTemplate.ShortDescription,
	)

	if err != nil {
		if isDevelopmentEnvironment() {
			log.Printf("Error retrieving diagram template: %v", err)
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Diagram not found or access denied"})

		return
	}

	// Return the diagram data
	c.JSON(http.StatusOK, gin.H{
		"data":    diagramTemplate,
		"message": "Diagram template retrieved successfully!",
	})
}

// This function retrieves an array of all diagram templates from across all project templates.
func GetInternalDiagramTemplates(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)

	if !ok {
		return
	}

	category := c.Query("category")
	projectTemplateID := c.Query("project_template_id")

	query := `
        SELECT
            id, user_id, tenant_id, project_template_id,
            title, diagram_type, diagram_status, category,
            p_design, created_at, updated_at, short_description
        FROM
            st_schema.diagram_templates
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

	if projectTemplateID != "" {
		argCounter++
		conditions = append(conditions, fmt.Sprintf("project_template_id = $%d", argCounter))
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve diagram templates: " + err.Error()})

		return
	}
	defer rows.Close()

	var diagramTemplates []models.DiagramTemplate

	for rows.Next() {
		var diagramTemplate models.DiagramTemplate
		err := rows.Scan(
			&diagramTemplate.ID,
			&diagramTemplate.UserID,
			&diagramTemplate.TenantID,
			&diagramTemplate.ProjectTemplateID,
			&diagramTemplate.Title,
			&diagramTemplate.DiagramType,
			&diagramTemplate.DiagramStatus,
			&diagramTemplate.Category,
			&diagramTemplate.Design,
			&diagramTemplate.CreatedAt,
			&diagramTemplate.UpdatedAt,
			&diagramTemplate.ShortDescription,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning diagram templates: " + err.Error()})

			return
		}
		diagramTemplates = append(diagramTemplates, diagramTemplate)
	}

	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating diagram templates: " + err.Error()})

		return
	}

	// Construct and send the response
	c.JSON(http.StatusOK, gin.H{
		"data":    diagramTemplates,
		"message": "Diagram templates retrieved successfully",
	})
}

// This function updates an existing diagram template
func UpdateInternalDiagramTemplate(c *gin.Context) {
	// Get the tenant ID from the context
	_, tenantID, ok := utilities.ProcessIdentity(c)

	if !ok {
		return
	}

	// Extracting project_template_id and diagram_template_id from the URL parameters
	projectTemplateID := c.Query("project_template_id")
	diagramTemplateID := c.Query("diagram_template_id")

	if projectTemplateID == "" || diagramTemplateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project Template ID and Diagram Template ID are required"})

		return
	}

	var updateData struct {
		Title            *string          `json:"title"`
		DiagramType      *string          `json:"diagram_type"`
		DiagramStatus    *string          `json:"diagram_status"`
		Category         *string          `json:"category"`
		Design           *json.RawMessage `json:"design"`
		RawDesign        []byte           `json:"raw_design"`
		ShortDescription *string          `json:"short_description"`
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
		setParts = append(setParts, fmt.Sprintf("p_design = $%d", argCounter))
		args = append(args, *updateData.Design)
		argCounter++
	}
	if updateData.RawDesign != nil {
		setParts = append(setParts, fmt.Sprintf("p_raw_design = $%d", argCounter))
		args = append(args, updateData.RawDesign)
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
            st_schema.diagram_templates
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
            title, diagram_type, diagram_status, category,
            p_design, created_at, updated_at, short_description
    `, setClause, argCounter, argCounter+1, argCounter+2)

	args = append(args, diagramTemplateID, projectTemplateID, tenantID)

	stmt, err := tenantManagement.DB.Prepare(query)
	if err != nil {
		if isDevelopmentEnvironment() {
			log.Printf("Error preparing statement: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})

		return
	}
	defer stmt.Close()

	var updatedTemplate models.DiagramTemplate
	err = stmt.QueryRow(args...).Scan(
		&updatedTemplate.ID,
		&updatedTemplate.UserID,
		&updatedTemplate.TenantID,
		&updatedTemplate.ProjectTemplateID,
		&updatedTemplate.Title,
		&updatedTemplate.DiagramType,
		&updatedTemplate.DiagramStatus,
		&updatedTemplate.Category,
		&updatedTemplate.Design,
		&updatedTemplate.CreatedAt,
		&updatedTemplate.UpdatedAt,
		&updatedTemplate.ShortDescription,
	)

	if err != nil {
		log.Printf(models.DatabaseError, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update the diagram template"})

		return
	}

	// Return the updated diagram template data
	c.JSON(http.StatusOK, gin.H{
		"data":    updatedTemplate,
		"message": "Diagram template updated successfully",
	})
}

// This function deletes an existing diagram template
func DeleteInternalDiagramTemplate(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)

	if !ok {
		return
	}

	projectTemplateID := c.Query("project_template_id")
	diagramTemplateID := c.Query("diagram_template_id")

	if projectTemplateID == "" || diagramTemplateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project Template ID and Diagram Template ID are required"})

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

	// Delete the diagram
	var diagram models.DiagramTemplate
	err = tx.QueryRow(`
        DELETE FROM st_schema.diagram_templates
        WHERE id = $1 AND project_template_id = $2 AND tenant_id = $3
        RETURNING id, user_id, tenant_id, project_template_id, 
            title, diagram_type, diagram_status, category,
            p_design, created_at, updated_at, short_description`,
		diagramTemplateID, projectTemplateID, tenantID,
	).Scan(
		&diagram.ID,
		&diagram.UserID,
		&diagram.TenantID,
		&diagram.ProjectTemplateID,
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
		log.Printf("Failed to delete diagram: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete the diagram"})

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
		"data":    diagram,
		"message": "Diagram template deleted successfully",
	})
}
