package images

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"sententiawebapi/handlers/apis/tenantManagement"
	"sententiawebapi/handlers/models"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/lib/pq"
)

type blobCopyTask struct {
	SrcURL   string
	DstBlob  *blob.Client
	Filename string
}

func resolveTable(documentType models.ResourceGroupType) (string, error) {
	switch documentType {
	case models.ResourceGroupProject:
		return "project_documents_files", nil
	case models.ResourceGroupTemplate:
		return "document_templates_files", nil
	case models.ResourceGroupCommunity:
		return "cm_document_templates_files", nil
	default:
		return "", fmt.Errorf("unknown document type: %s", documentType)
	}
}

func isPublic(documentType models.ResourceGroupType) bool {
	return documentType == models.ResourceGroupCommunity
}

func parallelCopyBlobs(ctx context.Context, tasks []blobCopyTask) error {
	maxParallelCopies := min(int(math.Ceil(float64(len(tasks))/4.0)), 16)
	sem := make(chan struct{}, maxParallelCopies)
	errCh := make(chan error, 1) // only care about the first error

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup

	for _, task := range tasks {
		wg.Add(1)
		sem <- struct{}{}

		go func(task blobCopyTask) {
			defer wg.Done()
			defer func() { <-sem }()

			select {
			case <-ctx.Done():
				return // don't proceed if already cancelled
			default:
				_, err := task.DstBlob.StartCopyFromURL(ctx, task.SrcURL, nil)
				if err != nil {
					select {
					case errCh <- fmt.Errorf("copy failed for %s: %w", task.Filename, err):
						cancel() // propagate cancellation
					default:
						// do nothing, first error already sent
					}
				}
			}
		}(task)
	}

	wg.Wait()

	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

type DocumentRef struct {
	Type models.ResourceGroupType
	ID   string
}

type CopyParams struct {
	SourceDocument      DocumentRef
	DestinationDocument DocumentRef
	TenantID            string
}

func CopyFiles(ctx context.Context, tx *sql.Tx, params CopyParams) error {
	sourceTable, err := resolveTable(params.SourceDocument.Type)
	if err != nil {
		return err
	}
	destinationTable, err := resolveTable(params.DestinationDocument.Type)
	if err != nil {
		return err
	}

	copyToPublic := isPublic(params.DestinationDocument.Type)
	copyFromPublic := isPublic(params.SourceDocument.Type)

	tenantContainerID, err := getBlobContainerID(params.TenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant container ID: %w", err)
	}

	rows, err := tenantManagement.DB.QueryContext(ctx, fmt.Sprintf(`
		SELECT filename FROM st_schema.%s WHERE document_id = $1 AND tenant_id = $2
	`, sourceTable), params.SourceDocument.ID, params.TenantID)

	if err != nil {
		return fmt.Errorf("query source images: %w", err)
	}
	defer rows.Close()

	helper, err := newBlobHelper()
	if err != nil {
		return err
	}

	storageClient, err := helper.newClient()
	if err != nil {
		return err
	}

	sourceContainer := tenantContainerID.String()
	destinationContainer := tenantContainerID.String()
	if copyToPublic {
		destinationContainer = "public"
	}
	if copyFromPublic {
		sourceContainer = "public"
	}

	srcContainerClient := storageClient.ServiceClient().NewContainerClient(sourceContainer)
	dstContainerClient := storageClient.ServiceClient().NewContainerClient(destinationContainer)

	insertBatch := NewInsertBatch(destinationTable, []string{"tenant_id", "document_id", "container_id", "filename"})

	var copyTasks []blobCopyTask

	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return fmt.Errorf("scan row: %w", err)
		}

		if sourceContainer != destinationContainer {
			srcBlob := srcContainerClient.NewBlobClient(filename)
			dstBlob := dstContainerClient.NewBlobClient(filename)

			copyTasks = append(copyTasks, blobCopyTask{
				SrcURL:   srcBlob.URL(),
				DstBlob:  dstBlob,
				Filename: filename,
			})
		}

		err = insertBatch.Append([]any{params.TenantID, params.DestinationDocument.ID, destinationContainer, filename})
		if err != nil {
			return err
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows iteration error: %w", err)
	}

	var copyErr error
	var insertErr error
	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()
		if len(copyTasks) > 0 {
			copyErr = parallelCopyBlobs(ctx, copyTasks)
		}
	}()

	go func() {
		defer wg.Done()
		if !insertBatch.IsEmpty() {
			insertQuery, arguments := insertBatch.FinalizeQuery()
			_, insertErr = tx.ExecContext(ctx, insertQuery, arguments...)
		}
	}()

	wg.Wait()

	if copyErr != nil {
		return copyErr
	}
	if insertErr != nil {
		return fmt.Errorf("failed to insert files: %w", insertErr)
	}

	return nil
}

