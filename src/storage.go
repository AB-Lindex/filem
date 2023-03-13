package main

import (
	"fmt"

	"github.com/AB-Lindex/filem/src/storage"
	"github.com/AB-Lindex/filem/src/storage/azure"
)

type storageConfig struct {
	Azure *azure.BlobConfig `yaml:"azure"`
}

var (
	errNoStorage = fmt.Errorf("no storage defined")
)

func (stcfg *storageConfig) Connect(dryRun bool) (storage.Storage, error) {
	if stcfg.Azure != nil {
		return stcfg.Azure.Connect(dryRun)
	}
	return nil, errNoStorage
}
