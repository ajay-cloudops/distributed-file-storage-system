package services

import (
	"io"
	"path/filepath"

	"distributed-file-storage/internal/storage"
)

func UploadFile(fileName string, file io.Reader) (string, error) {
	return storage.SaveFile(fileName, file)
}

func DeleteFile(fileName string) error {
	return storage.DeleteFile(fileName)
}

func ListFiles() ([]string, error) {
	return storage.ListFiles()
}

func GetFilePath(fileName string) string {
	return filepath.Join("storage", filepath.Base(fileName))
}
