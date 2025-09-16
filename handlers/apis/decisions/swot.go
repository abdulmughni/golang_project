package decisions

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"

	"sententiawebapi/handlers/apis/tenantManagement"
	models "sententiawebapi/handlers/models"
	"sententiawebapi/utilities"

	"github.com/gin-gonic/gin"
)

func NewSwot(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	var swot models.Swot
	swot.UserId = new(string)
	swot.TenantID = new(string)
	*swot.UserId = userID
	*swot.TenantID = tenantID

	swot.ProjectID = new(string)
	*swot.ProjectID, ok = utilities.ValidateQueryParam(c, "project_id")
	if !ok {
		return
	}

	if err := c.ShouldBindJSON(&swot); err != nil {
		log.Printf("BINDING ERROR: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal server error..."})
		return
	}

	row := tenantManagement.DB.QueryRow(
		`INSERT INTO st_schema.swot_analysis (
			user_id,
			tenant_id,
			project_id,
			title,
			swot_description,
			swot_status,
			category,
			assumptions,
			final_decision,
			architectural_decision_id,
			implications
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id`,
		swot.UserId,
		swot.TenantID,
		swot.ProjectID,
		swot.Title,
		swot.SwotDescription,
		swot.SwotStatus,
		swot.Category,
		swot.Assumptions,
		swot.FinalDecision,
		swot.ADecisionId,
		swot.Implications,
	)

	err := row.Scan(&swot.ID)
	if err != nil {
		log.Printf("SQL ERROR: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	data := map[string]interface{}{
		"id":      swot.ID,
		"details": swot,
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    data,
		"message": "Swot analysis resource was created successfully!",
	})
}

func GetSwot(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	var swot models.Swot
	swot.TenantID = new(string)
	*swot.TenantID = tenantID

	swot.ProjectID = new(string)
	*swot.ProjectID, ok = utilities.ValidateQueryParam(c, "project_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	swot.ID = new(string)
	*swot.ID, ok = utilities.ValidateQueryParam(c, "swot_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SWOT ID is required"})
		return
	}

	query := `
		SELECT
			id, title, swot_description, swot_status, category, assumptions,
			final_decision, architectural_decision_id, implications, project_id, user_id, tenant_id
		FROM
			st_schema.swot_analysis
		WHERE
			id = $1
		AND
			tenant_id = $2
		AND
			project_id = $3
	`

	err := tenantManagement.DB.QueryRow(query, swot.ID, swot.TenantID, swot.ProjectID).Scan(
		&swot.ID,
		&swot.Title,
		&swot.SwotDescription,
		&swot.SwotStatus,
		&swot.Category,
		&swot.Assumptions,
		&swot.FinalDecision,
		&swot.ADecisionId,
		&swot.Implications,
		&swot.ProjectID,
		&swot.UserId,
		&swot.TenantID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No SWOT analysis found for ID %s", *swot.ID)
			c.JSON(http.StatusNotFound, gin.H{"error": "SWOT analysis not found"})
			return
		}
		log.Printf("SQL ERROR: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve the analysis details"})
		return
	}

	data := map[string]interface{}{
		"id":      *swot.ID,
		"details": swot,
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    data,
		"message": "SWOT analysis resource retrieved successfully!",
	})
}

func GetSwots(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	projectID, ok := utilities.ValidateQueryParam(c, "project_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal server error..."})
		return
	}

	query := `
		SELECT
			id, user_id, tenant_id, title, swot_description, swot_status, category, assumptions,
			final_decision, architectural_decision_id, implications, project_id
		FROM
			st_schema.swot_analysis
		WHERE
			tenant_id = $1
		AND
			project_id = $2
	`

	rows, err := tenantManagement.DB.Query(query, tenantID, projectID)
	if err != nil {
		log.Printf("Failed to retrieve SWOT resources: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve SWOT resources"})
		return
	}
	defer rows.Close()

	var swots []models.Swot
	for rows.Next() {
		var analysis models.Swot
		err := rows.Scan(
			&analysis.ID,
			&analysis.UserId,
			&analysis.TenantID,
			&analysis.Title,
			&analysis.SwotDescription,
			&analysis.SwotStatus,
			&analysis.Category,
			&analysis.Assumptions,
			&analysis.FinalDecision,
			&analysis.ADecisionId,
			&analysis.Implications,
			&analysis.ProjectID,
		)
		if err != nil {
			log.Printf("Failed to scan SWOT resources: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan SWOT resources"})
			return
		}
		swots = append(swots, analysis)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Failed to iterate over SWOT analyses: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to iterate over SWOT analyses"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    swots,
		"message": "All SWOT analysis objects retrieved successfully!",
	})
}

