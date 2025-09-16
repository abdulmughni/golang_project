package images

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"sententiawebapi/handlers/apis/tenantManagement"
	"sententiawebapi/handlers/models"
	"sententiawebapi/utilities"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/h2non/filetype"
)

// NOTE: Currently using per-image signed URLs (SAS) via proxy with short-lived download token.
// This keeps things simple and secure for now.
//
// Future optimizations (if needed):
// - Use tenant-scoped SAS token (like Auth0 access token) to reduce roundtrips.
// - Use a Service Worker to handle caching and token injection outside the DOM.
// These would improve performance but add complexity — revisit if it becomes a bottleneck.

type BlobNameData struct {
	Ext       string
	ID        string
	FullName  string
	ShortName string
}

func validateImage(file multipart.File) error {
	const FILE_HEADER_SIZE = 261

	// Read only magic bytes
	buf := make([]byte, FILE_HEADER_SIZE)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return err
	}

	// Detect file type from buffer
	kind, err := filetype.Match(buf[:n])
	if err != nil {
		return err
	}

	// Accepted MIME types
	allowed := []string{
		"image/png",
		"image/jpeg",
		"image/gif",
		"image/webp",
	}

	valid := false
	for _, v := range allowed {
		if strings.EqualFold(kind.MIME.Value, v) {
			valid = true
			break
		}
	}

	if !valid {
		return errors.New("invalid image type")
	}

	// Rewind so the file can be used again
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	return nil
}

func generateUniqueBlobName(filename string) BlobNameData {
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".png" // fallback
	}
	blobId := uuid.New().String()
	blobName := fmt.Sprintf("img-%s%s", blobId, ext)
	blobShortname := fmt.Sprintf("img-%s%s", blobId[:8], ext)

	return BlobNameData{
		Ext:       ext,
		ID:        blobId,
		FullName:  blobName,
		ShortName: blobShortname,
	}
}

func linkBlobWithDocument(tx *sql.Tx, documentType models.ResourceGroupType, tenantID string, documentID string, containerID string, blobName string) error {
	var tableName string
	switch documentType {
	case models.ResourceGroupProject:
		tableName = "project_documents_files"
	case models.ResourceGroupTemplate:
		tableName = "document_templates_files"
	case models.ResourceGroupCommunity:
		tableName = "cm_document_templates_files"
	default:
		return fmt.Errorf("Unknown document type: %s", documentType)
	}

	query := fmt.Sprintf(`
		INSERT INTO st_schema.%s (
            tenant_id,
			document_id,
			container_id,
			filename
        ) VALUES ($1, $2, $3, $4)
	`, tableName)

	_, err := tx.Exec(query, tenantID, documentID, containerID, blobName)

	if err != nil {
		return fmt.Errorf("failed to link file (%s) with document (%s): %v", blobName, documentID, err)
	}

	return nil
}

func uploadToAzureBlob(file multipart.File, containerID string, blobName BlobNameData) error {
	helper, err := newBlobHelper()
	if err != nil {
		return err
	}

	client, err := helper.newClient()
	if err != nil {
		return err
	}

	containerClient := client.ServiceClient().NewContainerClient(containerID)
	blobClient := containerClient.NewBlockBlobClient(blobName.FullName)

	// Upload stream
	contentType := mime.TypeByExtension(blobName.Ext)
	_, err = blobClient.UploadStream(context.Background(), file, &azblob.UploadStreamOptions{
		BlockSize:   4 * 1024 * 1024, // 4MB chunks
		Concurrency: 3,
		HTTPHeaders: &blob.HTTPHeaders{
			BlobContentType:        &contentType,
			BlobContentDisposition: to.Ptr(fmt.Sprintf(`inline; filename="%s"`, blobName.ShortName)),
		},
	})
	if err != nil {
		return fmt.Errorf("Failed to upload file: %v", err)
	}

	return nil
}

func UploadImageHandler(c *gin.Context) {
	_, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	documentID := c.Query("document_id")
	documentType := c.Query("document_type")
	if documentID == "" || documentType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing document ID or type"})
		return
	}

	const MAX_UPLOAD_SIZE = 15 * 1024 * 1024 // 15 MB
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MAX_UPLOAD_SIZE)

	// Parse uploaded file (named "file" in FormData)
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		// Check is file size exceeded
		if strings.Contains(err.Error(), "http: request body too large") {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "File size is too big. Please make it at most 15MB",
			})
			return
		}

		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Missing or invalid file",
		})
		return
	}
	defer file.Close()

	log.Printf("Received file: %s (%d bytes)", header.Filename, header.Size)

	err = validateImage(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type"})
		return
	}

	blobContainerId, err := getBlobContainerID(tenantID)
	if err != nil {
		log.Print(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": models.InternalServerError})
		return
	}

	// Begin a new transaction
	tx, err := tenantManagement.DB.Begin()
	if err != nil {
		log.Printf("Database Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": models.InternalServerError})
		return
	}
	defer tx.Rollback()

	blobName := generateUniqueBlobName(header.Filename)

	err = linkBlobWithDocument(
		tx,
		models.ResourceGroupType(documentType),
		tenantID,
		documentID,
		blobContainerId.String(),
		blobName.FullName,
	)
	if err != nil {
		log.Print(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": models.InternalServerError,
		})
		tx.Rollback()
		return
	}

	err = uploadToAzureBlob(
		file,
		blobContainerId.String(),
		blobName,
	)
	if err != nil {
		log.Print(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": models.InternalServerError,
		})
		tx.Rollback()
		return
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Database Err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": models.InternalServerError})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url": "/images/" + blobName.FullName,
	})
}
