package services

import (
	"io"

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

func GetFile(fileName string) (io.ReadCloser, error) {
	return storage.GetFile(fileName)
}
