package routes

import (
	"sententiawebapi/handlers/apis/projects"
	"sententiawebapi/handlers/models"
	"sententiawebapi/middlewares"

	"github.com/gin-gonic/gin"
)

func InitProjectRoutes(router *gin.Engine, auth *middlewares.AuthMiddleware) {
	router.POST("/api/project", auth.RequireRole(models.UserRoleMember), projects.NewProject)
	router.GET("/api/project", auth.RequireRole(models.UserRoleMember), projects.GetProject)
	router.GET("/api/projects", auth.RequireRole(models.UserRoleMember), projects.GetProjects)
	router.PUT("/api/project", auth.RequireRole(models.UserRoleMember), projects.UpdateProject)
	router.DELETE("/api/project", auth.RequireRole(models.UserRoleMember), projects.DeleteProject)

	// Create project from private template
	router.POST("/api/projectFromTemplate", auth.RequireRole(models.UserRoleMember), projects.NewProjectFromTemplate)
	router.POST("/api/pub/projectFromPubTemplate", auth.RequireRole(models.UserRoleMember), projects.NewProjectFromPublicTemplate)

	// Documents Endpoints
	router.POST("/api/document", auth.RequireRole(models.UserRoleMember), projects.NewDocument)
	router.GET("/api/document", auth.RequireRole(models.UserRoleMember), projects.GetDocument)
	router.GET("/api/documents", auth.RequireRole(models.UserRoleMember), projects.GetDocuments)
	router.PUT("/api/document", auth.RequireRole(models.UserRoleMember), projects.UpdateDocument)
	router.DELETE("/api/document", auth.RequireRole(models.UserRoleMember), projects.DeleteDocument)
	router.POST("/api/document/clone", auth.RequireRole(models.UserRoleMember), projects.CloneDocument)

	router.POST("/api/conversation", auth.RequireRole(models.UserRoleMember), projects.NewConversation)
	router.GET("/api/conversation", auth.RequireRole(models.UserRoleMember), projects.GetConversation)
	router.GET("/api/conversations", auth.RequireRole(models.UserRoleMember), projects.GetConversations)
	router.PUT("/api/conversation", auth.RequireRole(models.UserRoleMember), projects.UpdateConversation)
	router.DELETE("/api/conversation", auth.RequireRole(models.UserRoleMember), projects.DeleteConversation)

	// Creates new document from private templates
	router.POST("/api/documentTemplate", auth.RequireRole(models.UserRoleMember), projects.NewDocumentFromTemplate)

	// Creates new document using public document templates
	router.POST("/api/pdt/docPubTemplate", auth.RequireRole(models.UserRoleMember), projects.NewDocumentFromPubTemplate)

	// Below handlers allow users to create new documents using private document template.
	router.POST("/api/documentFromTemplate", auth.RequireRole(models.UserRoleMember), projects.NewDocumentFromTemplate)
	router.POST("/api/pub/documentFromPubTemplate", auth.RequireRole(models.UserRoleMember), projects.NewDocumentFromPubTemplate)

	// Lists all project entities
	router.GET("/api/projectEntities", auth.RequireRole(models.UserRoleMember), projects.ListAllProjectEntities)

	// Project Requirements Endpoints
	router.GET("/api/projectRequirements", auth.RequireRole(models.UserRoleMember), projects.GetAllRequirementsHandler)
	router.POST("/api/projectRequirement", auth.RequireRole(models.UserRoleMember), projects.CreateRequirementHandler)
	router.PUT("/api/projectRequirement/:requirement_id", auth.RequireRole(models.UserRoleMember), projects.UpdateRequirementHandler)
	router.DELETE("/api/projectRequirement/:requirement_id", auth.RequireRole(models.UserRoleMember), projects.DeleteRequirementHandler)
}
