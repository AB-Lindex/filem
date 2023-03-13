package azure

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AB-Lindex/filem/src/storage"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
)

// BlobConfig is the Azure-Blob-config that is imported in the `message`-object in the app
type BlobConfig struct {
	Account    string        `yaml:"account"`
	AccountKey string        `yaml:"key"`
	Container  string        `yaml:"container"`
	SASTimeout time.Duration `yaml:"sasTimeout"`
}

// BlobStorage is the 'api' of the azre-blob-storage
type BlobStorage struct {
	config *BlobConfig
	client *azblob.Client
	sasKey string
}

func (store *BlobStorage) getSAS() (string, error) {
	if store.sasKey != "" {
		return store.sasKey, nil
	}

	types := sas.AccountResourceTypes{Object: true}
	perm := sas.AccountPermissions{Read: true}
	exp := time.Now().Add(store.config.SASTimeout)
	key, err := store.client.ServiceClient().GetSASURL(types, perm, exp, nil)
	if i := strings.IndexRune(key, '?'); i >= 0 {
		key = key[i:]
	}

	store.sasKey = key
	return key, err
}

var (
	errNoCredentials = fmt.Errorf("'key' not specified")
	errNoAccount     = fmt.Errorf("'account' not specified")
	errNoContainer   = fmt.Errorf("'container' not specified")
)

const urlFormat = "https://%s.blob.core.windows.net/%s/%s%s"
const clientFormat = "https://%s.blob.core.windows.net/"

// Connect is used to connect to the blob-storage
func (cfg *BlobConfig) Connect(dryRun bool) (*BlobStorage, error) {
	if cfg.Account == "" {
		return nil, errNoAccount
	}
	if cfg.Container == "" {
		return nil, errNoContainer
	}
	if cfg.AccountKey == "" {
		return nil, errNoCredentials
	}

	cred, err := azblob.NewSharedKeyCredential(cfg.Account, cfg.AccountKey)
	if err != nil {
		return nil, err
	}

	serviceURL := fmt.Sprintf(clientFormat, cfg.Account)
	client, err := azblob.NewClientWithSharedKeyCredential(serviceURL, cred, nil)
	if err != nil {
		return nil, err
	}

	return &BlobStorage{
		config: cfg,
		client: client,
	}, nil
}

// Save a file to the blob-storage
func (store *BlobStorage) Save(name, location, ct string, size int64, dryRun bool) (*storage.StoredObject, error) {
	var sas string

	file, err := os.OpenFile(location, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if dryRun {
		log.Info().Msgf("dry-run: upload %s -> '%s/%s'", location, store.config.Container, name)
	} else {
		sas, err = store.getSAS()
		if err != nil {
			return nil, err
		}

		var ctx = context.Background()
		resp, err := store.client.UploadFile(ctx, store.config.Container, name, file, nil)
		if err != nil {
			return nil, err
		}

		_ = resp
		log.Info().Msgf("azure-blob uploaded '%s/%s'", store.config.Container, name)
	}

	return &storage.StoredObject{
		URL: fmt.Sprintf(urlFormat, store.config.Account, store.config.Container, name, sas),
	}, nil
}
