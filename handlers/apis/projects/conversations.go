package projects

// This package contains the handlers for all conversation related operations.
// The handlers are:
// 1. NewConversation - Creates a new conversation entry in the 'conversation' table based on a template type that can come from 'community_prompt_config_templates', 'sp_prompt_config_templates', or 'prompt_config_template'.
// 2. GetConversation - Retrieves a specific conversation from the 'conversation' table using the provided user ID, project ID, and conversation ID.
// 3. GetConversations - Retrieves all conversations from the 'conversation' table for a given project ID and user ID.
// 4. UpdateConversation - Updates a conversation in the 'conversation' table based on the provided project ID, conversation ID, and user ID, setting a new 'conversation_config_template_id'.
// 5. DeleteConversation - Deletes a conversation from the 'conversation' table using the provided user ID, project ID, and conversation ID.

// Local Functions are:
// 1. checkExistingTemplate - Checks if a template already exists for this source ID and user.
// 2. createNewTemplateFromSource - Creates a new template from the source.
// 3. getConfigData - Fetches configuration data for title.

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"sententiawebapi/handlers/apis/tenantManagement"
	"sententiawebapi/handlers/models"
	"sententiawebapi/utilities"

	"github.com/gin-gonic/gin"
)

type UpdateConversationResource struct {
	UserID string `json:"user_id"`
	// Title                 *string `json:"title"`
	PromptConfigurationID string `json:"prompt_configuration"`
}

func NewConversation(c *gin.Context) {
	// Get the user ID and tenant ID from the context
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	var conversationObject models.Conversation

	projectId := c.Query("project_id")
	templateId := c.Query("template_id")
	communityTemplateId := c.Query("community_template_id")
	conversationConfigId := c.Query("conversation_configuration_id")
	templateType := c.Query("template_type")
	agentName := c.Query("agent_name")

	if err := c.ShouldBindJSON(&conversationObject); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if projectId == "" && templateId == "" && communityTemplateId == "" {
		log.Println("Conversation needs to belong to either project or private/community template")
		c.JSON(http.StatusBadRequest, gin.H{"error": models.ParameterRequired})
		return
	}

	if conversationConfigId != "" && templateType == "" {
		log.Println("Template type is missing")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template type is required when using custom conversation configuration"})
		return
	}

	var newConfigId *string
	var err error

	// Check if a template already exists for this source ID and user
	if conversationConfigId != "" && (templateType == "community" || templateType == "solutionPilot") {
		existingTemplateID, err := checkExistingTemplate(conversationConfigId, tenantID)
		if err != nil {
			log.Printf("Error checking existing template: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing template"})
			return
		}

		if existingTemplateID != "" {
			// Use the existing template
			newConfigId = &existingTemplateID
		} else if conversationConfigId != "" {
			// Create a new template from the source
			newConfigId, err = createNewTemplateFromSource(&conversationConfigId, userID, tenantID, templateType)
			if err != nil {
				log.Printf("Error creating new template from source: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing template"})
				return
			}
		}
	} else if conversationConfigId != "" && templateType == "private" {
		// For private templates, use the provided configuration ID
		newConfigId = &conversationConfigId
	} else if conversationConfigId != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template type"})
		return
	}

	conversationObject.ConversationConfigurationId = newConfigId
	conversationObject.UserID = userID
	conversationObject.TenantID = tenantID

	if projectId != "" {
		conversationObject.ProjectId = &projectId
	}

	if templateId != "" {
		conversationObject.TemplateId = &templateId
	}

	if communityTemplateId != "" {
		conversationObject.CommunityTemplateId = &communityTemplateId
	}

	if agentName != "" {
		conversationObject.AgentName = &agentName
	}

	if conversationObject.ConversationType == "" {
		conversationObject.ConversationType = "chat"
	}

	if conversationObject.Title == "" {
		if newConfigId == nil {
			conversationObject.Title = "Quick Chat"
		} else {
			// Fetch configuration data for title
			configData, err := getConfigData(*newConfigId)
			if err != nil {
				log.Printf("Error fetching configuration data: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": models.DatabaseError})
				return
			}

			conversationObject.Title = *configData.Title
		}
	}

	// Insert the new conversation into the database directly without a transaction
	row := tenantManagement.DB.QueryRow(`
        INSERT INTO
            st_schema.conversation (
                user_id,
                tenant_id,
                project_id,
                template_id,
				community_template_id,
                conversation_config_template_id,
				agent_name,
                title,
                conversation_type,
                description
            )
        VALUES
            ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        RETURNING
            id,
            project_id,
            template_id,
			community_template_id,
            user_id,
            tenant_id,
            conversation_config_template_id,
			agent_name,
            title,
            conversation_type,
            created_at,
            updated_at,
            description
        `,
		conversationObject.UserID,
		conversationObject.TenantID,
		conversationObject.ProjectId,
		conversationObject.TemplateId,
		conversationObject.CommunityTemplateId,
		conversationObject.ConversationConfigurationId,
		conversationObject.AgentName,
		conversationObject.Title,
		conversationObject.ConversationType,
		conversationObject.Description,
	)

	err = row.Scan(
		&conversationObject.ID,
		&conversationObject.ProjectId,
		&conversationObject.TemplateId,
		&conversationObject.CommunityTemplateId,
		&conversationObject.UserID,
		&conversationObject.TenantID,
		&conversationObject.ConversationConfigurationId,
		&conversationObject.AgentName,
		&conversationObject.Title,
		&conversationObject.ConversationType,
		&conversationObject.CreatedAt,
		&conversationObject.UpdatedAt,
		&conversationObject.Description,
	)
	if err != nil {
		log.Printf("Error scanning row: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": models.DatabaseError})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    conversationObject,
		"message": models.StatusSuccess,
	})
}

