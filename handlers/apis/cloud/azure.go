package cloud

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"sententiawebapi/handlers/apis/tenantManagement"
	"sententiawebapi/handlers/models"
	"sententiawebapi/utilities"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/gin-gonic/gin"
)

// getAzureCredential returns the appropriate Azure credential based on environment
func getAzureCredential() (azcore.TokenCredential, error) {
	env := os.Getenv("ENVIRONMENT")

	switch env {
	case "production", "prod":
		// Production environment - use managed identity
		clientID := azidentity.ClientID(os.Getenv("AZURE_MANAGED_IDENTITY_CLIENT_ID"))
		opts := &azidentity.ManagedIdentityCredentialOptions{ID: clientID}
		return azidentity.NewManagedIdentityCredential(opts)

	case "local", "dev", "development":
		// Development environments - use service principal
		return azidentity.NewClientSecretCredential(
			os.Getenv("ARM_TENANT_ID"),
			os.Getenv("ARM_CLIENT_ID"),
			os.Getenv("ARM_CLIENT_SECRET"),
			nil,
		)

	default:
		return nil, fmt.Errorf("unknown environment: %s. Must be one of: prod, production, local, dev, development", env)
	}
}

func GetAzureResourcesByTag(c *gin.Context) {
	// Get the raw tag string from query parameter, e.g. "environment=production"
	rawTag := c.Query("tag")
	if rawTag == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tag parameter is required"})
		return
	}

	// Get credential ID from query parameter
	credentialID := c.Query("credential_id")
	if credentialID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Credential ID is required"})
		return
	}

	// Get both userID and tenantID from context
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	// Get credential from database
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
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Credential not found or access denied"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch credential: %v", err)})
		return
	}

	// Create Azure credential
	azCred, err := getAzureCredential()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create Azure credential: %v", err)})
		return
	}

	// Create Key Vault client
	kvClient, err := azsecrets.NewClient(os.Getenv("AZURE_KEY_VAULT_URL"), azCred, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create Key Vault client: %v", err)})
		return
	}

	// Get secret from Key Vault
	secret, err := kvClient.GetSecret(context.Background(), credRef.KeyVaultSecretName, "", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get secret from Key Vault: %v", err)})
		return
	}

	// Parse secret JSON
	var creds struct {
		ClientSecret   string `json:"clientSecret"`
		ClientID       string `json:"clientId"`
		TenantID       string `json:"tenantId"`
		SubscriptionID string `json:"subscriptionId"`
		Provider       string `json:"provider"`
	}
	if err := json.Unmarshal([]byte(*secret.Value), &creds); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to parse credentials: %v", err)})
		return
	}

	// Create client with retrieved credentials
	clientCred, err := azidentity.NewClientSecretCredential(creds.TenantID, creds.ClientID, creds.ClientSecret, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create client credential: %v", err)})
		return
	}

	// Create client with retrieved credentials
	client, err := armresources.NewClient(creds.SubscriptionID, clientCred, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create client: %v", err)})
		return
	}

	// Parse tag string
	parts := strings.SplitN(rawTag, "=", 2)
	if len(parts) != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tag query must be in the format 'key=value'"})
		return
	}
	tagName := parts[0]
	tagValue := parts[1]

	// Construct the filter string for tag filtering
	// e.g. "tagName eq 'environment' and tagValue eq 'production'"
	filterStr := fmt.Sprintf("tagName eq '%s' and tagValue eq '%s'", tagName, tagValue)

	pager := client.NewListPager(&armresources.ClientListOptions{
		Filter: &filterStr,
	})

	var resources []map[string]interface{}
	for pager.More() {
		page, err := pager.NextPage(context.Background())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to get resources: %v", err),
			})
			return
		}

		for _, resource := range page.Value {
			// Find matching icon URL from AzureNodes
			iconURL := findIconURLForResourceType(safeString(resource.Type))

			resourceMap := map[string]interface{}{
				"id":            safeString(resource.ID),
				"name":          safeString(resource.Name),
				"type":          safeString(resource.Type),
				"kind":          safeString(resource.Kind),
				"location":      safeString(resource.Location),
				"tags":          resource.Tags,
				"properties":    resource.Properties,
				"iconURL":       iconURL,
				"complete_data": resource,
			}
			resources = append(resources, resourceMap)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": resources,
	})
}

// findIconURLForResourceType searches through AzureNodes to find matching resource type and returns its icon URL
func findIconURLForResourceType(resourceType string) string {
	for _, category := range AzureNodes {
		for _, item := range category.Items {
			if item.Type == resourceType {
				return item.URL
			}
		}
	}
	return "" // Return empty string if no matching icon is found
}

func safeString(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}
