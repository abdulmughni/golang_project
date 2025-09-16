package models

import (
	"encoding/json"
	"time"
)

// Private Templates
type ProjectTemplate struct {
	ID                *string    `json:"id"`
	UserID            *string    `json:"user_id"`
	TenantID          *string    `json:"tenant_id"`
	Title             *string    `json:"title"`
	Description       *string    `json:"description"`
	Category          *string    `json:"category"`
	DocumentTemplates *[]string  `json:"document_templates"`
	Complexity        *string    `json:"complexity"`
	CreatedAt         *time.Time `json:"created_at"`
	UpdatedAt         *time.Time `json:"updated_at"`
	Privacy           *bool      `json:"privacy"`
	Status            *string    `json:"status"`
	PublicTemplateRef *string    `json:"public_template_ref"` // This value is added
	DiagramTemplates  *[]string  `json:"diagram_templates,omitempty"`
	FirstName         *string    `json:"first_name"`
	LastName          *string    `json:"last_name"`
	UserPicture       *string    `json:"user_picture"`
	ShortDescription  *string    `json:"short_description"`
}

type DocumentTemplate struct {
	ID                *string          `json:"id"`
	UserID            *string          `json:"user_id"`
	TenantID          *string          `json:"tenant_id"`
	ProjectTemplateID *string          `json:"project_template_id"`
	Title             *string          `json:"title"`
	Complexity        *string          `json:"complexity"`
	Content           *json.RawMessage `json:"content"`
	RawContent        []byte           `json:"raw_content"`
	CreatedAt         *string          `json:"created_at"`
	UpdatedAt         *string          `json:"updated_at"`
	Privacy           *bool            `json:"privacy"`
	Category          *string          `json:"category"`
	Description       *string          `json:"description"`
	AiSuggestions     *bool            `json:"ai_suggestions"`
	FirstName         *string          `json:"first_name"`
	LastName          *string          `json:"last_name"`
	UserPicture       *string          `json:"user_picture"`
	DocumentType      *string          `json:"document_type"`
}

type DiagramTemplate struct {
	ID                *string          `json:"id"`
	UserID            *string          `json:"user_id"`
	TenantID          *string          `json:"tenant_id"`
	ProjectTemplateID *string          `json:"project_template_id"`
	Title             *string          `json:"title"`
	ShortDescription  *string          `json:"short_description"`
	DiagramType       *string          `json:"diagram_type"`
	DiagramStatus     *string          `json:"diagram_status"`
	Category          *string          `json:"category"`
	Design            *json.RawMessage `json:"design"`
	RawDesign         []byte           `json:"raw_design"`
	CreatedAt         *time.Time       `json:"created_at"`
	UpdatedAt         *time.Time       `json:"updated_at"`
	FirstName         *string          `json:"first_name"`
	LastName          *string          `json:"last_name"`
	UserPicture       *string          `json:"user_picture"`
}

// Public Templates
// Table st_schema.cm_project_templates
type PublicProjectTemplate struct {
	ID                *string    `json:"id"`                  // This is the ID of the project template
	ProjectTemplateID *string    `json:"project_template_id"` // This is the source project id
	UserID            *string    `json:"user_id"`             // Owner ID
	TenantID          *string    `json:"tenant_id"`           // Tenant ID
	Title             *string    `json:"title"`
	Version           *string    `json:"version"`
	Category          *string    `json:"category"`
	Description       *string    `json:"description"`
	DocumentTemplates *[]string  `json:"document_templates"`
	DiagramTemplates  *[]string  `json:"diagram_templates"`
	Complexity        *string    `json:"complexity"`
	CreatedAt         *time.Time `json:"created_at"`
	PublishedAt       *time.Time `json:"published_at"`
	LastUpdateAt      *time.Time `json:"last_update_at"`
	FirstName         *string    `json:"first_name"`
	LastName          *string    `json:"last_name"`
	UserPicture       *string    `json:"user_picture"`
	ShortDescription  *string    `json:"short_description"`
}

// Table st_schema.cm_document_templates
type PublicDocumentTemplate struct {
	ID                         *string          `json:"id"`
	CommunityProjectTemplateID *string          `json:"community_project_template_id"` // The ID of the associated project template
	UserID                     *string          `json:"user_id"`
	TenantID                   *string          `json:"tenant_id"`
	ProjectTemplateID          *string          `json:"project_template_id"`
	Title                      *string          `json:"title"`
	Description                *string          `json:"description"`
	Content                    *json.RawMessage `json:"content"`
	RawContent                 []byte           `json:"raw_content"`
	Complexity                 *string          `json:"complexity"`
	PublishedAt                *string          `json:"published_at"`
	UpdatedAt                  *string          `json:"updated_at"`
	Category                   *string          `json:"category"`
	LastUpdateAt               *string          `json:"last_update"`
	FirstName                  *string          `json:"first_name"`
	LastName                   *string          `json:"last_name"`
	UserPicture                *string          `json:"user_picture"`
	DocumentType               *string          `json:"document_type"`
}

// Table st_schema.cm_diagram_templates
type PublicDiagramTemplate struct {
	ID                         *string          `json:"id"`
	CommunityProjectTemplateID *string          `json:"community_project_template_id"`
	UserID                     *string          `json:"user_id"`
	TenantID                   *string          `json:"tenant_id"`
	ProjectTemplateID          *string          `json:"project_template_id"`
	Title                      *string          `json:"title"`
	ShortDescription           *string          `json:"short_description"`
	DiagramType                *string          `json:"diagram_type"`
	DiagramStatus              *string          `json:"diagram_status"`
	Category                   *string          `json:"category"`
	Design                     *json.RawMessage `json:"design"`
	RawDesign                  []byte           `json:"raw_design"`
	CreatedAt                  *string          `json:"created_at"`
	UpdatedAt                  *string          `json:"updated_at"`
	FirstName                  *string          `json:"first_name"`
	LastName                   *string          `json:"last_name"`
	UserPicture                *string          `json:"user_picture"`
}

type DocumentComponents struct {
	ID               *string `json:"id"`
	Owner            *string `json:"owner"`
	TenantID         *string `json:"tenant_id"`
	Title            *string `json:"title"`
	Category         *string `json:"category"`
	ShortDescription *string `json:"short_description"`
	Icon             *string `json:"icon"`
	IsFavorite       *bool   `json:"is_favorite"`
	FirstName        *string `json:"first_name"`
	LastName         *string `json:"last_name"`
	UserPicture      *string `json:"user_picture"`
}

type DocumentComponent struct {
	ID               *string   `json:"id"`
	Owner            *string   `json:"owner"`
	TenantID         *string   `json:"tenant_id"`
	Title            *string   `json:"title"`
	Category         *string   `json:"category"`
	Description      *string   `json:"description"`
	ShortDescription *string   `json:"short_description"`
	Content          *string   `json:"content"`
	Icon             *string   `json:"icon"`
	CreatedAt        time.Time `json:"created_at"`
	LastUpdateAt     time.Time `json:"last_update_at"`
	IsFavorite       *bool     `json:"is_favorite"` // New field
	FirstName        *string   `json:"first_name"`
	LastName         *string   `json:"last_name"`
	UserPicture      *string   `json:"user_picture"`
}
