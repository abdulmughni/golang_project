package routes

import (
	"sententiawebapi/handlers/apis/ai"
	aiFunctions "sententiawebapi/handlers/apis/ai/functions"
	"sententiawebapi/handlers/apis/community"
	"sententiawebapi/handlers/apis/templates"
	"sententiawebapi/handlers/models"
	"sententiawebapi/middlewares"

	"github.com/gin-gonic/gin"
)

// OpenAI Prompt Endpoints
// These endpoints are used for interacting with the OpenAI API. There are two types of prompts:
// 1. Assistant Prompts - These are Solution Pilot Prompts that function as agents and the configuration is handled by
// team the Soulution Pilot
// 2. Custom Prompts - These prompts allow user to pass in their own prompt configuration to OpenAI

func InitOpenAiPromptRoutes(router *gin.Engine, auth *middlewares.AuthMiddleware) {
	// New Ai Endpoints
	router.POST("/api/newPrompt", auth.RequireRole(models.UserRoleMember), ai.NewPromptHandler)
	router.GET("/api/newChatHistory", auth.RequireRole(models.UserRoleMember), ai.NewChatHistoryHandler)
	// router.GET("/api/newAiStream", auth.ValidateQueryParamToken("SENDGRID_KEY", models.UserRoleMember), ai.NewStreamHandler)

	router.POST("/api/documentSearch", auth.RequireRole(models.UserRoleMember), aiFunctions.DocumentSearchHandler)
	router.POST("/api/diagramSearch", auth.RequireRole(models.UserRoleMember), aiFunctions.DiagramSearchHandler)
	router.POST("/api/projectInfo", auth.RequireRole(models.UserRoleMember), aiFunctions.GetProjectInfoHandler)

	router.POST("api/databaseSchema", auth.RequireRole(models.UserRoleMember), ai.GenerateDatabaseDesignHandler)
}

func InitDocumentMagicianRoutes(router *gin.Engine, auth *middlewares.AuthMiddleware) {
	router.POST("/api/text/simplify", auth.RequireRole(models.UserRoleMember), ai.SimplifyTextHandler)
	router.POST("/api/text/fix-spelling-and-grammar", auth.RequireRole(models.UserRoleMember), ai.FixSpellingAndGrammarHandler)
	router.POST("/api/text/shorten", auth.RequireRole(models.UserRoleMember), ai.ShortenTextHandler)
	router.POST("/api/text/extend", auth.RequireRole(models.UserRoleMember), ai.ExtendTextHandler)
	router.POST("/api/text/adjust-tone", auth.RequireRole(models.UserRoleMember), ai.AdjustToneHandler)
	router.POST("/api/text/tldr", auth.RequireRole(models.UserRoleMember), ai.TldrHandler)
	router.POST("/api/text/prompt", auth.RequireRole(models.UserRoleMember), ai.AiWriterHandler)
	router.POST("/api/text/autocomplete", auth.RequireRole(models.UserRoleMember), ai.AutocompleteTextHandler)

	// Adding project description handler
	router.POST("/api/text/project-description", auth.RequireRole(models.UserRoleMember), ai.ProjectDescriptionHandler)
}

func InitAzOaiRoutes(router *gin.Engine, auth *middlewares.AuthMiddleware) {

	// All AI User template handlers, with publishing
	router.POST("/api/tenantAiTemplate", auth.RequireRole(models.UserRoleMember), templates.NewTenantAiTemplate)      // Creates new template
	router.GET("/api/tenantAiTemplate", auth.RequireRole(models.UserRoleMember), templates.GetTenantAiTemplate)       // Returns single template by ID
	router.GET("/api/tenantAiTemplates", auth.RequireRole(models.UserRoleMember), templates.GetTenantAiTemplates)     // Returns array of templates
	router.PUT("/api/tenantAiTemplate", auth.RequireRole(models.UserRoleMember), templates.UpdateTenantAiTemplate)    // Allows to update any field in the template
	router.DELETE("/api/tenantAiTemplate", auth.RequireRole(models.UserRoleMember), templates.DeleteTenantAiTemplate) // Releases the template

	// Handlers used for publishing and clonning templates into tenants private template repository
	router.PUT("/api/tenantAiTemplate/publish", auth.RequireRole(models.UserRoleMember), community.PublishTenantAiPromptTemplate)
	router.PUT("/api/tenantAiTemplate/unpublish", auth.RequireRole(models.UserRoleMember), community.UnpublishTenantAiPromptTemplate)
	router.POST("/api/tenantAiTemplate/clone", auth.RequireRole(models.UserRoleMember), community.ClonePublicAiPromptTemplate)

	// All Soulution Pilot AI template handlers
	router.GET("/api/spAiTemplate", auth.RequireRole(models.UserRoleMember), community.GetSpAiTemplate)
	router.GET("/api/spAiTemplates", auth.RequireRole(models.UserRoleMember), community.GetSpAiTemplates)

}

func InitAzOaiTemplateRoutes(router *gin.Engine, auth *middlewares.AuthMiddleware) {

	// cm_aiPromptConfigs.go
	// All Community AI template handlers, these ones require User to be logged in
	// These are the endpoints used in the Application
	router.GET("/api/publicTemplate", auth.ValidateJwt(), community.GetPublicPromptTemplate)
	router.GET("/api/publicTemplates", auth.ValidateJwt(), community.GetPublicPromptTemplates)

	// ai_spAiPromptConfigs.go
	// All Soulution Pilot AI template handlers (These are only Solution Pilot Templates)
	router.GET("/api/pspAiTemplate", community.GetSpAiTemplatePub)
	router.GET("/api/pspAiTemplates", community.GetSpAiTemplatesPub)

	// cm_aiPromptConfigs.go
	// All Community AI template handlers
	router.GET("/api/ppublicTemplate", community.GetPublicUserTemplatePub)
	router.GET("/api/ppublicTemplates", community.GetPublicUserTemplatesPub)
}
