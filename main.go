package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sententiawebapi/handlers/apis/community"
	"sententiawebapi/handlers/apis/tenantManagement"
	"sententiawebapi/handlers/routes"
	"sententiawebapi/middlewares"
	"sententiawebapi/utilities"

	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/microsoft/ApplicationInsights-Go/appinsights"
)

var (
	db              *sql.DB
	pdb             *sql.DB
	telemetryClient appinsights.TelemetryClient
)

func init() {
	// Extract just the instrumentation key from the connection string
	instrumentationKey := os.Getenv("INSTRUMENTATION_KEY")
	// Remove the "InstrumentationKey=" prefix if present
	if key := strings.TrimPrefix(instrumentationKey, "InstrumentationKey="); strings.Contains(key, ";") {
		instrumentationKey = strings.Split(key, ";")[0]
	}

	// Initialize Application Insights client with just the key
	telemetryClient = appinsights.NewTelemetryClient(instrumentationKey)
	telemetryClient.Context().Tags.Cloud().SetRole("goapi")

	// Configure Application Insights to be less verbose
	config := appinsights.NewTelemetryConfiguration(instrumentationKey)
	config.MaxBatchSize = 8192            // Increase batch size to reduce transmission frequency
	config.MaxBatchInterval = time.Minute // Only send telemetry once per minute

	// Only log errors and critical diagnostic messages
	appinsights.NewDiagnosticsMessageListener(func(msg string) error {
		if strings.Contains(strings.ToLower(msg), "error") ||
			strings.Contains(strings.ToLower(msg), "critical") {
			fmt.Printf("[%s] %s\n", time.Now().Format(time.RFC3339), msg)
		}
		return nil
	})

	// Check if .env file exists
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		log.Println("Loading data from the environment")
	} else {
		// Load environment variables from .env file
		if err := godotenv.Load(); err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_NAME"))

	var err error
	db, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatalf("Error initializing database connection: %v", err)
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(30 * time.Minute)

	// Test the connection
	if err = db.Ping(); err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	log.Println("Successfully connected to database with primary user...")

	pubConnectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_RO_USER_NAME"), os.Getenv("DB_RO_USER_PASS"), os.Getenv("DB_NAME"))

	pdb, err = sql.Open("postgres", pubConnectionString)
	if err != nil {
		log.Fatalf("Error initializing public database connection: %v", err)
	}

	// Test the public connection
	if err = pdb.Ping(); err != nil {
		log.Fatalf("Error connecting to database with public user: %v", err)
	}
	log.Println("Successfully connected to database with public user...")
}

