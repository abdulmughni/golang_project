package tiptap

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"sententiawebapi/handlers/apis/tenantManagement"
	"sententiawebapi/handlers/models"
	"sententiawebapi/utilities"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v5"
)

type RequestBody struct {
	ResourceType string `json:"resource_type" binding:"required"`
	ResourceId   string `json:"resource_id" binding:"required"`
}

type CollabTokenClaims struct {
	UserID          string `json:"user_id" binding:"required"`
	TenantID        string `json:"tenant_id" binding:"required"`
	BlobContainerId string `json:"blob_container_id" binding:"required"`
	jwt.RegisteredClaims
}

// validates that the user is an active member of the tenant,
// and returns the tenant's blob container ID. Does not require any specific member role.
func validateTenantAndGetContainerID(c *gin.Context, userID string, tenantID string) (*uuid.UUID, bool) {
	var containerID uuid.UUID
	err := tenantManagement.DB.QueryRow(`
        SELECT t.blob_container_id
        FROM st_schema.tenant_members tm
        JOIN st_schema.tenants t ON t.id = tm.tenant_id
        WHERE tm.user_id = $1
          AND tm.tenant_id = $2
          AND tm.status = 'Active'
    `, userID, tenantID).Scan(&containerID)

	if err != nil {
		if err == sql.ErrNoRows {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "You don't have access to this tenant",
				"code":  "TENANT_ACCESS_DENIED",
			})
		} else {
			log.Printf("Database error checking tenant membership: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
		}
		return nil, false
	}

	return &containerID, true
}

func validateUserAccess(tenantID string, resourceType string, resourceId string) (bool, error) {
	var query string
	switch resourceType {
	case "doc":
		query = `
		  SELECT 1
          FROM st_schema.project_documents
          WHERE id = $1 AND tenant_id = $2
		`
	case "diagram":
		query = `
		  SELECT 1
          FROM st_schema.diagrams
          WHERE id = $1 AND tenant_id = $2
		`
	case "doc-template":
		query = `
		  SELECT 1
          FROM st_schema.document_templates
          WHERE id = $1 AND tenant_id = $2
		`
	case "diagram-template":
		query = `
		  SELECT 1
          FROM st_schema.diagram_templates
          WHERE id = $1 AND tenant_id = $2
		`
	case "cm-doc-template":
		query = `
		  SELECT 1
          FROM st_schema.cm_document_templates
          WHERE id = $1 AND tenant_id = $2
		`
	case "cm-diagram-template":
		query = `
		  SELECT 1
          FROM st_schema.cm_diagram_templates
          WHERE id = $1 AND tenant_id = $2
		`
	default:
		log.Printf("Unknown resource type: %s", resourceType)
		return false, nil
	}

	var exists int
	err := tenantManagement.DB.QueryRow(query, resourceId, tenantID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("db error validating access: %w", err)
	}

	return exists == 1, nil
}

func CollabHandler(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	blobContainerId, isValidTenantMember := validateTenantAndGetContainerID(c, userID, tenantID)
	if !isValidTenantMember {
		return
	}
	if blobContainerId == nil {
		log.Print("Blob container id is nil")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": models.InternalServerError,
		})
		return
	}

	secret := os.Getenv("TIPTAP_COLLAB_SECRET")
	if secret == "" {
		log.Print("No collaboration token provided, please set TIPTAP_COLLAB_SECRET in your environment")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": models.InternalServerError,
		})
		return
	}

	var requestBody RequestBody
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	userHasAccess, err := validateUserAccess(tenantID, requestBody.ResourceType, requestBody.ResourceId)
	if err != nil {
		log.Print(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": models.InternalServerError})
		return
	}
	if !userHasAccess {
		log.Printf("Access denied: user_id=%s resource_type=%s resource_id=%s", userID, requestBody.ResourceType, requestBody.ResourceId)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You don't have access to this resource."})
		return
	}

	// Create the JWT claims
	expiresAt := time.Now().Add(10 * time.Minute).UTC()
	claims := CollabTokenClaims{
		UserID:          userID,
		TenantID:        tenantID,
		BlobContainerId: blobContainerId.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			Audience:  jwt.ClaimStrings{requestBody.ResourceType, requestBody.ResourceId},
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":      tokenString,
		"expires_at": expiresAt.Format(time.RFC3339), // ISO 8601
	})
}

func ValidateCollabTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.Query("token")
		audStr := c.Query("aud")

		if tokenString == "" || audStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing token or audience query parameter"})
			return
		}

		secret := os.Getenv("TIPTAP_COLLAB_SECRET")
		if secret == "" {
			log.Print("Missing TIPTAP_COLLAB_SECRET")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &CollabTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				log.Printf("Unexpected signing method: %v", token.Header["alg"])
				return nil, fmt.Errorf("unauthorized")
			}
			return []byte(secret), nil
		})

		if err != nil {
			log.Printf("Error parsing token: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		claims, ok := token.Claims.(*CollabTokenClaims)
		if !ok || !token.Valid {
			log.Printf("Invalid token claims")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Validate standard registered claims (exp, nbf, etc.)
		validator := jwt.NewValidator()
		if err := validator.Validate(claims); err != nil {
			log.Printf("Token validation error: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token expired or invalid", "message": err.Error()})
			return
		}

		// Parse and validate audience
		parts := strings.Split(audStr, "_")
		if len(parts) != 2 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid aud format"})
			return
		}
		expectedAud := jwt.ClaimStrings{parts[0], parts[1]}
		if !reflect.DeepEqual(claims.Audience, expectedAud) {
			log.Printf("Audience mismatch. Got: %v, Expected: %v", claims.Audience, expectedAud)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Audience does not match"})
			return
		}

		// Set values into context
		c.Set(models.UserId, claims.UserID)
		c.Set(models.TenantId, claims.TenantID)
		c.Set("BlobContainerID", claims.BlobContainerId)

		log.Printf("Collab token successfully validated")
		c.Next()
	}
}
