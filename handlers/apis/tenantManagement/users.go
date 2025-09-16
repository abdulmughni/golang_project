package tenantManagement

// Tenant Management Handlers

import (
	"database/sql"
	"log"
	"net/http"
	"sententiawebapi/handlers/models"
	"sententiawebapi/utilities"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type UserUpdateInput struct {
	FirstName  *string `json:"first_name"`
	LastName   *string `json:"last_name"`
	Occupation *string `json:"occupation"`
	Status     *string `json:"status"`
}

var DB *sql.DB // This will be initialized from main.go

func GetUser(c *gin.Context) {
	userID, _, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	var user models.User
	err := DB.QueryRow(`
		SELECT u.id, u.email, u.first_name, u.last_name, u.occupation, u.status,
			tm.tenant_id, tm.role
		FROM st_schema.users u
		JOIN st_schema.tenant_members tm ON u.id = tm.user_id
		WHERE u.id = $1`,
		userID).Scan(
		&user.ID, &user.Email, &user.FirstName, &user.LastName,
		&user.Occupation, &user.Status, &user.TenantID, &user.TenantRole)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(404, gin.H{"error": "User not found"})
			return
		}

		log.Printf(models.DatabaseError, err)
		c.JSON(500, gin.H{"error": "Failed to retrieve user"})
		return
	}

	c.JSON(200, gin.H{
		"data": user,
	})
}

func UpdateUser(c *gin.Context) {
	userID, _, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	var userInput UserUpdateInput
	if err := c.ShouldBindJSON(&userInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := "UPDATE st_schema.users SET"
	var updateArgs []interface{}
	var updates []string

	if userInput.FirstName != nil {
		updates = append(updates, " first_name = $"+strconv.Itoa(len(updateArgs)+1))
		updateArgs = append(updateArgs, *userInput.FirstName)
	}
	if userInput.LastName != nil {
		updates = append(updates, " last_name = $"+strconv.Itoa(len(updateArgs)+1))
		updateArgs = append(updateArgs, *userInput.LastName)
	}
	if userInput.Occupation != nil {
		updates = append(updates, " occupation = $"+strconv.Itoa(len(updateArgs)+1))
		updateArgs = append(updateArgs, *userInput.Occupation)
	}
	if userInput.Status != nil {
		updates = append(updates, " status = $"+strconv.Itoa(len(updateArgs)+1))
		updateArgs = append(updateArgs, *userInput.Status)
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	query += strings.Join(updates, ",")
	query += " WHERE id = $" + strconv.Itoa(len(updateArgs)+1)
	updateArgs = append(updateArgs, userID)

	_, err := DB.Exec(query, updateArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user details"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User details updated successfully"})
}
