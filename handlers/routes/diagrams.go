package routes

import (
	"sententiawebapi/handlers/apis/projects"
	"sententiawebapi/handlers/models"
	"sententiawebapi/middlewares"

	"github.com/gin-gonic/gin"
)

func InitDiagramsRoutes(router *gin.Engine, auth *middlewares.AuthMiddleware) {
	router.POST("/api/diagram", auth.RequireRole(models.UserRoleMember), projects.NewDiagram)
	router.GET("/api/diagram", auth.RequireRole(models.UserRoleMember), projects.GetDiagram)
	router.GET("/api/diagrams", auth.RequireRole(models.UserRoleMember), projects.GetDiagrams)
	router.PUT("/api/diagram", auth.RequireRole(models.UserRoleMember), projects.UpdateDiagram)
	router.DELETE("/api/diagram", auth.RequireRole(models.UserRoleMember), projects.DeleteDiagram)

	router.POST("/api/diagram/clone", auth.RequireRole(models.UserRoleMember), projects.CloneDiagram)
}
