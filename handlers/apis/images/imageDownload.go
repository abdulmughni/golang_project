package images

import (
	"fmt"
	"log"
	"net/http"
	"sententiawebapi/handlers/models"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
	"github.com/gin-gonic/gin"
)

func getSignedUrl(containerName string, filename string) (string, error) {
	helper, err := newBlobHelper()
	if err != nil {
		return "", err
	}

	now := time.Now().UTC()
	expiry := now.Add(10 * time.Minute)

	signatureValues := sas.BlobSignatureValues{
		Protocol:      sas.ProtocolHTTPS,
		StartTime:     now.Add(-2 * time.Minute),
		ExpiryTime:    expiry,
		Permissions:   (&sas.BlobPermissions{Read: true}).String(),
		ContainerName: containerName,
		BlobName:      filename,
		Version:       sas.Version,
	}

	blobSAS, err := signatureValues.SignWithSharedKey(helper.Credential)
	if err != nil {
		return "", fmt.Errorf("failed to sign SAS: %w", err)
	}

	signedURL := fmt.Sprintf(
		"https://%s.blob.core.windows.net/%s/%s?%s",
		helper.AccountName,
		containerName,
		filename,
		blobSAS.Encode(),
	)

	return signedURL, nil
}

func GetImageHandler(c *gin.Context) {
	filename := c.Param("filename")

	rawBlobContainerId, exists := c.Get("BlobContainerID")
	if !exists {
		log.Print("Blob container ID not found in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": models.InternalServerError})
	}
	blobContainerID, ok := rawBlobContainerId.(string)
	if !ok {
		log.Print("Blob container ID is not a string")
		c.JSON(http.StatusInternalServerError, gin.H{"error": models.InternalServerError})
		return
	}

	signedURL, err := getSignedUrl(
		blobContainerID,
		filename,
	)
	if err != nil {
		log.Printf("Failed to generate signed URL: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": models.InternalServerError,
		})
	}

	c.Redirect(http.StatusFound, signedURL)
}
