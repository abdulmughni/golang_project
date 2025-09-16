package decisions

import (
	"database/sql"
	"log"
	"net/http"

	"sententiawebapi/handlers/apis/tenantManagement"
	models "sententiawebapi/handlers/models"
	"sententiawebapi/utilities"

	"github.com/gin-gonic/gin"
)

func NewMatrix(c *gin.Context) {
	var matrix models.Matrix

	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}
	matrix.UserID = new(string)
	*matrix.UserID = userID
	matrix.TenantID = new(string)
	*matrix.TenantID = tenantID

	projectID, ok := utilities.ValidateQueryParam(c, "project_id")
	if !ok {
		return
	}
	matrix.ProjectID = new(string)
	*matrix.ProjectID = projectID

	if err := c.ShouldBindJSON(&matrix); err != nil {
		log.Printf("ERROR: Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal server error..."})
		return
	}

	// Begin a new transaction
	tx, err := tenantManagement.DB.Begin()
	if err != nil {
		log.Printf("ERROR: Failed to start a transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	row := tx.QueryRow(
		`INSERT INTO st_schema.matrix_analysis (
			user_id,
			tenant_id,
			project_id,
			title,
			matrix_description,
			matrix_status,
			category,
			assumptions,
			final_decision,
			architectural_decision_id,
			implications
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id`,
		matrix.UserID,
		matrix.TenantID,
		matrix.ProjectID,
		matrix.Title,
		matrix.MatrixDescription,
		matrix.MatrixStatus,
		matrix.Category,
		matrix.Assumptions,
		matrix.FinalDecision,
		matrix.ADecisionId,
		matrix.Implications,
	)

	err = row.Scan(&matrix.Id)
	if err != nil {
		log.Printf("ERROR: Failed to retrieve the analysis ID: %v", err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	if err = tx.Commit(); err != nil {
		log.Printf("ERROR: Failed to commit transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	data := map[string]interface{}{
		"id":      matrix.Id,
		"details": matrix,
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":    data,
		"message": "Matrix analysis created successfully!",
	})
}

func GetMatrix(c *gin.Context) {
	var matrix models.Matrix

	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	matrixID, ok := utilities.ValidateQueryParam(c, "matrix_id")
	if !ok {
		return
	}

	projectID, ok := utilities.ValidateQueryParam(c, "project_id")
	if !ok {
		return
	}

	row := tenantManagement.DB.QueryRow(`
		SELECT
			id, user_id, tenant_id, title, matrix_description, matrix_status, category, 
			assumptions, final_decision, architectural_decision_id, implications, project_id 
		FROM
			st_schema.matrix_analysis
		WHERE
			id = $1
		AND
			tenant_id = $2
		AND
			project_id = $3`, matrixID, tenantID, projectID)

	err := row.Scan(
		&matrix.Id,
		&matrix.UserID,
		&matrix.TenantID,
		&matrix.Title,
		&matrix.MatrixDescription,
		&matrix.MatrixStatus,
		&matrix.Category,
		&matrix.Assumptions,
		&matrix.FinalDecision,
		&matrix.ADecisionId,
		&matrix.Implications,
		&matrix.ProjectID,
	)
	if err == sql.ErrNoRows {
		log.Printf("ERROR: Matrix analysis not found: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Internal server error..."})
		return
	} else if err != nil {
		log.Printf("ERROR: Failed to retrieve matrix analysis: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	data := map[string]interface{}{
		"id":      matrix.Id,
		"details": matrix,
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    data,
		"message": "Matrix analysis retrieved successfully!",
	})
}

func GetAllMatrixs(c *gin.Context) {
	var matrixes []models.Matrix

	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	projectID, ok := utilities.ValidateQueryParam(c, "project_id")
	if !ok {
		return
	}

	rows, err := tenantManagement.DB.Query(`
		SELECT
			id, user_id, tenant_id, title, matrix_description, matrix_status, category, assumptions,
			final_decision, architectural_decision_id, implications, project_id
		FROM
			st_schema.matrix_analysis
		WHERE
			tenant_id = $1
		AND
			project_id = $2`, tenantID, projectID)
	if err != nil {
		log.Printf("ERROR: Failed to retrieve Matrix analyses: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var matrix models.Matrix
		if err := rows.Scan(
			&matrix.Id,
			&matrix.UserID,
			&matrix.TenantID,
			&matrix.Title,
			&matrix.MatrixDescription,
			&matrix.MatrixStatus,
			&matrix.Category,
			&matrix.Assumptions,
			&matrix.FinalDecision,
			&matrix.ADecisionId,
			&matrix.Implications,
			&matrix.ProjectID,
		); err != nil {
			log.Printf("ERROR: Failed to scan Matrix analysis: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
			return
		}
		matrixes = append(matrixes, matrix)
	}

	if err := rows.Err(); err != nil {
		log.Printf("ERROR: Failed to iterate over Matrix analyses: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    matrixes,
		"message": "All Matrix analyses retrieved successfully!",
	})
}

func UpdateMatrix(c *gin.Context) {
	var matrix models.Matrix

	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	matrixID, ok := utilities.ValidateQueryParam(c, "matrix_id")
	if !ok {
		return
	}

	projectID, ok := utilities.ValidateQueryParam(c, "project_id")
	if !ok {
		return
	}

	if err := c.ShouldBindJSON(&matrix); err != nil {
		log.Printf("ERROR: Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal server error..."})
		return
	}

	_, err := tenantManagement.DB.Exec(
		`UPDATE st_schema.matrix_analysis SET
			title = $1,
			matrix_description = $2,
			matrix_status = $3,
			category = $4,
			assumptions = $5,
			final_decision = $6,
			architectural_decision_id = $7,
			implications = $8
		WHERE
			id = $9
		AND
			tenant_id = $10
		AND
			project_id = $11`,
		matrix.Title,
		matrix.MatrixDescription,
		matrix.MatrixStatus,
		matrix.Category,
		matrix.Assumptions,
		matrix.FinalDecision,
		matrix.ADecisionId,
		matrix.Implications,
		matrixID,
		tenantID,
		projectID,
	)

	if err != nil {
		log.Printf("ERROR: Failed to update the matrix analysis: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	row := tenantManagement.DB.QueryRow(`
		SELECT
			id, user_id, tenant_id, title, matrix_description, matrix_status, category,
			assumptions, final_decision, architectural_decision_id, implications, project_id
		FROM
			st_schema.matrix_analysis
		WHERE
			id = $1
		AND
			tenant_id = $2
		AND
			project_id = $3`, matrixID, tenantID, projectID)

	err = row.Scan(
		&matrix.Id,
		&matrix.UserID,
		&matrix.TenantID,
		&matrix.Title,
		&matrix.MatrixDescription,
		&matrix.MatrixStatus,
		&matrix.Category,
		&matrix.Assumptions,
		&matrix.FinalDecision,
		&matrix.ADecisionId,
		&matrix.Implications,
		&matrix.ProjectID,
	)
	if err != nil {
		log.Printf("ERROR: Failed to retrieve updated matrix analysis: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	data := map[string]interface{}{
		"id":      matrix.Id,
		"details": matrix,
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    data,
		"message": "Matrix analysis resource was successfully updated!",
	})
}

func DeleteMatrix(c *gin.Context) {
	var matrix models.Matrix

	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	matrixID, ok := utilities.ValidateQueryParam(c, "matrix_id")
	if !ok {
		return
	}

	projectID, ok := utilities.ValidateQueryParam(c, "project_id")
	if !ok {
		return
	}

	row := tenantManagement.DB.QueryRow(`
		SELECT
			id, user_id, tenant_id, title, matrix_description, matrix_status, category,
			assumptions, final_decision, architectural_decision_id, implications, project_id
		FROM
			st_schema.matrix_analysis
		WHERE
			id = $1
		AND
			tenant_id = $2
		AND
			project_id = $3`, matrixID, tenantID, projectID)

	err := row.Scan(
		&matrix.Id,
		&matrix.UserID,
		&matrix.TenantID,
		&matrix.Title,
		&matrix.MatrixDescription,
		&matrix.MatrixStatus,
		&matrix.Category,
		&matrix.Assumptions,
		&matrix.FinalDecision,
		&matrix.ADecisionId,
		&matrix.Implications,
		&matrix.ProjectID,
	)
	if err != nil {
		log.Printf("ERROR: Matrix analysis not found: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Internal server error..."})
		return
	}

	tx, err := tenantManagement.DB.Begin()
	if err != nil {
		log.Printf("ERROR: Failed to start a transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	_, err = tx.Exec("DELETE FROM st_schema.matrix_analysis WHERE id = $1 AND tenant_id = $2", matrixID, tenantID)
	if err != nil {
		tx.Rollback()
		log.Printf("ERROR: Failed to delete the Matrix analysis: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("ERROR: Failed to commit transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	data := map[string]interface{}{
		"id":      matrix.Id,
		"details": matrix,
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    data,
		"message": "Matrix analysis and all associated criteria, concepts, user ratings, and arguments deleted successfully!",
	})
}

func NewMatrixCriteria(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	var criteria models.MatrixCriteria

	criteria.UserID = userID
	criteria.TenantID = tenantID

	matrixID, ok := utilities.ValidateQueryParam(c, "matrix_id")
	if !ok {
		return
	}
	criteria.MatrixID = matrixID

	if err := c.ShouldBindJSON(&criteria); err != nil {
		log.Printf("ERROR: Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal server error..."})
		return
	}

	tx, err := tenantManagement.DB.Begin()
	if err != nil {
		log.Printf("ERROR: Failed to start a transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	row := tx.QueryRow(
		`INSERT INTO st_schema.matrix_criteria (
            matrix_id,
            user_id,
            tenant_id,
            title,
            criteria_multiplier,
            criteria_multiplier_title
        ) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		criteria.MatrixID,
		criteria.UserID,
		criteria.TenantID,
		criteria.Title,
		criteria.CriteriaMultiplier,
		criteria.CriteriaMultiplierTitle,
	)

	err = row.Scan(&criteria.Id)
	if err != nil {
		tx.Rollback()
		log.Printf("ERROR: Failed to retrieve the criteria ID: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	if err = tx.Commit(); err != nil {
		log.Printf("ERROR: Failed to commit transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	data := map[string]interface{}{
		"id":      criteria.Id,
		"details": criteria,
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    data,
		"message": "Matrix criteria created successfully!",
	})
}

func GetMatrixCriteria(c *gin.Context) {
	var criteria models.MatrixCriteria

	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}
	criteria.UserID = userID
	criteria.TenantID = tenantID

	criteriaID, ok := utilities.ValidateQueryParam(c, "criteria_id")
	if !ok {
		return
	}
	criteria.Id = criteriaID

	matrixID, ok := utilities.ValidateQueryParam(c, "matrix_id")
	if !ok {
		return
	}
	criteria.MatrixID = matrixID

	row := tenantManagement.DB.QueryRow(
		`SELECT id, matrix_id, title, criteria_multiplier, criteria_multiplier_title
         FROM st_schema.matrix_criteria
         WHERE id = $1 AND matrix_id = $2 AND tenant_id = $3`,
		criteriaID,
		matrixID,
		tenantID,
	)

	if err := row.Scan(
		&criteria.Id,
		&criteria.MatrixID,
		&criteria.Title,
		&criteria.CriteriaMultiplier,
		&criteria.CriteriaMultiplierTitle,
	); err != nil {
		if err == sql.ErrNoRows {
			log.Printf("ERROR: Criteria not found: %v", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Internal server error..."})
			return
		}
		log.Printf("ERROR: Failed to retrieve criteria from the database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	rows, err := tenantManagement.DB.Query(
		`SELECT concept_id, user_rating
         FROM st_schema.matrix_user_ratings
         WHERE criteria_id = $1 AND tenant_id = $2`,
		criteriaID,
		tenantID,
	)
	if err != nil {
		log.Printf("ERROR: Failed to retrieve user ratings from the database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}
	defer rows.Close()

	var concepts []models.MatrixConcept
	for rows.Next() {
		var concept models.MatrixConcept
		if err := rows.Scan(&concept.Id, &concept.UserRating); err != nil {
			log.Printf("ERROR: Failed to retrieve user rating details from the database: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
			return
		}
		concepts = append(concepts, concept)
	}

	if err := rows.Err(); err != nil {
		log.Printf("ERROR: Failed to process user ratings from the database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	responseData := map[string]interface{}{
		"criteria": criteria,
		"concepts": concepts,
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    responseData,
		"message": "Matrix criteria and associated user ratings retrieved successfully!",
	})
}

func GetAllMatrixCriteria(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": models.ParameterRequired})
		return
	}

	// Get the matrix ID from the URL
	matrixID := c.Query("matrix_id")
	if matrixID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": models.ParameterRequired})
		return
	}

	// Query the database for all criteria associated with the matrix ID and tenant ID
	rows, err := tenantManagement.DB.Query(
		`SELECT id, title, criteria_multiplier, criteria_multiplier_title
         FROM st_schema.matrix_criteria
         WHERE matrix_id = $1 AND tenant_id = $2`,
		matrixID,
		tenantID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve criteria from the database"})
		return
	}
	defer rows.Close()

	var criteriasWithConcepts []map[string]interface{}

	// Iterate over the rows and populate the slice
	for rows.Next() {
		var criteria models.MatrixCriteria
		if err := rows.Scan(&criteria.Id, &criteria.Title, &criteria.CriteriaMultiplier, &criteria.CriteriaMultiplierTitle); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process criteria data"})
			return
		}
		criteria.MatrixID = matrixID
		criteria.UserID = userID
		criteria.TenantID = tenantID

		// Fetch associated user ratings for the current criteria
		ratingRows, err := tenantManagement.DB.Query(
			`SELECT criteria_id, concept_id, user_id, user_rating
			FROM st_schema.matrix_user_ratings
			WHERE criteria_id = $1 AND tenant_id = $2`,
			criteria.Id,
			tenantID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user ratings from the database"})
			return
		}

		var concepts []map[string]interface{}
		for ratingRows.Next() {
			var (
				fetchedCriteriaID string
				fetchedConceptID  string
				fetchedUserID     string
				fetchedRating     int
			)

			if err := ratingRows.Scan(&fetchedCriteriaID, &fetchedConceptID, &fetchedUserID, &fetchedRating); err != nil {
				ratingRows.Close()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user rating details from the database"})
				return
			}

			conceptData := map[string]interface{}{
				"id":          fetchedConceptID,
				"matrix_id":   matrixID,
				"user_id":     fetchedUserID,
				"criteria_id": fetchedCriteriaID,
				"title":       "", // As title is not fetched, it remains empty.
				"user_rating": fetchedRating,
			}
			concepts = append(concepts, conceptData)
		}
		ratingRows.Close()

		criteriaData := map[string]interface{}{
			"criteria": criteria,
			"concepts": concepts,
		}
		criteriasWithConcepts = append(criteriasWithConcepts, criteriaData)
	}

	// Handle any potential errors after iterating
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process criteria data"})
		return
	}

	// Return the list of criteria and associated concepts as the response
	c.JSON(http.StatusOK, gin.H{
		"data":    criteriasWithConcepts,
		"message": "Matrix criteria and associated user ratings retrieved successfully!",
	})
}

func UpdateMatrixCriteria(c *gin.Context) {
	var criteria models.MatrixCriteria

	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}
	criteria.UserID = userID
	criteria.TenantID = tenantID

	criteriaID, ok := utilities.ValidateQueryParam(c, "criteria_id")
	if !ok {
		return
	}
	criteria.Id = criteriaID

	matrixID, ok := utilities.ValidateQueryParam(c, "matrix_id")
	if !ok {
		return
	}
	criteria.MatrixID = matrixID

	if err := c.ShouldBindJSON(&criteria); err != nil {
		log.Printf("ERROR: Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal server error..."})
		return
	}

	_, err := tenantManagement.DB.Exec(`
		UPDATE
			st_schema.matrix_criteria
		SET
			title = $1,
			criteria_multiplier = $2,
			criteria_multiplier_title = $3
		WHERE
			id = $4
		AND
			matrix_id = $5
		AND
			tenant_id = $6
		`,
		criteria.Title,
		criteria.CriteriaMultiplier,
		criteria.CriteriaMultiplierTitle,
		criteria.Id,
		criteria.MatrixID,
		criteria.TenantID,
	)

	if err != nil {
		log.Printf("ERROR: Failed to update the criteria in the database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	row := tenantManagement.DB.QueryRow(`
		SELECT
			id, matrix_id, title, criteria_multiplier, criteria_multiplier_title, tenant_id
		FROM
			st_schema.matrix_criteria
		WHERE
			id = $1 AND matrix_id = $2 AND tenant_id = $3`,
		criteriaID,
		matrixID,
		tenantID,
	)

	err = row.Scan(
		&criteria.Id,
		&criteria.MatrixID,
		&criteria.Title,
		&criteria.CriteriaMultiplier,
		&criteria.CriteriaMultiplierTitle,
		&criteria.TenantID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("ERROR: Criteria not found: %v", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Internal server error..."})
			return
		}
		log.Printf("ERROR: Failed to retrieve criteria details from the database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	data := map[string]interface{}{
		"id":      criteria.Id,
		"details": criteria,
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    data,
		"message": "Matrix criteria updated successfully!",
	})
}

func DeleteMatrixCriteria(c *gin.Context) {
	var criteria models.MatrixCriteria

	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}
	criteria.UserID = userID
	criteria.TenantID = tenantID

	criteriaID, ok := utilities.ValidateQueryParam(c, "criteria_id")
	if !ok {
		return
	}
	criteria.Id = criteriaID

	matrixID, ok := utilities.ValidateQueryParam(c, "matrix_id")
	if !ok {
		return
	}
	criteria.MatrixID = matrixID

	row := tenantManagement.DB.QueryRow(`
		SELECT
			id, matrix_id, title, criteria_multiplier, criteria_multiplier_title, tenant_id
		FROM
			st_schema.matrix_criteria
		WHERE
			id = $1 AND matrix_id = $2 AND tenant_id = $3`,
		criteriaID,
		matrixID,
		tenantID,
	)

	err := row.Scan(
		&criteria.Id,
		&criteria.MatrixID,
		&criteria.Title,
		&criteria.CriteriaMultiplier,
		&criteria.CriteriaMultiplierTitle,
		&criteria.TenantID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("ERROR: Criteria not found: %v", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Internal server error..."})
			return
		}
		log.Printf("ERROR: Failed to retrieve criteria details from the database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	tx, err := tenantManagement.DB.Begin()
	if err != nil {
		log.Printf("ERROR: Failed to start a transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	_, err = tx.Exec(
		`DELETE FROM st_schema.matrix_criteria
         WHERE id = $1 AND matrix_id = $2 AND tenant_id = $3`,
		criteriaID,
		matrixID,
		tenantID,
	)

	if err != nil {
		tx.Rollback()
		log.Printf("ERROR: Failed to delete the criteria from the database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("ERROR: Failed to commit transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    criteria,
		"message": "Matrix criteria and associated user ratings deleted successfully!",
	})
}

func NewMatrixConcept(c *gin.Context) {
	var concept models.MatrixConcept

	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}
	concept.UserID = userID
	concept.TenantID = tenantID

	concept.MatrixID, ok = utilities.ValidateQueryParam(c, "matrix_id")
	if !ok {
		return
	}

	if err := c.ShouldBindJSON(&concept); err != nil {
		log.Printf("ERROR: Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal server error..."})
		return
	}

	row := tenantManagement.DB.QueryRow(
		`INSERT INTO st_schema.matrix_concepts (
            matrix_id,
            user_id,
            tenant_id,
            title,
            user_rating
        ) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		concept.MatrixID,
		concept.UserID,
		concept.TenantID,
		concept.Title,
		concept.UserRating,
	)

	err := row.Scan(&concept.Id)
	if err != nil {
		log.Printf("ERROR: Failed to retrieve the concept ID: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	data := map[string]interface{}{
		"id":      concept.Id,
		"details": concept,
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    data,
		"message": "Matrix concept created successfully!",
	})
}

func GetMatrixConcept(c *gin.Context) {
	var concept models.MatrixConcept

	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	concept.UserID = userID
	concept.TenantID = tenantID

	conceptID, ok := utilities.ValidateQueryParam(c, "concept_id")
	if !ok {
		return
	}
	concept.Id = conceptID

	concept.MatrixID, ok = utilities.ValidateQueryParam(c, "matrix_id")
	if !ok {
		return
	}

	row := tenantManagement.DB.QueryRow(`
		SELECT
			matrix_id, title, user_rating
		FROM
			st_schema.matrix_concepts
		WHERE
			id = $1 AND matrix_id = $2 AND tenant_id = $3`,
		conceptID,
		concept.MatrixID,
		tenantID,
	)

	err := row.Scan(&concept.MatrixID, &concept.Title, &concept.UserRating)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("ERROR: Concept not found: %v", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Internal server error..."})
			return
		}
		log.Printf("ERROR: Failed to retrieve concept details: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	data := map[string]interface{}{
		"id":      concept.Id,
		"details": concept,
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    data,
		"message": "Matrix concept retrieved successfully!",
	})
}

func GetAllMatrixConcepts(c *gin.Context) {
	var concepts []models.MatrixConcept

	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	matrixID, ok := utilities.ValidateQueryParam(c, "matrix_id")
	if !ok {
		return
	}

	// Update the SQL query to include tenant_id in the WHERE clause
	rows, err := tenantManagement.DB.Query(`
		SELECT
			id, title, user_rating
		FROM
			st_schema.matrix_concepts
		WHERE
			matrix_id = $1 AND tenant_id = $2`,
		matrixID,
		tenantID,
	)
	if err != nil {
		log.Printf("ERROR: Failed to retrieve the matrix concepts: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var concept models.MatrixConcept

		err := rows.Scan(&concept.Id, &concept.Title, &concept.UserRating)
		if err != nil {
			log.Printf("ERROR: Failed to scan matrix concept data: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
			return
		}
		concept.MatrixID = matrixID
		concept.UserID = userID
		concept.TenantID = tenantID
		concepts = append(concepts, concept)
	}

	if err = rows.Err(); err != nil {
		log.Printf("ERROR: Failed to iterate over matrix concept rows: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    concepts,
		"message": "All matrix concepts retrieved successfully!",
	})
}

func UpdateMatrixConcept(c *gin.Context) {
	var concept models.MatrixConcept

	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}
	concept.UserID = userID
	concept.TenantID = tenantID

	conceptID, ok := utilities.ValidateQueryParam(c, "concept_id")
	if !ok {
		return
	}
	concept.Id = conceptID

	concept.MatrixID, ok = utilities.ValidateQueryParam(c, "matrix_id")
	if !ok {
		return
	}

	if err := c.ShouldBindJSON(&concept); err != nil {
		log.Printf("ERROR: Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal server error..."})
		return
	}

	_, err := tenantManagement.DB.Exec(`
		UPDATE
			st_schema.matrix_concepts
		SET
			title = $1,
			user_rating = $2
		WHERE
			id = $3
		AND
			matrix_id = $4
		AND
			tenant_id = $5
		`,
		concept.Title,
		concept.UserRating,
		concept.Id,
		concept.MatrixID,
		concept.TenantID,
	)

	if err != nil {
		log.Printf("ERROR: Failed to update the concept in the database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	row := tenantManagement.DB.QueryRow(`
		SELECT
			id, matrix_id, title, user_rating, tenant_id
		FROM
			st_schema.matrix_concepts
		WHERE
			id = $1 AND matrix_id = $2 AND tenant_id = $3`,
		conceptID,
		concept.MatrixID,
		tenantID,
	)

	err = row.Scan(
		&concept.Id,
		&concept.MatrixID,
		&concept.Title,
		&concept.UserRating,
		&concept.TenantID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("ERROR: Concept not found: %v", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Internal server error..."})
			return
		}
		log.Printf("ERROR: Failed to retrieve concept details from the database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	data := map[string]interface{}{
		"id":      concept.Id,
		"details": concept,
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    data,
		"message": "Matrix concept updated successfully!",
	})
}

func DeleteMatrixConcept(c *gin.Context) {
	var concept models.MatrixConcept

	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}
	concept.UserID = userID
	concept.TenantID = tenantID

	conceptID, ok := utilities.ValidateQueryParam(c, "concept_id")
	if !ok {
		return
	}
	concept.Id = conceptID

	concept.MatrixID, ok = utilities.ValidateQueryParam(c, "matrix_id")
	if !ok {
		return
	}

	row := tenantManagement.DB.QueryRow(`
		SELECT
			id, matrix_id, title, user_rating, tenant_id
		FROM
			st_schema.matrix_concepts
		WHERE
			id = $1 AND matrix_id = $2 AND tenant_id = $3`,
		conceptID,
		concept.MatrixID,
		tenantID,
	)

	err := row.Scan(
		&concept.Id,
		&concept.MatrixID,
		&concept.Title,
		&concept.UserRating,
		&concept.TenantID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("ERROR: Concept not found: %v", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Internal server error..."})
			return
		}
		log.Printf("ERROR: Failed to retrieve the matrix concept for deletion: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	tx, err := tenantManagement.DB.Begin()
	if err != nil {
		log.Printf("ERROR: Failed to start transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	_, err = tx.Exec(`
		DELETE FROM
			st_schema.matrix_concepts
		WHERE
			id = $1 AND tenant_id = $2`,
		conceptID,
		tenantID,
	)
	if err != nil {
		tx.Rollback()
		log.Printf("ERROR: Failed to delete the matrix concept: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("ERROR: Failed to commit the transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error..."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    concept,
		"message": "Matrix concept and associated user ratings deleted successfully!",
	})
}

func UpdateMatrixUserRating(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	criteriaID := c.Query("criteria_id")

	if criteriaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "criteria_id query parameter is required."})
		return
	}

	var userRatingUpdate models.MatrixUserRating
	if err := c.ShouldBindJSON(&userRatingUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	row := tenantManagement.DB.QueryRow(
		`UPDATE st_schema.matrix_user_ratings
        SET user_rating = $1
        WHERE criteria_id = $2 AND concept_id = $3 AND tenant_id = $4
        RETURNING id`,
		userRatingUpdate.UserRating,
		criteriaID,
		userRatingUpdate.ConceptID,
		tenantID,
	)

	err := row.Scan(&userRatingUpdate.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			// If no rows were affected, insert a new record
			insertRow := tenantManagement.DB.QueryRow(
				`INSERT INTO st_schema.matrix_user_ratings (user_rating, criteria_id, concept_id, user_id, tenant_id)
				VALUES ($1, $2, $3, $4, $5) RETURNING id`,
				userRatingUpdate.UserRating,
				criteriaID,
				userRatingUpdate.ConceptID,
				userID,
				tenantID,
			)
			err = insertRow.Scan(&userRatingUpdate.Id)
			if err != nil {
				log.Printf(models.DatabaseError, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert a new user rating."})
				return
			}
		} else {
			log.Printf(models.DatabaseError, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update the user rating."})
			return
		}
	}

	// Set additional fields for response
	userRatingUpdate.CriteriaID = criteriaID
	userRatingUpdate.UserID = userID
	userRatingUpdate.TenantId = &tenantID

	c.JSON(http.StatusOK, gin.H{
		"data":    userRatingUpdate,
		"message": "User rating processed successfully!",
	})
}
