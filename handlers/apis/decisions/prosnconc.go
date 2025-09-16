package decisions

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"sententiawebapi/handlers/apis/tenantManagement"
	models "sententiawebapi/handlers/models"
	"sententiawebapi/utilities"
	"strings"

	"github.com/gin-gonic/gin"
)

// This function creates a new pros and cons analysis object under a project.
func NewPncAnalysis(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	var pnc models.PncAnalysis

	pnc.UserID = userID
	pnc.TenantID = tenantID
	pnc.ProjectID, ok = utilities.ValidateQueryParam(c, "project_id")
	if !ok {
		return
	}

	if err := c.ShouldBindJSON(&pnc); err != nil {
		log.Printf("ERROR: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal server error..."})
		return
	}

	row := tenantManagement.DB.QueryRow(`
		INSERT INTO st_schema.pnc_analysis (
			user_id,
			tenant_id,
			title,
			pnc_description,
			pnc_status,
			category,
			better_option,
			assumptions,
			final_decision,
			architectural_decision_id,
			implications,
			project_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id
	`,
		userID,
		tenantID,
		pnc.Title,
		pnc.PNCDescription,
		pnc.PNCStatus,
		pnc.Category,
		pnc.BetterOption,
		pnc.Assumptions,
		pnc.FinalDecision,
		pnc.ADecisionId,
		pnc.Implications,
		pnc.ProjectID,
	)

	err := row.Scan(&pnc.ID)
	if err != nil {
		log.Printf("ERROR: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	// Prepare the response data
	data := map[string]interface{}{
		"id":      pnc.ID,
		"details": pnc,
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    data,
		"message": "PNC analysis created successfully!",
	})
}

