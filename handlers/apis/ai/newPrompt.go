package ai

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	aiFunctions "sententiawebapi/handlers/apis/ai/functions"
	"sententiawebapi/handlers/apis/tenantManagement"
	"sententiawebapi/handlers/models"
	"sententiawebapi/utilities"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/responses"
)

type UserPromptParams struct {
	WebSearch              bool                       `json:"web_search,omitempty"`
	AdditionalInstructions string                     `json:"additional_instructions,omitempty"`
	Selections             []models.DocumentSelection `json:"selections,omitempty"`
}

type PromptBody struct {
	Prompt        string            `json:"prompt" binding:"required"`
	Params        *UserPromptParams `json:"params,omitempty"`
	Stream        bool              `json:"stream,omitempty"`
	ChunkedStream bool              `json:"chunked_stream,omitempty"`
}

type ConversationData struct {
	ID                           string `json:"id"`
	LastChatCompletionID         string `json:"last_chat_completion_id"`
	ConversationConfigTemplateId string `json:"conversation_config_template_id"`
}

type ChatCompletionData struct {
	Message        string                   `json:"message"`
	Status         responses.ResponseStatus `json:"status"`
	TotalTokens    int64                    `json:"total_tokens"`
	Params         *responses.ResponseNewParams
	ShouldContinue bool
}

// type GenerationState struct {
// 	Mutex    sync.Mutex
// 	TenantID string
// 	StreamID string
// 	Messages string // Collects the chunks in a stream
// 	Error    error
// 	Ended    bool
// }

// var newGenerationSessions = make(map[string]*GenerationState)

func getConversationData(tenantID string, conversationID string) (*ConversationData, error) {
	var conversationData ConversationData
	var lastChatCompletionID sql.NullString
	var conversationConfigTemplateID sql.NullString

	query := `
		SELECT
			id,
			last_chat_completion_id,
			conversation_config_template_id
		FROM
			st_schema.conversation
		WHERE
			tenant_id = $1
		AND id = $2
	`

	err := tenantManagement.DB.QueryRow(query, tenantID, conversationID).Scan(
		&conversationData.ID,
		&lastChatCompletionID,
		&conversationConfigTemplateID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Conversation does not exist")
			return nil, err
		}

		log.Printf("Failed to retrieve conversation data: %v", err)
		return nil, err
	}

	if conversationConfigTemplateID.Valid {
		conversationData.ConversationConfigTemplateId = conversationConfigTemplateID.String
	} else {
		conversationData.ConversationConfigTemplateId = ""
	}

	if lastChatCompletionID.Valid {
		conversationData.LastChatCompletionID = lastChatCompletionID.String
	} else {
		conversationData.LastChatCompletionID = ""
	}

	return &conversationData, nil
}

func getTemplateConfig(tenantID string, conversationID string) (*models.AiConfiguration, error) {
	query := `
		SELECT
			pt.configuration
		FROM
			st_schema.prompt_config_template pt
		INNER JOIN
			st_schema.conversation c ON pt.id = c.conversation_config_template_id
		WHERE
			c.id = $1 AND pt.tenant_id = $2;
	`

	var rawConfig []byte
	err := tenantManagement.DB.QueryRow(query, conversationID, tenantID).Scan(&rawConfig)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Prompt configuration does not exist: %v", err)
			return nil, err
		}
		log.Printf("Failed to get prompt configuration: %v", err)
		return nil, err
	}

	var config models.AiConfiguration
	if err := json.Unmarshal(rawConfig, &config); err != nil {
		log.Printf("Failed to unmarshal prompt configuration: %v", err)
		return nil, err
	}

	return &config, nil
}

// Return default parameters, used mostly for quick chat
func getDefaultResponseParams(userID string, userPrompt string) *responses.ResponseNewParams {
	return &responses.ResponseNewParams{
		Model:           "gpt-4o-mini",
		User:            openai.String(userID),
		Instructions:    openai.String("You are a helpful assistant that can answer questions and help with tasks."),
		Temperature:     openai.Float(0.5),
		TopP:            openai.Float(1),
		MaxOutputTokens: openai.Int(4000),
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(userPrompt),
		},
	}
}