func UpdateSwot(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	var swot models.Swot
	swot.UserId = new(string)
	swot.TenantID = new(string)
	*swot.UserId = userID
	*swot.TenantID = tenantID

	swot.ProjectID = new(string)
	*swot.ProjectID, ok = utilities.ValidateQueryParam(c, "project_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	swot.ID = new(string)
	*swot.ID, ok = utilities.ValidateQueryParam(c, "swot_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SWOT ID is required"})
		return
	}

	if err := c.ShouldBindJSON(&swot); err != nil {
		log.Printf("ERROR: Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	v := reflect.ValueOf(swot)
	t := v.Type()

	setParts := []string{}
	args := []interface{}{}
	argCounter := 1

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		if fieldValue.Kind() == reflect.Ptr && !fieldValue.IsNil() {
			fieldName := t.Field(i).Tag.Get("json")
			if idx := strings.Index(fieldName, ",omitempty"); idx != -1 {
				fieldName = fieldName[:idx]
			}
			setParts = append(setParts, fmt.Sprintf("%s = $%d", fieldName, argCounter))
			args = append(args, fieldValue.Interface())
			argCounter++
		}
	}

	if len(setParts) == 0 {
		log.Printf("ERROR: No fields to update")
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	setClause := strings.Join(setParts, ", ")
	query := fmt.Sprintf(`
		UPDATE
			st_schema.swot_analysis
		SET
			%s, updated_at = NOW()
		WHERE
			id = $%d
		AND
			tenant_id = $%d
		AND
			project_id = $%d
		RETURNING
			id, title, swot_description, swot_status, category, assumptions,
			final_decision, architectural_decision_id, implications, project_id, user_id, tenant_id
	`, setClause, argCounter, argCounter+1, argCounter+2)

	args = append(args, swot.ID, swot.TenantID, swot.ProjectID)

	err := tenantManagement.DB.QueryRow(query, args...).Scan(
		&swot.ID,
		&swot.Title,
		&swot.SwotDescription,
		&swot.SwotStatus,
		&swot.Category,
		&swot.Assumptions,
		&swot.FinalDecision,
		&swot.ADecisionId,
		&swot.Implications,
		&swot.ProjectID,
		&swot.UserId,
		&swot.TenantID,
	)

	if err != nil {
		log.Printf("Failed to update SWOT analysis: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update the SWOT analysis"})
		return
	}

	data := map[string]interface{}{
		"details": swot,
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    data,
		"message": "SWOT analysis updated successfully!",
	})
}

func DeleteSwot(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	var swot models.Swot
	swot.TenantID = new(string)
	*swot.TenantID = tenantID

	swot.ProjectID = new(string)
	*swot.ProjectID, ok = utilities.ValidateQueryParam(c, "project_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	swot.ID = new(string)
	*swot.ID, ok = utilities.ValidateQueryParam(c, "swot_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SWOT ID is required"})
		return
	}

	tx, err := tenantManagement.DB.Begin()
	if err != nil {
		log.Printf("ERROR: Failed to start a transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start a transaction"})
		return
	}

	err = tx.QueryRow(
		`DELETE FROM st_schema.swot_analysis
		 WHERE id = $1
		 AND tenant_id = $2
		 AND project_id = $3
		 RETURNING id, title, swot_description, swot_status, category, assumptions, 
		 final_decision, architectural_decision_id, implications, project_id, user_id, tenant_id`,
		swot.ID, swot.TenantID, swot.ProjectID,
	).Scan(
		&swot.ID,
		&swot.Title,
		&swot.SwotDescription,
		&swot.SwotStatus,
		&swot.Category,
		&swot.Assumptions,
		&swot.FinalDecision,
		&swot.ADecisionId,
		&swot.Implications,
		&swot.ProjectID,
		&swot.UserId,
		&swot.TenantID,
	)

	if err != nil {
		log.Printf("ERROR: Failed to delete SWOT analysis: %v", err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete the SWOT analysis"})
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("ERROR: Failed to commit transaction: %v", err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit the transaction"})
		return
	}

	data := map[string]interface{}{
		"details": swot,
		"id":      *swot.ID,
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    data,
		"message": "SWOT analysis and associated arguments deleted successfully!",
	})
}