type DocumentsRef struct {
	Type models.ResourceGroupType
	IDs  []string
}

type CopyProjectParams struct {
	SourceDocuments      DocumentsRef
	DestinationDocuments DocumentsRef
	TenantID             string
}

func CopyProjectFiles(ctx context.Context, tx *sql.Tx, params CopyProjectParams) error {
	sourceTable, err := resolveTable(params.SourceDocuments.Type)
	if err != nil {
		return err
	}
	destinationTable, err := resolveTable(params.DestinationDocuments.Type)
	if err != nil {
		return err
	}

	copyToPublic := isPublic(params.DestinationDocuments.Type)
	copyFromPublic := isPublic(params.SourceDocuments.Type)

	log.Printf("Copying files from %s to %s", sourceTable, destinationTable)

	if len(params.SourceDocuments.IDs) != len(params.DestinationDocuments.IDs) {
		return fmt.Errorf("source and destination ID lists must have the same length")
	}

	tenantContainerID, err := getBlobContainerID(params.TenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant container ID: %w", err)
	}

	sourceContainer := tenantContainerID.String()
	destinationContainer := tenantContainerID.String()
	if copyToPublic {
		destinationContainer = "public"
	}
	if copyFromPublic {
		sourceContainer = "public"
	}

	helper, err := newBlobHelper()
	if err != nil {
		return err
	}

	storageClient, err := helper.newClient()
	if err != nil {
		return err
	}

	srcContainerClient := storageClient.ServiceClient().NewContainerClient(sourceContainer)
	dstContainerClient := storageClient.ServiceClient().NewContainerClient(destinationContainer)

	// Build a map of source document ID -> destination document ID
	docMap := make(map[string]string)
	for i, srcID := range params.SourceDocuments.IDs {
		docMap[srcID] = params.DestinationDocuments.IDs[i]
	}

	insertBatch := NewInsertBatch(destinationTable, []string{"tenant_id", "document_id", "container_id", "filename"})

	// Query all matching images in a single statement
	query := fmt.Sprintf(`
		SELECT document_id, filename FROM st_schema.%s WHERE document_id = ANY($1) AND tenant_id = $2
	`, sourceTable)

	docRows, err := tenantManagement.DB.QueryContext(ctx, query, pq.Array(params.SourceDocuments.IDs), params.TenantID)
	if err != nil {
		return fmt.Errorf("query source images: %w", err)
	}
	defer docRows.Close()

	var copyTasks []blobCopyTask

	for docRows.Next() {
		var documentID, filename string
		if err := docRows.Scan(&documentID, &filename); err != nil {
			return fmt.Errorf("scan row: %w", err)
		}

		// Only copy if containers differ
		if sourceContainer != destinationContainer {
			srcBlob := srcContainerClient.NewBlobClient(filename)
			dstBlob := dstContainerClient.NewBlobClient(filename)

			copyTasks = append(copyTasks, blobCopyTask{
				SrcURL:   srcBlob.URL(),
				DstBlob:  dstBlob,
				Filename: filename,
			})
		}

		destDocID, ok := docMap[documentID]
		if !ok {
			return fmt.Errorf("no destination document found for source ID: %s", documentID)
		}

		err = insertBatch.Append([]any{params.TenantID, destDocID, destinationContainer, filename})
		if err != nil {
			return err
		}
	}

	if err := docRows.Err(); err != nil {
		return fmt.Errorf("rows iteration error: %w", err)
	}

	var copyErr error
	var insertErr error
	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()
		if len(copyTasks) > 0 {
			copyErr = parallelCopyBlobs(ctx, copyTasks)
		}
	}()

	go func() {
		defer wg.Done()
		if !insertBatch.IsEmpty() {
			insertQuery, arguments := insertBatch.FinalizeQuery()
			_, insertErr = tx.ExecContext(ctx, insertQuery, arguments...)
		}
	}()

	wg.Wait()

	if copyErr != nil {
		return copyErr
	}
	if insertErr != nil {
		return fmt.Errorf("failed to insert files: %w", insertErr)
	}

	return nil
}
