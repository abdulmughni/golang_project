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

// This function creates a new community diagram template
func NewPublicDiagramTemplate(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)

	if !ok {
		return
	}

	var diagram models.PublicDiagramTemplate

	// Extracting community project ID from the URL parameter
	communityProjectTemplateID := c.Query("community_project_template_id")

	if communityProjectTemplateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Community Project Template ID is required..."})

		return
	}

	if err := c.ShouldBindJSON(&diagram); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	diagram.UserID = &userID
	diagram.TenantID = &tenantID
	diagram.CommunityProjectTemplateID = &communityProjectTemplateID

	err := tenantManagement.DB.QueryRow(`
		INSERT INTO st_schema.cm_diagram_templates (
            user_id, tenant_id, community_project_template_id,
            title, short_description, diagram_type, diagram_status, category,
            design, raw_design, p_design, p_raw_design,
			published_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW()
        ) RETURNING id, community_project_template_id, user_id, tenant_id,
            title, short_description, diagram_type, category, p_design, published_at, last_update_at
		`,
		diagram.UserID, diagram.TenantID, diagram.CommunityProjectTemplateID,
		diagram.Title, diagram.ShortDescription, diagram.DiagramType, diagram.DiagramStatus, diagram.Category,
		diagram.Design, diagram.RawDesign, diagram.Design, diagram.RawDesign,
	).Scan(
		&diagram.ID,
		&diagram.CommunityProjectTemplateID,
		&diagram.UserID,
		&diagram.TenantID,
		&diagram.Title,
		&diagram.ShortDescription,
		&diagram.DiagramType,
		&diagram.Category,
		&diagram.Design,
		&diagram.CreatedAt,
		&diagram.UpdatedAt,
	)

	if err != nil {
		log.Printf("Database Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": models.DatabaseError})

		return
	}

	// Return the new diagram data
	c.JSON(200, gin.H{
		"data":    diagram,
		"message": "New community diagram template created successfully!",
	})
}

// GetPublicDiagramTemplate retrieves the diagram template details for authenticated users
func GetPublicDiagramTemplate(c *gin.Context) {
	// Extracting project ID and diagram ID from the URL parameters
	projectTemplateID := c.Query("cm_template_id")
	projectDiagramTemplateID := c.Query("cm_template_diagram_id")

	if projectTemplateID == "" || projectDiagramTemplateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parameters are missing..."})
		return
	}

	// Query to fetch the diagram for the specific template
	query := `
        SELECT
            id, community_project_template_id, title, diagram_type, category, p_design, published_at, last_update_at
        FROM
            st_schema.cm_diagram_templates
        WHERE
            community_project_template_id = $1
        AND
            id = $2`

	var diagram models.PublicDiagramTemplate
	err := tenantManagement.DB.QueryRow(query, projectTemplateID, projectDiagramTemplateID).Scan(
		&diagram.ID,
		&diagram.CommunityProjectTemplateID,
		&diagram.Title,
		&diagram.DiagramType,
		&diagram.Category,
		&diagram.Design,
		&diagram.CreatedAt,
		&diagram.UpdatedAt,
	)

	if err != nil {
		log.Printf(models.DatabaseError, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Diagram not found or access denied"})
		return
	}

	// Return the diagram data
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":                            diagram.ID,
			"community_project_template_id": diagram.CommunityProjectTemplateID,
			"title":                         diagram.Title,
			"diagram_type":                  diagram.DiagramType,
			"category":                      diagram.Category,
			"design":                        diagram.Design,
			"created_at":                    diagram.CreatedAt,
			"updated_at":                    diagram.UpdatedAt,
		},
		"message": "Project diagram retrieved successfully!",
	})
}

// This function retrieves an array of all diagram templates from a community project template
func GetPublicDiagramTemplates(c *gin.Context) {
	// Extracting community project ID from the URL parameter
	communityProjectTemplateID := c.Query("community_project_template_id")
	category := c.Query("category")

	if communityProjectTemplateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Community Project Template ID is required..."})

		return
	}

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

	query := `
        SELECT
            id, community_project_template_id, user_id, tenant_id,
            title, short_description, diagram_type, category, p_design,
            published_at, last_update_at
        FROM
            st_schema.cm_diagram_templates
    `

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	rows, err := tenantManagement.DB.Query(query, arguments...)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve diagram templates: " + err.Error()})

		return
	}
	defer rows.Close()

	var diagramTemplates []models.PublicDiagramTemplate

	for rows.Next() {
		var diagramTemplate models.PublicDiagramTemplate
		err := rows.Scan(
			&diagramTemplate.ID,
			&diagramTemplate.CommunityProjectTemplateID,
			&diagramTemplate.UserID,
			&diagramTemplate.TenantID,
			&diagramTemplate.Title,
			&diagramTemplate.ShortDescription,
			&diagramTemplate.DiagramType,
			&diagramTemplate.Category,
			&diagramTemplate.Design,
			&diagramTemplate.CreatedAt,
			&diagramTemplate.UpdatedAt,
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
		"message": "Community diagram templates retrieved successfully",
	})
}

