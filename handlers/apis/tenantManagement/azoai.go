// Primary function for enabling new tenants. The intent
// is to:
// 1. Register new tenant \ user in the database
// 2. Create client's Azure Resources
// 3. This function is to be called by Auth0

package tenantManagement

import (
	"log"
	"net/http"

	"sententiawebapi/handlers/models"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/gin-gonic/gin"
)

// 1. Select appropriate subscription based on available quota
// 2. Create Azure Resource Group for tenant if does not exist
// 3. Create Azure Ai Hub for tenant providers/Microsoft.MachineLearningServices/workspaces/
// 4. Create Azure Ai Servies (This is the ML Account) providers/Microsoft.CognitiveServices/accounts
// 5. Create Storage Account
// 6. Create Azure Key Vault
// 8. Create Azure Ai Project  /providers/Microsoft.MachineLearningServices/workspaces/2d5e31ca-de27-4be1
// 7. Create Azure Ai Search I think
// 9. Create Azure Deployment

func NewTenant(c *gin.Context) {

	var tenantOpenAiResource models.TenantOpenAiResource

	if err := c.ShouldBindJSON(&tenantOpenAiResource); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = CreateTenantResourceGroup(tenantOpenAiResource, cred) // Pass cred as a parameter
	if err != nil {
		log.Printf("Resource group validation error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// account, err := CreateAzOaiAccount(tenantOpenAiResource)
	// if err != nil {
	// 	log.Printf("Azure OpenAI account creation error: %v", err)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }

	// // Create the deployment for the OpenAI account
	// deployment, err := CreateAzOaiModelDeployment(tenantOpenAiResource)
	// if err != nil {
	// 	log.Printf("Azure OpenAI model deployment error: %v", err)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }

	// Return a success response
	c.JSON(http.StatusOK, gin.H{
		// "account":    account,
		// "deployment": deployment,
		"message": "Deployment initiated successfully",
	})
}
