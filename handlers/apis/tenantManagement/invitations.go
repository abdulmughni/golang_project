package tenantManagement

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sententiawebapi/handlers/models"
	"sententiawebapi/utilities"
	"time"

	"bytes"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type Inviter struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
}

type InvitationResponse struct {
	ID           string    `json:"id"`
	InviteeEmail string    `json:"invitee_email"`
	Role         string    `json:"role"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Expired      bool      `json:"expired"`
	Inviter      Inviter   `json:"inviter"`
}

type InvitationPayload struct {
	Email string          `json:"email" binding:"required,email"`
	Role  models.UserRole `json:"role" binding:"required"`
}

type FailedInvitation struct {
	Email string `json:"email"`
	Error string `json:"error"`
}

func InviteUsersToTenant(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	// Define a slice to hold the invitation entries
	var invitations []InvitationPayload

	// Parse the request body directly into the slice
	if err := c.ShouldBindJSON(&invitations); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format - please check email formats and required fields"})
		return
	}

	if len(invitations) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You need to provide at least one email address to invite"})
		return
	}

	inviter, err := getInviter(userID)
	if err != nil {
		fmt.Println("Error getting inviter data", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Get SendGrid API key from environment
	sendgridAPIKey := os.Getenv("SENDGRID_KEY")
	if sendgridAPIKey == "" {
		fmt.Println("SendGrid API key is not configured")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	sendgridClient := sendgrid.NewSendClient(sendgridAPIKey)

	var failedInvitations []FailedInvitation
	var successfulInvitations []InvitationResponse

	// Process each invitation
	for _, invitation := range invitations {
		response, err := inviteSingleUser(invitation.Email, invitation.Role, tenantID, userID, inviter, sendgridClient)
		if err != nil {
			failedInvitations = append(failedInvitations, FailedInvitation{
				Email: invitation.Email,
				Error: err.Error(),
			})
			continue
		}
		successfulInvitations = append(successfulInvitations, response)
	}

	// Rest of the function remains the same
	if len(successfulInvitations) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send any invitations"})
		return
	}

	if len(failedInvitations) > 0 {
		c.JSON(http.StatusMultiStatus, gin.H{
			"message":                fmt.Sprintf("Successfully sent %d invitations", len(successfulInvitations)),
			"failed_invitations":     failedInvitations,
			"successful_invitations": successfulInvitations,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message":                "All invitations sent successfully",
			"failed_invitations":     failedInvitations,
			"successful_invitations": successfulInvitations,
		})
	}
}

// gets the inviter's data
func getInviter(userID string) (inviter Inviter, err error) {
	err = DB.QueryRow(`
		SELECT
			id,
			first_name || ' ' || last_name as full_name,
			email
		FROM
			st_schema.users
		WHERE id = $1
	`, userID).Scan(&inviter.ID, &inviter.FullName, &inviter.Email)

	return inviter, err
}

// generates a secure token for the invitation
func generateSecureToken(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate secure random token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Function to generate formatted email content
func formatInvitationEmail(inviterName, inviteURL string) (string, string) {
	// Plain text version for clients that don't support HTML
	plainTextContent := fmt.Sprintf("%s invited you to join their tenant. Click on the link to join: %s",
		inviterName, inviteURL)

	// HTML version with the same styling as your OTP email
	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <style>
    body {
      font-family: 'Arial', sans-serif;
      line-height: 1.6;
      color: #333;
      max-width: 600px;
      margin: 0 auto;
      padding: 20px;
    }
    .logo {
      text-align: center;
      margin-bottom: 30px;
    }
    .logo img {
      max-width: 250px;
      height: auto;
    }
    .container {
      background-color: #f7f9fc;
      border-radius: 8px;
      padding: 30px;
      box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
    }
    .button {
      display: inline-block;
      background-color: #2d61d2;
      color: white;
      text-decoration: none;
      padding: 12px 24px;
      border-radius: 4px;
      font-weight: bold;
      margin: 20px 0;
      text-align: center;
    }
    .footer {
      font-size: 12px;
      color: #6f717c;
      text-align: center;
      margin-top: 30px;
      padding-top: 15px;
      border-top: 1px solid #e1e5ea;
    }
  </style>
</head>
<body>
  <div class="logo">
    <img src="https://solutionpilotmedia.blob.core.windows.net/branding/logo-blue-icon-black-text.svg" alt="Solution Pilot">
  </div>
  
  <div class="container">
    <h2>Invitation to Join</h2>
    <p>Hello,</p>
    <p><strong>{{.InviterName}}</strong> has invited you to join their tenant in Solution Pilot.</p>
    
    <div style="text-align: center;">
      <a href="{{.InviteURL}}" class="button">Click here to join</a>
    </div>
    
    <p>If you're unable to click the button above, copy and paste the following URL into your browser:</p>
	<br>
    <p style="word-break: break-all; font-size: 12px;">{{.InviteURL}}</p>
    <br>
    <p>Best regards,<br>
    The Solution Pilot Team</p>
  </div>
  
  <div class="footer">
    <p>This is an automated message, please do not reply to this email.</p>
    <p>&copy; 2025 Solution Pilot Inc. All rights reserved.</p>
  </div>
</body>
</html>
`

	// Create a template and parse the HTML
	tmpl, err := template.New("invitation").Parse(htmlTemplate)
	if err != nil {
		// Fallback to simple HTML if template parsing fails
		return plainTextContent, fmt.Sprintf("<p><strong>%s</strong> invited you to join their tenant.</p><p><a href='%s'>Click here to join</a></p>",
			inviterName, inviteURL)
	}

	// Prepare data for template
	data := struct {
		InviterName string
		InviteURL   string
	}{
		InviterName: inviterName,
		InviteURL:   inviteURL,
	}

	// Execute the template with the data
	var htmlBuffer bytes.Buffer
	if err := tmpl.Execute(&htmlBuffer, data); err != nil {
		// Fallback to simple HTML if template execution fails
		return plainTextContent, fmt.Sprintf("<p><strong>%s</strong> invited you to join their tenant.</p><p><a href='%s'>Click here to join</a></p>",
			inviterName, inviteURL)
	}

	return plainTextContent, htmlBuffer.String()
}

