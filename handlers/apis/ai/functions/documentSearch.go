package aiFunctions

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"sententiawebapi/handlers/apis/tenantManagement"
	"sententiawebapi/handlers/models"
	"sententiawebapi/utilities"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type DocSearchRequest struct {
	Query string   `json:"query" binding:"required"`
	Limit int      `json:"limit"`
	Scope []string `json:"scope"`
}

type DocSearchResult struct {
	Title    string  `json:"title"`
	Content  string  `json:"content"`
	Distance float64 `json:"distance"`
}

func selectDocSearchSource(
	req *DocSearchRequest,
	vectorStr string,
	chatCtx *models.ChatContext,
) (*sql.Rows, error) {
	var query string
	var args []any

	if chatCtx.ResourceGroupID == nil {
		return nil, fmt.Errorf("cannot perform semantic search without knowing project/template id")
	}

	// Replace "current" with actual document ID, or remove if not applicable
	scope := make([]string, 0, len(req.Scope))
	for _, id := range req.Scope {
		if id == "current" {
			if chatCtx.ResourceType != nil && *chatCtx.ResourceType == models.ResourceTypeDocument && chatCtx.ResourceID != nil {
				scope = append(scope, *chatCtx.ResourceID)
			}
		} else {
			scope = append(scope, id)
		}
	}

	switch chatCtx.ResourceGroupType {
	case models.ResourceGroupTemplate:
		query = `
			SELECT vector.content, vector.embedding <#> ` + vectorStr + `::vector AS distance, doc.title
			FROM st_schema.document_template_vectors vector
			JOIN st_schema.document_templates doc ON doc.id = vector.document_template_id
			WHERE vector.tenant_id = $1 AND vector.project_template_id = $2
		`
		args = append(args, chatCtx.TenantID, chatCtx.ResourceGroupID)

		if len(scope) > 0 {
			query += ` AND vector.document_template_id = ANY($3)`
			args = append(args, pq.Array(scope))
		}

	case models.ResourceGroupCommunity:
		query = `
			SELECT vector.content, vector.embedding <#> ` + vectorStr + `::vector AS distance, doc.title
			FROM st_schema.cm_document_template_vectors vector
			JOIN st_schema.cm_document_templates doc ON doc.id = vector.cm_document_template_id
			WHERE vector.tenant_id = $1 AND vector.cm_project_template_id = $2
		`
		args = append(args, chatCtx.TenantID, chatCtx.ResourceGroupID)

		if len(scope) > 0 {
			query += ` AND vector.cm_document_template_id = ANY($3)`
			args = append(args, pq.Array(scope))
		}

	default: // "project"
		query = `
			SELECT vector.content, vector.embedding <#> ` + vectorStr + `::vector AS distance, doc.title
			FROM st_schema.project_document_vectors vector
			JOIN st_schema.project_documents doc ON doc.id = vector.document_id
			WHERE vector.tenant_id = $1 AND vector.project_id = $2
		`
		args = append(args, chatCtx.TenantID, chatCtx.ResourceGroupID)

		if len(scope) > 0 {
			query += ` AND vector.document_id = ANY($3)`
			args = append(args, pq.Array(scope))
		}
	}

	query += fmt.Sprintf(` ORDER BY distance ASC LIMIT $%d`, len(args)+1)
	args = append(args, req.Limit)

	return tenantManagement.DB.Query(query, args...)
}

func DocumentSearch(openaiClient *openai.Client, req *DocSearchRequest, chatCtx *models.ChatContext) ([]DocSearchResult, error) {
	// 🪵 Pretty log input
	if payload, err := json.MarshalIndent(req, "", "  "); err == nil {
		log.Printf("[DocumentSearch] Request:\n%s", payload)
	} else {
		log.Printf("[DocumentSearch] Failed to marshal request: %v", err)
	}

	vectorStr, err := embedQuery(openaiClient, req.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding for a query message: %v", err)
	}

	rows, err := selectDocSearchSource(req, *vectorStr, chatCtx)
	if err != nil {
		return nil, fmt.Errorf("failed preparing query: %v", err)
	}
	defer rows.Close()

	var results []DocSearchResult
	for rows.Next() {
		var res DocSearchResult
		if err := rows.Scan(&res.Content, &res.Distance, &res.Title); err != nil {
			return nil, fmt.Errorf("failed to read db row: %v", err)
		}
		results = append(results, res)
	}

	return results, nil
}

func DocumentSearchHandler(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	projectID := c.Query("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	var req DocSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	openaiConfig, err := utilities.GetOpenAiConfig(tenantManagement.DB, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": models.InternalServerError})
	}

	openaiClient := openai.NewClient(
		option.WithAPIKey(openaiConfig.OpenAIApiKey),
	)

	results, err := DocumentSearch(&openaiClient, &req, &models.ChatContext{
		UserID:   userID,
		TenantID: tenantID,
		ResourceIdentifier: models.ResourceIdentifier{
			ResourceGroupType: models.ResourceGroupProject,
			ResourceGroupID:   &projectID,
			ResourceType:      utilities.Ptr(models.ResourceTypeDocument),
		},
	})
	if err != nil {
		log.Printf("Semantic search failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": models.InternalServerError})
		return
	}

	c.JSON(http.StatusOK, results)
}
