package decisions

// TODO: Calculate the better option and automatically set every time a new argument is added or updated or removed
// TODO: Simplify handling paramater lookup and validation
// TODO: Cleanup the code and remove unnecessary comments, remove and improve error and HTTP response, everything should be logged to to syslog
// TODO: Make sure to load all the parameter values into the data object file rather than declaring as vars

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

func NewTBar(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	var analysisWithOptions models.TBarAnalysisWithOptions
	analysisWithOptions.UserID = userID
	analysisWithOptions.TenantID = tenantID

	projectID := c.Query("project_id")
	if projectID == "" {
		log.Printf("ERROR: %v", "project_id is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal server error..."})
		return
	}

	// Set ProjectID
	analysisWithOptions.ProjectID = &projectID

	// Bind the JSON body to a struct
	if err := c.ShouldBindJSON(&analysisWithOptions); err != nil {
		log.Printf("ERROR: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal server error..."})
		return
	}

	tx, err := tenantManagement.DB.Begin()
	if err != nil {
		log.Printf("ERROR: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	// Create new TBar analysis resource in the database
	row := tx.QueryRow(`
		INSERT INTO
			st_schema.tbar_analysis
			(
				user_id,
				tenant_id,
				project_id,
				tbar_title,
				tbar_description,
				tbar_status,
				tbar_category,
				tbar_better_option,
				assumptions,
				final_decision,
				architectural_decision_id,
				implications
			)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`,
		analysisWithOptions.UserID,
		analysisWithOptions.TenantID,
		analysisWithOptions.ProjectID,
		analysisWithOptions.TBarTitle,
		analysisWithOptions.TBarDescription,
		analysisWithOptions.TBarStatus,
		analysisWithOptions.TBarCategory,
		analysisWithOptions.TBarBetterOption,
		analysisWithOptions.Assumptions,
		analysisWithOptions.FinalDecision,
		analysisWithOptions.ADecisionId,
		analysisWithOptions.Implications,
	)

	err = row.Scan(&analysisWithOptions.ID)
	if err != nil {
		tx.Rollback()
		log.Printf("SQL Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	// Insert into tbar_options for OptionA
	_, err = tx.Exec(`
		INSERT INTO
			st_schema.tbar_options (
				user_id,
				tenant_id,
				tbar_analysis_id,
				option_title
			)
		VALUES ($1, $2, $3, $4)
	`,
		analysisWithOptions.UserID,
		analysisWithOptions.TenantID,
		analysisWithOptions.ID,
		analysisWithOptions.OptionA,
	)
	if err != nil {
		tx.Rollback()
		log.Printf("SQL Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	// Insert into tbar_options for OptionB
	_, err = tx.Exec(`
		INSERT INTO
			st_schema.tbar_options (
				user_id,
				tenant_id,
				tbar_analysis_id,
				option_title
			)
		VALUES
			($1, $2, $3, $4)`,
		analysisWithOptions.UserID,
		analysisWithOptions.TenantID,
		analysisWithOptions.ID,
		analysisWithOptions.OptionB,
	)
	if err != nil {
		tx.Rollback()
		log.Printf("SQL Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	if err = tx.Commit(); err != nil {
		log.Printf("SQL Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	// Query options for the TBar analysis
	optionRows, err := tenantManagement.DB.Query(`
		SELECT
			id, option_title
		FROM
			st_schema.tbar_options
		WHERE
			tbar_analysis_id = $1
		AND
			tenant_id = $2
	`,
		analysisWithOptions.ID,
		analysisWithOptions.TenantID,
	)

	if err != nil {
		log.Printf("SQL Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}
	defer optionRows.Close()

	var options []map[string]string
	for optionRows.Next() {
		var optionID string
		var optionTitle string
		err = optionRows.Scan(&optionID, &optionTitle)
		if err != nil {
			log.Printf("SQL Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
			return
		}
		options = append(options, map[string]string{
			"option_id":    optionID,
			"option_title": optionTitle,
		})
	}

	data := map[string]interface{}{
		"tbar_analysis": map[string]interface{}{
			"id":         analysisWithOptions.ID,
			"project_id": analysisWithOptions.ProjectID,
			"details":    analysisWithOptions,
			"options":    options,
		},
	}
	c.JSON(http.StatusOK, gin.H{
		"data":    data,
		"message": "TBar analysis with options created successfully!",
	})
}

func GetTBars(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	projectID := c.Query("project_id")
	if projectID == "" {
		log.Printf("ERROR: %v", "project_id is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal server error..."})
		return
	}

	// Query to fetch all TBar analyses for the given user
	rows, err := tenantManagement.DB.Query(`
		SELECT
			a.id,
			a.tbar_title,
			a.tbar_description,
			a.tbar_status,
			a.tbar_category,
			a.tbar_better_option,
			a.project_id,
			o.id,
			o.option_title
		FROM
			st_schema.tbar_analysis AS a
		LEFT JOIN
			st_schema.tbar_options AS o ON a.id = o.tbar_analysis_id
		WHERE
			a.tenant_id = $1
		AND
			a.project_id = $2
		ORDER BY
			a.id,
			o.id
	`,
		tenantID,
		projectID)

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to retrieve TBar analyses"})
		return
	}

	defer rows.Close()

	// Data structure to hold the response
	data := make([]map[string]interface{}, 0)
	var prevAnalysisID string
	var analysis map[string]interface{}

	// Loop through the rows and build the response data with project data and options
	for rows.Next() {

		var tbar models.TBarAnalysisWithOptions
		var tbarOptions models.TBarOptions

		err = rows.Scan(
			&tbar.ID,
			&tbar.TBarTitle,
			&tbar.TBarDescription,
			&tbar.TBarStatus,
			&tbar.TBarCategory,
			&tbar.TBarBetterOption,
			&tbar.ProjectID,
			&tbarOptions.ID,
			&tbarOptions.OptionTitle,
		)

		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to retrieve TBar Data"})
			return
		}

		if tbar.ID != prevAnalysisID {
			// Initialize the analysis map with an empty options slice
			analysis = map[string]interface{}{
				"tbar_analysis": map[string]interface{}{
					"id":     tbar.ID,
					"UserID": tbar.UserID,
					"details": map[string]interface{}{
						"TBarTitle":        tbar.TBarTitle,
						"TBarDescription":  tbar.TBarDescription,
						"TBarStatus":       tbar.TBarStatus,
						"TBarCategory":     tbar.TBarCategory,
						"TBarBetterOption": tbar.TBarBetterOption,
					},
					"options": make([]map[string]string, 0), // Ensure options is initialized
				},
			}
			data = append(data, analysis)
			prevAnalysisID = tbar.ID
		}

		if tbarOptions.ID != "" {
			// Retrieve and update the options slice correctly
			options := analysis["tbar_analysis"].(map[string]interface{})["options"].([]map[string]string)
			options = append(options, map[string]string{
				"option_id":    tbarOptions.ID,
				"option_title": tbarOptions.OptionTitle,
			})
			analysis["tbar_analysis"].(map[string]interface{})["options"] = options
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    data,
		"message": "TBar analyses retrieved successfully!",
	})
}

func GetTBar(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	projectID := c.Query("project_id")
	if projectID == "" {
		log.Printf("ERROR: %v", "project_id is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal server error..."})
		return
	}

	tbarID := c.Query("tbar_id")
	if tbarID == "" {
		log.Printf("ERROR: %v", "tbar_id is required")
		c.JSON(400, gin.H{"error": "Internal server error..."})
		return
	}

	// Get TBar analysis details
	row := tenantManagement.DB.QueryRow(`
		SELECT
			id,
			user_id,
			tenant_id,
			tbar_title,
			tbar_description,
			tbar_status,
			tbar_category,
			tbar_better_option,
			assumptions,
			final_decision,
			architectural_decision_id,
			implications,
			project_id
		FROM
			st_schema.tbar_analysis
		WHERE
			id = $1
		AND
			tenant_id = $2
		AND
			project_id = $3
	`, tbarID, tenantID, projectID)

	var details models.TBarAnalysis
	err := row.Scan(
		&details.ID,
		&details.UserID,
		&details.TenantID,
		&details.TBarTitle,
		&details.TBarDescription,
		&details.TBarStatus,
		&details.TBarCategory,
		&details.TBarBetterOption,
		&details.Assumptions,
		&details.FinalDecision,
		&details.ADecisionId,
		&details.Implications,
		&details.ProjectID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(404, gin.H{"error": "TBar analysis not found"})
		} else {
			c.JSON(500, gin.H{"error": "Failed to retrieve TBar analysis details"})
		}
		return
	}

	// Get TBar options
	rows, err := tenantManagement.DB.Query(`
		SELECT
			id,
			option_title
		FROM
			st_schema.tbar_options
		WHERE
			tbar_analysis_id = $1
		AND
			tenant_id = $2
	`, tbarID, tenantID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to retrieve TBar options"})
		return
	}
	defer rows.Close()

	var options []map[string]string
	for rows.Next() {
		var optionID, optionTitle string
		err := rows.Scan(&optionID, &optionTitle)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to scan TBar options"})
			return
		}
		option := map[string]string{
			"option_id":    optionID,
			"option_title": optionTitle,
		}
		options = append(options, option)
	}

	// Create the response data
	responseData := gin.H{
		"data": gin.H{
			"tbar_analysis": gin.H{
				"details": gin.H{
					"ID":                        details.ID,
					"UserID":                    details.UserID,
					"TenantID":                  details.TenantID,
					"TBarTitle":                 details.TBarTitle,
					"TBarDescription":           details.TBarDescription,
					"TBarStatus":                details.TBarStatus,
					"TBarCategory":              details.TBarCategory,
					"TBarBetterOption":          details.TBarBetterOption,
					"assumptions":               details.Assumptions,
					"final_decision":            details.FinalDecision,
					"architectural_decision_id": details.ADecisionId,
					"implications":              details.Implications,
					"ProjectID":                 details.ProjectID, // include project_id in the response
				},
				"id":      details.ID,
				"options": options,
			},
		},
		"message": "TBar analysis retrieved successfully!",
	}

	c.JSON(200, responseData)
}

// Define option struct
type TBarOption struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// Modify updateData struct
var updateData struct {
	TBarTitle        *string     `json:"tbar_title" db:"tbar_title"`
	TBarDescription  *string     `json:"tbar_description" db:"tbar_description"`
	TBarStatus       *string     `json:"tbar_status" db:"tbar_status"`
	TBarCategory     *string     `json:"tbar_category" db:"tbar_category"`
	TBarBetterOption *string     `json:"tbar_better_option" db:"tbar_better_option"`
	Assumptions      *string     `json:"assumptions" db:"assumptions"`
	FinalDecision    *string     `json:"final_decision" db:"final_decision"`
	ADecisionId      *string     `json:"architectural_decision_id" db:"architectural_decision_id"`
	Implications     *string     `json:"implications" db:"implications"`
	OptionA          *TBarOption `json:"option_a" db:"-"`
	OptionB          *TBarOption `json:"option_b" db:"-"`
}

func UpdateTBar(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	projectID := c.Query("project_id")
	if projectID == "" {
		log.Printf("ERROR: %v", "project_id is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal server error..."})
		return
	}

	tbarID := c.Query("tbar_id")
	if tbarID == "" {
		log.Printf("ERROR: %v", "tbar_id is required")
		c.JSON(400, gin.H{"error": "Internal server error..."})
		return
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	// Start a transaction for the updates
	tx, err := tenantManagement.DB.Begin()
	if err != nil {
		log.Printf("ERROR: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	// Update TBar analysis fields using reflection
	v := reflect.ValueOf(updateData)
	t := v.Type()

	setParts := []string{}
	args := []interface{}{}
	argCounter := 1

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		dbFieldName := t.Field(i).Tag.Get("db")

		// Skip fields with "-" as the db tag
		if dbFieldName == "-" {
			continue
		}

		if !fieldValue.IsNil() {
			setParts = append(setParts, fmt.Sprintf("%s = $%d", dbFieldName, argCounter))
			args = append(args, fieldValue.Elem().Interface())
			argCounter++
		}
	}

	var updatedAnalysis models.TBarAnalysis

	if len(setParts) > 0 {
		setClause := strings.Join(setParts, ", ")
		query := fmt.Sprintf(`
			UPDATE
				st_schema.tbar_analysis
			SET
				%s, updated_at = NOW()
			WHERE
				tenant_id = $%d
			AND
				project_id = $%d
			AND
				id = $%d
			RETURNING
				id, user_id, tenant_id, tbar_title, tbar_description, tbar_status, tbar_category, tbar_better_option, assumptions, final_decision, architectural_decision_id, implications, updated_at, project_id
		`, setClause, argCounter, argCounter+1, argCounter+2)

		args = append(args, tenantID, projectID, tbarID)

		err := tx.QueryRow(query, args...).Scan(
			&updatedAnalysis.ID,
			&updatedAnalysis.UserID,
			&updatedAnalysis.TenantID,
			&updatedAnalysis.TBarTitle,
			&updatedAnalysis.TBarDescription,
			&updatedAnalysis.TBarStatus,
			&updatedAnalysis.TBarCategory,
			&updatedAnalysis.TBarBetterOption,
			&updatedAnalysis.Assumptions,
			&updatedAnalysis.FinalDecision,
			&updatedAnalysis.ADecisionId,
			&updatedAnalysis.Implications,
			&updatedAnalysis.UpdatedAt,
			&updatedAnalysis.ProjectID,
		)

		if err != nil {
			tx.Rollback()
			log.Printf("Failed to update TBar analysis: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update the TBar analysis"})
			return
		}
	}

	if updateData.OptionA != nil || updateData.OptionB != nil {
		if updateData.OptionA != nil {
			_, err = tx.Exec(`
				UPDATE st_schema.tbar_options
				SET option_title = $1
				WHERE id = $2 AND tenant_id = $3 AND tbar_analysis_id = $4
			`, updateData.OptionA.Title, updateData.OptionA.ID, tenantID, tbarID)

			if err != nil {
				tx.Rollback()
				log.Printf("Failed to update option A: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update option A"})
				return
			}
		}

		if updateData.OptionB != nil {
			_, err = tx.Exec(`
				UPDATE st_schema.tbar_options
				SET option_title = $1
				WHERE id = $2 AND tenant_id = $3 AND tbar_analysis_id = $4
			`, updateData.OptionB.Title, updateData.OptionB.ID, tenantID, tbarID)

			if err != nil {
				tx.Rollback()
				log.Printf("Failed to update option B: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update option B"})
				return
			}
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// Query for the options
	rows, err := tenantManagement.DB.Query(`
		SELECT
			id, user_id, tbar_analysis_id, option_title
		FROM
			st_schema.tbar_options
		WHERE
			tbar_analysis_id = $1 AND tenant_id = $2;
	`, tbarID, tenantID)
	if err != nil {
		log.Printf("Failed to retrieve TBar options: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve the TBar options"})
		return
	}
	defer rows.Close()

	var options []models.TBarOptions
	for rows.Next() {
		var option models.TBarOptions
		err = rows.Scan(&option.ID, &option.UserID, &option.TBarAnalysisID, &option.OptionTitle)
		if err != nil {
			log.Printf("Failed to scan option data: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan option data"})
			return
		}
		options = append(options, option)
	}

	// Check for errors from iterating over rows.
	if err = rows.Err(); err != nil {
		log.Printf("Failed to read options data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read options data"})
		return
	}

	// Return the updated TBar analysis data
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"tbar_analysis": gin.H{
				"details": updatedAnalysis,
				"options": options,
			},
		},
		"message": "TBar analysis updated successfully",
	})
}