// @Summary Retrieve a single conversation
// @Description Retrieves a specific conversation from the 'conversation' table.
// @Tags Conversations
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer [Token]"
// @Param project_id query string true "Project ID"
// @Param conversation_id query string true "Conversation ID"
// @Success 200 {object} Conversation "Conversation retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid input data"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/conversation [get]
func GetConversation(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	projectId := c.Query("project_id")
	templateId := c.Query("template_id")
	communityTemplateId := c.Query("community_template_id")

	var resourceId string

	if projectId != "" {
		resourceId = projectId
	} else if templateId != "" {
		resourceId = templateId
	} else if communityTemplateId != "" {
		resourceId = communityTemplateId
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Conversation needs to belong to either project or private/community template"})
		return
	}

	conversationId := c.Query("conversation_id")
	if conversationId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Conversation ID is required"})
		return
	}

	// Query to fetch the specific conversation
	query := `
		SELECT
			id,
			user_id,
			tenant_id,
			project_id,
			template_id,
			community_template_id,
			conversation_config_template_id,
			agent_name,
			title,
			conversation_type,
			created_at,
			updated_at,
			description
		FROM
			st_schema.conversation
		WHERE
			id = $1
			AND tenant_id = $2
			AND (project_id = $3 OR template_id = $3 OR community_template_id = $3)
	`

	row := tenantManagement.DB.QueryRow(query, conversationId, tenantID, resourceId)
	var conversation models.Conversation

	err := row.Scan(
		&conversation.ID,
		&conversation.UserID,
		&conversation.TenantID,
		&conversation.ProjectId,
		&conversation.TemplateId,
		&conversation.CommunityTemplateId,
		&conversation.ConversationConfigurationId,
		&conversation.AgentName,
		&conversation.Title,
		&conversation.ConversationType,
		&conversation.CreatedAt,
		&conversation.UpdatedAt,
		&conversation.Description,
	)

	if err != nil {
		log.Printf("Error scanning conversation: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": models.DatabaseError})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    conversation,
		"message": models.StatusSuccess,
	})
}