// Swot Arguments
func NewSwotArgument(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	var swotArgument models.SwotArgument
	swotArgument.UserID = new(string)
	swotArgument.TenantID = new(string)
	*swotArgument.UserID = userID
	*swotArgument.TenantID = tenantID

	swotID, ok := utilities.ValidateQueryParam(c, "swot_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SWOT ID is required"})
		return
	}
	swotArgument.SwotID = &swotID

	if err := c.ShouldBindJSON(&swotArgument); err != nil {
		log.Printf("ERROR: Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if *swotArgument.Argument == "" || *swotArgument.ArgumentWeight == 0 {
		log.Printf("ERROR: Argument and ArgumentWeight are required fields")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Argument and ArgumentWeight are required fields"})
		return
	}

	validSides := map[string]bool{"strength": true, "weakness": true, "opportunity": true, "threat": true}
	if !validSides[*swotArgument.Side] {
		log.Printf("ERROR: Invalid Side value: %s", *swotArgument.Side)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Side value must be either 'strength', 'weakness', 'opportunity', or 'threat'"})
		return
	}

	row := tenantManagement.DB.QueryRow(
		`INSERT INTO st_schema.swot_arguments (
			swot_id,
			user_id,
			tenant_id,
			argument,
			argument_weight,
			side,
			description
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, swot_id, user_id, tenant_id, argument, argument_weight, side, description`,
		swotArgument.SwotID,
		swotArgument.UserID,
		swotArgument.TenantID,
		swotArgument.Argument,
		swotArgument.ArgumentWeight,
		swotArgument.Side,
		swotArgument.Description,
	)

	err := row.Scan(
		&swotArgument.ID,
		&swotArgument.SwotID,
		&swotArgument.UserID,
		&swotArgument.TenantID,
		&swotArgument.Argument,
		&swotArgument.ArgumentWeight,
		&swotArgument.Side,
		&swotArgument.Description,
	)
	if err != nil {
		log.Printf("ERROR: Failed to retrieve the argument details: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create the Swot argument"})
		return
	}

	data := map[string]interface{}{
		"id":      swotArgument.ID,
		"details": swotArgument,
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    data,
		"message": "Swot argument created successfully!",
	})
}

func GetAllSwotArguments(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	swotID, ok := utilities.ValidateQueryParam(c, "swot_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SWOT ID is required"})
		return
	}

	rows, err := tenantManagement.DB.Query(
		`SELECT
			id, swot_id, user_id, tenant_id, argument, argument_weight, side, description
		 FROM
		 	st_schema.swot_arguments
		 WHERE
		 	swot_id = $1
		 AND
		 	tenant_id = $2`,
		swotID, tenantID,
	)
	if err != nil {
		log.Printf("ERROR: Failed to retrieve arguments: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve arguments"})
		return
	}
	defer rows.Close()

	var swotArguments []models.SwotArgument
	for rows.Next() {
		var argument models.SwotArgument
		if err := rows.Scan(
			&argument.ID,
			&argument.SwotID,
			&argument.UserID,
			&argument.TenantID,
			&argument.Argument,
			&argument.ArgumentWeight,
			&argument.Side,
			&argument.Description,
		); err != nil {
			log.Printf("ERROR: Failed to scan argument: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan argument"})
			return
		}
		swotArguments = append(swotArguments, argument)
	}

	if err := rows.Err(); err != nil {
		log.Printf("ERROR: Failed to iterate over arguments: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to iterate over arguments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    swotArguments,
		"message": "Arguments retrieved successfully!",
	})
}