// This function retrieves all pros and cons analysis objects under a project.
// This function retrieves a specific pros and cons analysis object under a project.
func GetPncAnalysis(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	var pnc models.PncAnalysis

	pnc.UserID = userID
	pnc.TenantID = tenantID

	// Validate and extract the project ID from the query parameters
	pnc.ProjectID, ok = utilities.ValidateQueryParam(c, "project_id")
	if !ok {
		return
	}

	// Validate and extract the analysis ID from the query parameters
	pnc.ID, ok = utilities.ValidateQueryParam(c, "pnc_id")
	if !ok {
		return
	}

	// Query the database for the specific PNC analysis
	query := `
		SELECT
			title,
			pnc_description,
			pnc_status,
			category,
			better_option,
			assumptions,
			final_decision,
			architectural_decision_id,
			implications
		FROM
			st_schema.pnc_analysis
		WHERE
			id = $1
		AND
			tenant_id = $2
		AND
			project_id = $3
	`

	err := tenantManagement.DB.QueryRow(query, pnc.ID, tenantID, pnc.ProjectID).Scan(
		&pnc.Title,
		&pnc.PNCDescription,
		&pnc.PNCStatus,
		&pnc.Category,
		&pnc.BetterOption,
		&pnc.Assumptions,
		&pnc.FinalDecision,
		&pnc.ADecisionId,
		&pnc.Implications,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("ERROR: Analysis not found for ID %s", pnc.ID)
			c.JSON(http.StatusNotFound, gin.H{"error": "Analysis not found"})
			return
		}
		log.Printf("ERROR: Failed to retrieve the analysis details: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	// Build and send the response
	c.JSON(http.StatusOK, gin.H{
		"data":    pnc,
		"message": "PNC analysis retrieved successfully!",
	})
}

// This function retrieves all pros and cons analysis objects under a project.
func GetAllPncAnalysis(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		log.Printf("ERROR: %v", "User ID is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal server error..."})
		return
	}

	projectID, ok := utilities.ValidateQueryParam(c, "project_id")
	if !ok {
		return
	}

	query := `
		SELECT
			id,
			user_id,
			tenant_id,
			title,
			pnc_description,
			pnc_status,
			category,
			better_option,
			assumptions,
			final_decision,
			architectural_decision_id,
			implications,
			project_id
		FROM
			st_schema.pnc_analysis
		WHERE
			tenant_id = $1
		AND
			project_id = $2
	`

	rows, err := tenantManagement.DB.Query(query, tenantID, projectID)
	if err != nil {
		log.Printf("ERROR: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}
	defer rows.Close()

	// Scan the results into a slice of PncAnalysis
	var analyses []models.PncAnalysis
	for rows.Next() {
		var analysis models.PncAnalysis
		if err := rows.Scan(
			&analysis.ID,
			&analysis.UserID,
			&analysis.TenantID,
			&analysis.Title,
			&analysis.PNCDescription,
			&analysis.PNCStatus,
			&analysis.Category,
			&analysis.BetterOption,
			&analysis.Assumptions,
			&analysis.FinalDecision,
			&analysis.ADecisionId,
			&analysis.Implications,
			&analysis.ProjectID,
		); err != nil {
			log.Printf("ERROR: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
			return
		}

		analyses = append(analyses, analysis)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		log.Printf("ERROR: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	// Build and send the response
	c.JSON(http.StatusOK, gin.H{
		"data":    analyses,
		"message": "All analysis objects retrieved successfully!",
	})
}

// Updates pros and cons analysis object under a project.
func UpdatePncAnalysis(c *gin.Context) {
	var pnc models.PncAnalysis

	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	pnc.UserID = userID
	pnc.TenantID = tenantID

	// Validate and extract the project ID from the query parameters
	pnc.ProjectID, ok = utilities.ValidateQueryParam(c, "project_id")
	if !ok {
		return
	}

	// Get the analysis ID from the URL parameters
	pnc.ID, ok = utilities.ValidateQueryParam(c, "pnc_id")
	if !ok {
		return
	}

	// Bind the JSON body to the PncAnalysis struct
	if err := c.ShouldBindJSON(&pnc); err != nil {
		log.Printf("ERROR: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal server error..."})
		return
	}

	// Use reflection to build the update query dynamically
	v := reflect.ValueOf(pnc)
	t := v.Type()

	setParts := []string{}
	args := []interface{}{}
	argCounter := 1

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		if fieldValue.Kind() == reflect.Ptr && !fieldValue.IsNil() {
			fieldName := t.Field(i).Tag.Get("json")
			setParts = append(setParts, fmt.Sprintf("%s = $%d", fieldName, argCounter))
			args = append(args, fieldValue.Elem().Interface())
			argCounter++
		}
	}

	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal server error..."})
		return
	}

	setClause := strings.Join(setParts, ", ")
	query := fmt.Sprintf(`
		UPDATE
			st_schema.pnc_analysis
		SET
			%s, updated_at = NOW()
		WHERE
			id = $%d
		AND
			tenant_id = $%d
		AND
			project_id = $%d
		RETURNING
			id, user_id, tenant_id, title, pnc_description, pnc_status, category, better_option, assumptions, final_decision, architectural_decision_id, implications, project_id
	`, setClause, argCounter, argCounter+1, argCounter+2)

	args = append(args, pnc.ID, pnc.TenantID, pnc.ProjectID)

	var updatedAnalysis models.PncAnalysis
	err := tenantManagement.DB.QueryRow(query, args...).Scan(
		&updatedAnalysis.ID,
		&updatedAnalysis.UserID,
		&updatedAnalysis.TenantID,
		&updatedAnalysis.Title,
		&updatedAnalysis.PNCDescription,
		&updatedAnalysis.PNCStatus,
		&updatedAnalysis.Category,
		&updatedAnalysis.BetterOption,
		&updatedAnalysis.Assumptions,
		&updatedAnalysis.FinalDecision,
		&updatedAnalysis.ADecisionId,
		&updatedAnalysis.Implications,
		&updatedAnalysis.ProjectID,
	)

	if err != nil {
		log.Printf("Failed to update PNC analysis: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	// Return the updated PNC analysis data
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"pnc_analysis": updatedAnalysis,
		},
		"message": "Pros & cons analysis updated successfully",
	})
}

