package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	appauth "distributed-file-storage/internal/auth"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type AdminFile struct {
	Key        string `json:"key"`
	FileName   string `json:"fileName"`
	OwnerSub   string `json:"ownerSub"`
	OwnerName  string `json:"ownerName"`
	OwnerEmail string `json:"ownerEmail,omitempty"`
	Size       int64  `json:"size"`
}

type DeletedFile struct {
	Key         string `json:"key"`
	FileName    string `json:"fileName"`
	OwnerSub    string `json:"ownerSub"`
	OwnerName   string `json:"ownerName"`
	OwnerEmail  string `json:"ownerEmail,omitempty"`
	DeletedBy   string `json:"deletedBy"`
	DeletedAt   string `json:"deletedAt"`
	OriginalKey string `json:"originalKey"`
	Size        int64  `json:"size"`
}

func userS3Prefix(sub string) string {
	return "users/" + sub + "/files/"
}

func userLocalDirectory(sub string) string {
	return filepath.Join(
		localStorageDir,
		"users",
		sub,
	)
}

func UserFileKey(sub string, fileName string) string {
	return userS3Prefix(sub) + filepath.Base(fileName)
}

func SaveUserFile(
	identity *appauth.Identity,
	fileName string,
	file io.Reader,
) (string, error) {

	fileName = filepath.Base(fileName)

	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	// -------------------------
	// Local device copy
	// -------------------------

	localDir := userLocalDirectory(identity.Sub)

	if err := os.MkdirAll(localDir, 0755); err != nil {
		return "", err
	}

	localPath := filepath.Join(localDir, fileName)

	if err := os.WriteFile(localPath, data, 0644); err != nil {
		return "", err
	}

	// -------------------------
	// AWS S3 copy
	// -------------------------

	client, err := getS3Client()
	if err != nil {
		return "", err
	}

	key := UserFileKey(identity.Sub, fileName)

	metadata := map[string]string{
		"owner-sub":  identity.Sub,
		"owner-name": identity.Name,
	}

	if identity.Email != "" {
		metadata["owner-email"] = identity.Email
	}

	_, err = client.PutObject(
		context.Background(),
		&s3.PutObjectInput{
			Bucket:        aws.String(bucketName),
			Key:           aws.String(key),
			Body:          bytes.NewReader(data),
			ContentLength: aws.Int64(int64(len(data))),
			Metadata:      metadata,
		},
	)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"local://%s + s3://%s/%s",
		localPath,
		bucketName,
		key,
	), nil
}

func ListUserLocalFiles(sub string) ([]string, error) {
	dir := userLocalDirectory(sub)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	files := []string{}

	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	sort.Strings(files)

	return files, nil
}

func ListUserS3Files(sub string) ([]string, error) {
	client, err := getS3Client()
	if err != nil {
		return nil, err
	}

	prefix := userS3Prefix(sub)

	result, err := client.ListObjectsV2(
		context.Background(),
		&s3.ListObjectsV2Input{
			Bucket: aws.String(bucketName),
			Prefix: aws.String(prefix),
		},
	)
	if err != nil {
		return nil, err
	}

	files := []string{}

	for _, object := range result.Contents {
		if object.Key == nil {
			continue
		}

		fileName := strings.TrimPrefix(
			*object.Key,
			prefix,
		)

		if fileName != "" {
			files = append(files, fileName)
		}
	}

	sort.Strings(files)

	return files, nil
}

func DeleteUserLocalFile(
	identity *appauth.Identity,
	fileName string,
) error {

	fileName = filepath.Base(fileName)

	path := filepath.Join(
		userLocalDirectory(identity.Sub),
		fileName,
	)

	err := os.Remove(path)

	if os.IsNotExist(err) {
		return nil
	}

	return err
}

func DeleteUserS3File(
	identity *appauth.Identity,
	fileName string,
) error {

	client, err := getS3Client()
	if err != nil {
		return err
	}

	fileName = filepath.Base(fileName)

	originalKey := UserFileKey(
		identity.Sub,
		fileName,
	)

	object, err := client.GetObject(
		context.Background(),
		&s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(originalKey),
		},
	)
	if err != nil {
		return err
	}
	defer object.Body.Close()

	data, err := io.ReadAll(object.Body)
	if err != nil {
		return err
	}

	deletedAt := time.Now().UTC()

	deletedKey := fmt.Sprintf(
		"deleted/%s/%d-%s",
		identity.Sub,
		deletedAt.UnixNano(),
		fileName,
	)

	deletedBy := identity.Name
	if deletedBy == "" {
		deletedBy = identity.Email
	}

	metadata := map[string]string{
		"owner-sub":    identity.Sub,
		"owner-name":   identity.Name,
		"owner-email":  identity.Email,
		"deleted-by":   deletedBy,
		"deleted-at":   deletedAt.Format(time.RFC3339),
		"original-key": originalKey,
		"file-name":    fileName,
	}

	_, err = client.PutObject(
		context.Background(),
		&s3.PutObjectInput{
			Bucket:        aws.String(bucketName),
			Key:           aws.String(deletedKey),
			Body:          bytes.NewReader(data),
			ContentLength: aws.Int64(int64(len(data))),
			Metadata:      metadata,
		},
	)
	if err != nil {
		return err
	}

	_, err = client.DeleteObject(
		context.Background(),
		&s3.DeleteObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(originalKey),
		},
	)

	return err
}

