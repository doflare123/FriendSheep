package storage

import (
	"bytes"
	"context"
	"fmt"
	conf "friendship/config"
	"friendship/logger"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	_ "golang.org/x/image/webp"
)

type S3Storage interface {
	UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, folder string) (string, error)
	DeleteFile(ctx context.Context, fileURL string) error
	GetFileURL(fileName string) string
}

type SelectelS3 struct {
	client      *s3.Client
	bucket      string
	endpoint    string
	region      string
	containerID string
}

func NewSelectelS3(conf *conf.Config, logger logger.Logger) (S3Storage, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(conf.S3.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			conf.S3.AccessKey,
			conf.S3.SecretKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(conf.S3.Endpoint)
		o.UsePathStyle = true
	})

	return &SelectelS3{
		client:      client,
		bucket:      conf.S3.Bucket,
		endpoint:    conf.S3.Endpoint,
		region:      conf.S3.Region,
		containerID: conf.S3.ContainerId,
	}, nil
}

func (s *SelectelS3) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, folder string) (string, error) {
	fileName := fmt.Sprintf("%s_%s.jpg", time.Now().Format("20060102_150405"), uuid.New().String()[:8])

	key := fileName
	if folder != "" {
		key = filepath.Join(folder, fileName)
	}

	jpegData, err := s.convertToJPEG(file, header)
	if err != nil {
		return "", fmt.Errorf("failed to convert image: %w", err)
	}

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(jpegData),
		ContentType: aws.String("image/jpeg"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	return s.GetFileURL(key), nil
}

func (s *SelectelS3) convertToJPEG(file multipart.File, header *multipart.FileHeader) ([]byte, error) {
	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return nil, fmt.Errorf("file is not an image: %s", contentType)
	}

	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	_ = format

	var buf bytes.Buffer
	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80})
	if err != nil {
		return nil, fmt.Errorf("failed to encode to JPEG: %w", err)
	}

	return buf.Bytes(), nil
}

func (s *SelectelS3) DeleteFile(ctx context.Context, fileURL string) error {
	key, err := s.extractKeyFromURL(fileURL)
	if err != nil {
		return err
	}

	_, err = s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

func (s *SelectelS3) GetFileURL(fileName string) string {
	return fmt.Sprintf("https://%s.selstorage.ru/%s", s.containerID, fileName)
}

func (s *SelectelS3) extractKeyFromURL(fileURL string) (string, error) {
	prefix := fmt.Sprintf("%s/%s/", s.endpoint, s.bucket)

	if len(fileURL) <= len(prefix) {
		return "", fmt.Errorf("invalid file URL")
	}

	return fileURL[len(prefix):], nil
}
