package middlewares

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sententiawebapi/handlers/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func ValidateQueryParamJwt(secretEnvKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.Query("token")

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing token query parameter"})
			return
		}

		secret := os.Getenv(secretEnvKey)
		if secret == "" {
			log.Printf("Missing secret for: %s", secretEnvKey)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &models.QueryParamToken{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				log.Printf("Unexpected signing method: %v", token.Header["alg"])

				return nil, fmt.Errorf("unauthorized")
			}

			return []byte(secret), nil
		})

		if err != nil {
			log.Printf("Error parsing token: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})

			return
		}

		claims, ok := token.Claims.(*models.QueryParamToken)

		if !ok || !token.Valid {
			log.Printf("Invalid token claims")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})

			return
		}

		// Check the token's registered claims (including expiry and not-before)
		validator := jwt.NewValidator()
		if err := validator.Validate(claims); err != nil {
			log.Printf("Token validation error: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token expired or invalid", "message": err.Error()})
			return
		}

		c.Set(models.UserId, claims.UserID)
		c.Set(models.TenantId, claims.TenantID)

		log.Printf("Token successfully validated")
	}
}