func AdminListBucketObjects() ([]AdminFile, error) {
	client, err := getS3Client()
	if err != nil {
		return nil, err
	}

	result, err := client.ListObjectsV2(
		context.Background(),
		&s3.ListObjectsV2Input{
			Bucket: aws.String(bucketName),
		},
	)
	if err != nil {
		return nil, err
	}

	files := []AdminFile{}

	for _, object := range result.Contents {
		if object.Key == nil {
			continue
		}

		files = append(files, AdminFile{
			Key:      *object.Key,
			FileName: filepath.Base(*object.Key),
			Size:     aws.ToInt64(object.Size),
		})
	}

	return files, nil
}

func AdminListAllFiles() ([]AdminFile, error) {
	client, err := getS3Client()
	if err != nil {
		return nil, err
	}

	result, err := client.ListObjectsV2(
		context.Background(),
		&s3.ListObjectsV2Input{
			Bucket: aws.String(bucketName),
			Prefix: aws.String("users/"),
		},
	)
	if err != nil {
		return nil, err
	}

	files := []AdminFile{}

	for _, object := range result.Contents {
		if object.Key == nil {
			continue
		}

		head, err := client.HeadObject(
			context.Background(),
			&s3.HeadObjectInput{
				Bucket: aws.String(bucketName),
				Key:    object.Key,
			},
		)
		if err != nil {
			continue
		}

		files = append(files, AdminFile{
			Key:        *object.Key,
			FileName:   filepath.Base(*object.Key),
			OwnerSub:   head.Metadata["owner-sub"],
			OwnerName:  head.Metadata["owner-name"],
			OwnerEmail: head.Metadata["owner-email"],
			Size:       aws.ToInt64(object.Size),
		})
	}

	return files, nil
}

func AdminListDeletedFiles() ([]DeletedFile, error) {
	client, err := getS3Client()
	if err != nil {
		return nil, err
	}

	result, err := client.ListObjectsV2(
		context.Background(),
		&s3.ListObjectsV2Input{
			Bucket: aws.String(bucketName),
			Prefix: aws.String("deleted/"),
		},
	)
	if err != nil {
		return nil, err
	}

	files := []DeletedFile{}

	for _, object := range result.Contents {
		if object.Key == nil {
			continue
		}

		head, err := client.HeadObject(
			context.Background(),
			&s3.HeadObjectInput{
				Bucket: aws.String(bucketName),
				Key:    object.Key,
			},
		)
		if err != nil {
			continue
		}

		files = append(files, DeletedFile{
			Key:         *object.Key,
			FileName:    head.Metadata["file-name"],
			OwnerSub:    head.Metadata["owner-sub"],
			OwnerName:   head.Metadata["owner-name"],
			OwnerEmail:  head.Metadata["owner-email"],
			DeletedBy:   head.Metadata["deleted-by"],
			DeletedAt:   head.Metadata["deleted-at"],
			OriginalKey: head.Metadata["original-key"],
			Size:        aws.ToInt64(object.Size),
		})
	}

	return files, nil
}

func AdminRestoreDeletedFile(deletedKey string) error {
	client, err := getS3Client()
	if err != nil {
		return err
	}

	if !strings.HasPrefix(deletedKey, "deleted/") {
		return fmt.Errorf("invalid deleted file key")
	}

	object, err := client.GetObject(
		context.Background(),
		&s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(deletedKey),
		},
	)
	if err != nil {
		return err
	}
	defer object.Body.Close()

	data, err := io.ReadAll(object.Body)
	if err != nil {
		return err
	}

	originalKey := object.Metadata["original-key"]

	if originalKey == "" {
		return fmt.Errorf("original file location missing")
	}

	metadata := map[string]string{
		"owner-sub":   object.Metadata["owner-sub"],
		"owner-name":  object.Metadata["owner-name"],
		"owner-email": object.Metadata["owner-email"],
	}

	_, err = client.PutObject(
		context.Background(),
		&s3.PutObjectInput{
			Bucket:        aws.String(bucketName),
			Key:           aws.String(originalKey),
			Body:          bytes.NewReader(data),
			ContentLength: aws.Int64(int64(len(data))),
			Metadata:      metadata,
		},
	)
	if err != nil {
		return err
	}

	_, err = client.DeleteObject(
		context.Background(),
		&s3.DeleteObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(deletedKey),
		},
	)

	return err
}
