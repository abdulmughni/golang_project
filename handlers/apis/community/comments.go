package community

import (
	"log"
	"net/http"

	"sententiawebapi/handlers/apis/tenantManagement"
	"sententiawebapi/handlers/models"
	"sententiawebapi/utilities"

	"github.com/gin-gonic/gin"
)

// PostPublicComment allows a user to post a comment on a public project template.
func PostPublicComment(c *gin.Context) {
	userID, _, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	templateID := c.Query("template_id")
	if templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template ID required"})
		return
	}

	var comment models.Comment
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if comment.Comment == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Comment content cannot be empty"})
		return
	}

	// Prepare a statement for inserting data
	stmt, err := tenantManagement.DB.Prepare(`
		INSERT INTO
			st_schema.cm_public_templates_comments (template_id, user_id, comment)
		VALUES
			($1, $2, $3)
			RETURNING id, template_id, user_id, comment, created_at
		`)

	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database preparation error"})
		return
	}
	defer stmt.Close()

	// Execute the prepared statement and capture the result values

	err = stmt.QueryRow(
		templateID,
		userID,
		comment.Comment,
	).Scan(
		&comment.ID,
		&comment.TemplateID,
		&comment.UserID,
		&comment.Comment,
		&comment.CreatedAt,
	)

	if err != nil {
		log.Printf("Error executing statement: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to post comment"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":    comment,
		"message": "Comment posted successfully",
	})
}

func GetPublicComments(c *gin.Context) {
	_, _, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	templateID := c.Query("template_id")
	if templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template ID required"})
		return
	}

	// Prepare the SQL query to fetch all comments ordered by creation date descending
	rows, err := tenantManagement.DB.Query(`
		SELECT 
			id, template_id, user_id, comment, created_at
		FROM 
			st_schema.cm_public_templates_comments
		WHERE	
			template_id = $1
		ORDER BY 
			created_at DESC
	`, templateID)
	if err != nil {
		log.Printf("Error preparing query: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query preparation error"})
		return
	}
	defer rows.Close()

	// Slice to hold all comments
	var comments []models.Comment

	// Iterate over the rows
	for rows.Next() {
		var comment models.Comment
		if err := rows.Scan(
			&comment.ID,
			&comment.TemplateID,
			&comment.UserID,
			&comment.Comment,
			&comment.CreatedAt); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue // Skip this row on error, but try to process others
		}
		comments = append(comments, comment)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		log.Printf("Error during rows iteration: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating database rows"})
		return
	}

	// Return the list of comments
	c.JSON(http.StatusOK, gin.H{
		"data":    comments,
		"message": "Comments retrieved successfully",
	})
}

func UpdatePublicComment(c *gin.Context) {
	userID, _, ok := utilities.ProcessIdentity(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	templateId := c.Query("template_id")
	if templateId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Comment ID required"})
		return
	}

	commentID := c.Query("comment_id")
	if commentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Comment ID required"})
		return
	}

	var commentUpdate models.Comment
	if err := c.ShouldBindJSON(&commentUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if commentUpdate.Comment == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Comment content cannot be empty"})
		return
	}

	// Prepare a statement for updating the comment
	stmt, err := tenantManagement.DB.Prepare(`
		UPDATE
			st_schema.cm_public_templates_comments
		SET
			comment = $1
		WHERE
			id = $2
		AND
			user_id = $3
		AND
			template_id = $4
		RETURNING id, template_id, user_id, comment, created_at
	`)

	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database preparation error"})
		return
	}
	defer stmt.Close()

	// Execute the prepared statement
	var comment models.Comment
	err = stmt.QueryRow(
		commentUpdate.Comment,
		commentID,
		userID,
		templateId,
	).Scan(
		&comment.ID,
		&comment.TemplateID,
		&comment.UserID,
		&comment.Comment,
		&comment.CreatedAt,
	)

	if err != nil {
		log.Printf("Error executing statement: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    comment,
		"message": "Comment updated successfully",
	})
}

func DeletePublicComment(c *gin.Context) {
	userID, _, ok := utilities.ProcessIdentity(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	commentID := c.Query("comment_id")
	if commentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Comment ID required"})
		return
	}

	// Prepare a statement for deleting the comment
	stmt, err := tenantManagement.DB.Prepare(`
		DELETE FROM
			st_schema.cm_public_templates_comments
		WHERE
			id = $1
		AND
			user_id = $2
		RETURNING id, template_id, user_id, comment, created_at
	`)

	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database preparation error"})
		return
	}
	defer stmt.Close()

	// Execute the prepared statement
	var comment models.Comment
	err = stmt.QueryRow(commentID, userID).Scan(
		&comment.ID,
		&comment.TemplateID,
		&comment.UserID,
		&comment.Comment,
		&comment.CreatedAt,
	)

	if err != nil {
		log.Printf("Error executing statement: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    comment,
		"message": "Comment deleted successfully",
	})
}
