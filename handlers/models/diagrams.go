package models

import (
	"encoding/json"
)

// Diagram represents a diagram within a project.
type Diagram struct {
	ID               *string          `json:"id"`
	UserID           *string          `json:"user_id"`
	TenantID         *string          `json:"tenant_id"`
	ProjectID        *string          `json:"project_id"`
	DocumentID       *string          `json:"document_id,omitempty"` // Optional field
	Title            *string          `json:"title"`
	DiagramType      *string          `json:"diagram_type"`
	DiagramStatus    *string          `json:"diagram_status"`
	Category         *string          `json:"category,omitempty"` // Optional field
	Design           *json.RawMessage `json:"design"`
	RawDesign        []byte           `json:"raw_design"`
	CreatedAt        *string          `json:"created_at"`
	UpdatedAt        *string          `json:"updated_at"`
	ShortDescription *string          `json:"short_description"` // Optional field
}
