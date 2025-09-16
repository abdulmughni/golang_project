package middlewares

import (
	"database/sql"
	"log"
	"net/http"
	"sententiawebapi/handlers/apis/tenantManagement"
	"sententiawebapi/handlers/models"
	"sententiawebapi/utilities"

	"github.com/gin-gonic/gin"
)

// satisfiesRole checks if the user's role meets the required role level
func satisfiesRole(userRole, requiredRole models.UserRole) bool {
	if userRole == models.UserRoleAdmin {
		return true
	}
	return userRole == requiredRole
}

// ValidateTenant checks if user has access to the tenant with sufficient role
func ValidateTenant(requiredRole models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, tenantID, ok := utilities.ProcessIdentity(c)
		if !ok {
			return
		}

		// Get the user's role directly
		var userRole models.UserRole
		err := tenantManagement.DB.QueryRow(`
            SELECT role
            FROM st_schema.tenant_members
            WHERE user_id = $1
            AND tenant_id = $2
            AND status = 'Active'`,
			userID, tenantID).Scan(&userRole)

		if err != nil {
			if err == sql.ErrNoRows {
				// User is not a member of this tenant
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error": "You don't have access to this tenant",
					"code":  "TENANT_ACCESS_DENIED",
				})
				return
			}

			// Database error
			log.Printf("Database error checking tenant membership: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Check if the user's role is sufficient
		if !satisfiesRole(userRole, requiredRole) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
				"code":  "INSUFFICIENT_ROLE",
			})
			return
		}

		// Store the user's role in context for potential use in handlers
		c.Set("userRole", userRole)
		c.Next()
	}
}
