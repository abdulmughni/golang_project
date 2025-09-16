package routes

import (
	"sententiawebapi/handlers/apis/tiptap"
	"sententiawebapi/middlewares"

	"github.com/gin-gonic/gin"
)

func InitTiptapRoutes(router *gin.Engine, auth *middlewares.AuthMiddleware) {
	// Tenant membership validation happens inside the CollabHandler
	router.POST("/api/tiptap/collab", auth.ValidateJwt(), tiptap.CollabHandler)
}