func DeleteTBar(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	projectID := c.Query("project_id")
	if projectID == "" {
		log.Printf("ERROR: %v", "project_id is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal server error..."})
		return
	}

	tbarID := c.Query("tbar_id")
	if tbarID == "" {
		log.Printf("ERROR: %v", "tbar_id is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal server error..."})
		return
	}

	// Start a transaction
	tx, err := tenantManagement.DB.Begin()
	if err != nil {
		log.Printf("ERROR: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	// Retrieve the TBar analysis and options before deletion
	var analysisWithOptions models.TBarAnalysisWithOptions

	// Query to retrieve the TBar analysis details
	row := tx.QueryRow(`
		SELECT
			id, tbar_title, tbar_description, tbar_status, tbar_category,
			tbar_better_option, assumptions, final_decision, architectural_decision_id,
			implications, project_id
		FROM
			st_schema.tbar_analysis
		WHERE
			id = $1 AND tenant_id = $2 AND project_id = $3
	`, tbarID, tenantID, projectID)

	err = row.Scan(
		&analysisWithOptions.ID,
		&analysisWithOptions.TBarTitle,
		&analysisWithOptions.TBarDescription,
		&analysisWithOptions.TBarStatus,
		&analysisWithOptions.TBarCategory,
		&analysisWithOptions.TBarBetterOption,
		&analysisWithOptions.Assumptions,
		&analysisWithOptions.FinalDecision,
		&analysisWithOptions.ADecisionId,
		&analysisWithOptions.Implications,
		&analysisWithOptions.ProjectID,
	)
	if err != nil {
		tx.Rollback()
		log.Printf("ERROR: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve TBar analysis"})
		return
	}

	// Query to retrieve the TBar options
	optionRows, err := tx.Query(`
		SELECT
			id, option_title
		FROM
			st_schema.tbar_options
		WHERE
			tbar_analysis_id = $1 AND tenant_id = $2
	`, tbarID, tenantID)

	if err != nil {
		tx.Rollback()
		log.Printf("ERROR: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve TBar options"})
		return
	}
	defer optionRows.Close()

	var options []map[string]string
	for optionRows.Next() {
		var optionID string
		var optionTitle string
		err = optionRows.Scan(&optionID, &optionTitle)
		if err != nil {
			log.Printf("ERROR: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan TBar options"})
			return
		}
		options = append(options, map[string]string{
			"option_id":    optionID,
			"option_title": optionTitle,
		})
	}

	// Delete the TBar itself
	_, err = tx.Exec(`
        DELETE FROM st_schema.tbar_analysis
        WHERE
			id = $1 AND tenant_id = $2 AND project_id = $3
    `, tbarID, tenantID, projectID)
	if err != nil {
		tx.Rollback()
		log.Printf("ERROR: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete TBar analysis"})
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("ERROR: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// Return the response in the same structure as NewTBar
	data := map[string]interface{}{
		"tbar_analysis": map[string]interface{}{
			"id":         analysisWithOptions.ID,
			"project_id": analysisWithOptions.ProjectID,
			"details":    analysisWithOptions,
			"options":    options,
		},
	}
	c.JSON(http.StatusOK, gin.H{
		"data":    data,
		"message": "TBar analysis and associated records deleted successfully!",
	})
}

