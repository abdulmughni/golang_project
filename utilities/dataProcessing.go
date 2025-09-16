package utilities

import (
	"net/http"
	"sententiawebapi/handlers/models"

	"github.com/gin-gonic/gin"
)

type ResponseData struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

// Response is a utility function that generates
// a common JSON response
func Response(c *gin.Context, statusCode int, message string, data interface{}) {
	response := ResponseData{
		Data:    data,
		Message: message,
	}
	c.JSON(statusCode, response)
}

// ProcessIdentity is a utility function
// that retrieves the user ID from the context instead of keeping
// the logic in the handler functions.
func ProcessIdentity(c *gin.Context) (userID string, tenantID string, ok bool) {
	userIDInterface, ok := c.Get(models.UserId)
	if !ok {
		Response(c, http.StatusBadRequest, models.UserIdError, nil)
		return "", "", false
	}

	// Assuming userID is stored as a string in the context
	userID, ok = userIDInterface.(string)
	if !ok {
		Response(c, http.StatusBadRequest, "Invalid user ID format", nil)
		return "", "", false
	}

	tenantIDInterface, ok := c.Get(models.TenantId)
	if !ok {
		Response(c, http.StatusBadRequest, models.TenantIdError, nil)
		return "", "", false
	}

	// Assuming tenantID is stored as a string in the context
	tenantID, ok = tenantIDInterface.(string)
	if !ok {
		Response(c, http.StatusBadRequest, "Invalid tenant ID format", nil)
		return "", "", false
	}

	return userID, tenantID, true
}
