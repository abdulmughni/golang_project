package middlewares

import (
	"log"
	"sententiawebapi/handlers/models"

	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	validator *auth0Validator
}

func NewAuthMiddleware() *AuthMiddleware {
	validator, err := NewValidator()
	if err != nil {
		log.Fatalf("Failed to create JWT validation middleware: %v", err)
	}

	return &AuthMiddleware{
		validator: validator,
	}
}

// a middleware that only validates JWT
func (auth *AuthMiddleware) ValidateJwt() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth.validator.ValidateJwt()(c)
		if c.IsAborted() {
			return
		}
		c.Next()
	}
}

// a middleware that validates both JWT and tenant access with the specified role
func (auth *AuthMiddleware) RequireRole(requiredRole models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First validate JWT
		auth.validator.ValidateJwt()(c)
		if c.IsAborted() {
			return
		}

		// Then validate tenant access
		ValidateTenant(requiredRole)(c)
		if c.IsAborted() {
			return
		}

		// If all is good, continue
		c.Next()
	}
}

func (auth *AuthMiddleware) ValidateQueryParamToken(secretEnvKey string, requiredRole models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First validate JWT in query params
		ValidateQueryParamJwt(secretEnvKey)(c)
		if c.IsAborted() {
			return
		}

		// Then validate tenant access
		ValidateTenant(requiredRole)(c)
		if c.IsAborted() {
			return
		}

		// If all is good, continue
		c.Next()
	}
}

// gets the claims from the token
func (auth *AuthMiddleware) GetTokenClaims(c *gin.Context) validator.ValidatedClaims {
	return auth.validator.GetTokenClaims(c)
}
