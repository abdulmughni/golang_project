package templates

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"sententiawebapi/handlers/apis/tenantManagement"
	"sententiawebapi/handlers/models"
	"sententiawebapi/utilities"

	"github.com/gin-gonic/gin"
)

func GetDocumentComponent(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	componentID := c.Query("id")
	if componentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Component ID required"})
		return
	}

	stmt, err := tenantManagement.DB.Prepare(`
		SELECT
			c.id, c.owner, c.tenant_id, c.title, c.category,
			c.description, c.short_description,
			c.content, c.icon, c.created_at, c.last_update_at,
			CASE WHEN tdc.component_id IS NOT NULL THEN true ELSE false END AS is_favorite
		FROM
			st_schema.cm_document_components c
		LEFT JOIN
			st_schema.tenant_document_components tdc
		ON
			c.id = tdc.component_id AND tdc.tenant_id = $1
		WHERE
			c.id = $2
	`)
	if err != nil {
		log.Printf("Database Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	defer stmt.Close()

	var component models.DocumentComponent
	var isFavorite sql.NullBool
	var contentJSON string
	err = stmt.QueryRow(tenantID, componentID).Scan(
		&component.ID,
		&component.Owner,
		&component.TenantID,
		&component.Title,
		&component.Category,
		&component.Description,
		&component.ShortDescription,
		&contentJSON,
		&component.Icon,
		&component.CreatedAt,
		&component.LastUpdateAt,
		&isFavorite,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Document component not found"})
		} else {
			log.Printf("Database Err: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	if isFavorite.Valid {
		component.IsFavorite = &isFavorite.Bool
	} else {
		component.IsFavorite = nil
	}

	// Unmarshal content from JSON string to plain string
	var contentStruct struct {
		Template string `json:"template"`
	}
	err = json.Unmarshal([]byte(contentJSON), &contentStruct)
	if err != nil {
		if isDevelopmentEnvironment() {
			log.Printf("JSON Unmarshal Err: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	component.Content = &contentStruct.Template

	c.JSON(http.StatusOK, gin.H{
		"data":    component,
		"message": models.StatusSuccess,
	})
}

func GetDocumentComponents(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	stmt, err := tenantManagement.DB.Prepare(`
		SELECT
			c.id, c.owner, c.tenant_id, c.title, c.category,
			c.short_description, c.icon,
			CASE WHEN tdc.component_id IS NOT NULL THEN true ELSE false END AS is_favorite
		FROM
			st_schema.cm_document_components c
		LEFT JOIN
			st_schema.tenant_document_components tdc
		ON
			c.id = tdc.component_id AND tdc.tenant_id = $1
	`)
	if err != nil {
		if isDevelopmentEnvironment() {
			log.Printf("Database Err: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	defer stmt.Close()

	rows, err := stmt.Query(tenantID)
	if err != nil {
		if isDevelopmentEnvironment() {
			log.Printf("Database Err: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	defer rows.Close()

	var components []models.DocumentComponents
	for rows.Next() {
		var component models.DocumentComponents
		var isFavorite sql.NullBool
		err := rows.Scan(
			&component.ID,
			&component.Owner,
			&component.TenantID,
			&component.Title,
			&component.Category,
			&component.ShortDescription,
			&component.Icon,
			&isFavorite,
		)
		if err != nil {
			log.Printf("Database Err: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		if isFavorite.Valid {
			component.IsFavorite = &isFavorite.Bool
		} else {
			component.IsFavorite = nil
		}

		components = append(components, component)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Database Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    components,
		"message": models.StatusSuccess,
	})
}

func GetFavoriteDocumentComponents(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	stmt, err := tenantManagement.DB.Prepare(`
		SELECT
			c.id, c.owner, c.tenant_id, c.title, c.category,
			c.description, c.short_description,
			c.content, c.icon, c.created_at, c.last_update_at,
			true AS is_favorite
		FROM
			st_schema.cm_document_components c
		INNER JOIN
			st_schema.tenant_document_components tdc
		ON
			c.id = tdc.component_id AND tdc.tenant_id = $1
	`)
	if err != nil {
		if isDevelopmentEnvironment() {
			log.Printf("Database Err: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	defer stmt.Close()

	rows, err := stmt.Query(tenantID)
	if err != nil {
		if isDevelopmentEnvironment() {
			log.Printf("Database Err: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	defer rows.Close()

	var components []models.DocumentComponent
	for rows.Next() {
		var component models.DocumentComponent
		var isFavorite sql.NullBool
		var contentJSON string
		err := rows.Scan(
			&component.ID,
			&component.Owner,
			&component.TenantID,
			&component.Title,
			&component.Category,
			&component.Description,
			&component.ShortDescription,
			&contentJSON,
			&component.Icon,
			&component.CreatedAt,
			&component.LastUpdateAt,
			&isFavorite,
		)
		if err != nil {
			if isDevelopmentEnvironment() {
				log.Printf("Database Err: %v", err)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		if isFavorite.Valid {
			component.IsFavorite = &isFavorite.Bool
		} else {
			component.IsFavorite = nil
		}

		// Unmarshal content from JSON string to plain string
		var contentStruct struct {
			Template string `json:"template"`
		}
		err = json.Unmarshal([]byte(contentJSON), &contentStruct)
		if err != nil {
			if isDevelopmentEnvironment() {
				log.Printf("JSON Unmarshal Err: %v", err)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		component.Content = &contentStruct.Template

		components = append(components, component)
	}

	if err := rows.Err(); err != nil {
		if isDevelopmentEnvironment() {
			log.Printf("Database Err: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    components,
		"message": models.StatusSuccess,
	})
}

func PinDocumentComponent(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	componentID := c.Query("id")
	if componentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Component ID required"})
		return
	}

	// Prepare insert statement, ON CONFLICT means skip if the component is already pinned
	stmt, err := tenantManagement.DB.Prepare(`
		INSERT INTO
			st_schema.tenant_document_components (user_id, tenant_id, component_id)
		VALUES
			($1, $2, $3)
		ON CONFLICT
			(tenant_id, component_id) DO NOTHING
	`)
	if err != nil {
		if isDevelopmentEnvironment() {
			log.Printf("Database Prepare Err: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	defer stmt.Close()

	// Execute the query
	_, err = stmt.Exec(userID, tenantID, componentID)
	if err != nil {
		log.Printf("Database Exec Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Retrieve the pinned document component
	stmt, err = tenantManagement.DB.Prepare(`
		SELECT
			c.id, c.owner, c.tenant_id, c.title, c.category,
			c.description, c.short_description,
			c.content, c.icon, c.created_at, c.last_update_at,
			CASE WHEN tdc.component_id IS NOT NULL THEN true ELSE false END AS is_favorite
		FROM
			st_schema.cm_document_components c
		LEFT JOIN
			st_schema.tenant_document_components tdc
		ON
			c.id = tdc.component_id AND tdc.tenant_id = $1
		WHERE
			c.id = $2
	`)
	if err != nil {
		log.Printf("Database Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	defer stmt.Close()

	var component models.DocumentComponent
	var isFavorite sql.NullBool
	err = stmt.QueryRow(tenantID, componentID).Scan(
		&component.ID,
		&component.Owner,
		&component.TenantID,
		&component.Title,
		&component.Category,
		&component.Description,
		&component.ShortDescription,
		&component.Content,
		&component.Icon,
		&component.CreatedAt,
		&component.LastUpdateAt,
		&isFavorite,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Document component not found"})
		} else {
			log.Printf("Database Err: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	if isFavorite.Valid {
		component.IsFavorite = &isFavorite.Bool
	} else {
		component.IsFavorite = nil
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Document component successfully pinned",
		"data":    component,
	})
}

func UnpinDocumentComponent(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	componentID := c.Query("id")
	if componentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Component ID required"})
		return
	}

	// Prepare delete statement
	stmt, err := tenantManagement.DB.Prepare(`
		DELETE FROM
			st_schema.tenant_document_components
		WHERE
			tenant_id = $1
		AND
			component_id = $2
	`)
	if err != nil {
		if isDevelopmentEnvironment() {
			log.Printf("Database Prepare Err: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(tenantID, componentID)
	if err != nil {
		if isDevelopmentEnvironment() {
			log.Printf("Database Exec Err: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Check if any row was affected
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		if isDevelopmentEnvironment() {
			log.Printf("RowsAffected Err: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document component not found or already unpinned"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Document component successfully unpinned",
		"data":    map[string]string{"id": componentID},
	})
}
