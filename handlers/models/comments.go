package models

import "time"

type Comment struct {
	ID         string    `json:"id"`
	TemplateID string    `json:"template_id"`
	UserID     string    `json:"user_id"`
	Comment    string    `json:"comment"`
	CreatedAt  time.Time `json:"created_at"`
}
