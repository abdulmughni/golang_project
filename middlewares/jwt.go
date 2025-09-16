package middlewares

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sententiawebapi/handlers/models"
	"time"

	"github.com/gin-gonic/gin"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

type CustomClaims struct {
	UserId   string `json:"userId"`
	TenantId string `json:"tenantId"`
}

func (c CustomClaims) Validate(_ context.Context) error {
	if c.UserId == "" {
		return fmt.Errorf("missing required claim 'userId'")
	}
	if c.TenantId == "" {
		return fmt.Errorf("missing required claim 'tenantId'")
	}
	return nil
}

type auth0Validator struct {
	jwtValidator *validator.Validator
}

func checkRequiredVariables() error {
	requiredEnvMap := map[string]error{
		"AUTH0_DOMAIN":   fmt.Errorf("missing 'AUTH0_DOMAIN' env variable"),
		"AUTH0_AUDIENCE": fmt.Errorf("missing 'AUTH0_AUDIENCE' env variable"),
	}

	for key, err := range requiredEnvMap {
		if k := os.Getenv(key); k == "" {
			return err
		}
	}
	return nil
}

func NewValidator() (*auth0Validator, error) {
	if err := checkRequiredVariables(); err != nil {
		return nil, err
	}

	issuerURL, err := url.Parse("https://" + os.Getenv("AUTH0_DOMAIN") + "/")
	if err != nil {
		return nil, fmt.Errorf("failed to parse the issuer url: %v", err)
	}

	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)
	log.Printf("JWKS Provider URL: %s", issuerURL)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{os.Getenv("AUTH0_AUDIENCE")},
		validator.WithAllowedClockSkew(time.Minute),
		validator.WithCustomClaims(func() validator.CustomClaims {
			return &CustomClaims{}
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set up the jwt validator: %v", err)
	}

	return &auth0Validator{
		jwtValidator: jwtValidator,
	}, nil
}

func (a auth0Validator) GetTokenClaims(c *gin.Context) validator.ValidatedClaims {
	claims := c.Request.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
	return *claims
}

func (a auth0Validator) ValidateJwt() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var encounteredError error
		errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
			bodyBytes, _ := ioutil.ReadAll(r.Body)
			log.Printf("Encountered error while validating JWT: %v, response body: %s", err, string(bodyBytes))
			encounteredError = err
		}

		middleware := jwtmiddleware.New(
			a.jwtValidator.ValidateToken,
			jwtmiddleware.WithErrorHandler(errorHandler),
		)

		handler := NewHandlerFuncAdapter(func(w http.ResponseWriter, r *http.Request) {
			ctx.Request = r
			if encounteredError == nil {
				claims := a.GetTokenClaims(ctx)
				customClaims := claims.CustomClaims.(*CustomClaims)
				ctx.Set(models.UserId, customClaims.UserId)
				ctx.Set(models.TenantId, customClaims.TenantId)
			}
		})

		middleware.CheckJWT(handler).ServeHTTP(ctx.Writer, ctx.Request)

		if encounteredError != nil {
			log.Printf("JWT validation failed: %v", encounteredError)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "code": "UNAUTHORIZED_ACCESS"})
			return
		}
	}
}

type HandlerFuncAdapter struct {
	handlerFunc func(w http.ResponseWriter, r *http.Request)
}

func (h HandlerFuncAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handlerFunc(w, r)
}

func NewHandlerFuncAdapter(handlerFunc func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return HandlerFuncAdapter{handlerFunc: handlerFunc}
}
