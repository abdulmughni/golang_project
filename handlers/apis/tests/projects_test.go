package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"sententiawebapi/handlers/apis/decisions"
	"sententiawebapi/handlers/apis/projects"
	"sententiawebapi/handlers/apis/templates"
	"sententiawebapi/handlers/apis/tenantManagement"
	"sententiawebapi/middlewares"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

var (

	//Database Vars
	testDB *sql.DB

	// Project Vars
	createdProjectID         string
	createdConversationID    string
	createdDocumentID        string
	createdAzOAITemplate     string
	createdProjectTemplateID string
	createdDiagramID         string

	// TBar Vars
	createdTBarID         string
	createdTBarOptionAID  string
	createdTBarOptionBID  string
	createdTBarArgumentID string

	// Pnc Vars
	createdPncID          string
	createdPncArgumentID  string
	createdSwotID         string
	createdSwotArgumentID string

	// Assistant Vars
	createdAssistantID string

	// ANSI color codes
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorReset  = "\033[0m"

	failedTests []string
)

// TestMain runs before all tests
func TestMain(m *testing.M) {
	gin.ForceConsoleColor()
	testDB = initTestDB()
	code := m.Run()

	if testDB != nil {
		testDB.Close()
	}

	if len(failedTests) > 0 {
		fmt.Println("\n==================== Failed Tests Summary ====================")
		fmt.Println("|        Test Name         |")
		fmt.Println("|--------------------------|")
		for _, name := range failedTests {
			fmt.Printf("| %-24s |\n", name)
		}
		fmt.Println("=============================================================")
	} else {
		fmt.Println("\nAll tests passed!")
	}

	os.Exit(code)
}

// Initialaze db for tests its also reused in other test packages.
func initTestDB() *sql.DB {
	// Assuming you have a similar .env or configuration setup for test environment
	// Check if .env file exists
	if _, err := os.Stat("../../.env"); os.IsNotExist(err) {
		log.Println("Loading data from the environment")
		log.Println(os.Getenv("DB_USER"))
	} else {
		// Load environment variables from .env file
		if err := godotenv.Load(); err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_NAME"))

	var err error
	tenantManagement.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	return tenantManagement.DB
}

// Add this helper function to your test file
func prettyPrintJSON(t *testing.T, rr *httptest.ResponseRecorder) {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, rr.Body.Bytes(), "", "    ")
	if err != nil {
		t.Logf("%sFailed to format JSON: %v%s", colorRed, err, colorReset)
		return
	}

	statusColor := colorGreen
	statusText := "SUCCESS"
	if rr.Code >= 400 {
		statusColor = colorRed
		statusText = "FAILED"
	}

	log.Printf("\nStatus: %s%s%s (%d)\nResponse Body:\n%s\n",
		statusColor,
		statusText,
		colorReset,
		rr.Code,
		prettyJSON.String())
}

func logTestName(name string) {
	log.Printf("\n%s=== RUN %s%s\n", colorYellow, name, colorReset)
}

// Helper functions to simplify test setup and execution
type TestRequest struct {
	Method      string
	Path        string
	Body        interface{}
	QueryParams map[string]string
}