func main() {
	router := gin.Default()

	// Add Application Insights middleware before other middleware
	router.Use(func(c *gin.Context) {
		startTime := time.Now()

		// Process request
		c.Next()

		// Create request telemetry after processing
		duration := time.Since(startTime)
		requestTelemetry := appinsights.NewRequestTelemetry(
			c.Request.Method,
			c.Request.URL.String(),
			duration,
			fmt.Sprintf("%d", c.Writer.Status()),
		)

		// Add custom properties
		requestTelemetry.Properties["userAgent"] = c.Request.UserAgent()
		requestTelemetry.Properties["clientIP"] = c.ClientIP()
		requestTelemetry.Properties["path"] = c.FullPath()

		// Track the request
		telemetryClient.Track(requestTelemetry)

		// Track any errors if status code is 4xx or 5xx
		if c.Writer.Status() >= 400 {
			telemetryClient.TrackTrace(
				fmt.Sprintf("Request failed: %s %s", c.Request.Method, c.Request.URL.String()),
				appinsights.Error,
			)
		}
	})

	// Configure CORS based on environment variable
	env := os.Getenv("ENVIRONMENT")
	crs := cors.DefaultConfig()

	switch env {
	case "local":
		crs.AllowOrigins = []string{
			"http://localhost:3001",
			"http://localhost:3000",
			"https://app.solutionpilot.ai",
			"https://solutionpilot.ai",
			"https://checkout.stripe.com",
			"https://*.stripe.com",
		}
	case "dev":
		crs.AllowOrigins = []string{
			"http://localhost:3001",
			"http://localhost:3000",
			"https://devapp.solutionpilot.ai",
			"https://dev.solutionpilot.ai",
			"https://checkout.stripe.com",
			"https://*.stripe.com",
			"https://golang-project-abdul-mughnis-projects-916437c0.vercel.app",
		}
	case "prod", "production":
		crs.AllowOrigins = []string{
			"http://localhost:3001",
			"http://localhost:3000",
			"https://app.solutionpilot.ai",
			"https://solutionpilot.ai",
			"https://checkout.stripe.com",
			"https://*.stripe.com",
			"http://localhost:3001",
			"http://localhost:3000",
			"https://golang-project-abdul-mughnis-projects-916437c0.vercel.app",
		}
	default:
		crs.AllowOrigins = []string{
			"http://localhost:3001",
			"http://localhost:3000",
			"https://devapp.solutionpilot.ai",
			"https://dev.solutionpilot.ai",
			"https://checkout.stripe.com",
			"https://*.stripe.com",
			"https://golang-project-abdul-mughnis-projects-916437c0.vercel.app",
		} // Specify a default set of origins or leave it empty to disallow all
	}

	crs.AllowCredentials = true
	// Allow Private Network Access preflight (Chrome) when calling local API from HTTPS sites
	crs.AllowPrivateNetwork = true
	crs.AllowHeaders = []string{
		"Origin",
		"Content-Type",
		"Accept",
		"Authorization",
		"Stripe-Version",   // Add these
		"Stripe-Signature", // Stripe-specific
		"X-Requested-With",
		"X-Tenant-Id",
		"X-Tenant-ID",
		"Accept-Language",
		"Cache-Control",
		"Pragma",
	}
	crs.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}
	crs.ExposeHeaders = []string{"Content-Length"}
	// Allow Vercel preview and SolutionPilot subdomains dynamically
	crs.AllowOriginFunc = func(origin string) bool {
		o := strings.ToLower(strings.TrimSpace(origin))
		if o == "" {
			return false
		}
		if strings.HasSuffix(o, ".vercel.app") {
			return true
		}
		if strings.HasSuffix(o, ".solutionpilot.ai") {
			return true
		}
		for _, allowed := range crs.AllowOrigins {
			if o == strings.ToLower(strings.TrimSuffix(allowed, "/")) {
				return true
			}
		}
		return false
	}

	router.Use(cors.New(crs))

	// Ensure all preflight requests are short-circuited before hitting auth middlewares
	router.OPTIONS("/*path", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	// Add middleware to log CORS issues
	router.Use(func(c *gin.Context) {
		c.Next()
		if c.Writer.Status() == http.StatusForbidden {
			log.Printf("CORS blocked request from origin: %s", c.Request.Header.Get("Origin"))
		}
	})

	// Set GIN Mode
	if env := os.Getenv("ENVIRONMENT"); env == "prod" {
		gin.SetMode(gin.DebugMode)
		log.Printf("Running in %s mode", env)
	} else {
		gin.SetMode(gin.DebugMode)
		log.Printf("Running in %s mode", env)
	}

	// Initialize handlers' database connection
	tenantManagement.DB = db
	community.PDB = pdb

	// Create the combined auth middleware
	// Validates both JWT and tenant access
	auth := middlewares.NewAuthMiddleware()

	// Auth Endpoints
	router.GET("/api/jwt", auth.ValidateJwt(), func(context *gin.Context) {
		context.JSON(http.StatusOK, auth.GetTokenClaims(context))
	})

	// Health Endpoints
	router.GET("/api/health", utilities.HealthCheck)

	// Tenant Management
	routes.InitTenantManagement(router, auth)

	// Projects Endpoints
	routes.InitProjectRoutes(router, auth)
	routes.InitProjectTemplateRoutes(router, auth)
	routes.InitDocumentTemplateRoutes(router, auth)
	routes.InitDecisionRoutes(router, auth)
	routes.InitDiagramsRoutes(router, auth)

	// Ai Features Endpoints
	// Azure Open AI

	routes.InitAzOaiRoutes(router, auth)
	routes.InitAzOaiTemplateRoutes(router, auth)
	routes.InitPublicTemplateRouters(router, auth)
	routes.InitPublicTemplatePubRouters(router, auth)

	routes.InitOpenAiPromptRoutes(router, auth)
	routes.InitDocumentMagicianRoutes(router, auth)

	// Community Endpoints
	routes.InitCommentsRoutes(router, auth)

	// Cloud Endpoints
	routes.InitAzureRoutes(router, auth)

	// Tiptap Endpoints
	routes.InitTiptapRoutes(router, auth)

	// Image Storage Endpoints
	routes.InitImagesRoutes(router, auth)

	// Track API startup
	telemetryClient.TrackEvent("APIStartup")
	telemetryClient.TrackTrace("Starting Sententia Prime API", appinsights.Information)

	// Print specific environment variables before starting the server
	log.Println("=== Environment Variables ===")
	// Application Related
	log.Printf("[VAR]GIN_MODE=%s", os.Getenv("GIN_MODE"))
	log.Printf("[VAR]ENVIRONMENT=%s", os.Getenv("ENVIRONMENT"))

	// Database Related
	log.Println("=== Database Variables ===")
	log.Printf("[VAR]DB_HOST=%s", os.Getenv("DB_HOST"))
	log.Printf("[VAR]DB_PORT=%s", os.Getenv("DB_PORT"))
	log.Printf("[VAR]DB_USER=%s", os.Getenv("DB_USER"))
	if os.Getenv("DB_PASS") != "" {
		log.Printf("[VAR]DB_PASS=[CONFIGURED]")
	} else {
		log.Printf("[VAR]DB_PASS=[NOT CONFIGURED]")
	}
	log.Printf("[VAR]DB_NAME=%s", os.Getenv("DB_NAME"))

	log.Printf("[VAR]DB_RO_USER_NAME=%s", os.Getenv("DB_RO_USER_NAME"))
	log.Printf("[VAR]DB_RO_USER_PASS=%s", os.Getenv("DB_RO_USER_PASS"))

	// Azure Key Vault Related
	log.Println("=== Azure Variables ===")
	log.Printf("[VAR]AZURE_KEY_VAULT_URL=%s", os.Getenv("AZURE_KEY_VAULT_URL"))
	if os.Getenv("ENVIRONMENT") == "prod" {
		log.Printf("[VAR]ARM_TENANT_ID=%s", os.Getenv("ARM_TENANT_ID"))
		log.Printf("[VAR]ARM_CLIENT_ID=%s", os.Getenv("ARM_CLIENT_ID"))
		if os.Getenv("ARM_CLIENT_SECRET") != "" {
			log.Printf("[VAR]ARM_CLIENT_SECRET=[CONFIGURED]")
		} else {
			log.Printf("[VAR]ARM_CLIENT_SECRET=[NOT CONFIGURED]")
		}
	} else {
		log.Printf("[VAR]AZURE_MANAGED_IDENTITY_CLIENT_ID=%s", os.Getenv("AZURE_MANAGED_IDENTITY_CLIENT_ID"))
	}

	// Stripe Related
	log.Println("=== Stripe Variables ===")
	if os.Getenv("STRIPE_SECRET_KEY") != "" {
		log.Printf("[VAR]STRIPE_SECRET_KEY=[CONFIGURED]")
	} else {
		log.Printf("[VAR]STRIPE_SECRET_KEY=[NOT CONFIGURED]")
	}
	// Auth0 Related
	log.Println("=== Auth0 Variables ===")
	log.Printf("[VAR]AUTH0_DOMAIN=%s", os.Getenv("AUTH0_DOMAIN"))
	log.Printf("[VAR]AUTH0_AUDIENCE=%s", os.Getenv("AUTH0_AUDIENCE"))
	log.Printf("[VAR]AUTH0_CLIENT_ID=%s", os.Getenv("AUTH0_CLIENT_ID"))
	if os.Getenv("AUTH0_CLIENT_SECRET") != "" {
		log.Printf("[VAR]AUTH0_CLIENT_SECRET=[CONFIGURED]")
	} else {
		log.Printf("[VAR]AUTH0_CLIENT_SECRET=[NOT CONFIGURED]")
	}

	log.Println("==========================")

	// Verify database connections before starting
	log.Println("=== Verifying Database Connections ===")

	// Test primary connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Error connecting to database with primary user: %v", err)
	}
	log.Println("✓ Primary database connection verified")

	// Test public connection
	if err := pdb.Ping(); err != nil {
		log.Fatalf("Error connecting to database with public user: %v", err)
	}
	log.Println("✓ Public database connection verified")

	// Log CORS configuration
	log.Println("=== CORS Configuration ===")
	log.Printf("Allowed Origins: %v", crs.AllowOrigins)
	log.Printf("Allowed Methods: %v", crs.AllowMethods)
	log.Printf("Allowed Headers: %v", crs.AllowHeaders)
	log.Println("==========================")

	// Start the server
	if os.Getenv("ENVIRONMENT") == "prod" || os.Getenv("ENVIRONMENT") == "dev" {
		// Production needs to be port 80
		router.Run(":80")
	} else {
		// Development needs to be port 8080
		router.Run(":8080")
	}
}
