package images

import (
	"fmt"
	"os"
	"sententiawebapi/handlers/apis/tenantManagement"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/gofrs/uuid"
)

type BlobHelper struct {
	AccountName string
	Credential  *azblob.SharedKeyCredential
}

func newBlobHelper() (*BlobHelper, error) {
	accountName := os.Getenv("STORAGE_ACCOUNT_NAME")
	accountKey := os.Getenv("STORAGE_ACCOUNT_KEY")

	if accountName == "" || accountKey == "" {
		return nil, fmt.Errorf("STORAGE_ACCOUNT_NAME and/or STORAGE_ACCOUNT_KEY are missing")
	}

	cred, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create shared key credential: %w", err)
	}

	return &BlobHelper{
		AccountName: accountName,
		Credential:  cred,
	}, nil
}

func (b *BlobHelper) newClient() (*azblob.Client, error) {
	client, err := azblob.NewClientWithSharedKeyCredential(
		fmt.Sprintf("https://%s.blob.core.windows.net/", b.AccountName),
		b.Credential, nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create blob client: %w", err)
	}
	return client, nil
}

func getBlobContainerID(tenantID string) (*uuid.UUID, error) {
	var containerID uuid.UUID
	err := tenantManagement.DB.QueryRow(`
        SELECT blob_container_id
        FROM st_schema.tenants
        WHERE id = $1
    `, tenantID).Scan(&containerID)

	if err != nil {
		return nil, fmt.Errorf("Failed to get blob container id: %v", err)
	}

	return &containerID, nil
}
