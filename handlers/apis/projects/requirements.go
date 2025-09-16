package projects

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"sententiawebapi/handlers/apis/tenantManagement"
	"sententiawebapi/handlers/models"
	"sententiawebapi/utilities"
	"strings"

	"github.com/gin-gonic/gin"
)

type RequirementsApi struct {
	ProjectID string
	TenantID  string
	UserID    string
}

type DBExecutor interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
}

func NewRequirementsApi(tenantID string, userID string, projectID string) *RequirementsApi {
	return &RequirementsApi{
		ProjectID: projectID,
		TenantID:  tenantID,
		UserID:    userID,
	}
}

func (r *RequirementsApi) AddOne(db DBExecutor, req models.Requirement) (models.Requirement, error) {
	if err := validateCategory(req.Category); err != nil {
		return models.Requirement{}, err
	}
	if err := validateStatus(req.Status); err != nil {
		return models.Requirement{}, err
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		return models.Requirement{}, fmt.Errorf("requirement title is required")
	}

	query := `
		INSERT INTO st_schema.project_requirements (
			project_id, tenant_id, title, details, category, status, owner, start_date, target_date
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		) RETURNING id
	`

	var id string
	err := db.QueryRow(
		query,
		r.ProjectID,
		r.TenantID,
		title,
		coalesce(req.Details),
		req.Category,
		req.Status,
		req.Owner,
		req.StartDate,
		req.TargetDate,
	).Scan(&id)
	if err != nil {
		return models.Requirement{}, err
	}

	req.ID = id
	return req, nil
}

func (r *RequirementsApi) AddMany(db DBExecutor, requirements []models.Requirement) ([]models.Requirement, error) {
	inserted := make([]models.Requirement, 0, len(requirements))

	for _, req := range requirements {
		req.Owner = &r.UserID
		insertedReq, err := r.AddOne(db, req)
		if err != nil {
			return nil, err
		}
		inserted = append(inserted, insertedReq)
	}

	return inserted, nil
}

func (r *RequirementsApi) UpdateOne(db DBExecutor, id string, req models.Requirement) (*models.Requirement, error) {
	setClauses := []string{}
	args := []interface{}{}
	argPos := 1

	if req.Title != "" {
		title := strings.TrimSpace(req.Title)
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", argPos))
		args = append(args, title)
		argPos++
	}
	if req.Details != nil {
		setClauses = append(setClauses, fmt.Sprintf("details = $%d", argPos))
		args = append(args, *req.Details)
		argPos++
	}
	if req.Category != nil {
		if err := validateCategory(req.Category); err != nil {
			return nil, err
		}
		setClauses = append(setClauses, fmt.Sprintf("category = $%d", argPos))
		args = append(args, *req.Category)
		argPos++
	}
	if req.Status != nil {
		if err := validateStatus(req.Status); err != nil {
			return nil, err
		}
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argPos))
		args = append(args, *req.Status)
		argPos++
	}
	if req.StartDate != nil {
		setClauses = append(setClauses, fmt.Sprintf("start_date = $%d", argPos))
		args = append(args, *req.StartDate)
		argPos++
	}
	if req.TargetDate != nil {
		setClauses = append(setClauses, fmt.Sprintf("target_date = $%d", argPos))
		args = append(args, *req.TargetDate)
		argPos++
	}
	if req.Owner != nil {
		setClauses = append(setClauses, fmt.Sprintf("owner = $%d", argPos))
		args = append(args, *req.Owner)
		argPos++
	}

	if len(setClauses) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	query := fmt.Sprintf(`
		UPDATE st_schema.project_requirements
		SET %s
		WHERE id = $%d AND tenant_id = $%d
		RETURNING id, title, details, category, status, owner, start_date, target_date
	`, strings.Join(setClauses, ", "), argPos, argPos+1)

	args = append(args, id, r.TenantID)

	var updated models.Requirement
	err := db.QueryRow(query, args...).Scan(
		&updated.ID,
		&updated.Title,
		&updated.Details,
		&updated.Category,
		&updated.Status,
		&updated.Owner,
		&updated.StartDate,
		&updated.TargetDate,
	)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

