package cloud

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"sententiawebapi/handlers/apis/tenantManagement"
	"sententiawebapi/handlers/models"
	"sententiawebapi/utilities"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

func CreateTenantCredential(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	// Parse request body
	var account models.CloudCredentials
	if err := c.ShouldBindJSON(&account); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create secret name using Azure Key Vault compliant format
	baseName := fmt.Sprintf("creds-%s-%s",
		strings.ToLower(strings.ReplaceAll(account.Name, " ", "-")),
		strings.ReplaceAll(tenantID, "-", ""))

	if len(baseName) > 127 {
		baseName = baseName[:127]
	}
	secretName := baseName
	fmt.Println("Created secret name: ", secretName)

	var cred azcore.TokenCredential
	var err error

	// Check environment
	if os.Getenv("ENV") == "prod" {
		fmt.Println("Local environment detected, using service principal credentials")
		cred, err = azidentity.NewClientSecretCredential(
			os.Getenv("ARM_TENANT_ID"),
			os.Getenv("ARM_CLIENT_ID"),
			os.Getenv("ARM_CLIENT_SECRET"),
			nil,
		)
	} else {
		fmt.Println("Production environment detected, using managed identity")
		clientID := azidentity.ClientID(os.Getenv("AZURE_MANAGED_IDENTITY_CLIENT_ID"))
		opts := &azidentity.ManagedIdentityCredentialOptions{ID: clientID}
		cred, err = azidentity.NewManagedIdentityCredential(opts)
	}

	if err != nil {
		fmt.Println("Failed to create credential:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create credential: %v", err)})
		return
	}
	fmt.Println("Successfully created credential")

	// Create Key Vault client
	keyVaultURL := os.Getenv("AZURE_KEY_VAULT_URL")
	fmt.Println("Attempting to create Key Vault client with URL:", keyVaultURL)
	if keyVaultURL == "" {
		fmt.Println("ERROR: AZURE_KEY_VAULT_URL environment variable is not set!")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Key Vault URL is not configured"})
		return
	}
	client, err := azsecrets.NewClient(keyVaultURL, cred, nil)
	if err != nil {
		fmt.Println("Failed to create Key Vault client:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create Key Vault client: %v", err)})
		return
	}
	fmt.Println("Successfully created Key Vault client")

	// Check if secret already exists
	fmt.Println("Checking if secret exists:", secretName)
	_, err = client.GetSecret(context.Background(), secretName, "", nil)
	if err != nil {
		fmt.Printf("GetSecret error for '%s': %v\n", secretName, err)
		// Check for SecretNotFound or similar error
		if strings.Contains(err.Error(), "SecretNotFound") || strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "could not be found") {
			fmt.Println("Secret does not exist, proceeding with creation:", secretName)
			// continue to creation
		} else {
			fmt.Println("Unexpected error checking secret existence:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to check secret existence: %v", err)})
			return
		}
	} else {
		fmt.Println("Secret exists, fetching credential details:", secretName)
		// Fetch the existing credential reference
		var ref models.CredentialReference
		query := `
			SELECT id, name, provider, key_vault_secret_name, created_at, updated_at
			FROM st_schema.tenant_credentials
			WHERE tenant_id = $1 AND name = $2`

		err = tenantManagement.DB.QueryRow(query, tenantID, account.Name).Scan(
			&ref.ID, &ref.Name, &ref.Provider, &ref.KeyVaultSecretName, &ref.CreatedAt, &ref.UpdatedAt,
		)
		if err != nil {
			fmt.Printf("Failed to fetch credential reference for existing secret '%s': %v\n", secretName, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": fmt.Sprintf("Failed to fetch credential reference: %v", err),
			})
			return
		}

		c.JSON(http.StatusConflict, gin.H{
			"message":    fmt.Sprintf("Cloud credential with name '%s' already exists", account.Name),
			"credential": ref,
		})
		return
	}

	// If error is "secret not found", continue with creation
	if strings.Contains(err.Error(), "SecretNotFound") {
		fmt.Println("Secret does not exist, proceeding with creation:", secretName)
	} else {
		// If error is something else, return error
		fmt.Println("Unexpected error checking secret existence:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to check secret existence: %v", err)})
		return
	}

	// Store credentials as JSON string
	secretValue := fmt.Sprintf(`{
		"clientSecret": "%s",
		"clientId": "%s",
		"tenantId": "%s",
		"subscriptionId": "%s",
		"provider": "%s"
	}`, account.ClientSecret, account.ClientID, account.TenantID, account.SubscriptionID, account.Provider)

	// Set secret in Key Vault
	_, err = client.SetSecret(context.Background(), secretName, azsecrets.SetSecretParameters{
		Value: &secretValue,
	}, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to store credentials: %v", err)})
		return
	}

	// Check if credential already exists for this user and name
	var existingRef models.CredentialReference
	checkQuery := `
		SELECT id, name, provider, key_vault_secret_name, created_at, updated_at
		FROM st_schema.tenant_credentials
		WHERE tenant_id = $1 AND name = $2`

	err = tenantManagement.DB.QueryRow(checkQuery, tenantID, account.Name).Scan(
		&existingRef.ID,
		&existingRef.Name,
		&existingRef.Provider,
		&existingRef.KeyVaultSecretName,
		&existingRef.CreatedAt,
		&existingRef.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Credential doesn't exist, proceed with insertion
		query := `
			INSERT INTO st_schema.tenant_credentials (user_id, tenant_id, name, provider, key_vault_secret_name)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, name, provider, key_vault_secret_name, created_at, updated_at`

		var ref models.CredentialReference
		err = tenantManagement.DB.QueryRow(query, userID, tenantID, account.Name, account.Provider, secretName).Scan(
			&ref.ID, &ref.Name, &ref.Provider, &ref.KeyVaultSecretName, &ref.CreatedAt, &ref.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to store credential reference: %v", err)})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":    "Cloud credentials created successfully",
			"credential": ref,
		})
	} else if err != nil {
		// Some other error occurred
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to check existing credentials: %v", err)})
		return
	} else {
		// Credential exists in DB, check if it exists in Key Vault
		_, err = client.GetSecret(context.Background(), existingRef.KeyVaultSecretName, "", nil)
		if err != nil && strings.Contains(err.Error(), "SecretNotFound") {
			// Secret was deleted from Key Vault, update with new secret
			_, err = client.SetSecret(context.Background(), secretName, azsecrets.SetSecretParameters{
				Value: &secretValue,
			}, nil)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to store credentials: %v", err)})
				return
			}

			// Update the reference in database with new secret name
			updateQuery := `
				UPDATE st_schema.tenant_credentials
				SET key_vault_secret_name = $1, updated_at = NOW()
				WHERE id = $2
				RETURNING id, name, provider, key_vault_secret_name, created_at, updated_at`

			err = tenantManagement.DB.QueryRow(updateQuery, secretName, existingRef.ID).Scan(
				&existingRef.ID,
				&existingRef.Name,
				&existingRef.Provider,
				&existingRef.KeyVaultSecretName,
				&existingRef.CreatedAt,
				&existingRef.UpdatedAt,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update credential reference: %v", err)})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message":    "Cloud credentials created successfully",
				"credential": existingRef,
			})
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to check Key Vault secret: %v", err)})
			return
		} else {
			// Both DB reference and Key Vault secret exist
			c.JSON(http.StatusConflict, gin.H{
				"message":    fmt.Sprintf("Cloud credential with name '%s' already exists", account.Name),
				"credential": existingRef,
			})
		}
	}
}