// This function updates an existing community diagram template
func UpdatePublicDiagramTemplate(c *gin.Context) {
	// Get the tenant ID from the context
	userID, tenantID, ok := utilities.ProcessIdentity(c)

	if !ok {
		return
	}

	// Extracting community_project_template_id and diagram_template_id from the URL parameters
	communityProjectTemplateID := c.Query("community_project_template_id")
	diagramTemplateID := c.Query("diagram_template_id")

	if communityProjectTemplateID == "" || diagramTemplateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Community Project Template ID and Diagram Template ID are required"})

		return
	}

	var updateData struct {
		Title            *string          `json:"title"`
		ShortDescription *string          `json:"short_description"`
		DiagramType      *string          `json:"diagram_type"`
		DiagramStatus    *string          `json:"diagram_status"`
		Category         *string          `json:"category"`
		Design           *json.RawMessage `json:"design"`
		RawDesign        []byte           `json:"raw_design"`
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

	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No updatable fields provided"})

		return
	}

	// Always add last_update_at when updating
	setParts = append(setParts, "last_update_at = NOW()")

	setClause := strings.Join(setParts, ", ")
	query := fmt.Sprintf(`
        UPDATE
            st_schema.cm_diagram_templates
        SET
            %s
        WHERE
            id = $%d
        AND
            community_project_template_id = $%d
        AND
            user_id = $%d
        AND
            tenant_id = $%d
        RETURNING
            id, community_project_template_id, user_id, tenant_id,
            title, short_description, diagram_type, category, p_design,
            published_at, last_update_at
    `, setClause, argCounter, argCounter+1, argCounter+2, argCounter+3)

	args = append(args, diagramTemplateID, communityProjectTemplateID, userID, tenantID)

	stmt, err := tenantManagement.DB.Prepare(query)
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})

		return
	}
	defer stmt.Close()

	var updatedTemplate models.PublicDiagramTemplate
	err = stmt.QueryRow(args...).Scan(
		&updatedTemplate.ID,
		&updatedTemplate.CommunityProjectTemplateID,
		&updatedTemplate.UserID,
		&updatedTemplate.TenantID,
		&updatedTemplate.Title,
		&updatedTemplate.ShortDescription,
		&updatedTemplate.DiagramType,
		&updatedTemplate.Category,
		&updatedTemplate.Design,
		&updatedTemplate.CreatedAt,
		&updatedTemplate.UpdatedAt,
	)

	if err != nil {
		log.Printf(models.DatabaseError, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update the diagram template or not authorized"})

		return
	}

	// Return the updated diagram template data
	c.JSON(http.StatusOK, gin.H{
		"data":    updatedTemplate,
		"message": "Community diagram template updated successfully",
	})
}

// This function deletes an existing community diagram template
func DeletePublicDiagramTemplate(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)

	if !ok {
		return
	}

	communityProjectTemplateID := c.Query("community_project_template_id")
	diagramTemplateID := c.Query("diagram_template_id")

	if communityProjectTemplateID == "" || diagramTemplateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Community Project Template ID and Diagram Template ID are required"})

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
	var diagram models.PublicDiagramTemplate
	err = tx.QueryRow(`
        DELETE FROM st_schema.cm_diagram_templates
        WHERE id = $1
        AND community_project_template_id = $2
        AND user_id = $3
        AND tenant_id = $4
        RETURNING id, community_project_template_id, user_id, tenant_id,
            title, short_description, diagram_type, category, p_design,
            published_at, last_update_at`,
		diagramTemplateID, communityProjectTemplateID, userID, tenantID,
	).Scan(
		&diagram.ID,
		&diagram.CommunityProjectTemplateID,
		&diagram.UserID,
		&diagram.TenantID,
		&diagram.Title,
		&diagram.ShortDescription,
		&diagram.DiagramType,
		&diagram.Category,
		&diagram.Design,
		&diagram.CreatedAt,
		&diagram.UpdatedAt,
	)

	if err != nil {
		log.Printf("Failed to delete diagram: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete the diagram or not authorized"})

		return
	}

	// Update community project template timestamp
	updateProjectTemplateQuery := `
        UPDATE st_schema.cm_project_templates
        SET last_update_at = NOW()
        WHERE id = $1
    `
	_, err = tx.Exec(updateProjectTemplateQuery, communityProjectTemplateID)

	if err != nil {
		log.Printf("Failed to update community project template timestamp: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update the community project template timestamp"})

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
		"message": "Community diagram template deleted successfully",
	})
}

// GetWebPublicProjectTemplateDiagram retrieves the diagram template details for unauthenticated users
func GetWebPublicProjectTemplateDiagram(c *gin.Context) {
	// Extracting project ID and diagram ID from the URL parameters
	projectTemplateID := c.Query("cm_template_id")
	projectDiagramTemplateID := c.Query("cm_template_diagram_id")

	if projectTemplateID == "" || projectDiagramTemplateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parameters are missing..."})
		return
	}

	// Query to fetch the diagram for the specific template
	query := `
        SELECT
            id, community_project_template_id, title, diagram_type, category, p_design, published_at, last_update_at
        FROM
            st_schema.cm_diagram_templates
        WHERE
            community_project_template_id = $1
        AND
            id = $2`

	var diagram models.PublicDiagramTemplate
	err := tenantManagement.DB.QueryRow(query, projectTemplateID, projectDiagramTemplateID).Scan(
		&diagram.ID,
		&diagram.CommunityProjectTemplateID,
		&diagram.Title,
		&diagram.DiagramType,
		&diagram.Category,
		&diagram.Design,
		&diagram.CreatedAt,
		&diagram.UpdatedAt,
	)

	if err != nil {
		log.Printf(models.DatabaseError, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Diagram not found or access denied"})
		return
	}

	// Return the diagram data
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":                            diagram.ID,
			"community_project_template_id": diagram.CommunityProjectTemplateID,
			"title":                         diagram.Title,
			"diagram_type":                  diagram.DiagramType,
			"category":                      diagram.Category,
			"design":                        diagram.Design,
			"created_at":                    diagram.CreatedAt,
			"updated_at":                    diagram.UpdatedAt,
		},
		"message": "Project diagram retrieved successfully!",
	})
}
