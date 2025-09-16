package utilities

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"

	"github.com/gin-gonic/gin"

	"sententiawebapi/handlers/models"
)

var DB *sql.DB // This will be initialized from main.go

func ValidateEmail(email string) bool {
	// A simple regex for email validation
	var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,6}$`)
	return emailRegex.MatchString(email)
}

func HealthCheck(c *gin.Context) {

	// _, err := DB.Query("SELECT 1")
	// if err != nil {
	// 	c.JSON(500, gin.H{"error": "Service is unhealthy"})
	// 	return
	// }
	c.JSON(200, gin.H{"message": "Service is healthy"})
}

// Used for resolving test response status
func ResolveStatus(rr *httptest.ResponseRecorder) {

	if rr.Code != http.StatusOK && rr.Code != http.StatusCreated {
		log.Fatalf("Test failed with status code: %d, response body: %s", rr.Code, rr.Body.String())
		log.Println(models.Red + "FAILED" + models.Reset)
	} else {
		log.Printf("Test succeeded with status code: %d, response body: %s", rr.Code, rr.Body.String())
		log.Println(models.Green + "SUCCESS" + models.Reset)
	}
}

// Helper function to get a query parameter and check if it's valid
func ValidateQueryParam(c *gin.Context, param string) (string, bool) {
	value := c.Query(param)
	if value == "" {
		log.Printf("ERROR: %v", param+" is required...")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal server error..."})
		return "", false
	}
	return value, true
}

func ResolveResourceIdentifier(c *gin.Context) (*models.ResourceIdentifier, error) {
	rgt := models.ResourceGroupType(c.Query("rgt"))
	rgi := c.Query("rgi")
	rt := models.ResourceType(c.Query("rt"))
	ri := c.Query("ri")

	if rgt == "" || !rgt.IsValid() {
		return nil, fmt.Errorf("invalid or missing resource group type (rgt)")
	}

	identifier := &models.ResourceIdentifier{
		ResourceGroupType: rgt,
	}

	if rgi != "" {
		identifier.ResourceGroupID = &rgi
	}

	if rt != "" {
		if !rt.IsValid() {
			return nil, fmt.Errorf("invalid resource type (rt)")
		}
		identifier.ResourceType = &rt
	}

	if ri != "" {
		identifier.ResourceID = &ri
	}

	return identifier, nil
}

func Ptr[T any](v T) *T {
	return &v
}
