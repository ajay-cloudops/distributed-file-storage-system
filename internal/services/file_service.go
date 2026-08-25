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

func ListFileVersions(fileName string) ([]storage.FileVersion, error) {
	return storage.ListFileVersions(fileName)
}

func RestoreFileVersion(fileName string, versionID string) error {
	return storage.RestoreFileVersion(fileName, versionID)
}

func ListLocalFiles() ([]string, error) {
	return storage.ListLocalFiles()
}

func ListS3Files() ([]string, error) {
	return storage.ListS3Files()
}

func DeleteLocalFile(fileName string) error {
	return storage.DeleteLocalFile(fileName)
}

func DeleteS3File(fileName string) error {
	return storage.DeleteS3File(fileName)
}
