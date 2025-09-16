package projects

import (
	"log"
	"net/http"
	"sententiawebapi/handlers/apis/tenantManagement"
	"sententiawebapi/handlers/models"
	"sententiawebapi/utilities"

	"github.com/gin-gonic/gin"
)

type ProjectEntity struct {
	ID           string  `json:"id"`
	UserID       string  `json:"user_id"`
	Title        string  `json:"title"`
	UpdatedAt    string  `json:"updated_at"`
	Type         *string `json:"type,omitempty"`
	ProjectID    string  `json:"project_id"`
	DocumentType *string `json:"document_type,omitempty"`
	Category     *string `json:"category,omitempty"`
}

type GroupedProjectEntities struct {
	Documents []ProjectEntity `json:"documents"`
	Diagrams  []ProjectEntity `json:"diagrams"`
	Decisions []ProjectEntity `json:"decisions"`
}

func ListAllProjectEntities(c *gin.Context) {
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

	query := `
		SELECT 
			id, 
			user_id, 
			title, 
			updated_at, 
			'document' AS entity_type, 
			project_id,
			document_type AS category
		FROM st_schema.project_documents
		WHERE tenant_id = $1 AND project_id = $2

		UNION ALL

		SELECT 
			id, 
			user_id, 
			title, 
			updated_at, 
			'diagram' AS entity_type, 
			project_id,
			diagram_type AS category
		FROM st_schema.diagrams
		WHERE tenant_id = $1 AND project_id = $2

		UNION ALL

		SELECT 
			id, 
			user_id, 
			tbar_title AS title, 
			updated_at, 
			'tchart' AS entity_type, 
			project_id,
			tbar_category AS category
		FROM st_schema.tbar_analysis
		WHERE tenant_id = $1 AND project_id = $2

		UNION ALL

		SELECT 
			id, 
			user_id, 
			title, 
			updated_at, 
			'pnc' AS entity_type, 
			project_id,
			category
		FROM st_schema.pnc_analysis
		WHERE tenant_id = $1 AND project_id = $2

		UNION ALL

		SELECT 
			id, 
			user_id, 
			title, 
			updated_at, 
			'swot' AS entity_type, 
			project_id,
			category
		FROM st_schema.swot_analysis
		WHERE tenant_id = $1 AND project_id = $2

		UNION ALL

		SELECT 
			id, 
			user_id, 
			title, 
			updated_at, 
			'matrix' AS entity_type, 
			project_id,
			category
		FROM st_schema.matrix_analysis
		WHERE tenant_id = $1 AND project_id = $2

		ORDER BY updated_at DESC
	`

	// Execute the query
	rows, err := tenantManagement.DB.Query(query, tenantID, projectID)
	if err != nil {
		log.Printf(models.DatabaseError, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	defer rows.Close()

	// Initialize listing with empty arrays
	listing := GroupedProjectEntities{
		Documents: []ProjectEntity{},
		Diagrams:  []ProjectEntity{},
		Decisions: []ProjectEntity{},
	}

	for rows.Next() {
		var entity ProjectEntity
		var entityType string

		err := rows.Scan(&entity.ID, &entity.UserID, &entity.Title, &entity.UpdatedAt, &entityType, &entity.ProjectID, &entity.Category)

		if err != nil {
			log.Printf("Error scanning row: %v", err)
		} else {
			switch entityType {
			case "document":
				listing.Documents = append(listing.Documents, entity)
			case "diagram":
				listing.Diagrams = append(listing.Diagrams, entity)
			case "tchart", "pnc", "swot", "matrix":
				entity.Type = &entityType
				listing.Decisions = append(listing.Decisions, entity)
			default:
				log.Printf("Unknown entity type: %s", entityType)
			}
		}
	}

	c.JSON(http.StatusOK, listing)
}
