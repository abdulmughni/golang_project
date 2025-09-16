package routes

import (
	"sententiawebapi/handlers/apis/community"
	"sententiawebapi/handlers/models"
	"sententiawebapi/middlewares"

	"github.com/gin-gonic/gin"
)

func InitCommentsRoutes(router *gin.Engine, auth *middlewares.AuthMiddleware) {

	router.POST("/api/publicComment", auth.RequireRole(models.UserRoleMember), community.PostPublicComment)
	router.GET("/api/publicComments", auth.RequireRole(models.UserRoleMember), community.GetPublicComments)
	router.PUT("/api/publicComment", auth.RequireRole(models.UserRoleMember), community.UpdatePublicComment)
	router.DELETE("/api/publicComment", auth.RequireRole(models.UserRoleMember), community.DeletePublicComment)

}
