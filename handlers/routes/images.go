package routes

import (
	"sententiawebapi/handlers/apis/images"
	"sententiawebapi/handlers/apis/tiptap"
	"sententiawebapi/handlers/models"
	"sententiawebapi/middlewares"

	"github.com/gin-gonic/gin"
)

func InitImagesRoutes(router *gin.Engine, auth *middlewares.AuthMiddleware) {
	router.POST("/api/images", auth.RequireRole(models.UserRoleMember), images.UploadImageHandler)
	router.GET("/api/images/:filename", tiptap.ValidateCollabTokenMiddleware(), images.GetImageHandler)
}