func useConversationConfig(chatCtx *models.ChatContext, responseParams *responses.ResponseNewParams) (err error) {
	conversationData, err := getConversationData(chatCtx.TenantID, *chatCtx.ConversationID)
	if err != nil {
		return err
	}

	// OpenAI needs to store messages for history
	responseParams.Store = openai.Bool(true)

	// Set the configuration
	if conversationData.ConversationConfigTemplateId != "" {
		// Use user-defined configuration
		templateConfig, err := getTemplateConfig(chatCtx.TenantID, *chatCtx.ConversationID)
		if err != nil {
			return err
		}

		responseParams.Temperature = openai.Float(templateConfig.AiTemperature)
		responseParams.TopP = openai.Float(templateConfig.TopP)
		responseParams.MaxOutputTokens = openai.Int(int64(templateConfig.MaxTokens))
		responseParams.Instructions = openai.String(templateConfig.SystemConfig)
	}

	// Attach last response id for history
	if conversationData.LastChatCompletionID != "" {
		responseParams.PreviousResponseID = openai.String(conversationData.LastChatCompletionID)
	}

	// Add functions
	responseParams.Tools = append(responseParams.Tools, aiFunctions.GetFunctionDefinitions(chatCtx)...)

	return nil
}

func useAssistantConfig(assistantName string, openAiConfig *models.OpenAiConfig, responseParams *responses.ResponseNewParams) (err error) {
	var rawConfig []byte

	query := `
		SELECT
			config
		FROM
			st_schema.assistants
		WHERE
			name = $1
		AND is_active = true
	`

	err = tenantManagement.DB.QueryRow(query, assistantName).Scan(&rawConfig)
	if err != nil {
		log.Printf("Failed to get assistant config: %v", err)
		return err
	}

	var assistantParams models.AssistantParams
	if err := json.Unmarshal(rawConfig, &assistantParams); err != nil {
		log.Printf("Failed to unmarshal assistant config: %v", err)
		return err
	}

	mergeAssistantParams(responseParams, &assistantParams, openAiConfig)

	return nil
}

func useUserConfig(userParams *UserPromptParams, responseParams *responses.ResponseNewParams) (err error) {
	if userParams.WebSearch {
		responseParams.Tools = append(responseParams.Tools, responses.ToolUnionParam{
			OfWebSearch: &responses.WebSearchToolParam{
				Type:              responses.WebSearchToolTypeWebSearchPreview2025_03_11,
				SearchContextSize: responses.WebSearchToolSearchContextSizeMedium,
			},
		})

		responseParams.ToolChoice = responses.ResponseNewParamsToolChoiceUnion{
			OfHostedTool: &responses.ToolChoiceTypesParam{
				Type: responses.ToolChoiceTypesTypeWebSearchPreview,
			},
		}
	}

	if userParams.AdditionalInstructions != "" {
		if responseParams.Instructions.IsPresent() {
			responseParams.Instructions = openai.String(responseParams.Instructions.Value + "\n" + userParams.AdditionalInstructions)
		} else {
			responseParams.Instructions = openai.String(userParams.AdditionalInstructions)
		}
	}

	if len(userParams.Selections) > 0 {
		currInput := responseParams.Input.OfString.Value
		var selectionsText strings.Builder

		for i, selection := range userParams.Selections {
			selectionsText.WriteString(fmt.Sprintf("Document selection %d: ", i+1))
			selectionsText.WriteString("```html\n")
			selectionsText.WriteString(selection.Data)
			selectionsText.WriteString("\n```\n")
		}

		newInput := selectionsText.String() + "\n\n" + currInput
		responseParams.Input = responses.ResponseNewParamsInputUnion{
			OfString: openai.String(newInput),
		}

		selectionsInstructions := "You will find document selections in user input. In most cases, ignore any HTML tags or labels like 'Document Selection 1'; focus on what the text says as that's the only thing user sees."

		if responseParams.Instructions.IsPresent() {
			responseParams.Instructions = openai.String(responseParams.Instructions.Value + "\n" + selectionsInstructions)
		} else {
			responseParams.Instructions = openai.String(selectionsInstructions)
		}
	}

	return nil
}

func updateLastChatCompletionID(tenantID string, conversationID string, lastChatCompletionID string) error {
	query := `
		UPDATE
			st_schema.conversation
		SET
			last_chat_completion_id = $1, updated_at = NOW()
		WHERE
			tenant_id = $2
		AND id = $3
	`

	_, err := tenantManagement.DB.Exec(query, lastChatCompletionID, tenantID, conversationID)

	if err != nil {
		log.Printf("Failed to update last chat completion ID: %v", err)
		return err
	}

	return nil
}

