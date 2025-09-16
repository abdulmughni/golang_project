package routes

import (
	"sententiawebapi/handlers/apis/cloud"
	"sententiawebapi/handlers/models"
	"sententiawebapi/middlewares"

	"github.com/gin-gonic/gin"
)

func InitAzureRoutes(router *gin.Engine, auth *middlewares.AuthMiddleware) {
	// Route to get Azure resources by tag
	router.POST("/api/cloud/resources", auth.RequireRole(models.UserRoleMember), cloud.GetAzureResourcesByTag)

	// Routes for managing Client Secrets
	router.GET("/api/cloud/credentials", auth.RequireRole(models.UserRoleMember), cloud.GetTenantCredentials)
	router.POST("/api/cloud/credentials", auth.RequireRole(models.UserRoleMember), cloud.CreateTenantCredential2)
	router.DELETE("/api/cloud/credentials", auth.RequireRole(models.UserRoleMember), cloud.DeleteTenantCredential)

}
