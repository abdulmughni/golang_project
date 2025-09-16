package tenantManagement

import (
	"context"
	"log"
	"time"

	"sententiawebapi/handlers/models"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cognitiveservices/armcognitiveservices"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/gin-gonic/gin"
)

// CreateTenantResourceGroup creates a new resource group for the tenant if it does not exis
func CreateTenantResourceGroup(tenantOpenAiResource models.TenantOpenAiResource, cred azcore.TokenCredential) error {

	// Set a 2 minute timeout for the deployment
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	clientFactory, err := armresources.NewClientFactory(tenantOpenAiResource.SubscriptionID, cred, nil)
	if err != nil {
		return err
	}
	resourceGroupClient := clientFactory.NewResourceGroupsClient()

	// Check if the resource group exists
	response, err := resourceGroupClient.CheckExistence(ctx, tenantOpenAiResource.ResourceGroupName, nil)

	if err == nil && response.Success {
		// Resource group exists
		return nil
	}

	// Resource group does not exist or there was an error, create it
	_, err = resourceGroupClient.CreateOrUpdate(
		ctx,
		tenantOpenAiResource.ResourceGroupName,
		armresources.ResourceGroup{
			Location: to.Ptr(tenantOpenAiResource.Location),
		},
		nil,
	)
	if err != nil {
		return err
	}

	return nil
}

func CreateAzOaiAccount(tenantOpenAiResource models.TenantOpenAiResource) (*armcognitiveservices.Account, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}

	// Set a 2 minute timeout for the deployment
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	clientFactory, err := armcognitiveservices.NewClientFactory(tenantOpenAiResource.SubscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}
	poller, err := clientFactory.NewAccountsClient().BeginCreate(
		ctx,
		tenantOpenAiResource.ResourceGroupName,
		tenantOpenAiResource.OpenAIResourceName, armcognitiveservices.Account{
			Identity: &armcognitiveservices.Identity{
				Type: to.Ptr(armcognitiveservices.ResourceIdentityTypeSystemAssigned),
			},
			Kind:     tenantOpenAiResource.Account.Kind,
			Location: &tenantOpenAiResource.Location,
			SKU: &armcognitiveservices.SKU{
				Name: tenantOpenAiResource.Account.SKU.Name,
			},
			Tags: tenantOpenAiResource.Account.Tags,
		}, nil)
	if err != nil {
		return nil, err
	}
	res, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &res.Account, nil
}

func CreateAzOaiModelDeployment(tenantOpenAiResource models.TenantOpenAiResource) (*armcognitiveservices.Deployment, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute) // Set a 30 minute timeout
	defer cancel()

	clientFactory, err := armcognitiveservices.NewClientFactory(tenantOpenAiResource.SubscriptionID, cred, nil)
	if err != nil {
		log.Printf("Error creating client factory: %v", err)
		return nil, err
	}
	poller, err := clientFactory.NewDeploymentsClient().BeginCreateOrUpdate(
		ctx,
		tenantOpenAiResource.ResourceGroupName,
		tenantOpenAiResource.OpenAIResourceName,
		tenantOpenAiResource.Deployment.Name, armcognitiveservices.Deployment{
			Properties: &armcognitiveservices.DeploymentProperties{
				Model: &armcognitiveservices.DeploymentModel{
					Name:    to.Ptr(tenantOpenAiResource.Deployment.Properties.Model.Name),
					Format:  to.Ptr(tenantOpenAiResource.Deployment.Properties.Model.Format),
					Version: to.Ptr(tenantOpenAiResource.Deployment.Properties.Model.Version),
				},
			},
			SKU: &armcognitiveservices.SKU{
				Name:     to.Ptr(tenantOpenAiResource.Deployment.Properties.Sku.Name),
				Capacity: to.Ptr(tenantOpenAiResource.Deployment.Properties.Sku.Capacity),
			},
		}, nil)
	if err != nil {
		log.Printf("Error creating deployment: %v", err)
		return nil, err
	}
	res, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		log.Printf("Error polling deployment: %v", err)
		return nil, err
	}

	return &res.Deployment, nil
}

func CreateAzBlobStorageAccount(c *gin.Context, tenantOpenAiResource models.TenantOpenAiResource) (*armstorage.Account, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}

	// Set a 2 minute timeout for the deployment
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	clientFactory, err := armstorage.NewClientFactory(tenantOpenAiResource.SubscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}
	accountsClient := clientFactory.NewAccountsClient()

	// Create the storage account
	pollerResp, err := accountsClient.BeginCreate(
		ctx,
		tenantOpenAiResource.ResourceGroupName,
		tenantOpenAiResource.StorageAccount.Name,
		armstorage.AccountCreateParameters{
			Kind: to.Ptr(armstorage.Kind(tenantOpenAiResource.StorageAccount.Kind)),
			SKU: &armstorage.SKU{
				Name: to.Ptr(armstorage.SKUName(tenantOpenAiResource.StorageAccount.SKU)),
			},
			Location: to.Ptr(tenantOpenAiResource.Location),
			Properties: &armstorage.AccountPropertiesCreateParameters{
				AccessTier: to.Ptr(armstorage.AccessTier(tenantOpenAiResource.StorageAccount.AccessTier)),
				Encryption: &armstorage.Encryption{
					Services: &armstorage.EncryptionServices{
						File: &armstorage.EncryptionService{
							KeyType: to.Ptr(armstorage.KeyTypeAccount),
							Enabled: to.Ptr(true),
						},
						Blob: &armstorage.EncryptionService{
							KeyType: to.Ptr(armstorage.KeyTypeAccount),
							Enabled: to.Ptr(true),
						},
						Queue: &armstorage.EncryptionService{
							KeyType: to.Ptr(armstorage.KeyTypeAccount),
							Enabled: to.Ptr(true),
						},
						Table: &armstorage.EncryptionService{
							KeyType: to.Ptr(armstorage.KeyTypeAccount),
							Enabled: to.Ptr(true),
						},
					},
					KeySource: to.Ptr(armstorage.KeySourceMicrosoftStorage),
				},
			},
		}, nil)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}

	// Create the blob container
	blobContainerClient := clientFactory.NewBlobContainersClient()
	_, err = blobContainerClient.Create(
		ctx,
		tenantOpenAiResource.ResourceGroupName,
		tenantOpenAiResource.StorageAccount.Name,
		tenantOpenAiResource.StorageAccount.ContainerName,
		armstorage.BlobContainer{},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &resp.Account, nil
}
