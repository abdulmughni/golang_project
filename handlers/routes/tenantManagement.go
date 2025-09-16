package routes

import (
	"sententiawebapi/handlers/apis/stripe"
	"sententiawebapi/handlers/apis/tenantManagement"
	"sententiawebapi/handlers/models"
	"sententiawebapi/middlewares"

	"github.com/gin-gonic/gin"
)

func InitTenantManagement(router *gin.Engine, auth *middlewares.AuthMiddleware) {
	// Adjust the route to include the object_id as a URL path parameter
	router.GET("/api/user", auth.ValidateJwt(), tenantManagement.GetUser)
	router.PATCH("/api/user", auth.ValidateJwt(), tenantManagement.UpdateUser)
	router.POST("/api/m/NewAzOiAccount", auth.RequireRole(models.UserRoleAdmin), tenantManagement.NewTenant)
	router.POST("/api/paymentSession", auth.RequireRole(models.UserRoleAdmin), stripe.CreateStripeCheckoutSession)

	// Tenant management endpoints
	router.GET("/api/tenant", auth.RequireRole(models.UserRoleMember), tenantManagement.GetTenantWithMembers)
	router.PUT("/api/tenant", auth.RequireRole(models.UserRoleAdmin), tenantManagement.UpdateTenant)

	// Tenant members endpoints
	router.PUT("/api/tenant/members/:member_id", auth.RequireRole(models.UserRoleAdmin), tenantManagement.UpdateTenantMember)
	router.DELETE("/api/tenant/members/:member_id", auth.RequireRole(models.UserRoleAdmin), tenantManagement.RemoveTenantMember)

	// Invitations management endpoints
	router.GET("/api/tenant/invitations", auth.RequireRole(models.UserRoleAdmin), tenantManagement.GetTenantInvitations)
	router.POST("/api/tenant/invitations", auth.RequireRole(models.UserRoleAdmin), tenantManagement.InviteUsersToTenant)
	router.DELETE("/api/tenant/invitations/:invitation_id", auth.RequireRole(models.UserRoleAdmin), tenantManagement.DeleteTenantInvitation)
}
