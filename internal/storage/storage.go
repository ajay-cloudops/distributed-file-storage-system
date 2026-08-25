package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const bucketName = "ajay-distributed-file-storage-2026"
const region = "ap-south-1"

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

func SaveFile(fileName string, file io.Reader) (string, error) {
	client, err := getS3Client()
	if err != nil {
		return "", err
	}

	fileName = filepath.Base(fileName)

	_, err = client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName),
		Body:   file,
	})
	if err != nil {
		return "", err
	}

	return "s3://" + bucketName + "/" + fileName, nil
}

func DeleteFile(fileName string) error {
	client, err := getS3Client()
	if err != nil {
		return err
	}

	fileName = filepath.Base(fileName)

	_, err = client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName),
	})

	return err
}

func ListFiles() ([]string, error) {
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

	files := []string{}

	for _, obj := range result.Contents {
		if obj.Key != nil {
			files = append(files, *obj.Key)
		}
	}

	return files, nil
}

func DownloadFile(fileName string) (string, error) {
	client, err := getS3Client()
	if err != nil {
		return "", err
	}

	fileName = filepath.Base(fileName)

	result, err := client.GetObject(
		context.Background(),
		&s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(fileName),
		},
	)
	if err != nil {
		return "", err
	}
	defer result.Body.Close()

	err = os.MkdirAll("storage", 0755)
	if err != nil {
		return "", err
	}

	localPath := filepath.Join("storage", fileName)

	dst, err := os.Create(localPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	_, err = io.Copy(dst, result.Body)
	if err != nil {
		return "", err
	}

	return localPath, nil
}

func GetFile(fileName string) (io.ReadCloser, error) {
	client, err := getS3Client()
	if err != nil {
		return nil, err
	}

	fileName = filepath.Base(fileName)

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

	return result.Body, nil
}
