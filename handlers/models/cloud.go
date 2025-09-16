package models

import (
	"time"

	"github.com/google/uuid"
)

type CloudProvider string

const (
	ProviderAzure    CloudProvider = "azure"
	ProviderAWS      CloudProvider = "aws"
	ProviderGCP      CloudProvider = "gcp"
	ProviderPostgres CloudProvider = "postgres"
)

// CloudCredentials represents the base credentials structure for most providers
type CloudCredentials struct {
	Name           string        `json:"name" binding:"required"`
	Provider       CloudProvider `json:"provider" binding:"required"`
	ClientSecret   string        `json:"client_secret" binding:"required"`
	ClientID       string        `json:"client_id" binding:"required"`
	TenantID       string        `json:"tenant_id" binding:"required"`
	SubscriptionID string        `json:"subscription_id,omitempty"` // Only required for Azure
}

type PostgresCredentials struct {
	Name     string        `json:"name" binding:"required"`
	Provider CloudProvider `json:"provider" binding:"required"`
	Username string        `json:"username" binding:"required"`
	Password string        `json:"password" binding:"required"`
	Host     string        `json:"host" binding:"required"`
	Port     string        `json:"port" binding:"required"`
	Database string        `json:"database" binding:"required"`
	SSLMode  bool          `json:"sslmode"` // true = require, false = disable
}

// CredentialReference represents the stored reference to cloud credentials
type CredentialReference struct {
	ID                 uuid.UUID     `json:"id"`
	Name               string        `json:"name"`
	Provider           CloudProvider `json:"provider"`
	KeyVaultSecretName string        `json:"key_vault_secret_name"`
	CreatedAt          time.Time     `json:"created_at"`
	UpdatedAt          time.Time     `json:"updated_at"`
}

type AzureCredentials struct {
	Name           string `json:"name" binding:"required"`
	ClientSecret   string `json:"client_secret" binding:"required"`
	ClientID       string `json:"client_id" binding:"required"`
	TenantID       string `json:"tenant_id" binding:"required"`
	SubscriptionID string `json:"subscription_id" binding:"required"`
}

// AzureResource represents an Azure resource with its metadata
type AzureResource struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Kind       string                 `json:"kind"`
	Location   string                 `json:"location"`
	Tags       map[string]string      `json:"tags"`
	Properties map[string]interface{} `json:"properties"`
	IconURL    string                 `json:"iconURL"`
}

type Secret struct {
	Provider CloudProvider
	Name     string
	Value    string
}