func newTokenUsageResource(tokenUsageRequest *models.TenantTokenUsageRequest) error {
	// Prepare the database statement to insert the AI response
	stmt, err := tenantManagement.DB.Prepare(`
        INSERT INTO st_schema.tenant_token_usage (
            tenant_id,
            user_id,
            conversation_id,
			ai_vendor,
            ai_model,
			configuration,
            prompt_tokens,
            completion_tokens
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8
        ) RETURNING
            id,
            tenant_id,
            user_id,
            conversation_id,
            ai_vendor,
            ai_model,
            prompt_tokens,
            completion_tokens,
            created_at
    `)

	// If the service failed to prepare the statement, return an error
	if err != nil {
		log.Printf("Failed to prepare SQL statement: %v", err)
		return err
	}

	defer stmt.Close()

	toolsJSON, err := json.Marshal(tokenUsageRequest.Tools)
	if err != nil {
		log.Printf("Failed to marshal tools: %v", err)
		return err
	}

	// Execute the statement and return the result
	var tokenUsageResource models.TenantTokenUsageResource
	err = stmt.QueryRow(
		tokenUsageRequest.TenantID,
		tokenUsageRequest.UserID,
		tokenUsageRequest.ConversationID,
		tokenUsageRequest.AiVendor,
		tokenUsageRequest.AiModel,
		toolsJSON,
		tokenUsageRequest.PromptTokens,
		tokenUsageRequest.CompletionTokens,
	).Scan(
		&tokenUsageResource.ID,
		&tokenUsageResource.TenantID,
		&tokenUsageResource.UserID,
		&tokenUsageResource.ConversationID,
		&tokenUsageResource.AiVendor,
		&tokenUsageResource.AiModel,
		&tokenUsageResource.PromptTokens,
		&tokenUsageResource.CompletionTokens,
		&tokenUsageResource.CreatedAt,
	)

	if err != nil {
		log.Printf("Failed to execute SQL insertion: %v", err)
		return err
	}

	return nil
}

func handleCompletion(chatCtx *models.ChatContext, response *responses.Response, params *responses.ResponseNewParams, client *openai.Client) (data *ChatCompletionData, err error) {
	// clear the previous input entirely
	params.Input = responses.ResponseNewParamsInputUnion{
		OfInputItemList: []responses.ResponseInputItemUnionParam{},
	}

	functionOutputs, err := aiFunctions.ExecuteFunctionCallsParallel(context.Background(), client, chatCtx, response.Output)
	if err != nil {
		return nil, fmt.Errorf("function call failed: %v", err)
	}

	params.Input.OfInputItemList = append(
		params.Input.OfInputItemList,
		functionOutputs...,
	)

	hasFunctionCall := len(functionOutputs) > 0

	if chatCtx.ConversationID != nil {
		params.PreviousResponseID = openai.String(response.ID)
		if !hasFunctionCall {
			updateLastChatCompletionID(chatCtx.TenantID, *chatCtx.ConversationID, response.ID)
		}
	}

	var outputText = response.OutputText()

	now := time.Now()
	defer func() {
		err := newTokenUsageResource(&models.TenantTokenUsageRequest{
			TenantID:         chatCtx.TenantID,
			UserID:           chatCtx.UserID,
			ConversationID:   chatCtx.ConversationID,
			AiVendor:         "openai",
			AiModel:          response.Model,
			Tools:            map[string]interface{}{},
			PromptTokens:     int32(response.Usage.InputTokens),
			CompletionTokens: int32(response.Usage.OutputTokens),
		})
		if err == nil {
			log.Printf("Token usage stored successfully")
		}

		if chatCtx.ConversationID != nil {
			// Save AI response for history
			_, err = newCompletionResource(&models.CompletionRequest{
				UserID:           chatCtx.UserID,
				TenantID:         chatCtx.TenantID,
				ConversationID:   *chatCtx.ConversationID,
				Prompt:           outputText,
				Tools:            map[string]interface{}{},
				CreatedAt:        &now,
				PromptTokens:     int32(response.Usage.InputTokens),
				CompletionTokens: int32(response.Usage.OutputTokens),
				TotalTokens:      int32(response.Usage.TotalTokens),
			})
			if err == nil {
				log.Printf("Completion resource stored successfully")
			}
		}
	}()

	tokenUsage := gin.H{
		"prompt_tokens":     response.Usage.InputTokens,
		"completion_tokens": response.Usage.OutputTokens,
		"total_tokens":      response.Usage.TotalTokens,
	}

	log.Printf("Token usage: %v", tokenUsage)

	if outputText == "" && !hasFunctionCall {
		log.Printf("No output text available in response")
		return nil, fmt.Errorf("no output text available in response")
	}

	return &ChatCompletionData{
		Message:        outputText,
		Status:         response.Status,
		TotalTokens:    response.Usage.TotalTokens,
		Params:         params,
		ShouldContinue: hasFunctionCall,
	}, nil
}