// Create new TBar argument for TBar Analysis Option
func NewTBarArgument(c *gin.Context) {
	var argument models.TBarArgument

	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	argument.OptionID, ok = utilities.ValidateQueryParam(c, "option_id")
	if !ok {
		return
	}

	if err := c.ShouldBindJSON(&argument); err != nil {
		log.Printf("ERROR: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal server error..."})
		return
	}

	argument.UserID = userID
	argument.TenantID = tenantID

	err := tenantManagement.DB.QueryRow(`
		INSERT INTO
			st_schema.tbar_arguments (user_id, tenant_id, option_id, argument_name, argument_weight, description)
		VALUES
			($1, $2, $3, $4, $5, $6)
		RETURNING id;
	`,
		argument.UserID,
		argument.TenantID,
		argument.OptionID,
		argument.ArgumentName,
		argument.ArgumentWeight,
		argument.Description,
	).Scan(&argument.ID)

	if err != nil {
		log.Printf("SQL Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "TBar argument was created successfully",
		"data": gin.H{
			"id":              argument.ID,
			"option_id":       argument.OptionID,
			"argument_name":   argument.ArgumentName,
			"argument_weight": argument.ArgumentWeight,
			"description":     argument.Description,
		},
	})
}

func GetTBarArguments(c *gin.Context) {
	var argument models.TBarArgument

	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}
	argument.UserID = userID
	argument.TenantID = tenantID

	argument.OptionID, ok = utilities.ValidateQueryParam(c, "option_id")
	if !ok {
		return
	}

	query := `
		SELECT
			id, argument_name, argument_weight, description
		FROM
			st_schema.tbar_arguments
		WHERE
			option_id = $1
		AND
			tenant_id = $2;
	`

	// Execute the query
	rows, err := tenantManagement.DB.Query(
		query,
		argument.OptionID,
		argument.TenantID,
	)
	if err != nil {
		log.Printf("ERROR: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}
	defer rows.Close()

	// Loop through the results and append them to a slice
	var arguments []map[string]interface{}
	for rows.Next() {
		var arg models.TBarArgument

		err = rows.Scan(&arg.ID, &arg.ArgumentName, &arg.ArgumentWeight, &arg.Description)
		if err != nil {
			log.Printf("ERROR: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
			return
		}

		arguments = append(arguments, map[string]interface{}{
			"id":             arg.ID,
			"argumentName":   arg.ArgumentName,
			"argumentWeight": arg.ArgumentWeight,
			"description":    arg.Description,
			"optionId":       argument.OptionID,
			"userId":         argument.UserID,
		})
	}

	// Check for errors from iterating over rows
	if err = rows.Err(); err != nil {
		log.Printf("ERROR: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to iterate over the SQL rows"})
		return
	}

	// Return the arguments as JSON
	c.JSON(http.StatusOK, gin.H{
		"message": "TBar arguments retrieved successfully",
		"data":    arguments,
	})
}

func UpdateTBarArgument(c *gin.Context) {
	var argument models.TBarArgument

	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}
	argument.TenantID = tenantID

	argument.OptionID, ok = utilities.ValidateQueryParam(c, "option_id")
	if !ok {
		return
	}

	argument.ID, ok = utilities.ValidateQueryParam(c, "argument_id")
	if !ok {
		return
	}

	// Bind the JSON body to a struct
	var updateData struct {
		ArgumentName   string `json:"argument_name" binding:"required"`
		ArgumentWeight int    `json:"argument_weight" binding:"required"`
		Description    string `json:"description"` // Optional
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		log.Printf("ERROR: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal server error..."})
		return
	}

	err := tenantManagement.DB.QueryRow(`
		UPDATE
			st_schema.tbar_arguments
		SET
			argument_name = $1,
			argument_weight = $2,
			description = $3
		WHERE
			id = $4
		AND
			tenant_id = $5
		AND
			option_id = $6
		RETURNING
			id, argument_name, argument_weight, description, option_id;
	`,
		updateData.ArgumentName,
		updateData.ArgumentWeight,
		updateData.Description,
		argument.ID,
		argument.TenantID,
		argument.OptionID,
	).Scan(
		&argument.ID,
		&argument.ArgumentName,
		&argument.ArgumentWeight,
		&argument.Description,
		&argument.OptionID,
	)

	if err != nil {
		log.Printf("ERROR: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update the TBar argument"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "TBar argument was updated successfully",
		"data": gin.H{
			"id":             argument.ID,
			"argumentName":   argument.ArgumentName,
			"argumentWeight": argument.ArgumentWeight,
			"description":    argument.Description,
			"optionId":       argument.OptionID,
		},
	})
}

func DeleteTBarArgument(c *gin.Context) {
	var argument models.TBarArgument

	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}
	argument.TenantID = tenantID

	argument.OptionID, ok = utilities.ValidateQueryParam(c, "option_id")
	if !ok {
		return
	}

	argument.ID, ok = utilities.ValidateQueryParam(c, "argument_id")
	if !ok {
		return
	}

	err := tenantManagement.DB.QueryRow(`
		DELETE FROM
			st_schema.tbar_arguments
		WHERE
			id = $1
		AND
			tenant_id = $2
		AND
			option_id = $3
		RETURNING
			argument_name, argument_weight, description;
	`, argument.ID, argument.TenantID, argument.OptionID).Scan(
		&argument.ArgumentName,
		&argument.ArgumentWeight,
		&argument.Description,
	)

	if err != nil {
		log.Printf("ERROR: Failed to delete TBar argument: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete TBar argument"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "TBar argument deleted successfully",
		"data": gin.H{
			"id":             argument.ID,
			"argumentName":   argument.ArgumentName,
			"argumentWeight": argument.ArgumentWeight,
			"description":    argument.Description,
			"optionId":       argument.OptionID,
		},
	})
}
