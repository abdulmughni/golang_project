package cloud

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"sententiawebapi/handlers/apis/tenantManagement"
	"sententiawebapi/handlers/models"

	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

func GetPostgresCredentials(tenantID, credentialID string) (*models.PostgresCredentials, error) {
	var credRef models.CredentialReference
	query := `
		SELECT
			id, name, provider, key_vault_secret_name
		FROM
			st_schema.tenant_credentials
		WHERE
			id = $1
		AND
			tenant_id = $2`

	err := tenantManagement.DB.QueryRow(query, credentialID, tenantID).Scan(
		&credRef.ID,
		&credRef.Name,
		&credRef.Provider,
		&credRef.KeyVaultSecretName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("credential not found or access denied")
		}
		return nil, fmt.Errorf("failed to fetch credential: %w", err)
	}

	// Create Azure credential
	azCred, err := getAzureCredential()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Azure identity credential: %w", err)
	}

	// Create Key Vault client
	keyVaultURL := os.Getenv("AZURE_KEY_VAULT_URL")
	kvClient, err := azsecrets.NewClient(keyVaultURL, azCred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Key Vault client: %w", err)
	}

	// Get secret from Key Vault
	secret, err := kvClient.GetSecret(context.Background(), credRef.KeyVaultSecretName, "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret from Key Vault: %w", err)
	}

	// Parse secret JSON into PostgresCredentials
	var creds models.PostgresCredentials
	if err := json.Unmarshal([]byte(*secret.Value), &creds); err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	return &creds, nil
}

func ConnectToPostgres(tenantID, credentialID string) (*sql.DB, error) {
	creds, err := GetPostgresCredentials(tenantID, credentialID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Postgres credentials: %v", err)
	}

	sslMode := "require"
	if !creds.SSLMode {
		sslMode = "prefer"
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		creds.Host,
		creds.Port,
		creds.Username,
		creds.Password,
		creds.Database,
		sslMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	return db, nil
}
