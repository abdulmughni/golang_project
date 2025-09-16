package ai

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"sententiawebapi/handlers/apis/tenantManagement"
	"sententiawebapi/handlers/models"
	"sententiawebapi/utilities"
	"strings"

	"github.com/gin-gonic/gin"
)

func newPromptResource(promptRequest *models.PromptRequest) (*models.PromptResource, error) {
	// Local function not a handler
	// Endpoint: NewPromptResource
	// Return: PromptResource data object
	// Table: st_schema.user_prompt

	// Functionality: Saves the user prompt message in the database in the database
	// in user_prompt table. This will be used to generate the AI prompt
	// for the Azure OpenAI API

	// Marshal selections to JSON before sending to DB
	var selectionsJSON interface{} = nil
	var err error

	if len(promptRequest.Selections) > 0 {
		selectionsJSON, err = json.Marshal(promptRequest.Selections)

		if err != nil {
			log.Printf("[newPromptResource] Error marshaling selections: %v", err)

			return nil, err
		}
	}

	// Prepare the database statement to insert the user prompt
	// We also need to return it from the function
	stmt, err := tenantManagement.DB.Prepare(`
        INSERT INTO st_schema.user_prompt (
            user_id,
            tenant_id,
            conversation_id,
            prompt,
			selections,
			created_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6
        ) RETURNING
            id,
            user_id,
            tenant_id,
            conversation_id,
			'user' AS prompt_role,
            prompt,
            created_at,
            updated_at
    `)

	if err != nil {
		log.Printf("[newPromptResource] Error preparing statement: %v", err)
		return nil, err
	}

	defer stmt.Close()

	var promptResource models.PromptResource

	err = stmt.QueryRow(
		promptRequest.UserID,
		promptRequest.TenantID,
		promptRequest.ConversationID,
		promptRequest.Message,
		selectionsJSON,
		promptRequest.CreatedAt,
	).Scan(
		&promptResource.ID,
		&promptResource.UserID,
		&promptResource.TenantID,
		&promptResource.ConversationID,
		&promptResource.PromptRole,
		&promptResource.Prompt,
		&promptResource.CreatedAt,
		&promptResource.UpdatedAt,
	)

	promptResource.Selections = promptRequest.Selections

	if err != nil {
		log.Printf("[newPromptResource] Error executing statement: %v", err)
		return nil, err
	}

	return &promptResource, nil
}

func newCompletionResource(completionRequest *models.CompletionRequest) (*models.CompletionRequestResource, error) {
	// Check if the prompt string starts and ends with a double quote
	if strings.HasPrefix(completionRequest.Prompt, "\"") && strings.HasSuffix(completionRequest.Prompt, "\"") {
		// If it does, remove the first and last double quote from the string
		completionRequest.Prompt = strings.TrimPrefix(completionRequest.Prompt, "\"")
		completionRequest.Prompt = strings.TrimSuffix(completionRequest.Prompt, "\"")
	}

	// Prepare the database statement to insert the AI response
	stmt, err := tenantManagement.DB.Prepare(`
        INSERT INTO st_schema.completion_prompt (
            user_id,
            tenant_id,
            conversation_id,
            prompt,
			tools,
            prompt_tokens,
            completion_tokens,
			total_tokens,
			created_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9
        ) RETURNING
            id,
            user_id,
            tenant_id,
            conversation_id,
            'assistant' AS prompt_role,
            prompt,
            created_at,
            updated_at,
            prompt_tokens,
            completion_tokens,
			total_tokens
    `)

	// If the service failed to prepare the statement, return an error
	if err != nil {
		log.Printf("[newCompletionResource] Error preparing statement: %v", err)
		return nil, err
	}

	defer stmt.Close()

	toolsJSON, err := json.Marshal(completionRequest.Tools)
	if err != nil {
		log.Printf("Failed to marshal tools: %v", err)
		return nil, err
	}

	// Execute the statement and return the result
	var completionRequestResource models.CompletionRequestResource
	err = stmt.QueryRow(
		completionRequest.UserID,
		completionRequest.TenantID,
		completionRequest.ConversationID,
		completionRequest.Prompt,
		toolsJSON,
		completionRequest.PromptTokens,
		completionRequest.CompletionTokens,
		completionRequest.TotalTokens,
		completionRequest.CreatedAt,
	).Scan(
		&completionRequestResource.ID,
		&completionRequestResource.UserID,
		&completionRequestResource.TenantID,
		&completionRequestResource.ConversationID,
		&completionRequestResource.PromptRole,
		&completionRequestResource.Prompt,
		&completionRequestResource.CreatedAt,
		&completionRequestResource.UpdatedAt,
		&completionRequestResource.PromptTokens,
		&completionRequestResource.CompletionTokens,
		&completionRequestResource.TotalTokens,
	)

	if err != nil {
		log.Printf("[newCompletionResource] Error executing statement: %v", err)
		return nil, err
	}

	// Return complete resource
	return &completionRequestResource, nil
}

func getCombinedMessages(tenantID string, conversationID string) ([]models.MessageResource, error) {
	var messages []models.MessageResource

	query := `
		SELECT
			id, prompt, created_at, 'user' as role, selections
		FROM
			st_schema.user_prompt
		WHERE
			tenant_id = $1 AND conversation_id = $2
		UNION ALL
		SELECT
			id, prompt, created_at, 'assistant' as role, NULL as selections
		FROM
			st_schema.completion_prompt
		WHERE
			tenant_id = $1 AND conversation_id = $2
		ORDER BY
			created_at ASC
	`

	rows, err := tenantManagement.DB.Query(query, tenantID, conversationID)
	if err != nil {
		log.Printf("Error fetching combined messages: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var message models.MessageResource
		var selectionsJSON *json.RawMessage

		err := rows.Scan(&message.ID, &message.Content, &message.CreatedAt, &message.Role, &selectionsJSON)
		if err != nil {
			if err == sql.ErrNoRows {
				return messages, nil
			}

			log.Printf("Error scanning combined message row: %v", err)
			return nil, err
		}

		// Parse selections if they exist
		if selectionsJSON != nil {
			err = json.Unmarshal(*selectionsJSON, &message.Selections)

			if err != nil {
				log.Printf("Error unmarshaling selections: %v", err)

				return nil, err
			}
		} else if message.Role == "user" {
			// assign empty array
			message.Selections = []models.DocumentSelection{}
		}

		messages = append(messages, message)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error in combined messages query result: %v", err)
		return nil, err
	}

	return messages, nil
}

func NewChatHistoryHandler(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	conversationID := c.Query("conversation_id")
	if conversationID == "" {
		log.Printf("Conversation ID is not provided")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Conversation ID is required"})
		return
	}

	messages, err := getCombinedMessages(tenantID, conversationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	// Ensure messages is an empty array if nil
	if messages == nil {
		messages = []models.MessageResource{}
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
	})
}