// func generateToken(userID string, tenantID string) (string, error) {
// 	claims := models.QueryParamToken{
// 		UserID:   userID,
// 		TenantID: tenantID,
// 		RegisteredClaims: jwt.RegisteredClaims{
// 			ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Minute)),
// 			Subject:   userID,
// 		},
// 	}

// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

// 	tokenString, err := token.SignedString([]byte(os.Getenv("SENDGRID_KEY")))

// 	if err != nil {
// 		log.Println("Could not generate token", err)

// 		return "", err
// 	}

// 	return tokenString, nil
// }

// func streamResponse(c *gin.Context, client *openai.Client, idMap *IdMap, responseParams *responses.ResponseNewParams) {
// 	stream := client.Responses.NewStreaming(context.Background(), *responseParams)

// 	streamID := uuid.New().String()
// 	session := &GenerationState{
// 		TenantID: idMap.TenantID,
// 		StreamID: streamID,
// 		Messages: "",
// 		Ended:    false,
// 	}
// 	newGenerationSessions[streamID] = session

// 	// Start streaming in a goroutine
// 	go func() {
// 		var completeText string

// 		for stream.Next() {
// 			event := stream.Current()

// 			if event.JSON.Text.IsPresent() {
// 				log.Printf("Content stream finished: %s", event.Text)
// 				completeText = event.Text

// 				session.Mutex.Lock()
// 				session.Ended = true
// 				session.Mutex.Unlock()
// 			}

// 			if !session.Ended {
// 				session.Mutex.Lock()
// 				session.Messages += strings.ReplaceAll(event.Delta, " ", "&nbsp;")
// 				session.Mutex.Unlock()
// 			}
// 		}

// 		if err := stream.Err(); err != nil {
// 			log.Printf("Error during streaming: %v", err)

// 			session.Mutex.Lock()
// 			session.Error = err
// 			session.Mutex.Unlock()
// 		}

// 		log.Printf("Complete text: %s", completeText)

// 		response := stream.Current().Response
// 		handleCompletion(idMap, &response, responseParams, client)
// 	}()

// 	token, err := generateToken(idMap.UserID, idMap.TenantID)
// 	if err != nil {
// 		log.Printf("Failed to generate token: %v", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
// 		return
// 	}

// 	c.JSON(http.StatusAccepted, gin.H{
// 		"status":     "Generation started",
// 		"stream_url": fmt.Sprintf("newAiStream?stream_id=%s&token=%s", streamID, token),
// 	})
// }

func streamChunks(c *gin.Context, client *openai.Client, chatCtx *models.ChatContext, responseParams *responses.ResponseNewParams) error {
	// Set headers for chunked streaming
	c.Header("Content-Type", "text/plain; charset=UTF-8")
	c.Header("Transfer-Encoding", "chunked")

	// Write the HTTP status code before streaming begins
	c.Writer.WriteHeader(http.StatusOK)

	// Make sure the ResponseWriter supports streaming by checking for the Flusher interface.
	writer := c.Writer
	flusher, ok := writer.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming not supported")
	}

	for {
		stream := client.Responses.NewStreaming(context.Background(), *responseParams)

		for stream.Next() {
			event := stream.Current()

			if event.Type == "response.output_text.delta" {
				writer.Write([]byte(event.Delta))
				flusher.Flush()
			}
		}
		if err := stream.Err(); err != nil {
			return fmt.Errorf("error during streaming: %v", err)
		}

		if b, err := json.MarshalIndent(stream.Current().Response, "", "  "); err == nil {
			log.Println("==== Response ====")
			log.Println(string(b))
			log.Println("==================")

		} else {
			log.Println("Failed to marshal response:", err)
		}

		response := stream.Current().Response
		completion, err := handleCompletion(chatCtx, &response, responseParams, client)
		if err != nil {
			return fmt.Errorf("failed to process the openai response: %v", err)
		}

		if completion != nil && completion.ShouldContinue {
			continue
		}

		break
	}

	return nil
}

