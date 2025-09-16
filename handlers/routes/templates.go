package routes

import (
	"sententiawebapi/handlers/apis/community"
	"sententiawebapi/handlers/apis/templates"
	"sententiawebapi/handlers/models"
	"sententiawebapi/middlewares"

	"github.com/gin-gonic/gin"
)

// pdt stands for Public Document Templates

func InitProjectTemplateRoutes(router *gin.Engine, auth *middlewares.AuthMiddleware) {
	// Project Template APIs
	router.POST("/api/projectTemplate", auth.RequireRole(models.UserRoleMember), templates.CreateProjectTemplate)
	router.GET("/api/projectTemplate", auth.RequireRole(models.UserRoleMember), templates.GetProjectTemplate)
	router.GET("/api/projectTemplates", auth.RequireRole(models.UserRoleMember), templates.GetProjectTemplates)
	router.PUT("/api/projectTemplate", auth.RequireRole(models.UserRoleMember), templates.UpdateProjectTemplate)
	router.DELETE("/api/projectTemplate", auth.RequireRole(models.UserRoleMember), templates.DeleteProjectTemplate)
}

func InitDocumentTemplateRoutes(router *gin.Engine, auth *middlewares.AuthMiddleware) {
	// Internal document template handlers, these are the endpoints for interacting with internal project templates
	// they include listing documents, publishing and unpublishing documents also editing

	// Private document template endpoints (GET ALL, GET ONE, UPDATE, DELETE)
	router.POST("/api/idt/documentTemplate", auth.RequireRole(models.UserRoleMember), templates.NewInternalDocumentTemplate)
	router.GET("/api/idt/documentTemplate", auth.RequireRole(models.UserRoleMember), templates.GetInternalDocumentTemplate)
	router.GET("/api/idt/documentTemplates", auth.RequireRole(models.UserRoleMember), templates.GetInternalDocumentTemplates)
	router.PUT("/api/idt/documentTemplate", auth.RequireRole(models.UserRoleMember), templates.UpdateInternalDocumentTemplate)
	router.DELETE("/api/idt/documentTemplate", auth.RequireRole(models.UserRoleMember), templates.DeleteInternalDocumentTemplate)

	// Community document template endpoints
	router.POST("/api/publicDocumentTemplate", auth.RequireRole(models.UserRoleMember), community.NewPublicTemplateDocument)
	router.GET("/api/publicDocumentTemplate", community.GetPublicTemplateDocument)
	router.GET("/api/publicDocumentTemplates", community.GetPublicTemplateDocuments)
	router.PUT("/api/publicDocumentTemplate", auth.RequireRole(models.UserRoleMember), community.UpdatePublicTemplateDocument)
	router.DELETE("/api/publicDocumentTemplate", auth.RequireRole(models.UserRoleMember), community.DeletePublicTemplateDocument)

	// Private diagram template endpoints
	router.POST("/api/idt/diagramTemplate", auth.RequireRole(models.UserRoleMember), templates.NewInternalDiagramTemplate)
	router.GET("/api/idt/diagramTemplate", auth.RequireRole(models.UserRoleMember), templates.GetInternalDiagramTemplate)
	router.GET("/api/idt/diagramTemplates", auth.RequireRole(models.UserRoleMember), templates.GetInternalDiagramTemplates)
	router.PUT("/api/idt/diagramTemplate", auth.RequireRole(models.UserRoleMember), templates.UpdateInternalDiagramTemplate)
	router.DELETE("/api/idt/diagramTemplate", auth.RequireRole(models.UserRoleMember), templates.DeleteInternalDiagramTemplate)

	// Community diagram template endpoints
	router.POST("/api/publicDiagramTemplate", auth.RequireRole(models.UserRoleMember), community.NewPublicDiagramTemplate)
	router.GET("/api/publicDiagramTemplate", community.GetPublicDiagramTemplate)
	router.GET("/api/publicDiagramTemplates", community.GetPublicDiagramTemplates)
	router.PUT("/api/publicDiagramTemplate", auth.RequireRole(models.UserRoleMember), community.UpdatePublicDiagramTemplate)
	router.DELETE("/api/publicDiagramTemplate", auth.RequireRole(models.UserRoleMember), community.DeletePublicDiagramTemplate)

	// Public document component templates
	router.GET("/api/dcm/component", auth.RequireRole(models.UserRoleMember), templates.GetDocumentComponent)
	router.GET("/api/dcm/components", auth.RequireRole(models.UserRoleMember), templates.GetDocumentComponents)
	router.GET("/api/dcm/favoriteComponents", auth.RequireRole(models.UserRoleMember), templates.GetFavoriteDocumentComponents)

	router.POST("/api/dcm/pinComponent", auth.RequireRole(models.UserRoleMember), templates.PinDocumentComponent)
	router.POST("/api/dcm/unpinComponent", auth.RequireRole(models.UserRoleMember), templates.UnpinDocumentComponent)
}

// Publishing APIs, below handlers allow users to publish their project templates
// to the community
func InitPublicTemplateRouters(router *gin.Engine, auth *middlewares.AuthMiddleware) {
	router.POST("/api/publishProjectTemplate", auth.RequireRole(models.UserRoleMember), community.PublishProjectTemplate)
	router.POST("/api/unpublishProjectTemplate", auth.RequireRole(models.UserRoleMember), community.UnpublishProjectTemplate)

}

func InitPublicTemplatePubRouters(router *gin.Engine, auth *middlewares.AuthMiddleware) {
	// Document Template details
	// router.GET("/api/documentTemplate", jwtMiddleware, templates.GetProjectTemplateDocument)
	router.GET("/api/publicProjectTemplate", auth.RequireRole(models.UserRoleMember), community.GetPublicProjectTemplate)
	router.GET("/api/publicProjectTemplates", auth.RequireRole(models.UserRoleMember), community.GetPublicProjectTemplates)
	router.GET("/api/publicProjectDocumentTemplate", auth.RequireRole(models.UserRoleMember), community.GetPublicTemplateDocument)
	router.GET("/api/publicProjectDiagramTemplate", auth.RequireRole(models.UserRoleMember), community.GetPublicDiagramTemplate)
	router.PUT("/api/publicProjectTemplate", auth.RequireRole(models.UserRoleMember), templates.UpdatePublicProjectTemplate)

	// Community Template APIs for Website ( don't require JWT Token )
	router.GET("/api/pub/publicProjectTemplate", community.GetWebPublicProjectTemplate)
	router.GET("/api/pub/publicProjectTemplates", community.GetWebPublicProjectTemplates)
	router.GET("/api/pub/publicProjectTemplatesPag", community.GetWebPublicProjectTemplatesPagination)

	router.GET("/api/pub/publicProjectDocumentTemplate", community.GetWebPublicProjectTemplateDocument)
	router.GET("/api/pub/publicProjectDiagramTemplate", community.GetWebPublicProjectTemplateDiagram)

	// Clone public project template
	router.POST("/api/clonePublicProjectTemplate", auth.RequireRole(models.UserRoleMember), community.ClonePublicProjectTemplate)

}
