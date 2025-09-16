package utilities

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sententiawebapi/handlers/models"
)

func GetOpenAiConfig(db *sql.DB, tenantID string) (*models.OpenAiConfig, error) {
	var configJSON string

	err := db.QueryRow(`
		SELECT config_schema
		FROM st_schema.ai_providers
		WHERE tenant_id = $1 AND name = 'openai' AND is_active = true
	`, tenantID).Scan(&configJSON)

	if err != nil {
		log.Printf("Failed to retrieve OpenAI configuration: %v", err)
		return nil, fmt.Errorf("failed to retrieve OpenAI configuration: %v", err)
	}

	// Parse the JSON configuration
	var config models.OpenAiConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		log.Printf("Failed to parse OpenAI configuration: %v", err)
		return nil, fmt.Errorf("failed to parse OpenAI configuration: %v", err)
	}

	return &config, nil
}