func executeRequest(t *testing.T, req TestRequest) *httptest.ResponseRecorder {
	// Set up gin in test mode
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Set up JWT middleware
	jwt, err := middlewares.NewValidator()
	if err != nil {
		t.Fatalf("Failed to create JWT validation middleware: %v", err)
	}
	jwtMiddleware := jwt.ValidateJwt()

	// Register the route with the appropriate handler
	switch req.Method + req.Path {

	// Project Routes
	case "POST/project":
		r.POST("/project", jwtMiddleware, projects.NewProject)
	case "GET/api/project":
		r.GET("/api/project", jwtMiddleware, projects.GetProject)
	case "GET/api/projects":
		r.GET("/api/projects", jwtMiddleware, projects.GetProjects)
	case "PUT/project":
		r.PUT("/project", jwtMiddleware, projects.UpdateProject)
	case "DELETE/project":
		r.DELETE("/project", jwtMiddleware, projects.DeleteProject)

	// Document Routes
	case "POST/api/document":
		r.POST("/api/document", jwtMiddleware, projects.NewDocument)
	case "GET/api/documents":
		r.GET("/api/documents", jwtMiddleware, projects.GetDocuments)
	case "GET/api/document":
		r.GET("/api/document", jwtMiddleware, projects.GetDocument)
	case "PUT/api/document":
		r.PUT("/api/document", jwtMiddleware, projects.UpdateDocument)
	case "DELETE/api/document":
		r.DELETE("/api/document", jwtMiddleware, projects.DeleteDocument)
	case "POST/api/document/clone":
		r.POST("/api/document/clone", jwtMiddleware, projects.CloneDocument)

	// Conversation Routes
	case "POST/api/conversation":
		r.POST("/api/conversation", jwtMiddleware, projects.NewConversation)
	case "GET/api/conversation":
		r.GET("/api/conversation", jwtMiddleware, projects.GetConversation)
	case "GET/api/conversations":
		r.GET("/api/conversations", jwtMiddleware, projects.GetConversations)
	case "DELETE/api/conversation":
		r.DELETE("/api/conversation", jwtMiddleware, projects.DeleteConversation)

	// Diagram Routes
	case "POST/api/diagram":
		r.POST("/api/diagram", jwtMiddleware, projects.NewDiagram)
	case "GET/api/diagram":
		r.GET("/api/diagram", jwtMiddleware, projects.GetDiagram)
	case "GET/api/diagrams":
		r.GET("/api/diagrams", jwtMiddleware, projects.GetDiagrams)
	case "PUT/api/diagram":
		r.PUT("/api/diagram", jwtMiddleware, projects.UpdateDiagram)
	case "DELETE/api/diagram":
		r.DELETE("/api/diagram", jwtMiddleware, projects.DeleteDiagram)
	case "POST/api/diagram/clone":
		r.POST("/api/diagram/clone", jwtMiddleware, projects.CloneDiagram)

	// TBar Routes
	case "POST/api/tbar":
		r.POST("/api/tbar", jwtMiddleware, decisions.NewTBar)
	case "GET/api/tbar":
		r.GET("/api/tbar", jwtMiddleware, decisions.GetTBar)
	case "GET/api/tbars":
		r.GET("/api/tbars", jwtMiddleware, decisions.GetTBars)
	case "PUT/api/tbar":
		r.PUT("/api/tbar", jwtMiddleware, decisions.UpdateTBar)
	case "DELETE/api/tbar":
		r.DELETE("/api/tbar", jwtMiddleware, decisions.DeleteTBar)

	// TBar Argument Routes
	case "POST/api/tbar/argument":
		r.POST("/api/tbar/argument", jwtMiddleware, decisions.NewTBarArgument)
	case "GET/api/tbar/arguments":
		r.GET("/api/tbar/arguments", jwtMiddleware, decisions.GetTBarArguments)
	case "PUT/api/tbar/argument":
		r.PUT("/api/tbar/argument", jwtMiddleware, decisions.UpdateTBarArgument)
	case "DELETE/api/tbar/argument":
		r.DELETE("/api/tbar/argument", jwtMiddleware, decisions.DeleteTBarArgument)

	// Pnc Routes
	case "POST/api/pnc":
		r.POST("/api/pnc", jwtMiddleware, decisions.NewPncAnalysis)
	case "GET/api/pnc":
		r.GET("/api/pnc", jwtMiddleware, decisions.GetPncAnalysis)
	case "GET/api/pncs":
		r.GET("/api/pncs", jwtMiddleware, decisions.GetAllPncAnalysis)
	case "PUT/api/pnc":
		r.PUT("/api/pnc", jwtMiddleware, decisions.UpdatePncAnalysis)
	case "DELETE/api/pnc":
		r.DELETE("/api/pnc", jwtMiddleware, decisions.DeletePncAnalysis)

	// Pnc Argument Routes
	case "POST/api/pncArgument":
		r.POST("/api/pncArgument", jwtMiddleware, decisions.NewPncArgument)
	case "GET/api/pncArguments":
		r.GET("/api/pncArguments", jwtMiddleware, decisions.GetAllPncArguments)
	case "PUT/api/pncArgument":
		r.PUT("/api/pncArgument", jwtMiddleware, decisions.UpdatePncArgument)
	case "DELETE/api/pncArgument":
		r.DELETE("/api/pncArgument", jwtMiddleware, decisions.DeletePncArgument)

	// Swot Routes
	case "POST/api/swot":
		r.POST("/api/swot", jwtMiddleware, decisions.NewSwot)
	case "GET/api/swot":
		r.GET("/api/swot", jwtMiddleware, decisions.GetSwot)
	case "GET/api/swots":
		r.GET("/api/swots", jwtMiddleware, decisions.GetSwots)
	case "PUT/api/swot":
		r.PUT("/api/swot", jwtMiddleware, decisions.UpdateSwot)
	case "DELETE/api/swot":
		r.DELETE("/api/swot", jwtMiddleware, decisions.DeleteSwot)

	// Swot Argument Routes
	case "POST/api/swotArgument":
		r.POST("/api/swotArgument", jwtMiddleware, decisions.NewSwotArgument)
	case "GET/api/swotArguments":
		r.GET("/api/swotArguments", jwtMiddleware, decisions.GetAllSwotArguments)
	case "PUT/api/swotArgument":
		r.PUT("/api/swotArgument", jwtMiddleware, decisions.UpdateSwotArgument)
	case "DELETE/api/swotArgument":
		r.DELETE("/api/swotArgument", jwtMiddleware, decisions.DeleteSwotArgument)

	// Tenant AI Template Routes
	case "POST/api/tenantAiTemplate":
		r.POST("/api/tenantAiTemplate", jwtMiddleware, templates.NewTenantAiTemplate)
	case "GET/api/tenantAiTemplate":
		r.GET("/api/tenantAiTemplate", jwtMiddleware, templates.GetTenantAiTemplate)
	case "GET/api/tenantAiTemplates":
		r.GET("/api/tenantAiTemplates", jwtMiddleware, templates.GetTenantAiTemplates)
	case "PUT/api/tenantAiTemplate":
		r.PUT("/api/tenantAiTemplate", jwtMiddleware, templates.UpdateTenantAiTemplate)
	case "DELETE/api/tenantAiTemplate":
		r.DELETE("/api/tenantAiTemplate", jwtMiddleware, templates.DeleteTenantAiTemplate)
	}

	// Build the URL with query parameters
	url := req.Path
	if len(req.QueryParams) > 0 {
		params := make([]string, 0)
		for key, value := range req.QueryParams {
			params = append(params, fmt.Sprintf("%s=%s", key, value))
		}
		url = fmt.Sprintf("%s?%s", url, strings.Join(params, "&"))
	}

	// Create the request
	var httpReq *http.Request
	if req.Body != nil {
		bodyBytes, _ := json.Marshal(req.Body)
		httpReq, err = http.NewRequest(req.Method, url, bytes.NewBuffer(bodyBytes))
	} else {
		httpReq, err = http.NewRequest(req.Method, url, nil)
	}
	assert.NoError(t, err)

	// Set headers
	token := os.Getenv("TEST_JWT_TOKEN")
	httpReq.Header.Set("Authorization", "Bearer "+token)
	if req.Body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// Execute request
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, httpReq)

	// Pretty print response
	prettyPrintJSON(t, rr)
	return rr
}

func extractIDFromResponse(t *testing.T, rr *httptest.ResponseRecorder) string {
	var responseMap map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &responseMap)
	if err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}

	// Handle TBar specific response structure
	if data, ok := responseMap["data"].(map[string]interface{}); ok {
		if tbarAnalysis, ok := data["tbar_analysis"].(map[string]interface{}); ok {
			if id, ok := tbarAnalysis["id"].(string); ok {
				return id
			}
		}
		// Fall back to regular structure
		if id, ok := data["id"].(string); ok {
			return id
		}
		t.Fatalf("ID not found in response structure")
	}
	t.Fatalf("Response 'data' field not found or is not an object")
	return ""
}