// Deletes a pros and cons analysis object under a project.
func DeletePncAnalysis(c *gin.Context) {
	var pnc models.PncAnalysis

	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	pnc.UserID = userID
	pnc.TenantID = tenantID
	pnc.ProjectID, ok = utilities.ValidateQueryParam(c, "project_id")
	if !ok {
		return
	}

	pnc.ID, ok = utilities.ValidateQueryParam(c, "pnc_id")
	if !ok {
		return
	}

	// Start a transaction
	tx, err := tenantManagement.DB.Begin()
	if err != nil {
		log.Printf("ERROR: Failed to start a transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	// Select the PNC analysis data before deletion
	err = tx.QueryRow(`
		SELECT id, user_id, tenant_id, title, pnc_description, pnc_status, category, better_option, assumptions, final_decision, architectural_decision_id, implications, project_id
		FROM st_schema.pnc_analysis
		WHERE id = $1 AND tenant_id = $2 AND project_id = $3
	`, pnc.ID, pnc.TenantID, pnc.ProjectID).Scan(
		&pnc.ID,
		&pnc.UserID,
		&pnc.TenantID,
		&pnc.Title,
		&pnc.PNCDescription,
		&pnc.PNCStatus,
		&pnc.Category,
		&pnc.BetterOption,
		&pnc.Assumptions,
		&pnc.FinalDecision,
		&pnc.ADecisionId,
		&pnc.Implications,
		&pnc.ProjectID,
	)
	if err != nil {
		log.Printf("ERROR: Failed to retrieve PNC analysis before deletion: %v", err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	// Delete the PNC analysis
	_, err = tx.Exec("DELETE FROM st_schema.pnc_analysis WHERE id = $1 AND tenant_id = $2 AND project_id = $3", pnc.ID, pnc.TenantID, pnc.ProjectID)
	if err != nil {
		log.Printf("ERROR: Failed to delete the PNC analysis: %v", err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		log.Printf("ERROR: Failed to commit the transaction: %v", err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	// Respond with the deleted PNC analysis data and a success message
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"pnc_analysis": pnc,
		},
		"message": "PNC analysis and associated arguments deleted successfully!",
	})
}

// This function creates a new pros and cons argument object under a project PNC.
func NewPncArgument(c *gin.Context) {
	var pncArgument models.PncArgument

	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	pncArgument.UserID = userID
	pncArgument.TenantID = tenantID

	pncArgument.PncID, ok = utilities.ValidateQueryParam(c, "pnc_id")
	if !ok {
		return
	}

	if err := c.ShouldBindJSON(&pncArgument); err != nil {
		log.Printf("ERROR: Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if pncArgument.Side != "con" && pncArgument.Side != "pro" {
		log.Printf("ERROR: Side value must be either 'pro' or 'con'")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal server error ..."})
		return
	}

	row := tenantManagement.DB.QueryRow(
		`INSERT INTO st_schema.pnc_arguments (
            pnc_id,
            user_id,
            tenant_id,
            argument,
            argument_weight,
            side,
			description
        ) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		pncArgument.PncID,
		pncArgument.UserID,
		pncArgument.TenantID,
		pncArgument.Argument,
		pncArgument.ArgumentWeight,
		pncArgument.Side,
		pncArgument.Description,
	)

	err := row.Scan(&pncArgument.ID)
	if err != nil {
		log.Printf("ERROR: Failed to retrieve the argument ID: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	data := map[string]interface{}{
		"id":      pncArgument.ID,
		"details": pncArgument,
	}

	// Respond with the created PNC argument data
	c.JSON(http.StatusOK, gin.H{
		"data":    data,
		"message": "PNC argument created successfully!",
	})
}

// This function retrieves a specific pros and cons argument object under a project PNC.
func GetAllPncArguments(c *gin.Context) {
	var arguments []models.PncArgument

	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	// Validate and extract the PNC ID from the query parameters
	pncID, ok := utilities.ValidateQueryParam(c, "pnc_id")
	if !ok {
		return
	}

	// Query the database to get all arguments associated with the PNC ID and User ID
	rows, err := tenantManagement.DB.Query(`
		SELECT
			id,
			pnc_id,
			user_id,
			tenant_id,
			argument,
			argument_weight,
			side,
			description
		FROM
			st_schema.pnc_arguments
		WHERE
			pnc_id = $1
		AND
			tenant_id = $2`,
		pncID, tenantID)
	if err != nil {
		log.Printf("ERROR: Failed to retrieve arguments: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	defer rows.Close()

	// Scan the results into a slice of PncArgument
	for rows.Next() {
		var argument models.PncArgument
		if err := rows.Scan(
			&argument.ID,
			&argument.PncID,
			&argument.UserID,
			&argument.TenantID,
			&argument.Argument,
			&argument.ArgumentWeight,
			&argument.Side,
			&argument.Description,
		); err != nil {
			log.Printf("ERROR: Failed to scan argument: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		arguments = append(arguments, argument)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		log.Printf("ERROR: Failed to iterate over arguments: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Build and send the response
	c.JSON(http.StatusOK, gin.H{
		"data":    arguments,
		"message": "Arguments retrieved successfully!",
	})
}

// This function retrieves a specific pros and cons argument object under a project PNC.
func UpdatePncArgument(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	var pncArgument models.PncArgument
	pncArgument.UserID = userID
	pncArgument.TenantID = tenantID

	pncArgument.PncID, ok = utilities.ValidateQueryParam(c, "pnc_id")
	if !ok {
		return
	}

	// Validate and extract the argument ID from the query parameters
	pncArgument.ID, ok = utilities.ValidateQueryParam(c, "argument_id")
	if !ok {
		return
	}

	// Bind the JSON body to the PncArgument struct
	if err := c.ShouldBindJSON(&pncArgument); err != nil {
		log.Printf("ERROR: Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Use reflection to build the update query dynamically
	v := reflect.ValueOf(pncArgument)
	t := v.Type()

	setParts := []string{}
	args := []interface{}{}
	argCounter := 1

	// Loop through the fields of the PncArgument struct
	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		fieldName := t.Field(i).Tag.Get("json")

		// Remove `,omitempty` if it exists in the JSON tag
		if idx := strings.Index(fieldName, ",omitempty"); idx != -1 {
			fieldName = fieldName[:idx]
		}

		// Check if the field is a string and not empty
		if fieldValue.Kind() == reflect.String && fieldValue.String() != "" {
			setParts = append(setParts, fmt.Sprintf("%s = $%d", fieldName, argCounter))
			args = append(args, fieldValue.Interface())
			argCounter++
		} else if fieldValue.Kind() != reflect.String && !fieldValue.IsZero() {
			// Handle non-string fields
			setParts = append(setParts, fmt.Sprintf("%s = $%d", fieldName, argCounter))
			args = append(args, fieldValue.Interface())
			argCounter++
		}
	}

	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	setClause := strings.Join(setParts, ", ")
	query := fmt.Sprintf(`
        UPDATE
            st_schema.pnc_arguments
        SET
            %s
        WHERE
            id = $%d
        AND
            tenant_id = $%d
        AND
            pnc_id = $%d
        RETURNING
            id, argument, argument_weight, side, user_id, tenant_id, pnc_id, description
    `, setClause, argCounter, argCounter+1, argCounter+2)

	args = append(args, pncArgument.ID, pncArgument.TenantID, pncArgument.PncID)

	var updatedArgument models.PncArgument
	err := tenantManagement.DB.QueryRow(query, args...).Scan(
		&updatedArgument.ID,
		&updatedArgument.Argument,
		&updatedArgument.ArgumentWeight,
		&updatedArgument.Side,
		&updatedArgument.UserID,
		&updatedArgument.TenantID,
		&updatedArgument.PncID,
		&updatedArgument.Description,
	)

	if err != nil {
		log.Printf("Failed to update PNC argument: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update the argument"})
		return
	}

	// Return the updated PNC argument data
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"pnc_argument": updatedArgument,
		},
		"message": "Argument updated successfully!",
	})
}

// Deletes the argument object under a project PNC.
func DeletePncArgument(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	var pncArgument models.PncArgument
	pncArgument.TenantID = tenantID

	pncArgument.ID, ok = utilities.ValidateQueryParam(c, "argument_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal server error..."})
		return
	}

	// Use RETURNING clause to get the details of the deleted argument
	var deletedArgument models.PncArgument
	err := tenantManagement.DB.QueryRow(`
		DELETE FROM
			st_schema.pnc_arguments
		WHERE
			id = $1
		AND
			tenant_id = $2
		RETURNING
			id, argument, argument_weight, side, user_id, tenant_id, pnc_id, description`,
		pncArgument.ID,
		pncArgument.TenantID,
	).Scan(
		&deletedArgument.ID,
		&deletedArgument.Argument,
		&deletedArgument.ArgumentWeight,
		&deletedArgument.Side,
		&deletedArgument.UserID,
		&deletedArgument.TenantID,
		&deletedArgument.PncID,
		&deletedArgument.Description,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("ERROR: Argument not found for ID %s", pncArgument.ID)
			c.JSON(http.StatusNotFound, gin.H{"error": "Internal server error..."})
		} else {
			log.Printf("ERROR: Failed to delete the argument: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		}
		return
	}

	// Respond with a success message and details of the deleted argument
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"details": gin.H{
				"id":              deletedArgument.ID,
				"argument":        deletedArgument.Argument,
				"argument_weight": deletedArgument.ArgumentWeight,
				"side":            deletedArgument.Side,
				"user_id":         deletedArgument.UserID,
				"tenant_id":       deletedArgument.TenantID,
				"pnc_id":          deletedArgument.PncID,
				"description":     deletedArgument.Description,
			},
			"id": deletedArgument.ID,
		},
		"message": "Argument deleted successfully!",
	})
}