func (r *RequirementsApi) DeleteOne(db DBExecutor, reqID string) error {
	query := `
		DELETE FROM st_schema.project_requirements
		WHERE id = $1 AND tenant_id = $2
	`

	res, err := db.Exec(query, reqID, r.TenantID)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not determine rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("no requirement found with id %s", reqID)
	}

	return nil
}

func (r *RequirementsApi) GetAll(db DBExecutor) ([]models.Requirement, error) {
	query := `
		SELECT
			r.id,
			r.title,
			r.details,
			r.category,
			r.status,
			r.owner,
			r.start_date,
			r.target_date,
			u.first_name,
			u.last_name,
			u.user_picture
		FROM st_schema.project_requirements r
		LEFT JOIN st_schema.users u ON r.owner = u.id
		WHERE r.project_id = $1 AND r.tenant_id = $2
	`

	rows, err := db.Query(query, r.ProjectID, r.TenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requirements := []models.Requirement{}
	for rows.Next() {
		var req models.Requirement
		var firstName, lastName, userPicture sql.NullString

		err := rows.Scan(
			&req.ID,
			&req.Title,
			&req.Details,
			&req.Category,
			&req.Status,
			&req.Owner,
			&req.StartDate,
			&req.TargetDate,
			&firstName,
			&lastName,
			&userPicture,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan requirement: %w", err)
		}

		if firstName.Valid {
			req.OwnerData = &models.Owner{
				FirstName: nullableStringPtr(firstName),
				LastName:  nullableStringPtr(lastName),
				Picture:   nullableStringPtr(userPicture),
			}
		}

		requirements = append(requirements, req)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return requirements, nil
}

func validateCategory(category *string) error {
	if category == nil {
		return nil
	}
	valid := map[string]struct{}{
		"Design": {}, "Development": {}, "Deployment": {}, "Business": {}, "Compliance": {},
		"Implementation": {}, "Document": {}, "Diagram": {}, "Decision": {}, "Strategy": {}, "Tactical": {},
	}
	if _, ok := valid[*category]; !ok {
		return fmt.Errorf("invalid category: %s", *category)
	}
	return nil
}

func validateStatus(status *string) error {
	if status == nil {
		return nil
	}
	valid := map[string]struct{}{
		"Not Started": {}, "In Progress": {}, "Completed": {}, "Delayed": {}, "Blocked": {},
	}
	if _, ok := valid[*status]; !ok {
		return fmt.Errorf("invalid status: %s", *status)
	}
	return nil
}

func coalesce(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

func nullableStringPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

// =============================
//         Route Handlers
// =============================

func CreateRequirementHandler(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	projectID := c.Query("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	var req models.Requirement
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	requirementsApi := NewRequirementsApi(tenantID, userID, projectID)
	created, err := requirementsApi.AddOne(tenantManagement.DB, req)
	if err != nil {
		log.Printf("Failed to create requirement: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": models.InternalServerError})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    created,
		"message": "Requirement created successfully!",
	})
}

func GetAllRequirementsHandler(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	projectID := c.Query("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	requirementsApi := NewRequirementsApi(tenantID, userID, projectID)
	requirements, err := requirementsApi.GetAll(tenantManagement.DB)
	if err != nil {
		log.Printf("Failed to fetch requirements: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": models.InternalServerError})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    requirements,
		"message": "Requirements retrieved successfully!",
	})
}

func UpdateRequirementHandler(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	requirementID := c.Param("requirement_id")
	if requirementID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Requirement ID is required"})
		return
	}

	projectID := c.Query("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	var req models.Requirement
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	requirementsApi := NewRequirementsApi(tenantID, userID, projectID)
	updated, err := requirementsApi.UpdateOne(tenantManagement.DB, requirementID, req)
	if err != nil {
		log.Printf("Failed to update requirement: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": models.InternalServerError})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    updated,
		"message": "Requirement updated successfully!",
	})
}

func DeleteRequirementHandler(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	requirementID := c.Param("requirement_id")
	if requirementID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Requirement ID is required"})
		return
	}

	projectID := c.Query("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	requirementsApi := NewRequirementsApi(tenantID, userID, projectID)
	err := requirementsApi.DeleteOne(tenantManagement.DB, requirementID)
	if err != nil {
		log.Printf("Failed to delete the requirement: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete the requirement"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Requirement deleted successfully",
		"id":      requirementID,
	})
}