func CreateTenantCredential2(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	providerType := c.Query("provider")
	if providerType == "" {
		providerType = "azure"
	}

	var secret models.Secret

	switch providerType {
	case "postgres":
		var postgresCredentials models.PostgresCredentials
		if err := c.ShouldBindJSON(&postgresCredentials); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Store credentials as JSON string
		secretValue := fmt.Sprintf(`{
			"username": "%s",
			"password": "%s",
			"host": "%s",
			"port": "%s",
			"database": "%s",
			"sslmode": %t,
			"provider": "%s"
		}`,
			postgresCredentials.Username,
			postgresCredentials.Password,
			postgresCredentials.Host,
			postgresCredentials.Port,
			postgresCredentials.Database,
			postgresCredentials.SSLMode,
			postgresCredentials.Provider,
		)

		secret = models.Secret{
			Provider: postgresCredentials.Provider,
			Name:     postgresCredentials.Name,
			Value:    secretValue,
		}
	default:
		// Parse request body
		var account models.CloudCredentials
		if err := c.ShouldBindJSON(&account); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Store credentials as JSON string
		secretValue := fmt.Sprintf(`{
			"clientSecret": "%s",
			"clientId": "%s",
			"tenantId": "%s",
			"subscriptionId": "%s",
			"provider": "%s"
		}`, account.ClientSecret, account.ClientID, account.TenantID, account.SubscriptionID, account.Provider)

		secret = models.Secret{
			Provider: account.Provider,
			Name:     account.Name,
			Value:    secretValue,
		}
	}

	// Create secret name using Azure Key Vault compliant format
	baseName := fmt.Sprintf("creds-%s-%s",
		strings.ToLower(strings.ReplaceAll(secret.Name, " ", "-")),
		strings.ReplaceAll(tenantID, "-", ""))

	if len(baseName) > 127 {
		baseName = baseName[:127]
	}
	secretName := baseName
	fmt.Println("Created secret name: ", secretName)

	var cred azcore.TokenCredential
	var err error

	// Check environment
	if os.Getenv("ENVIRONMENT") == "local" {
		fmt.Println("Local environment detected, using service principal credentials")
		cred, err = azidentity.NewClientSecretCredential(
			os.Getenv("ARM_TENANT_ID"),
			os.Getenv("ARM_CLIENT_ID"),
			os.Getenv("ARM_CLIENT_SECRET"),
			nil,
		)
	} else {
		fmt.Println("Production environment detected, using managed identity")
		clientID := azidentity.ClientID(os.Getenv("AZURE_MANAGED_IDENTITY_CLIENT_ID"))
		opts := &azidentity.ManagedIdentityCredentialOptions{ID: clientID}
		cred, err = azidentity.NewManagedIdentityCredential(opts)
	}

	if err != nil {
		fmt.Println("Failed to initialize Azure identity credential:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create credential")})
		return
	}
	fmt.Println("Successfully initialized Azure TokenCredential")

	// Create Key Vault client
	keyVaultURL := os.Getenv("AZURE_KEY_VAULT_URL")
	fmt.Println("Attempting to create Key Vault client with URL:", keyVaultURL)
	if keyVaultURL == "" {
		fmt.Println("ERROR: AZURE_KEY_VAULT_URL environment variable is not set!")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create credential"})
		return
	}
	client, err := azsecrets.NewClient(keyVaultURL, cred, nil)
	if err != nil {
		fmt.Println("Failed to create Key Vault client:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create credential"})
		return
	}
	fmt.Println("Successfully created Key Vault client")

	// Start transaction
	tx, err := tenantManagement.DB.Begin()
	if err != nil {
		log.Printf("Failed to start DB transaction: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create credential"})
		return
	}

	// Prepare insert query
	insertQuery := `
		INSERT INTO st_schema.tenant_credentials (
			user_id,
			tenant_id,
			name,
			provider,
			key_vault_secret_name
		)
		VALUES (
			$1, $2, $3, $4, $5
		)
		RETURNING
			id,
			name,
			provider,
			key_vault_secret_name,
			created_at,
			updated_at
	`

	var ref models.CredentialReference

	// Attempt DB insert
	err = tx.QueryRow(
		insertQuery,
		userID,
		tenantID,
		secret.Name,
		secret.Provider,
		secretName,
	).Scan(
		&ref.ID,
		&ref.Name,
		&ref.Provider,
		&ref.KeyVaultSecretName,
		&ref.CreatedAt,
		&ref.UpdatedAt,
	)

	if err != nil {
		tx.Rollback()

		// Try to detect unique violation
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			log.Printf("Duplicate credential reference for tenant: %s, name: %s", tenantID, secret.Name)

			c.JSON(http.StatusConflict, gin.H{
				"error": fmt.Sprintf("A credential with name '%s' already exists.", secret.Name),
			})
			return
		}

		log.Printf("Failed to insert credential reference (tenant: %s, name: %s): %v", tenantID, secret.Name, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create credential",
		})
		return
	}

	fmt.Printf("Successfully inserted credential reference for secret '%s'\n", secretName)

	// Now try to store secret in Azure Key Vault
	fmt.Printf("Storing secret '%s' in Key Vault...\n", secretName)

	_, err = client.SetSecret(context.Background(), secretName, azsecrets.SetSecretParameters{
		Value: &secret.Value,
		Tags: map[string]*string{
			"provider": to.Ptr(providerType),
			"env":      to.Ptr(os.Getenv("ENVIRONMENT")),
		},
	}, nil)

	if err != nil {
		fmt.Printf("Failed to store secret in Key Vault: %v\n", err)
		tx.Rollback()

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create credential",
		})
		return
	}

	fmt.Printf("Successfully stored secret '%s' in Key Vault\n", secretName)

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		log.Printf("Failed to commit transaction: %v\n", err)

		// Attempt to delete the secret from Key Vault
		_, delErr := client.DeleteSecret(context.Background(), secretName, nil)
		if delErr != nil {
			log.Printf("WARNING: Failed to delete orphaned secret '%s' from Key Vault: %v\n", secretName, delErr)
		} else {
			fmt.Printf("Rolled back Key Vault secret '%s' after failed DB commit\n", secretName)
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to finalize credential creation",
		})
		return
	}

	// All done
	c.JSON(http.StatusOK, gin.H{
		"message":    "Credential created successfully",
		"credential": ref,
	})
}

