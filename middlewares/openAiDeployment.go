package middlewares

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/gin-gonic/gin"
)

type DeploymentRequest struct {
	// Define the structure according to Azure OpenAI deployment request requirements
}

func NewOpenAIDeployment(c *gin.Context) {
	// Extract deployment details from the request
	var deployReq DeploymentRequest
	if err := c.BindJSON(&deployReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set up Azure credentials using service principal
	cred, err := azidentity.NewClientSecretCredential(
		os.Getenv("AZURE_TENANT_ID"),
		os.Getenv("AZURE_CLIENT_ID"),
		os.Getenv("AZURE_CLIENT_SECRET"),
		nil,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create credentials"})
		return
	}

	token, err := cred.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: []string{"https://management.azure.com/.default"},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get access token"})
		return
	}

	// Construct the Azure OpenAI API request
	requestURL := "https://management.azure.com/<your-azure-openai-deployment-url>"
	requestBody, err := json.Marshal(deployReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal request body"})
		return
	}

	req, err := http.NewRequest("PUT", requestURL, bytes.NewBuffer(requestBody))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
		return
	}

	req.Header.Add("Authorization", "Bearer "+token.Token)
	req.Header.Add("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send request"})
		return
	}
	defer resp.Body.Close()

	// Handle the response
	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "API request failed"})
		return
	}

	// Return success
	c.JSON(http.StatusOK, gin.H{"message": "Deployment created successfully"})
}
