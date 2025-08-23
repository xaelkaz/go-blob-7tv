// services/storage/storage.go
package storage

import (
	"context"
	"log"
	"strings"
	"sync"

	"gokeki/config"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
)

var (
	serviceClient *azblob.Client
	accountName   string
	initOnce      sync.Once
	available     bool
)

func InitAzureStorage() bool {
	initOnce.Do(func() {
		cfg := config.LoadConfig()

		if cfg.AzureConnStr == "" {
			log.Println("⚠️  Azure Storage disabled (no connection string)")
			return
		}

		var err error
		serviceClient, err = azblob.NewClientFromConnectionString(cfg.AzureConnStr, nil)
		if err != nil {
			log.Printf("❌ Failed to create Azure client: %v", err)
			return
		}

		accountName = getAccountNameFromConnStr(cfg.AzureConnStr)
		if accountName == "" {
			log.Println("❌ Failed to parse account name from connection string")
			return
		}

		available = true
		log.Printf("✅ Azure Storage initialized (Account: %s, Container: %s)", accountName, cfg.ContainerName)
	})
	return available
}

func getAccountNameFromConnStr(connStr string) string {
	parts := strings.Split(connStr, ";")
	for _, part := range parts {
		if strings.HasPrefix(part, "AccountName=") {
			return strings.TrimPrefix(part, "AccountName=")
		}
	}
	return ""
}

func AzureStorageAvailable() bool {
	return InitAzureStorage()
}

func ContainerURL() string {
	return "https://" + accountName + ".blob.core.windows.net/" + config.LoadConfig().ContainerName
}

func UploadToAzureBlob(fileData []byte, blobName string, contentType string) (string, error) {
	if !AzureStorageAvailable() {
		return "", nil
	}

	cfg := config.LoadConfig()

	// Check existence with a head request (download with range 0-0)
	_, err := serviceClient.DownloadStream(context.Background(), cfg.ContainerName, blobName, &azblob.DownloadStreamOptions{
		Range: azblob.HTTPRange{Count: 1},
	})
	if err == nil {
		url := "https://" + accountName + ".blob.core.windows.net/" + cfg.ContainerName + "/" + blobName
		return url, nil // Already exists
	}

	if !bloberror.HasCode(err, bloberror.BlobNotFound) {
		return "", err
	}

	// Upload if not exists
	_, err = serviceClient.UploadBuffer(context.Background(), cfg.ContainerName, blobName, fileData, &azblob.UploadBufferOptions{
		HTTPHeaders: &blob.HTTPHeaders{
			BlobContentType: to.Ptr(contentType),
		},
	})
	if err != nil {
		return "", err
	}

	url := "https://" + accountName + ".blob.core.windows.net/" + cfg.ContainerName + "/" + blobName
	return url, nil
}

func ListBlobsWithPrefix(prefix string) ([]*container.BlobItem, error) {
	if !AzureStorageAvailable() {
		return nil, nil
	}

	var blobs []*container.BlobItem
	pager := serviceClient.NewListBlobsFlatPager(config.LoadConfig().ContainerName, &azblob.ListBlobsFlatOptions{
		Prefix: to.Ptr(prefix),
	})

	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, err
		}
		blobs = append(blobs, resp.Segment.BlobItems...)
	}
	return blobs, nil
}
