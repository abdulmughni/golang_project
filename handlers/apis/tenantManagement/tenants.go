package tenantManagement

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sententiawebapi/handlers/models"
	"sententiawebapi/utilities"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type TenantMemberResponse struct {
	UserID      string          `json:"user_id"`
	TenantID    string          `json:"tenant_id"`
	Email       string          `json:"email"`
	FirstName   string          `json:"first_name"`
	LastName    string          `json:"last_name"`
	Occupation  string          `json:"occupation"`
	CompanyName string          `json:"company_name,omitempty"`
	UserPicture string          `json:"user_picture,omitempty"`
	Role        models.UserRole `json:"role"`
	Status      string          `json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
}

type TenantResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TenantWithMembersResponse struct {
	Tenant  TenantResponse         `json:"tenant"`
	Members []TenantMemberResponse `json:"members"`
}

func GetTenantWithMembers(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	// First get tenant information
	var tenant TenantResponse
	err := DB.QueryRow(`
		SELECT id, name, description, status, created_at, updated_at
		FROM st_schema.tenants
		WHERE id = $1
	`, tenantID).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Description,
		&tenant.Status,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)

	if err != nil {
		fmt.Println("Error retrieving tenant:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Then get all members of the tenant with their details
	rows, err := DB.Query(`
		SELECT
			u.id AS user_id,
			tm.tenant_id,
			u.email,
			u.first_name,
			u.last_name,
			u.occupation,
			u.company_name,
			u.user_picture,
			tm.role,
			tm.status,
			tm.created_at
		FROM
			st_schema.tenant_members tm
		JOIN
			st_schema.users u ON tm.user_id = u.id
		WHERE
			tm.tenant_id = $1
		ORDER BY
			tm.created_at DESC
	`, tenantID)

	if err != nil {
		fmt.Println("Error querying tenant members:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	defer rows.Close()

	var members []TenantMemberResponse
	for rows.Next() {
		var member TenantMemberResponse
		var companyName, userPicture sql.NullString

		err := rows.Scan(
			&member.UserID,
			&member.TenantID,
			&member.Email,
			&member.FirstName,
			&member.LastName,
			&member.Occupation,
			&companyName,
			&userPicture,
			&member.Role,
			&member.Status,
			&member.CreatedAt,
		)

		if err != nil {
			fmt.Println("Error scanning member row:", err)
			continue
		}

		if companyName.Valid {
			member.CompanyName = companyName.String
		}

		if userPicture.Valid {
			member.UserPicture = userPicture.String
		}

		members = append(members, member)
	}

	if err = rows.Err(); err != nil {
		fmt.Println("Error iterating member rows:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	response := TenantWithMembersResponse{
		Tenant:  tenant,
		Members: members,
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    response,
		"message": "Tenant and members fetched successfully",
	})
}

func UpdateTenant(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	// Parse the request body
	var updateRequest struct {
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
	}
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	// Validate required fields
	if updateRequest.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant name is required"})
		return
	}

	// Prepare the update statement
	query := `
		UPDATE st_schema.tenants
		SET
			name = $1,
			description = $2,
			updated_at = NOW()
		WHERE
			id = $3
		RETURNING id, name, description, status, created_at, updated_at
	`

	// Execute the update
	var tenant TenantResponse
	err := DB.QueryRow(
		query,
		updateRequest.Name,
		updateRequest.Description,
		tenantID,
	).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Status,
		&tenant.Description,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)

	if err != nil {
		fmt.Println("Error updating tenant:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tenant":  tenant,
		"message": "Tenant updated successfully",
	})
}

func RemoveTenantMember(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	// Get the member ID to remove from the URL parameter
	memberID := c.Param("member_id")
	if memberID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Member ID is required"})
		return
	}

	// Check if user is trying to remove themselves
	// Prevents also removing the last admin from the tenant
	if memberID == userID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "You cannot remove yourself from the tenant",
		})
		return
	}

	path := fmt.Sprintf("tenantMembers/%s?tenantID=%s", memberID, url.QueryEscape(tenantID))
	_, err := callServiceApi("DELETE", path, nil)
	if err != nil {
		log.Printf("Failed to remove tenant member: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Return success
	c.JSON(http.StatusOK, gin.H{
		"message": "Member removed successfully",
		"data": gin.H{
			"user_id": memberID,
		},
	})
}

// updates a tenant member's role
func UpdateTenantMember(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	memberID := c.Param("member_id")
	if memberID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Member ID is required"})
		return
	}

	// Prevent user to change their own role
	// Prevents also changing the last admin to a member role
	if memberID == userID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "You cannot change your own role",
		})
		return
	}

	// Parse the new role from the request body
	var input struct {
		Role models.UserRole `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate the role
	if input.Role != models.UserRoleAdmin && input.Role != models.UserRoleMember {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role. Must be 'admin' or 'member'"})
		return
	}

	// Update the member's role
	result, err := DB.Exec(`
		UPDATE st_schema.tenant_members
		SET role = $1
		WHERE user_id = $2 AND tenant_id = $3
	`, input.Role, memberID, tenantID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Check if any rows were affected (i.e., if the member exists)
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Member not found in this tenant"})
		return
	}

	// Return success
	c.JSON(http.StatusOK, gin.H{
		"message": "Member role updated successfully",
		"data": gin.H{
			"user_id": memberID,
			"role":    input.Role,
		},
	})
}

func callServiceApi(method, path string, payload interface{}) ([]byte, error) {
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		return nil, fmt.Errorf("ENVIRONMENT needs to be set")
	}

	// Map of known environments
	serviceURLs := map[string]string{
		"local": "http://localhost:8081",
		"dev":   "https://devserviceapi.solutionpilot.ai",
		"prod":  "https://serviceapi.solutionpilot.ai",
	}

	serviceApiUrl, ok := serviceURLs[environment]
	if !ok {
		return nil, fmt.Errorf("unknown environment: %s", environment)
	}

	fullURL := serviceApiUrl + "/api/" + strings.TrimLeft(path, "/")

	var bodyReader io.Reader
	if payload != nil {
		body, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		bodyReader = bytes.NewBuffer(body)
	}

	req, err := http.NewRequest(method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	serviceApiKey := os.Getenv("SERVICE_API_KEY")
	if serviceApiKey == "" {
		return nil, fmt.Errorf("SERVICE_API_KEY needs to be set")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", serviceApiKey)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for standard error field
	var errCheck struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(respBody, &errCheck); err == nil && errCheck.Error != "" {
		return nil, fmt.Errorf("remote service error: %s", errCheck.Error)
	}

	return respBody, nil
}
