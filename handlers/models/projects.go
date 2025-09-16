package models

import (
	"encoding/json"
	"time"
)

type Project struct {
	ID               *string `json:"id"`
	UserID           *string `json:"user_id"`
	TenantID         *string `json:"tenant_id"`
	Title            *string `json:"title"`
	Description      *string `json:"description"`
	Complexity       *string `json:"complexity"`
	Status           *string `json:"status"`
	Category         *string `json:"category"`
	CreatedAt        *string `json:"created_at"`
	UpdatedAt        *string `json:"updated_at"`
	UserPicture      *string `json:"user_picture"`
	FirstName        *string `json:"first_name"`
	LastName         *string `json:"last_name"`
	ShortDescription *string `json:"short_description"`
}

type Document struct {
	ID            *string          `json:"id"`
	UserID        *string          `json:"user_id"`
	TenantId      *string          `json:"tenant_id"`
	ProjectID     *string          `json:"project_id"`
	Title         *string          `json:"title"`
	Complexity    *string          `json:"complexity"`
	Content       *json.RawMessage `json:"content_json"`
	RawContent    []byte           `json:"raw_content"`
	PRawContent   []byte           `json:"p_raw_content"`
	CreatedAt     *string          `json:"created_at"`
	UpdatedAt     *string          `json:"updated_at"`
	AiSuggestions *bool            `json:"ai_suggestions"`
	DocumentType  *string          `json:"document_type"`
}

type Conversation struct {
	ID                          string    `json:"id"`
	UserID                      string    `json:"user_id"`
	TenantID                    string    `json:"tenant_id"`
	ProjectId                   *string   `json:"project_id"`
	TemplateId                  *string   `json:"template_id"`
	CommunityTemplateId         *string   `json:"community_template_id"`
	ConversationConfigurationId *string   `json:"conversation_configuration_id"`
	LastChatCompletionId        *string   `json:"last_chat_completion_id"`
	AgentName                   *string   `json:"agent_name"`
	Title                       string    `json:"title"`
	ConversationType            string    `json:"conversation_type"`
	CreatedAt                   time.Time `json:"created_at"`
	UpdatedAt                   time.Time `json:"updated_at"`
	Description                 *string   `json:"description"`
}

// Table st_schema.cm_public_templates_comments
type PublicObjectConversation struct {
	ID         *string `json:"id"`
	TemplateId *string `json:"template_id"` // This is the ID of the template that the conversation belongs to
	UserID     *string `json:"user_id"`
	Comment    *string `json:"comment"`
	CreatedAt  *string `json:"created_at"`
}

type Owner struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Picture   *string `json:"picture"`
}

type Requirement struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	Details    *string `json:"details"`
	Category   *string `json:"category"`
	Status     *string `json:"status"`
	Owner      *string `json:"owner"`
	OwnerData  *Owner  `json:"owner_data"`
	StartDate  *string `json:"start_date"`
	TargetDate *string `json:"target_date"`
}
