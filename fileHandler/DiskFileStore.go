package fileHandler

import (
	"bytes"
	"fmt"
	"gbb.go/gvp/utils"
	"os"
	"sync"
)

type DiskFileStore struct {
	mutex      sync.RWMutex
	fileFolder string
}

type FileInfo struct {
	fileName      string
	fileType      string
	fileExtension string
	fileSize      int
	checksum      string
}

func NewDiskFileStore(fileFolder string) *DiskFileStore {
	return &DiskFileStore{
		fileFolder: fileFolder,
	}
}

func (store *DiskFileStore) SaveToDisk(fileName string, fileExtension string, fileData bytes.Buffer) (string, error) {
	fileID := utils.GenerateUUID()
	name := fmt.Sprintf("%s_%s.%s", fileID, fileName, fileExtension)

	filePath := fmt.Sprintf("%s/%s", store.fileFolder, name)

	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("cannot create file: %w", err)
	}

	_, err = fileData.WriteTo(file)
	if err != nil {
		return "", fmt.Errorf("cannot write data to file: %w", err)
	}

	return name, nil
}

func (store *DiskFileStore) SaveToDiskV2(fileName string, fileData bytes.Buffer) (string, error) {
	filePath := fmt.Sprintf("%s/%s", store.fileFolder, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("cannot create file: %w", err)
	}

	_, err = fileData.WriteTo(file)
	if err != nil {
		return "", fmt.Errorf("cannot write data to file: %w", err)
	}

	return fileName, nil
}
