package routes

import (
	"sententiawebapi/handlers/apis/decisions"
	"sententiawebapi/handlers/models"
	"sententiawebapi/middlewares"

	"github.com/gin-gonic/gin"
)

func InitDecisionRoutes(router *gin.Engine, auth *middlewares.AuthMiddleware) {
	// T-Chart Endpoints
	router.GET("/api/tbars", auth.RequireRole(models.UserRoleMember), decisions.GetTBars)
	router.POST("/api/tbar", auth.RequireRole(models.UserRoleMember), decisions.NewTBar)
	router.GET("/api/tbar", auth.RequireRole(models.UserRoleMember), decisions.GetTBar)
	router.PUT("/api/tbar", auth.RequireRole(models.UserRoleMember), decisions.UpdateTBar)
	router.DELETE("/api/tbar", auth.RequireRole(models.UserRoleMember), decisions.DeleteTBar)

	router.POST("/api/tbar/argument", auth.RequireRole(models.UserRoleMember), decisions.NewTBarArgument)
	router.GET("/api/tbar/arguments", auth.RequireRole(models.UserRoleMember), decisions.GetTBarArguments)
	router.PUT("/api/tbar/argument", auth.RequireRole(models.UserRoleMember), decisions.UpdateTBarArgument)
	router.DELETE("/api/tbar/argument", auth.RequireRole(models.UserRoleMember), decisions.DeleteTBarArgument)

	// Pros & Cons Endpoints
	router.POST("/api/pnc", auth.RequireRole(models.UserRoleMember), decisions.NewPncAnalysis)
	router.PUT("/api/pnc", auth.RequireRole(models.UserRoleMember), decisions.UpdatePncAnalysis)
	router.GET("/api/pnc", auth.RequireRole(models.UserRoleMember), decisions.GetPncAnalysis)
	router.GET("/api/pncs", auth.RequireRole(models.UserRoleMember), decisions.GetAllPncAnalysis)
	router.DELETE("/api/pnc", auth.RequireRole(models.UserRoleMember), decisions.DeletePncAnalysis)

	router.POST("/api/pncArgument", auth.RequireRole(models.UserRoleMember), decisions.NewPncArgument)
	router.GET("/api/pncArguments", auth.RequireRole(models.UserRoleMember), decisions.GetAllPncArguments)
	router.PUT("/api/pncArgument", auth.RequireRole(models.UserRoleMember), decisions.UpdatePncArgument)
	router.DELETE("/api/pncArgument", auth.RequireRole(models.UserRoleMember), decisions.DeletePncArgument)
	// SWOT Endpoints
	router.POST("/api/swot", auth.RequireRole(models.UserRoleMember), decisions.NewSwot)
	router.GET("/api/swot", auth.RequireRole(models.UserRoleMember), decisions.GetSwot)
	router.GET("/api/swots", auth.RequireRole(models.UserRoleMember), decisions.GetSwots)
	router.PUT("/api/swot", auth.RequireRole(models.UserRoleMember), decisions.UpdateSwot)
	router.DELETE("/api/swot", auth.RequireRole(models.UserRoleMember), decisions.DeleteSwot)

	router.POST("/api/swotArgument", auth.RequireRole(models.UserRoleMember), decisions.NewSwotArgument)
	router.GET("/api/swotArguments", auth.RequireRole(models.UserRoleMember), decisions.GetAllSwotArguments)
	router.PUT("/api/swotArgument", auth.RequireRole(models.UserRoleMember), decisions.UpdateSwotArgument)
	router.DELETE("/api/swotArgument", auth.RequireRole(models.UserRoleMember), decisions.DeleteSwotArgument)

	// Decision Matrix Endpoints
	// Matrix Object Enpoint
	router.POST("/api/matrix", auth.RequireRole(models.UserRoleMember), decisions.NewMatrix)
	router.GET("/api/matrix", auth.RequireRole(models.UserRoleMember), decisions.GetMatrix)
	router.GET("/api/matrixs", auth.RequireRole(models.UserRoleMember), decisions.GetAllMatrixs)
	router.PUT("/api/matrix", auth.RequireRole(models.UserRoleMember), decisions.UpdateMatrix)
	router.DELETE("/api/matrix", auth.RequireRole(models.UserRoleMember), decisions.DeleteMatrix)

	// Matrix Criteria Endpoints
	router.POST("/api/matrixCriteria", auth.RequireRole(models.UserRoleMember), decisions.NewMatrixCriteria)
	router.PUT("/api/matrixCriteria", auth.RequireRole(models.UserRoleMember), decisions.UpdateMatrixCriteria)
	router.DELETE("/api/matrixCriteria", auth.RequireRole(models.UserRoleMember), decisions.DeleteMatrixCriteria)
	router.GET("/api/matrixCriterias", auth.RequireRole(models.UserRoleMember), decisions.GetAllMatrixCriteria)
	router.GET("/api/matrixCriteria", auth.RequireRole(models.UserRoleMember), decisions.GetMatrixCriteria)

	// Matrix Concepts Endpoints
	router.POST("/api/matrixConcept", auth.RequireRole(models.UserRoleMember), decisions.NewMatrixConcept)
	router.GET("/api/matrixConcept", auth.RequireRole(models.UserRoleMember), decisions.GetMatrixConcept)
	router.GET("/api/matrixConcepts", auth.RequireRole(models.UserRoleMember), decisions.GetAllMatrixConcepts)
	router.PUT("/api/matrixConcept", auth.RequireRole(models.UserRoleMember), decisions.UpdateMatrixConcept)
	router.DELETE("/api/matrixConcept", auth.RequireRole(models.UserRoleMember), decisions.DeleteMatrixConcept)

	// Matrix User Rating
	router.PUT("/api/matrixUserRating", auth.RequireRole(models.UserRoleMember), decisions.UpdateMatrixUserRating)
}
