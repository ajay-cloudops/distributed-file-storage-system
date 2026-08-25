package storage

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const bucketName = "ajay-distributed-file-storage-2026"
const region = "ap-south-1"
const localStorageDir = "storage"

type FileVersion struct {
	VersionID    string `json:"versionId"`
	IsLatest     bool   `json:"isLatest"`
	LastModified string `json:"lastModified"`
}

func getS3Client() (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, err
	}

	return s3.NewFromConfig(cfg), nil
}

// SaveFile stores the same file in:
// 1. Local storage folder
// 2. AWS S3 bucket
func SaveFile(fileName string, file io.Reader) (string, error) {
	fileName = filepath.Base(fileName)

	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	// Create local storage folder
	err = os.MkdirAll(localStorageDir, 0755)
	if err != nil {
		return "", err
	}

	// Save local copy
	localPath := filepath.Join(localStorageDir, fileName)

	err = os.WriteFile(localPath, data, 0644)
	if err != nil {
		return "", err
	}

	// Save cloud copy in S3
	client, err := getS3Client()
	if err != nil {
		return "", err
	}

	_, err = client.PutObject(
		context.Background(),
		&s3.PutObjectInput{
			Bucket:        aws.String(bucketName),
			Key:           aws.String(fileName),
			Body:          bytes.NewReader(data),
			ContentLength: aws.Int64(int64(len(data))),
		},
	)
	if err != nil {
		return "", err
	}

	return "local://" + localPath + " + s3://" + bucketName + "/" + fileName, nil
}

// DeleteFile removes file from local storage and S3.
func DeleteFile(fileName string) error {
	fileName = filepath.Base(fileName)

	localPath := filepath.Join(localStorageDir, fileName)

	// Remove local copy.
	// Ignore error if local file does not exist.
	if err := os.Remove(localPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	client, err := getS3Client()
	if err != nil {
		return err
	}

	_, err = client.DeleteObject(
		context.Background(),
		&s3.DeleteObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(fileName),
		},
	)

	return err
}

// ListFiles returns unique files found in local storage or S3.
func ListFiles() ([]string, error) {
	fileSet := make(map[string]bool)

	// Local files
	err := os.MkdirAll(localStorageDir, 0755)
	if err != nil {
		return nil, err
	}

	localFiles, err := os.ReadDir(localStorageDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range localFiles {
		if !entry.IsDir() {
			fileSet[entry.Name()] = true
		}
	}

	// S3 files
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

	for _, obj := range result.Contents {
		if obj.Key != nil {
			fileSet[*obj.Key] = true
		}
	}

	files := make([]string, 0, len(fileSet))

	for fileName := range fileSet {
		files = append(files, fileName)
	}

	sort.Strings(files)

	return files, nil
}

// GetFile first checks local storage.
// If missing locally, it downloads from S3 and recreates the local copy.
func GetFile(fileName string) (io.ReadCloser, error) {
	fileName = filepath.Base(fileName)

	err := os.MkdirAll(localStorageDir, 0755)
	if err != nil {
		return nil, err
	}

	localPath := filepath.Join(localStorageDir, fileName)

	// Try local storage first
	localFile, err := os.Open(localPath)
	if err == nil {
		return localFile, nil
	}

	if !os.IsNotExist(err) {
		return nil, err
	}

	// Local file missing -> fetch from S3
	client, err := getS3Client()
	if err != nil {
		return nil, err
	}

	result, err := client.GetObject(
		context.Background(),
		&s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(fileName),
		},
	)
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}

	// Restore local copy automatically
	err = os.WriteFile(localPath, data, 0644)
	if err != nil {
		return nil, err
	}

	return io.NopCloser(bytes.NewReader(data)), nil
}

func DownloadFile(fileName string) (string, error) {
	fileName = filepath.Base(fileName)

	file, err := GetFile(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	err = os.MkdirAll(localStorageDir, 0755)
	if err != nil {
		return "", err
	}

	localPath := filepath.Join(localStorageDir, fileName)

	err = os.WriteFile(localPath, data, 0644)
	if err != nil {
		return "", err
	}

	return localPath, nil
}

func ListFileVersions(fileName string) ([]FileVersion, error) {
	client, err := getS3Client()
	if err != nil {
		return nil, err
	}

	fileName = filepath.Base(fileName)

	result, err := client.ListObjectVersions(
		context.Background(),
		&s3.ListObjectVersionsInput{
			Bucket: aws.String(bucketName),
			Prefix: aws.String(fileName),
		},
	)
	if err != nil {
		return nil, err
	}

	versions := []FileVersion{}

	for _, version := range result.Versions {
		if aws.ToString(version.Key) != fileName {
			continue
		}

		lastModified := ""

		if version.LastModified != nil {
			lastModified = version.LastModified.Format("2006-01-02 15:04:05")
		}

		versions = append(versions, FileVersion{
			VersionID:    aws.ToString(version.VersionId),
			IsLatest:     aws.ToBool(version.IsLatest),
			LastModified: lastModified,
		})
	}

	return versions, nil
}

// RestoreFileVersion restores an old S3 version and
// also updates the local copy with the restored content.
func RestoreFileVersion(fileName string, versionID string) error {
	client, err := getS3Client()
	if err != nil {
		return err
	}

	fileName = filepath.Base(fileName)

	oldVersion, err := client.GetObject(
		context.Background(),
		&s3.GetObjectInput{
			Bucket:    aws.String(bucketName),
			Key:       aws.String(fileName),
			VersionId: aws.String(versionID),
		},
	)
	if err != nil {
		return err
	}
	defer oldVersion.Body.Close()

	data, err := io.ReadAll(oldVersion.Body)
	if err != nil {
		return err
	}

	// Restore S3 version
	_, err = client.PutObject(
		context.Background(),
		&s3.PutObjectInput{
			Bucket:        aws.String(bucketName),
			Key:           aws.String(fileName),
			Body:          bytes.NewReader(data),
			ContentLength: aws.Int64(int64(len(data))),
		},
	)
	if err != nil {
		return err
	}

	// Update local copy
	err = os.MkdirAll(localStorageDir, 0755)
	if err != nil {
		return err
	}

	localPath := filepath.Join(localStorageDir, fileName)

	return os.WriteFile(localPath, data, 0644)
}