func GetTenantCredentials(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	// Query database for user's credentials
	query := `
		SELECT id, name, provider, key_vault_secret_name, created_at, updated_at
		FROM st_schema.tenant_credentials
		WHERE tenant_id = $1`

	rows, err := tenantManagement.DB.Query(query, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch credentials: %v", err)})
		return
	}
	defer rows.Close()

	var credentials []models.CredentialReference
	for rows.Next() {
		var cred models.CredentialReference
		err := rows.Scan(
			&cred.ID,
			&cred.Name,
			&cred.Provider,
			&cred.KeyVaultSecretName,
			&cred.CreatedAt,
			&cred.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to scan credential: %v", err)})
			return
		}
		credentials = append(credentials, cred)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Cloud credentials retrieved successfully",
		"credentials": credentials,
	})
}

func DeleteTenantCredential(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	// Get credential ID from URL parameter
	credentialID := c.Query("id")
	if credentialID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Credential ID is required"})
		return
	}

	// First, get the credential details from the database
	var credential models.CredentialReference
	query := `
		SELECT id, name, provider, key_vault_secret_name, created_at, updated_at
		FROM st_schema.tenant_credentials
		WHERE id = $1 AND tenant_id = $2`

	err := tenantManagement.DB.QueryRow(query, credentialID, tenantID).Scan(
		&credential.ID,
		&credential.Name,
		&credential.Provider,
		&credential.KeyVaultSecretName,
		&credential.CreatedAt,
		&credential.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Credential not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch credential: %v", err)})
		return
	}

	// Create Azure credential based on environment
	var cred azcore.TokenCredential
	if os.Getenv("ENVIRONMENT") == "prod" {
		fmt.Println("Local environment detected, using service principal credentials")
		cred, err = azidentity.NewClientSecretCredential(
			os.Getenv("ARM_TENANT_ID"),
			os.Getenv("ARM_CLIENT_ID"),
			os.Getenv("ARM_CLIENT_SECRET"),
			nil,
		)
	} else {
		fmt.Println("Production environment detected, using managed identity")
		clientID := azidentity.ClientID(os.Getenv("AZURE_MANAGED_IDENTITY_CLIENT_ID"))
		opts := &azidentity.ManagedIdentityCredentialOptions{ID: clientID}
		cred, err = azidentity.NewManagedIdentityCredential(opts)
	}

	if err != nil {
		fmt.Println("Failed to create credential:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create credential: %v", err)})
		return
	}

	// Create Key Vault client
	client, err := azsecrets.NewClient(os.Getenv("AZURE_KEY_VAULT_URL"), cred, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create Key Vault client: %v", err)})
		return
	}

	// Delete from Key Vault
	_, err = client.DeleteSecret(context.Background(), credential.KeyVaultSecretName, nil)
	if err != nil && !strings.Contains(err.Error(), "SecretNotFound") {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete from Key Vault: %v", err)})
		return
	}

	// Delete from database
	deleteQuery := `
		DELETE FROM st_schema.tenant_credentials
		WHERE id = $1 AND tenant_id = $2`

	result, err := tenantManagement.DB.Exec(deleteQuery, credentialID, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete from database: %v", err)})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get rows affected: %v", err)})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Credential not found"})
		return
	}

	// Return success with the deleted credential details
	c.JSON(http.StatusOK, gin.H{
		"message":    "Cloud credential deleted successfully",
		"credential": credential,
	})
}