// handles the email creation and sending process
func sendInvitationEmail(client *sendgrid.Client, inviteeEmail, inviterName, token string) error {
	// Set up email template configuration
	from := mail.NewEmail("Solution Pilot", "service@mail.solutionpilot.ai")
	to := mail.NewEmail("", inviteeEmail)
	subject := "You have been invited to join a tenant in Solution Pilot"

	var baseURL string
	if os.Getenv("ENVIRONMENT") == "local" {
		baseURL = "http://localhost:3000/accept-invite"
	} else if os.Getenv("ENVIRONMENT") == "dev" {
		baseURL = "https://devapp.solutionpilot.ai/accept-invite"
	} else {
		baseURL = "https://app.solutionpilot.ai/accept-invite"
	}
	inviteURL := fmt.Sprintf("%s?invite_token=%s", baseURL, token)

	plainTextContent, htmlContent := formatInvitationEmail(inviterName, inviteURL)

	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)

	// Sending the email
	response, err := client.Send(message)
	if err != nil {
		return err
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("email service returned error code: %d", response.StatusCode)
	}

	return nil
}

// handles the invitation process for a single email
func inviteSingleUser(email string, role models.UserRole, tenantID, userID string, inviter Inviter, sendgridClient *sendgrid.Client) (InvitationResponse, error) {
	var invitationResponse InvitationResponse

	// First check if the email is already a member of the tenant
	var isMember bool
	err := DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM st_schema.users u
			JOIN st_schema.tenant_members tm ON u.id = tm.user_id
			WHERE u.email = $1 AND tm.tenant_id = $2
		)
	`, email, tenantID).Scan(&isMember)

	if err != nil {
		fmt.Println("Error checking if email is already a member:", err)
		return invitationResponse, fmt.Errorf("failed to verify membership status")
	}

	if isMember {
		fmt.Printf("Email %s is already a member of the tenant\n", email)
		return invitationResponse, fmt.Errorf("User is already a member of this tenant")
	}

	// Now start the transaction only if the user isn't already a member
	tx, err := DB.Begin()
	if err != nil {
		fmt.Println("Error starting transaction:", err)
		return invitationResponse, fmt.Errorf("invitation processing failed")
	}

	// Generate a unique token for this invitation
	invitationToken, err := generateSecureToken(32)
	if err != nil {
		fmt.Println("Error generating secure token:", err)
		tx.Rollback()
		return invitationResponse, fmt.Errorf("invitation processing failed")
	}

	// Set expiration date (e.g., 30 days from now)
	expiresAt := time.Now().AddDate(0, 0, 30)

	// Insert invitation record into database within the transaction
	err = tx.QueryRow(`
		INSERT INTO st_schema.tenant_invitations
		(tenant_id, inviter_id, invitee_email, role, status, invitation_token, expires_at)
		VALUES ($1, $2, $3, $4, 'Pending', $5, $6)
		RETURNING id, status, created_at, updated_at
	`, tenantID, userID, email, role, invitationToken, expiresAt).Scan(
		&invitationResponse.ID,
		&invitationResponse.Status,
		&invitationResponse.CreatedAt,
		&invitationResponse.UpdatedAt,
	)

	if err != nil {
		// Check if this is a unique constraint violation (email already invited)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			fmt.Printf("Email %s already has a pending invitation\n", email)
			tx.Rollback()
			return invitationResponse, fmt.Errorf("User already has a pending invitation")
		}

		// Other error
		fmt.Println("Error inserting invitation:", err)
		tx.Rollback()
		return invitationResponse, fmt.Errorf("invitation processing failed")
	}

	// Send invitation email
	err = sendInvitationEmail(sendgridClient, email, inviter.FullName, invitationToken)
	if err != nil {
		fmt.Println("Error sending email to", email, err)
		tx.Rollback() // rollback database update when email fails
		return invitationResponse, fmt.Errorf("failed to send the email")
	}

	// Commit this invitation's transaction
	if err := tx.Commit(); err != nil {
		fmt.Println("Error committing transaction:", err)
		return invitationResponse, fmt.Errorf("invitation processing failed")
	}

	// Populate the rest of the response fields
	invitationResponse.InviteeEmail = email
	invitationResponse.Role = string(role)
	invitationResponse.Expired = false
	invitationResponse.Inviter = inviter

	return invitationResponse, nil
}

// GetTenantInvitations returns all invitations for a tenant with formatted fields
func GetTenantInvitations(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	// Query to get invitations with inviter information
	rows, err := DB.Query(`
		SELECT
			ti.id,
			ti.invitee_email,
			ti.role,
			ti.status,
			ti.created_at,
			ti.updated_at,
			u.first_name || ' ' || u.last_name AS inviter_name,
			u.email AS inviter_email,
			u.id AS inviter_id,
			CASE WHEN ti.expires_at < NOW() THEN true ELSE false END AS expired
		FROM
			st_schema.tenant_invitations ti
		JOIN
			st_schema.users u ON ti.inviter_id = u.id
		WHERE
			ti.tenant_id = $1
		ORDER BY
			ti.created_at DESC
	`, tenantID)

	if err != nil {
		fmt.Println("Error querying invitations:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	defer rows.Close()

	var invitations []InvitationResponse
	for rows.Next() {
		var invitation InvitationResponse

		err := rows.Scan(
			&invitation.ID,
			&invitation.InviteeEmail,
			&invitation.Role,
			&invitation.Status,
			&invitation.CreatedAt,
			&invitation.UpdatedAt,
			&invitation.Inviter.FullName,
			&invitation.Inviter.Email,
			&invitation.Inviter.ID,
			&invitation.Expired,
		)

		if err != nil {
			fmt.Println("Error scanning invitation row:", err)
			continue
		}

		invitations = append(invitations, invitation)
	}

	if err = rows.Err(); err != nil {
		fmt.Println("Error iterating invitation rows:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"invitations": invitations,
		"message":     "Invitations fetched successfully",
	})
}

// DeleteTenantInvitation completely removes an invitation from the database
func DeleteTenantInvitation(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	// Get invitation ID from the URL parameter
	invitationID := c.Param("invitation_id")
	if invitationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invitation ID is required"})
		return
	}

	// Delete the invitation
	result, err := DB.Exec(`
		DELETE FROM st_schema.tenant_invitations
		WHERE id = $1 AND tenant_id = $2
	`, invitationID, tenantID)

	if err != nil {
		fmt.Println("Error deleting invitation:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete invitation"})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Println("Error getting rows affected:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invitation not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"invitation_id": invitationID, "message": "Invitation deleted successfully"})
}