func TestPostProject(t *testing.T) {
	logTestName("TestPostProject")

	projectData := map[string]interface{}{
		"title":       "New Test Project",
		"status":      "In Progress",
		"category":    "Personal",
		"complexity":  "complex",
		"description": "This project describes an entire Entra as internal service design.",
	}

	rr := executeRequest(t, TestRequest{
		Method: "POST",
		Path:   "/project",
		Body:   projectData,
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	createdProjectID = extractIDFromResponse(t, rr)
	log.Printf("Created Project ID: %s", createdProjectID)
	recordFailure(t, "TestPostProject")
}

func TestGetProject(t *testing.T) {
	logTestName("TestGetProject")

	rr := executeRequest(t, TestRequest{
		Method: "GET",
		Path:   "/api/project",
		QueryParams: map[string]string{
			"id": createdProjectID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestGetProject")
}

func TestGetProjects(t *testing.T) {
	logTestName("TestGetProjects")

	rr := executeRequest(t, TestRequest{
		Method: "GET",
		Path:   "/api/projects",
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestGetProjects")
}

func TestPutProject(t *testing.T) {
	logTestName("TestUpdateProject")

	updateProjectData := map[string]interface{}{
		"title":         "Hanlers Test Project 2 Updated",
		"status":        "In Progress",
		"category":      "Technology",
		"project_order": 1,
		"complexity":    "complex",
		"description":   "This project describes an entire Entra as internal service design.",
	}

	rr := executeRequest(t, TestRequest{
		Method: "PUT",
		Path:   "/project",
		Body:   updateProjectData,
		QueryParams: map[string]string{
			"id": createdProjectID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestPutProject")
}

func TestPostDocument(t *testing.T) {
	logTestName("TestPostDocument")
	log.Printf("Using Project ID: %s", createdProjectID)

	documentData := map[string]interface{}{
		"title":          "Azure DataPower Test",
		"content":        "<h1>This is a test document</h1>",
		"complexity":     "complex",
		"document_order": 1,
	}

	rr := executeRequest(t, TestRequest{
		Method: "POST",
		Path:   "/api/document",
		Body:   documentData,
		QueryParams: map[string]string{
			"project_id": createdProjectID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	createdDocumentID = extractIDFromResponse(t, rr)
	log.Printf("Created Document ID: %s", createdDocumentID)
	recordFailure(t, "TestPostDocument")
}

func TestGetDocuments(t *testing.T) {
	logTestName("TestGetDocuments")

	rr := executeRequest(t, TestRequest{
		Method: "GET",
		Path:   "/api/documents",
		QueryParams: map[string]string{
			"project_id": createdProjectID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestGetDocuments")
}

func TestGetDocument(t *testing.T) {
	logTestName("TestGetDocument")

	rr := executeRequest(t, TestRequest{
		Method: "GET",
		Path:   "/api/document",
		QueryParams: map[string]string{
			"project_id":  createdProjectID,
			"document_id": createdDocumentID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestGetDocument")
}

func TestPutDocument(t *testing.T) {
	logTestName("TestPutDocument")

	updateData := map[string]interface{}{
		"title":          "Azure DataPower Test Updated",
		"content":        "<h1>This is a test document with updated content</h1>",
		"complexity":     "complex",
		"document_order": 1,
	}

	rr := executeRequest(t, TestRequest{
		Method: "PUT",
		Path:   "/api/document",
		Body:   updateData,
		QueryParams: map[string]string{
			"project_id":  createdProjectID,
			"document_id": createdDocumentID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestPutDocument")
}

func TestDeleteDocument(t *testing.T) {
	logTestName("TestDeleteDocument")

	rr := executeRequest(t, TestRequest{
		Method: "DELETE",
		Path:   "/api/document",
		QueryParams: map[string]string{
			"project_id":  createdProjectID,
			"document_id": createdDocumentID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestDeleteDocument")
}

func TestPostConversation(t *testing.T) {
	logTestName("TestPostConversation")

	conversationData := map[string]interface{}{
		"title":             "",
		"conversation_type": "chat",
		"description":       "",
	}

	rr := executeRequest(t, TestRequest{
		Method: "POST",
		Path:   "/api/conversation",
		Body:   conversationData,
		QueryParams: map[string]string{
			"project_id":                    createdProjectID,
			"conversation_configuration_id": "",
			"agent_name":                    "",
			"template_type":                 "community",
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	createdConversationID = extractIDFromResponse(t, rr)
	log.Printf("Created Conversation ID: %s", createdConversationID)
	recordFailure(t, "TestPostConversation")
}

// TODO: TestGetConversation is failing because of PSQL issues.
// https://dev.azure.com/rautenbergsoftware/Solution%20Pilot/_build/results?buildId=1853&view=logs&j=6e95d870-caf2-5e72-1a9b-4c4e9f0af78a&t=ae85350b-fed3-558f-7a38-d258170f064f&l=393
func TestGetConversation(t *testing.T) {
	logTestName("TestGetConversation")

	rr := executeRequest(t, TestRequest{
		Method: "GET",
		Path:   "/api/conversation",
		QueryParams: map[string]string{
			"project_id":      createdProjectID,
			"conversation_id": createdConversationID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestGetConversation")
}

func TestGetConversations(t *testing.T) {
	logTestName("TestGetConversations")

	rr := executeRequest(t, TestRequest{
		Method: "GET",
		Path:   "/api/conversations",
		QueryParams: map[string]string{
			"project_id": createdProjectID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestGetConversations")
}

func TestDeleteConversation(t *testing.T) {
	logTestName("TestDeleteConversation")

	rr := executeRequest(t, TestRequest{
		Method: "DELETE",
		Path:   "/api/conversation",
		QueryParams: map[string]string{
			"project_id":      createdProjectID,
			"conversation_id": createdConversationID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestDeleteConversation")
}

func TestPostDiagram(t *testing.T) {
	logTestName("TestPostDiagram")
	log.Printf("Using Project ID: %s", createdProjectID)

	diagramData := map[string]interface{}{
		"title":          "Test Architecture Diagram",
		"diagram_type":   "architecture",
		"diagram_status": "draft",
		"category":       "system",
		"design":         "{\"nodes\":[],\"edges\":[]}",
	}

	rr := executeRequest(t, TestRequest{
		Method: "POST",
		Path:   "/api/diagram",
		Body:   diagramData,
		QueryParams: map[string]string{
			"project_id": createdProjectID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	createdDiagramID = extractIDFromResponse(t, rr)
	log.Printf("Created Diagram ID: %s", createdDiagramID)
	recordFailure(t, "TestPostDiagram")
}

// TODO: We are not using get diagram ? How is this happening?
func TestGetDiagram(t *testing.T) {
	logTestName("TestGetDiagram")
	log.Printf("Getting diagram with Project ID: %s and Diagram ID: %s", createdProjectID, createdDiagramID)

	rr := executeRequest(t, TestRequest{
		Method: "GET",
		Path:   "/api/diagram",
		QueryParams: map[string]string{
			"project_id": createdProjectID,
			"diagram_id": createdDiagramID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestGetDiagram")
}

func TestGetDiagrams(t *testing.T) {
	logTestName("TestGetDiagrams")
	log.Printf("Getting diagrams for Project ID: %s", createdProjectID)

	rr := executeRequest(t, TestRequest{
		Method: "GET",
		Path:   "/api/diagrams",
		QueryParams: map[string]string{
			"project_id": createdProjectID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestGetDiagrams")
}

func TestUpdateDiagram(t *testing.T) {
	logTestName("TestUpdateDiagram")
	log.Printf("Updating diagram with Project ID: %s and Diagram ID: %s", createdProjectID, createdDiagramID)

	updateData := map[string]interface{}{
		"title":          "New Diagram Name",
		"diagram_type":   "Security",
		"diagram_status": "Approved",
		"design":         "{\"nodes\":[{\"id\":\"node_6731.032420065595\",\"type\":\"node\",\"position\":{\"x\":167,\"y\":452},\"selected\":false,\"data\":{\"label\":\"App Services\",\"icon_url\":\"https://spnodedata.blob.core.windows.net/nodes/azure_clean/web/app-services.svg\",\"text_alignments\":\"\",\"border_color\":\"\",\"border_style\":\"\",\"color\":\"\",\"background\":\"\"},\"style\":{\"zIndex\":1000},\"width\":120,\"measured\":{\"width\":120,\"height\":132}}],\"edges\":[]}",
	}

	rr := executeRequest(t, TestRequest{
		Method: "PUT",
		Path:   "/api/diagram",
		Body:   updateData,
		QueryParams: map[string]string{
			"project_id": createdProjectID,
			"diagram_id": createdDiagramID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestUpdateDiagram")
}

func TestDeleteDiagram(t *testing.T) {
	logTestName("TestDeleteDiagram")
	log.Printf("Deleting diagram with Project ID: %s and Diagram ID: %s", createdProjectID, createdDiagramID)

	rr := executeRequest(t, TestRequest{
		Method: "DELETE",
		Path:   "/api/diagram",
		QueryParams: map[string]string{
			"project_id": createdProjectID,
			"diagram_id": createdDiagramID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestDeleteDiagram")
}

func TestPostTBar(t *testing.T) {
	logTestName("TestPostTBar")
	log.Printf("Using Project ID: %s", createdProjectID)

	tbarData := map[string]interface{}{
		"tbar_title":       "Test TBar Analysis",
		"tbar_description": "Comparing two architectural approaches",
		"tbar_status":      "In Progress",
		"tbar_category":    "Architecture",
		"option_a":         "Microservices Architecture",
		"option_b":         "Monolithic Architecture",
		"assumptions":      "Both approaches must handle high scalability",
		"final_decision":   "",
		"implications":     "The choice will affect deployment and maintenance",
	}

	rr := executeRequest(t, TestRequest{
		Method: "POST",
		Path:   "/api/tbar",
		Body:   tbarData,
		QueryParams: map[string]string{
			"project_id": createdProjectID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	createdTBarID = extractIDFromResponse(t, rr)
	log.Printf("Created TBar ID: %s", createdTBarID)
	recordFailure(t, "TestPostTBar")

	// Extract option IDs from response
	var responseMap map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &responseMap)
	if err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}

	if data, ok := responseMap["data"].(map[string]interface{}); ok {
		if tbarAnalysis, ok := data["tbar_analysis"].(map[string]interface{}); ok {
			if options, ok := tbarAnalysis["options"].([]interface{}); ok && len(options) == 2 {
				optionA := options[0].(map[string]interface{})
				optionB := options[1].(map[string]interface{})
				createdTBarOptionAID = optionA["option_id"].(string)
				createdTBarOptionBID = optionB["option_id"].(string)
				log.Printf("Created TBar Option A ID: %s", createdTBarOptionAID)
				log.Printf("Created TBar Option B ID: %s", createdTBarOptionBID)
			}
		}
	}
}

func TestGetTBar(t *testing.T) {
	logTestName("TestGetTBar")
	log.Printf("Getting TBar with Project ID: %s and TBar ID: %s", createdProjectID, createdTBarID)

	rr := executeRequest(t, TestRequest{
		Method: "GET",
		Path:   "/api/tbar",
		QueryParams: map[string]string{
			"project_id": createdProjectID,
			"tbar_id":    createdTBarID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestGetTBar")
}

func TestGetTBars(t *testing.T) {
	logTestName("TestGetTBars")
	log.Printf("Getting TBars for Project ID: %s", createdProjectID)

	rr := executeRequest(t, TestRequest{
		Method: "GET",
		Path:   "/api/tbars",
		QueryParams: map[string]string{
			"project_id": createdProjectID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestGetTBars")
}

func TestPostTBarArgument(t *testing.T) {
	logTestName("TestPostTBarArgument")
	log.Printf("Creating argument for TBar Option ID: %s", createdTBarOptionAID)

	argumentData := map[string]interface{}{
		"argument_name":   "High Scalability",
		"argument_weight": 5,
		"description":     "Easier to scale individual components",
	}

	rr := executeRequest(t, TestRequest{
		Method: "POST",
		Path:   "/api/tbar/argument",
		Body:   argumentData,
		QueryParams: map[string]string{
			"option_id": createdTBarOptionAID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)

	// Extract argument ID from response
	var responseMap map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &responseMap)
	if err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}

	if data, ok := responseMap["data"].(map[string]interface{}); ok {
		createdTBarArgumentID = data["id"].(string)
		log.Printf("Created TBar Argument ID: %s", createdTBarArgumentID)
	}
	recordFailure(t, "TestPostTBarArgument")
}

func TestGetTBarArguments(t *testing.T) {
	logTestName("TestGetTBarArguments")
	log.Printf("Getting arguments for TBar Option ID: %s", createdTBarOptionAID)

	rr := executeRequest(t, TestRequest{
		Method: "GET",
		Path:   "/api/tbar/arguments",
		QueryParams: map[string]string{
			"option_id": createdTBarOptionAID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestGetTBarArguments")
}

func TestUpdateTBarArgument(t *testing.T) {
	logTestName("TestUpdateTBarArgument")
	log.Printf("Updating argument ID: %s for TBar Option ID: %s", createdTBarArgumentID, createdTBarOptionAID)

	updateData := map[string]interface{}{
		"argument_name":   "Excellent Scalability",
		"argument_weight": 8,
		"description":     "Much easier to scale individual components independently",
	}

	rr := executeRequest(t, TestRequest{
		Method: "PUT",
		Path:   "/api/tbar/argument",
		Body:   updateData,
		QueryParams: map[string]string{
			"option_id":   createdTBarOptionAID,
			"argument_id": createdTBarArgumentID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestUpdateTBarArgument")
}

func TestUpdateTBar(t *testing.T) {
	logTestName("TestUpdateTBar")
	log.Printf("Updating TBar ID: %s", createdTBarID)

	updateData := map[string]interface{}{
		"tbar_title":         "Updated TBar Analysis",
		"tbar_description":   "Updated comparison of architectural approaches",
		"tbar_status":        "Completed",
		"tbar_category":      "Architecture",
		"tbar_better_option": "Option A",
		"final_decision":     "Selected Microservices Architecture",
		"implications":       "Will require more DevOps expertise",
	}

	rr := executeRequest(t, TestRequest{
		Method: "PUT",
		Path:   "/api/tbar",
		Body:   updateData,
		QueryParams: map[string]string{
			"project_id": createdProjectID,
			"tbar_id":    createdTBarID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestUpdateTBar")
}

func TestDeleteTBarArgument(t *testing.T) {
	logTestName("TestDeleteTBarArgument")
	log.Printf("Deleting argument ID: %s for TBar Option ID: %s", createdTBarArgumentID, createdTBarOptionAID)

	rr := executeRequest(t, TestRequest{
		Method: "DELETE",
		Path:   "/api/tbar/argument",
		QueryParams: map[string]string{
			"option_id":   createdTBarOptionAID,
			"argument_id": createdTBarArgumentID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestDeleteTBarArgument")
}

func TestDeleteTBar(t *testing.T) {
	logTestName("TestDeleteTBar")
	log.Printf("Deleting TBar ID: %s", createdTBarID)

	rr := executeRequest(t, TestRequest{
		Method: "DELETE",
		Path:   "/api/tbar",
		QueryParams: map[string]string{
			"project_id": createdProjectID,
			"tbar_id":    createdTBarID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestDeleteTBar")
}

func TestPostPnc(t *testing.T) {
	logTestName("TestPostPnc")
	log.Printf("Using Project ID: %s", createdProjectID)

	pncData := map[string]interface{}{
		"title":                     "Test Pros & Cons Analysis",
		"pnc_description":           "Analyzing deployment options",
		"pnc_status":                "In Progress",
		"category":                  "Technology",
		"better_option":             nil,
		"assumptions":               "Both options must support high availability",
		"final_decision":            "",
		"architectural_decision_id": nil,
		"implications":              "The choice will affect operational costs",
	}

	rr := executeRequest(t, TestRequest{
		Method: "POST",
		Path:   "/api/pnc",
		Body:   pncData,
		QueryParams: map[string]string{
			"project_id": createdProjectID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	createdPncID = extractIDFromResponse(t, rr)
	log.Printf("Created PNC ID: %s", createdPncID)
	recordFailure(t, "TestPostPnc")
}

func TestGetPnc(t *testing.T) {
	logTestName("TestGetPnc")
	log.Printf("Getting PNC with Project ID: %s and PNC ID: %s", createdProjectID, createdPncID)

	rr := executeRequest(t, TestRequest{
		Method: "GET",
		Path:   "/api/pnc",
		QueryParams: map[string]string{
			"project_id": createdProjectID,
			"pnc_id":     createdPncID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestGetPnc")
}

func TestGetPncs(t *testing.T) {
	logTestName("TestGetPncs")
	log.Printf("Getting PNCs for Project ID: %s", createdProjectID)

	rr := executeRequest(t, TestRequest{
		Method: "GET",
		Path:   "/api/pncs",
		QueryParams: map[string]string{
			"project_id": createdProjectID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestGetPncs")
}

func TestPostPncArgument(t *testing.T) {
	logTestName("TestPostPncArgument")
	log.Printf("Creating argument for PNC ID: %s", createdPncID)

	argumentData := map[string]interface{}{
		"argument":        "Cost Effective",
		"argument_weight": 4,
		"side":            "pro",
		"description":     "This option has lower operational costs",
	}

	rr := executeRequest(t, TestRequest{
		Method: "POST",
		Path:   "/api/pncArgument",
		Body:   argumentData,
		QueryParams: map[string]string{
			"pnc_id": createdPncID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	createdPncArgumentID = extractIDFromResponse(t, rr)
	log.Printf("Created PNC Argument ID: %s", createdPncArgumentID)
	recordFailure(t, "TestPostPncArgument")
}

func TestGetPncArguments(t *testing.T) {
	logTestName("TestGetPncArguments")
	log.Printf("Getting arguments for PNC ID: %s", createdPncID)

	rr := executeRequest(t, TestRequest{
		Method: "GET",
		Path:   "/api/pncArguments",
		QueryParams: map[string]string{
			"pnc_id": createdPncID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestGetPncArguments")
}

func TestUpdatePncArgument(t *testing.T) {
	logTestName("TestUpdatePncArgument")
	log.Printf("Updating argument ID: %s for PNC ID: %s", createdPncArgumentID, createdPncID)

	updateData := map[string]interface{}{
		"argument":        "Very Cost Effective",
		"argument_weight": 5,
		"description":     "This option has significantly lower operational costs",
	}

	rr := executeRequest(t, TestRequest{
		Method: "PUT",
		Path:   "/api/pncArgument",
		Body:   updateData,
		QueryParams: map[string]string{
			"pnc_id":      createdPncID,
			"argument_id": createdPncArgumentID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestUpdatePncArgument")
}

func TestUpdatePnc(t *testing.T) {
	logTestName("TestUpdatePnc")
	log.Printf("Updating PNC ID: %s", createdPncID)

	updateData := map[string]interface{}{
		"title":           "Updated Pros & Cons Analysis",
		"pnc_description": "Updated analysis of deployment options",
		"pnc_status":      "Completed",
		"category":        "Technology",
		"better_option":   "Option A",
		"final_decision":  "Selected cloud deployment",
		"implications":    "Will require cloud expertise",
	}

	rr := executeRequest(t, TestRequest{
		Method: "PUT",
		Path:   "/api/pnc",
		Body:   updateData,
		QueryParams: map[string]string{
			"project_id": createdProjectID,
			"pnc_id":     createdPncID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestUpdatePnc")
}

func TestDeletePncArgument(t *testing.T) {
	logTestName("TestDeletePncArgument")
	log.Printf("Deleting argument ID: %s for PNC ID: %s", createdPncArgumentID, createdPncID)

	rr := executeRequest(t, TestRequest{
		Method: "DELETE",
		Path:   "/api/pncArgument",
		QueryParams: map[string]string{
			"argument_id": createdPncArgumentID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestDeletePncArgument")
}

func TestDeletePnc(t *testing.T) {
	logTestName("TestDeletePnc")
	log.Printf("Deleting PNC ID: %s", createdPncID)

	rr := executeRequest(t, TestRequest{
		Method: "DELETE",
		Path:   "/api/pnc",
		QueryParams: map[string]string{
			"project_id": createdProjectID,
			"pnc_id":     createdPncID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestDeletePnc")
}

func TestPostSwot(t *testing.T) {
	logTestName("TestPostSwot")
	log.Printf("Using Project ID: %s", createdProjectID)

	swotData := map[string]interface{}{
		"title":                     "Test SWOT Analysis",
		"swot_description":          "Analyzing cloud architecture approach",
		"swot_status":               "In Progress",
		"category":                  "Technology",
		"assumptions":               "Cloud infrastructure must be scalable",
		"final_decision":            "",
		"architectural_decision_id": nil,
		"implications":              "Will affect our deployment strategy",
	}

	rr := executeRequest(t, TestRequest{
		Method: "POST",
		Path:   "/api/swot",
		Body:   swotData,
		QueryParams: map[string]string{
			"project_id": createdProjectID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	createdSwotID = extractIDFromResponse(t, rr)
	log.Printf("Created SWOT ID: %s", createdSwotID)
	recordFailure(t, "TestPostSwot")
}

func TestGetSwot(t *testing.T) {
	logTestName("TestGetSwot")
	log.Printf("Getting SWOT with Project ID: %s and SWOT ID: %s", createdProjectID, createdSwotID)

	rr := executeRequest(t, TestRequest{
		Method: "GET",
		Path:   "/api/swot",
		QueryParams: map[string]string{
			"project_id": createdProjectID,
			"swot_id":    createdSwotID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestGetSwot")
}

func TestGetSwots(t *testing.T) {
	logTestName("TestGetSwots")
	log.Printf("Getting SWOTs for Project ID: %s", createdProjectID)

	rr := executeRequest(t, TestRequest{
		Method: "GET",
		Path:   "/api/swots",
		QueryParams: map[string]string{
			"project_id": createdProjectID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestGetSwots")
}

func TestPostSwotArgument(t *testing.T) {
	logTestName("TestPostSwotArgument")
	log.Printf("Creating argument for SWOT ID: %s", createdSwotID)

	argumentData := map[string]interface{}{
		"argument":        "Strong Cloud Expertise",
		"argument_weight": 5,
		"side":            "strength",
		"description":     "Team has extensive experience with cloud platforms",
	}

	rr := executeRequest(t, TestRequest{
		Method: "POST",
		Path:   "/api/swotArgument",
		Body:   argumentData,
		QueryParams: map[string]string{
			"swot_id": createdSwotID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	createdSwotArgumentID = extractIDFromResponse(t, rr)
	log.Printf("Created SWOT Argument ID: %s", createdSwotArgumentID)
	recordFailure(t, "TestPostSwotArgument")
}

func TestGetSwotArguments(t *testing.T) {
	logTestName("TestGetSwotArguments")
	log.Printf("Getting arguments for SWOT ID: %s", createdSwotID)

	rr := executeRequest(t, TestRequest{
		Method: "GET",
		Path:   "/api/swotArguments",
		QueryParams: map[string]string{
			"swot_id": createdSwotID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestGetSwotArguments")
}

func TestUpdateSwotArgument(t *testing.T) {
	logTestName("TestUpdateSwotArgument")
	log.Printf("Updating argument ID: %s for SWOT ID: %s", createdSwotArgumentID, createdSwotID)

	updateData := map[string]interface{}{
		"argument":        "Exceptional Cloud Expertise",
		"argument_weight": 8,
		"side":            "strength",
		"description":     "Team has extensive experience with multiple cloud platforms",
	}

	rr := executeRequest(t, TestRequest{
		Method: "PUT",
		Path:   "/api/swotArgument",
		Body:   updateData,
		QueryParams: map[string]string{
			"swot_id":     createdSwotID,
			"argument_id": createdSwotArgumentID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestUpdateSwotArgument")
}

func TestUpdateSwot(t *testing.T) {
	logTestName("TestUpdateSwot")
	log.Printf("Updating SWOT ID: %s", createdSwotID)

	updateData := map[string]interface{}{
		"title":            "Updated SWOT Analysis",
		"swot_description": "Updated cloud architecture analysis",
		"swot_status":      "Completed",
		"category":         "Technology",
		"assumptions":      "Cloud infrastructure must be highly scalable",
		"final_decision":   "Proceed with cloud deployment",
		"implications":     "Will require additional cloud training",
	}

	rr := executeRequest(t, TestRequest{
		Method: "PUT",
		Path:   "/api/swot",
		Body:   updateData,
		QueryParams: map[string]string{
			"project_id": createdProjectID,
			"swot_id":    createdSwotID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestUpdateSwot")
}

func TestDeleteSwotArgument(t *testing.T) {
	logTestName("TestDeleteSwotArgument")
	log.Printf("Deleting argument ID: %s for SWOT ID: %s", createdSwotArgumentID, createdSwotID)

	rr := executeRequest(t, TestRequest{
		Method: "DELETE",
		Path:   "/api/swotArgument",
		QueryParams: map[string]string{
			"swot_id":     createdSwotID,
			"argument_id": createdSwotArgumentID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestDeleteSwotArgument")
}

func TestDeleteSwot(t *testing.T) {
	logTestName("TestDeleteSwot")
	log.Printf("Deleting SWOT ID: %s", createdSwotID)

	rr := executeRequest(t, TestRequest{
		Method: "DELETE",
		Path:   "/api/swot",
		QueryParams: map[string]string{
			"project_id": createdProjectID,
			"swot_id":    createdSwotID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestDeleteSwot")
}

func TestDeleteProject(t *testing.T) {
	logTestName("TestDeleteProject")

	rr := executeRequest(t, TestRequest{
		Method: "DELETE",
		Path:   "/project",
		QueryParams: map[string]string{
			"id": createdProjectID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestDeleteProject")
}

func TestPostAiTemplate(t *testing.T) {
	logTestName("TestPostAiTemplate")

	assistantData := map[string]interface{}{
		"title":       "This is assistant title",
		"description": "assistant description",
		"category":    "GCP Architecture",
		"ai_model":    "Open AI GPT-4",
		"ai_vendor":   "Azure",
		"configuration": map[string]interface{}{
			"ai_temperature": 1,
			"prompt_role":    "system",
			"system_config":  "This is system prompt",
			"top_p":          1,
			"max_tokens":     2400,
		},
		"privacy":      true,
		"published_by": "userTemplate",
	}

	rr := executeRequest(t, TestRequest{
		Method: "POST",
		Path:   "/api/tenantAiTemplate",
		Body:   assistantData,
		// Add QueryParams here if you need to pass any, e.g.:
		// QueryParams: map[string]string{
		//     "some_id": createdAssistantID,
		// },
	})

	createdAssistantID = extractIDFromResponse(t, rr)
	log.Printf("Created Assistant ID: %s", createdAssistantID)

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestPostAssistant")
}

func TestGetAiAssistant(t *testing.T) {
	logTestName("TestAiGetAssistant")

	rr := executeRequest(t, TestRequest{
		Method: "GET",
		Path:   "/api/tenantAiTemplate",
		QueryParams: map[string]string{
			"id": createdAssistantID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestGetAssistant")
}

func TestGetAiAssistants(t *testing.T) {
	logTestName("TestGetAiAssistants")

	rr := executeRequest(t, TestRequest{
		Method: "GET",
		Path:   "/api/tenantAiTemplates",
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestGetAssistants")
}

func TestPutAiAssistant(t *testing.T) {
	logTestName("TestPutAiAssistant")

	updateData := map[string]interface{}{
		"title":       "Updated Assistant Title",
		"description": "Updated assistant description",
		"category":    "Updated Category",
		"ai_model":    "Open AI GPT-4",
		"ai_vendor":   "Azure",
		"configuration": map[string]interface{}{
			"ai_temperature": 0.5,
			"prompt_role":    "system",
			"system_config":  "You are a helpful assistant",
			"top_p":          0.95,
			"max_tokens":     4048,
		},
		"privacy":      true,
		"published_by": "unitTest",
	}

	rr := executeRequest(t, TestRequest{
		Method: "PUT",
		Path:   "/api/tenantAiTemplate",
		Body:   updateData,
		QueryParams: map[string]string{
			"id": createdAssistantID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	recordFailure(t, "TestPutAssistant")
}

func TestCloneDocument(t *testing.T) {
	logTestName("TestCloneDocument")
	log.Printf("Cloning document with Project ID: %s and Document ID: %s", createdProjectID, createdDocumentID)

	cloneData := map[string]interface{}{
		"title":         "Cloned Document Title",
		"complexity":    "medium-complexity",
		"document_type": "doctype-security",
	}

	rr := executeRequest(t, TestRequest{
		Method: "POST",
		Path:   "/api/document/clone",
		Body:   cloneData,
		QueryParams: map[string]string{
			"project_id":  createdProjectID,
			"document_id": createdDocumentID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	// Optionally extract and store the cloned document ID if needed
	log.Printf("Cloned Document Response: %s", rr.Body.String())
	recordFailure(t, "TestCloneDocument")
}

func TestCloneDiagram(t *testing.T) {
	logTestName("TestCloneDiagram")
	log.Printf("Cloning diagram with Project ID: %s and Diagram ID: %s", createdProjectID, createdDiagramID)

	cloneData := map[string]interface{}{
		"title":             "Cloned Diagram Title",
		"diagram_type":      "Network",
		"diagram_status":    "in-progress",
		"category":          "Security",
		"short_description": "Cloned short description",
	}

	rr := executeRequest(t, TestRequest{
		Method: "POST",
		Path:   "/api/diagram/clone",
		Body:   cloneData,
		QueryParams: map[string]string{
			"project_id": createdProjectID,
			"diagram_id": createdDiagramID,
		},
	})

	assert.Equal(t, http.StatusOK, rr.Code)
	// Optionally extract and store the cloned diagram ID if needed
	log.Printf("Cloned Diagram Response: %s", rr.Body.String())
	recordFailure(t, "TestCloneDiagram")
}

func recordFailure(t *testing.T, testName string) {
	if t.Failed() {
		failedTests = append(failedTests, testName)
	}
}

// ============================================================================
// TODO: The following endpoints are NOT covered by tests in this file.
//       Use these as a checklist for expanding your test coverage.
// ============================================================================

// -------------------- Project/Document Template Creation --------------------
/*
POST   /api/projectFromTemplate
POST   /api/pub/projectFromPubTemplate
POST   /api/documentTemplate
POST   /api/pdt/docPubTemplate
POST   /api/documentFromTemplate
POST   /api/pub/documentFromPubTemplate
GET    /api/projectEntities
*/

// -------------------- Decision Matrix Endpoints --------------------
/*
POST   /api/matrix
GET    /api/matrix
GET    /api/matrixs
PUT    /api/matrix
DELETE /api/matrix

// Matrix Criteria
POST   /api/matrixCriteria
PUT    /api/matrixCriteria
DELETE /api/matrixCriteria
GET    /api/matrixCriterias
GET    /api/matrixCriteria

// Matrix Concepts
POST   /api/matrixConcept
GET    /api/matrixConcept
GET    /api/matrixConcepts
PUT    /api/matrixConcept
DELETE /api/matrixConcept

// Matrix User Rating
PUT    /api/matrixUserRating
*/

// -------------------- AI Template Publishing/Cloning --------------------
/*
PUT    /api/tenantAiTemplate/publish
PUT    /api/tenantAiTemplate/unpublish
POST   /api/tenantAiTemplate/clone

// Solution Pilot AI Templates
GET    /api/spAiTemplate
GET    /api/spAiTemplates

// Community AI Templates
GET    /api/publicTemplate
GET    /api/publicTemplates
GET    /api/pspAiTemplate
GET    /api/pspAiTemplates
GET    /api/ppublicTemplate
GET    /api/ppublicTemplates
*/

// -------------------- Project Template APIs --------------------
/*
POST   /api/projectTemplate
GET    /api/projectTemplate
GET    /api/projectTemplates
PUT    /api/projectTemplate
DELETE /api/projectTemplate
*/

// -------------------- Internal Document Template APIs --------------------
/*
POST   /api/idt/documentTemplate
GET    /api/idt/documentTemplate
GET    /api/idt/documentTemplates
PUT    /api/idt/documentTemplate
DELETE /api/idt/documentTemplate
*/

// -------------------- Community Document Template APIs --------------------
/*
POST   /api/publicDocumentTemplate
GET    /api/publicDocumentTemplate
GET    /api/publicDocumentTemplates
PUT    /api/publicDocumentTemplate
DELETE /api/publicDocumentTemplate
*/

// -------------------- Internal Diagram Template APIs --------------------
/*
POST   /api/idt/diagramTemplate
GET    /api/idt/diagramTemplate
GET    /api/idt/diagramTemplates
PUT    /api/idt/diagramTemplate
DELETE /api/idt/diagramTemplate
*/

// -------------------- Community Diagram Template APIs --------------------
/*
POST   /api/publicDiagramTemplate
GET    /api/publicDiagramTemplate
GET    /api/publicDiagramTemplates
PUT    /api/publicDiagramTemplate
DELETE /api/publicDiagramTemplate
*/

// -------------------- Document Component APIs --------------------
/*
GET    /api/dcm/component
GET    /api/dcm/components
GET    /api/dcm/favoriteComponents
POST   /api/dcm/pinComponent
POST   /api/dcm/unpinComponent
*/

// -------------------- Public Template Publishing --------------------
/*
POST   /api/publishProjectTemplate
POST   /api/unpublishProjectTemplate
*/

// -------------------- Public Template Public Routers --------------------
/*
GET    /api/publicProjectTemplate
GET    /api/publicProjectTemplates
GET    /api/publicProjectDocumentTemplate
GET    /api/publicProjectDiagramTemplate
PUT    /api/publicProjectTemplate
GET    /api/pub/publicProjectTemplate
GET    /api/pub/publicProjectTemplates
GET    /api/pub/publicProjectTemplatesPag
GET    /api/pub/publicProjectDocumentTemplate
GET    /api/pub/publicProjectDiagramTemplate
POST   /api/clonePublicProjectTemplate
*/

// -------------------- Public Comments --------------------
/*
POST   /api/publicComment
GET    /api/publicComments
PUT    /api/publicComment
DELETE /api/publicComment
*/

// -------------------- Cloud Resource Management --------------------
/*
POST   /api/cloud/resources
GET    /api/cloud/credentials
POST   /api/cloud/credentials
DELETE /api/cloud/credentials
*/

// -------------------- Tiptap Collaboration --------------------
/*
GET    /api/tiptap/collab
*/

// -------------------- Tenant/User Management --------------------
/*
GET    /api/user
PATCH  /api/user
POST   /api/m/NewAzOiAccount
POST   /api/paymentSession
GET    /api/tenant
PUT    /api/tenant
PUT    /api/tenant/members/:member_id
DELETE /api/tenant/members/:member_id
GET    /api/tenant/invitations
POST   /api/tenant/invitations
DELETE /api/tenant/invitations/:invitation_id
*/

// ============================================================================
// End of missing route checklist
// ============================================================================