// @Summary Retrieve all conversations for a project
// @Description Retrieves all conversations for a given project ID and tenant ID.
// @Tags Conversations
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer [Token]"
// @Param project_id query string true "Project ID"
// @Success 200 {array} Conversation "List of conversations"
// @Failure 400 {object} map[string]string "Invalid input data"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/conversations [get]
func GetConversations(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	projectId := c.Query("project_id")
	templateId := c.Query("template_id")
	communityTemplateId := c.Query("community_template_id")

	var resourceId string

	if projectId != "" {
		resourceId = projectId
	} else if templateId != "" {
		resourceId = templateId
	} else if communityTemplateId != "" {
		resourceId = communityTemplateId
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Conversations needs to belong to either project or private/community template"})
		return
	}

	// Prepare the SQL query - update to include tenant_id and filter by tenant_id
	stmt, err := tenantManagement.DB.Prepare(`
        SELECT
            id,
            user_id,
            tenant_id,
            project_id,
			template_id,
			community_template_id,
            conversation_config_template_id,
			agent_name,
            title,
            conversation_type,
            created_at,
            updated_at,
            description
        FROM
            st_schema.conversation
        WHERE
            tenant_id = $1
        	AND (project_id = $2 OR template_id = $2 OR community_template_id = $2)
    `)
	if err != nil {
		if isDevelopmentEnvironment() {
			log.Printf("Database Err: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	defer stmt.Close()

	// Execute the prepared statement with tenant ID
	rows, err := stmt.Query(tenantID, resourceId)
	if err != nil {
		if isDevelopmentEnvironment() {
			log.Printf("Database Err: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	defer rows.Close()

	var conversations []map[string]interface{}
	for rows.Next() {
		var conversation models.Conversation
		err := rows.Scan(
			&conversation.ID,
			&conversation.UserID,
			&conversation.TenantID,
			&conversation.ProjectId,
			&conversation.TemplateId,
			&conversation.CommunityTemplateId,
			&conversation.ConversationConfigurationId,
			&conversation.AgentName,
			&conversation.Title,
			&conversation.ConversationType,
			&conversation.CreatedAt,
			&conversation.UpdatedAt,
			&conversation.Description,
		)
		if err != nil {
			if isDevelopmentEnvironment() {
				log.Printf("Database Err: %v", err)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		conversations = append(conversations, map[string]interface{}{
			"id":                            conversation.ID,
			"user_id":                       conversation.UserID,
			"tenant_id":                     conversation.TenantID,
			"project_id":                    conversation.ProjectId,
			"template_id":                   conversation.TemplateId,
			"community_template_id":         conversation.CommunityTemplateId,
			"conversation_configuration_id": conversation.ConversationConfigurationId,
			"agent_name":                    conversation.AgentName,
			"title":                         conversation.Title,
			"conversation_type":             conversation.ConversationType,
			"created_at":                    conversation.CreatedAt,
			"updated_at":                    conversation.UpdatedAt,
			"description":                   conversation.Description,
		})
	}
	if err = rows.Err(); err != nil {
		if isDevelopmentEnvironment() {
			log.Printf("Database Err: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    conversations,
		"message": models.StatusSuccess,
	})
}

// TODO: We don't need this ?
func UpdateConversation(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	projectId := c.Query("project_id")
	templateId := c.Query("template_id")
	communityTemplateId := c.Query("community_template_id")

	var resourceId string

	if projectId != "" {
		resourceId = projectId
	} else if templateId != "" {
		resourceId = templateId
	} else if communityTemplateId != "" {
		resourceId = communityTemplateId
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Conversation needs to belong to either project or private/community template"})
		return
	}

	conversationID := c.Query("conversation_id")
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Conversation ID is required"})
		return
	}

	// Parse the request body
	var updateReq UpdateConversationResource
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// A transaction is not strictly necessary here since we're only doing one operation
	// However, keeping it consistent with the delete operation which also uses a transaction

	// Prepare the update statement with tenant_id in WHERE and RETURNING clauses
	query := `
		UPDATE
			st_schema.conversation
		SET
			conversation_config_template_id = $1, updated_at = NOW()
		WHERE
			id = $2 AND tenant_id = $3 AND (project_id = $4 OR template_id = $4 OR community_template_id = $4)
		RETURNING
			id,
			project_id,
			template_id,
			community_template_id,
			user_id,
			tenant_id,
			conversation_config_template_id,
			agent_name,
			title,
			conversation_type,
			created_at,
			updated_at,
			description
	`

	// Execute the update statement directly without a transaction
	var conversationObject models.Conversation

	err := tenantManagement.DB.QueryRow(
		query,
		updateReq.PromptConfigurationID,
		conversationID,
		tenantID,
		resourceId,
	).Scan(
		&conversationObject.ID,
		&conversationObject.ProjectId,
		&conversationObject.TemplateId,
		&conversationObject.CommunityTemplateId,
		&conversationObject.UserID,
		&conversationObject.TenantID,
		&conversationObject.ConversationConfigurationId,
		&conversationObject.AgentName,
		&conversationObject.Title,
		&conversationObject.ConversationType,
		&conversationObject.CreatedAt,
		&conversationObject.UpdatedAt,
		&conversationObject.Description,
	)

	if err != nil {
		log.Printf("Error updating conversation: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": models.DatabaseError})
		return
	}

	// Respond with the updated conversation object
	c.JSON(http.StatusOK,
		gin.H{
			"data":    conversationObject,
			"message": models.StatusUpdated,
		})
}

// @Summary Delete a conversation
// @Description Deletes a conversation using the provided tenant ID, project ID, and conversation ID.
// @Tags Conversations
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer [Token]"
// @Param project_id query string true "Project ID"
// @Param conversation_id query string true "Conversation ID"
// @Success 200 {object} Conversation "Conversation deleted successfully"
// @Failure 400 {object} map[string]string "Invalid input data"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/conversation [delete]
func DeleteConversation(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	projectId := c.Query("project_id")
	templateId := c.Query("template_id")
	communityTemplateId := c.Query("community_template_id")

	var resourceId string

	if projectId != "" {
		resourceId = projectId
	} else if templateId != "" {
		resourceId = templateId
	} else if communityTemplateId != "" {
		resourceId = communityTemplateId
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Conversation needs to belong to either project or private/community template"})
		return
	}

	conversationId := c.Query("conversation_id")
	if conversationId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Conversation ID is required"})
		return
	}

	// Prepare the delete statement with tenant_id filter and including tenant_id in RETURNING
	query := `
        DELETE FROM
            st_schema.conversation
        WHERE
            id = $1
        	AND tenant_id = $2
        	AND (project_id = $3 OR template_id = $3 OR community_template_id = $3)
        RETURNING
            id,
            user_id,
            tenant_id,
            project_id,
            template_id,
            community_template_id,
            conversation_config_template_id,
			agent_name,
            title,
            conversation_type,
            created_at,
            updated_at,
            description
    `

	// Execute the delete statement directly without a transaction
	var conversation models.Conversation
	err := tenantManagement.DB.QueryRow(
		query,
		conversationId,
		tenantID,
		resourceId,
	).Scan(
		&conversation.ID,
		&conversation.UserID,
		&conversation.TenantID,
		&conversation.ProjectId,
		&conversation.TemplateId,
		&conversation.CommunityTemplateId,
		&conversation.ConversationConfigurationId,
		&conversation.AgentName,
		&conversation.Title,
		&conversation.ConversationType,
		&conversation.CreatedAt,
		&conversation.UpdatedAt,
		&conversation.Description,
	)

	if err != nil {
		log.Printf("Error deleting conversation: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": models.DatabaseError})
		return
	}

	// Respond with the deleted conversation object
	c.JSON(http.StatusOK, gin.H{
		"data":    conversation,
		"message": models.StatusDeleted,
	})
}

func checkExistingTemplate(sourceID, tenantID string) (string, error) {
	if sourceID == "" {
		return "", nil
	}

	// Prepare the SQL query to check if a template with the given sourceID and tenantID exists
	query := `
        SELECT
			id
        FROM
			st_schema.prompt_config_template
        WHERE
			source_id = $1
		AND
			tenant_id = $2;
    `

	// Execute the query
	var existingTemplateID string
	err := tenantManagement.DB.QueryRow(query, sourceID, tenantID).Scan(&existingTemplateID)
	if err != nil {
		if err == sql.ErrNoRows {
			// No existing template found
			return "", nil
		}
		// Handle other types of errors (e.g., database errors)
		log.Printf("Error checking for existing template: %v", err)
		return "", err
	}

	// Return the ID of the existing template
	return existingTemplateID, nil
}

func createNewTemplateFromSource(sourceID *string, userID, tenantID, templateType string) (*string, error) {
	var query string
	if templateType == "community" {
		query = `
            INSERT INTO st_schema.prompt_config_template
                (user_id, tenant_id, source_id, title, description, category, ai_vendor, ai_model, configuration, privacy, original_publisher, published_by, created_at, updated_at)
			SELECT
				$2, $3, id, title, description, category, ai_vendor, ai_model, configuration, TRUE, user_id, 'communityTemplate', NOW(), NOW()
			FROM
				st_schema.community_prompt_config_templates
			WHERE
				id = $1
            RETURNING id`
	} else { // solutionPilot
		query = `
            INSERT INTO st_schema.prompt_config_template
                (user_id, tenant_id, source_id, title, description, category, ai_vendor, ai_model, configuration, privacy, original_publisher, published_by, created_at, updated_at)
            SELECT
                $2, $3, id, title, description, category, ai_vendor, ai_model, configuration, TRUE, user_id, 'solutionPilotTemplate', NOW(), NOW()
            FROM
                st_schema.sp_prompt_config_templates
            WHERE
                id = $1
            RETURNING id`
	}

	var newConfigId string
	err := tenantManagement.DB.QueryRow(query, sourceID, userID, tenantID).Scan(&newConfigId)
	if err != nil {
		log.Printf("Error creating new template from source: %v", err)
		return nil, err
	}

	return &newConfigId, nil
}

func getConfigData(configurationId string) (*models.TenantAiTemplate, error) {
	// This function needs to be normalized and refactored
	// Right now there is another function like this in the assistant.go handled

	if configurationId == "" {
		return nil, fmt.Errorf("configuration ID is required")
	}

	query := `
        SELECT
            id, title, description, category, ai_vendor, ai_model, configuration,
            privacy, original_publisher, published_by, created_at, updated_at
        FROM
            st_schema.prompt_config_template
        WHERE
            id = $1;
    `

	stmt, err := tenantManagement.DB.Prepare(query)
	if err != nil {
		log.Printf("Database Error: %v", err)
		return nil, fmt.Errorf("database error: %v", err) // Consider returning a generic error for production
	}
	defer stmt.Close()

	var template models.TenantAiTemplate
	var rawConfig []byte // To store raw JSON configuration

	err = stmt.QueryRow(configurationId).Scan(
		&template.ID, &template.Title, &template.Description, &template.Category,
		&template.AiVendor, &template.AiModel, &rawConfig,
		&template.Privacy, &template.OriginalPublisher, &template.PublishedBy,
		&template.CreatedAt, &template.UpdatedAt,
	)

	if err != nil {
		log.Printf("Database Error: %v", err)
		return nil, fmt.Errorf("database error: %v", err)
	}

	var config models.AiConfiguration
	if err := json.Unmarshal(rawConfig, &config); err != nil {
		log.Printf("JSON Unmarshal Error: %v", err)
		return nil, fmt.Errorf("json unmarshal error: %v", err)
	}
	template.Configuration = &config

	return &template, nil
}

func isDevelopmentEnvironment() bool {
	env := os.Getenv("ENVIRONMENT")
	return env == "dev" || env == "local"
}
