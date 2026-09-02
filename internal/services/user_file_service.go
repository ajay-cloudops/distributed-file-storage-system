package services

import (
	"io"

	appauth "distributed-file-storage/internal/auth"
	"distributed-file-storage/internal/storage"
)

func SaveUserFile(
	user *appauth.Identity,
	fileName string,
	file io.Reader,
) (string, error) {
	return storage.SaveUserFile(user, fileName, file)
}

func ListUserLocalFiles(sub string) ([]string, error) {
	return storage.ListUserLocalFiles(sub)
}

func ListUserS3Files(sub string) ([]string, error) {
	return storage.ListUserS3Files(sub)
}

func DeleteUserLocalFile(
	user *appauth.Identity,
	fileName string,
) error {
	return storage.DeleteUserLocalFile(user, fileName)
}

func DeleteUserS3File(
	user *appauth.Identity,
	fileName string,
) error {
	return storage.DeleteUserS3File(user, fileName)
}

func AdminListAllFiles() ([]storage.AdminFile, error) {
	return storage.AdminListAllFiles()
}

func AdminListDeletedFiles() ([]storage.DeletedFile, error) {
	return storage.AdminListDeletedFiles()
}

func AdminRestoreDeletedFile(key string) error {
	return storage.AdminRestoreDeletedFile(key)
}

func AdminListBucketObjects() ([]storage.AdminFile, error) {
	return storage.AdminListBucketObjects()
}