func UpdateSwotArgument(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	var swotArgument models.SwotArgument
	swotArgument.UserID = new(string)
	swotArgument.TenantID = new(string)
	*swotArgument.UserID = userID
	*swotArgument.TenantID = tenantID

	swotID, ok := utilities.ValidateQueryParam(c, "swot_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SWOT ID is required"})
		return
	}
	swotArgument.SwotID = new(string)
	*swotArgument.SwotID = swotID

	argumentID, ok := utilities.ValidateQueryParam(c, "argument_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Argument ID is required"})
		return
	}
	swotArgument.ID = new(string)
	*swotArgument.ID = argumentID

	if err := c.ShouldBindJSON(&swotArgument); err != nil {
		log.Printf("ERROR: Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if swotArgument.Argument == nil || *swotArgument.Argument == "" || swotArgument.ArgumentWeight == nil || *swotArgument.ArgumentWeight == 0 {
		log.Printf("ERROR: Argument and ArgumentWeight are required fields")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Argument and ArgumentWeight are required fields"})
		return
	}

	validSides := map[string]bool{"strength": true, "weakness": true, "opportunity": true, "threat": true}
	if swotArgument.Side == nil || !validSides[*swotArgument.Side] {
		log.Printf("ERROR: Invalid Side value: %s", *swotArgument.Side)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Side value must be either 'strength', 'weakness', 'opportunity', or 'threat'"})
		return
	}

	var updatedArgument models.SwotArgument

	err := tenantManagement.DB.QueryRow(
		`UPDATE st_schema.swot_arguments SET
			argument = $1,
			argument_weight = $2,
			side = $3,
			description = $4
		WHERE
			id = $5
		AND
			tenant_id = $6
		AND
			swot_id = $7
		RETURNING id, swot_id, user_id, tenant_id, argument, argument_weight, side, description`,
		swotArgument.Argument,
		swotArgument.ArgumentWeight,
		swotArgument.Side,
		swotArgument.Description,
		swotArgument.ID,
		swotArgument.TenantID,
		swotArgument.SwotID,
	).Scan(
		&updatedArgument.ID,
		&updatedArgument.SwotID,
		&updatedArgument.UserID,
		&updatedArgument.TenantID,
		&updatedArgument.Argument,
		&updatedArgument.ArgumentWeight,
		&updatedArgument.Side,
		&updatedArgument.Description,
	)
	if err != nil {
		log.Printf("ERROR: Failed to update the argument: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update the argument"})
		return
	}

	data := map[string]interface{}{
		"id":      updatedArgument.ID,
		"details": updatedArgument,
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    data,
		"message": "Argument updated successfully!",
	})
}

func DeleteSwotArgument(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	var swotArgument models.SwotArgument
	swotArgument.TenantID = new(string)
	*swotArgument.TenantID = tenantID

	swotID, ok := utilities.ValidateQueryParam(c, "swot_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SWOT ID is required"})
		return
	}
	swotArgument.SwotID = new(string)
	*swotArgument.SwotID = swotID

	argumentID, ok := utilities.ValidateQueryParam(c, "argument_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Argument ID is required"})
		return
	}
	swotArgument.ID = new(string)
	*swotArgument.ID = argumentID

	var deletedArgument models.SwotArgument

	err := tenantManagement.DB.QueryRow(
		`DELETE FROM st_schema.swot_arguments
		WHERE
			id = $1
		AND
			tenant_id = $2
		AND
			swot_id = $3
		RETURNING id, swot_id, user_id, tenant_id, argument, argument_weight, side, description`,
		swotArgument.ID,
		swotArgument.TenantID,
		swotArgument.SwotID,
	).Scan(
		&deletedArgument.ID,
		&deletedArgument.SwotID,
		&deletedArgument.UserID,
		&deletedArgument.TenantID,
		&deletedArgument.Argument,
		&deletedArgument.ArgumentWeight,
		&deletedArgument.Side,
		&deletedArgument.Description,
	)
	if err != nil {
		log.Printf("ERROR: Failed to delete the argument: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete the argument"})
		return
	}

	data := map[string]interface{}{
		"id":      deletedArgument.ID,
		"details": deletedArgument,
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    data,
		"message": "Argument deleted successfully!",
	})
}