func regularResponse(client *openai.Client, chatCtx *models.ChatContext, responseParams *responses.ResponseNewParams) (*ChatCompletionData, error) {
	response, err := client.Responses.New(
		context.Background(),
		*responseParams,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat response: %v", err)
	}

	completion, err := handleCompletion(chatCtx, response, responseParams, client)
	if err != nil {
		return nil, fmt.Errorf("failed to process openai response: %v", err)
	}

	return completion, nil
}

func NewPromptHandler(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	conversationID := c.Query("conversation_id")
	assistantName := c.Query("assistant")

	resourceIdentifier, err := utilities.ResolveResourceIdentifier(c)
	if err != nil {
		log.Printf("Resource identifier err: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown or missing resource identifier"})
		return
	}

	chatCtx := &models.ChatContext{
		UserID:             userID,
		TenantID:           tenantID,
		ResourceIdentifier: *resourceIdentifier,
	}

	if conversationID != "" {
		chatCtx.ConversationID = &conversationID
	}

	var promptBody PromptBody
	if err := c.ShouldBindJSON(&promptBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if conversationID != "" {
		now := time.Now()

		defer func() {
			// Save user message for history
			promptRequest := &models.PromptRequest{
				UserID:         userID,
				TenantID:       tenantID,
				ConversationID: conversationID,
				Message:        promptBody.Prompt,
				Selections:     promptBody.Params.Selections,
				CreatedAt:      &now,
			}

			_, err := newPromptResource(promptRequest)
			if err == nil {
				log.Printf("Prompt resource stored successfully")
			}
		}()
	}

	client, openAiConfig, err := GetOpenAiClient(tenantID)
	if err != nil {
		log.Printf("Failed to get OpenAI client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": models.InternalServerError})
	}

	var responseParams = getDefaultResponseParams(userID, promptBody.Prompt)

	var configError error
	if conversationID != "" {
		configError = useConversationConfig(chatCtx, responseParams)
	}

	if assistantName != "" {
		configError = useAssistantConfig(assistantName, openAiConfig, responseParams)
	}

	if promptBody.Params != nil {
		configError = useUserConfig(promptBody.Params, responseParams)
	}

	if configError != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	if promptBody.ChunkedStream {
		err := streamChunks(c, client, chatCtx, responseParams)
		if err != nil {
			log.Printf("Streaming error: %v", err)
		}
	} else {
		var running = true
		var totalTokens int64 = 0
		var lastCompletion *ChatCompletionData

		for running {
			completion, err := regularResponse(client, chatCtx, responseParams)
			if err != nil {
				log.Print(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": models.InternalServerError})
				return
			}

			totalTokens += completion.TotalTokens
			running = completion.ShouldContinue
			lastCompletion = completion
		}

		c.JSON(http.StatusOK, gin.H{
			"message":      lastCompletion.Message,
			"status":       lastCompletion.Status,
			"total_tokens": totalTokens,
		})
	}
}

// func NewStreamHandler(c *gin.Context) {
// 	streamID := c.Query("stream_id")

// 	if streamID == "" {
// 		log.Printf("Stream ID is not provided")
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Stream ID is required"})
// 		return
// 	}

// 	// Retrieve session for the stream ID
// 	session, exists := newGenerationSessions[streamID]
// 	if !exists {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "Stream not found in active sessions"})
// 		return
// 	}

// 	c.Header("Content-Type", "text/event-stream")
// 	c.Header("Cache-Control", "no-cache")
// 	c.Header("Connection", "keep-alive")
// 	c.Header("Transfer-Encoding", "chunked")

// 	flusher, ok := c.Writer.(http.Flusher)
// 	if !ok {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming unsupported"})
// 		return
// 	}

// 	// Stream the response in real-time
// 	c.Stream(func(w io.Writer) bool {
// 		session.Mutex.Lock()
// 		defer session.Mutex.Unlock()

// 		// If an error occurred during streaming
// 		if session.Error != nil {
// 			c.SSEvent("error", session.Error.Error())
// 			return false
// 		}

// 		// Stream the current buffer
// 		if len(session.Messages) > 0 {
// 			c.SSEvent("message", session.Messages)
// 			session.Messages = "" // Clear the buffer after streaming
// 			flusher.Flush()
// 		}

// 		// Check if the end-of-stream marker is present
// 		if session.Ended {
// 			c.SSEvent("end-of-stream", "Stream completed")
// 			return false
// 		}

// 		// If no new data, keep the connection alive
// 		select {
// 		case <-c.Request.Context().Done():
// 			return false
// 		default:
// 			return true
// 		}
// 	})
// }

func lines(parts ...string) string {
	return strings.Join(parts, "\n")
}
