package storage

import (
	"io"
	"os"
	"path/filepath"
)

func SaveFile(fileName string, file io.Reader) (string, error) {
	storageDir := "storage"

	err := os.MkdirAll(storageDir, 0755)
	if err != nil {
		return "", err
	}

	filePath := filepath.Join(storageDir, filepath.Base(fileName))

	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

func DeleteFile(fileName string) error {
	filePath := filepath.Join("storage", filepath.Base(fileName))

	return os.Remove(filePath)
}

func ListFiles() ([]string, error) {
	storageDir := "storage"

	err := os.MkdirAll(storageDir, 0755)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(storageDir)
	if err != nil {
		return nil, err
	}

	files := []string{}

	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}
