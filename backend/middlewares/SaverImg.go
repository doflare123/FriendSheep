package middlewares

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

var (
	s3Client   *s3.S3
	bucketName string
	s3Endpoint string
	region     string
)

func InitS3(accessKey, secretKey string) error {
	return InitS3WithConfig(accessKey, secretKey, "https://s3.ru-3.storage.selcloud.ru", "ru-3", "friendsheep")
}

// InitS3WithConfig инициализирует соединение с S3 с полной конфигурацией
func InitS3WithConfig(accessKey, secretKey, endpoint, reg, bucket string) error {
	s3Endpoint = endpoint
	region = reg
	bucketName = bucket

	fmt.Printf("Инициализация S3 с параметрами:\n")
	fmt.Printf("  Endpoint: %s\n", s3Endpoint)
	fmt.Printf("  Region: %s\n", region)
	fmt.Printf("  Bucket: %s\n", bucketName)
	fmt.Printf("  AccessKey: %s...\n", accessKey[:10])

	awsConfig := &aws.Config{
		Endpoint:         aws.String(s3Endpoint),
		Region:           aws.String(region),
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		S3ForcePathStyle: aws.Bool(true),
		DisableSSL:       aws.Bool(false),
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return fmt.Errorf("не удалось создать S3 сессию: %v", err)
	}

	sess.Config.Region = aws.String(region)

	s3Client = s3.New(sess, awsConfig)

	fmt.Printf("✓ S3 клиент успешно инициализирован\n")
	return nil
}

// UploadImageToS3 загружает изображение в S3 и возвращает публичный URL
func UploadImageToS3(file io.Reader, originalFilename, subfolder string) (string, error) {
	if s3Client == nil {
		return "", fmt.Errorf("S3 клиент не инициализирован. Вызовите InitS3() перед загрузкой файлов")
	}

	fmt.Printf("Начинаем загрузку файла: %s в подпапку: %s\n", originalFilename, subfolder)

	ext := filepath.Ext(originalFilename)

	if !isValidImageType(ext) {
		return "", fmt.Errorf("неподдерживаемый формат изображения: %s", ext)
	}

	img, format, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("не удалось декодировать изображение: %v", err)
	}
	fmt.Printf("Изображение декодировано, формат: %s\n", format)

	var buf bytes.Buffer
	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})
	if err != nil {
		return "", fmt.Errorf("не удалось сконвертировать в JPEG: %v", err)
	}
	fmt.Printf("Изображение сконвертировано в JPEG, размер: %d байт\n", buf.Len())

	filename := fmt.Sprintf("%s_%d.jpg", uuid.New().String(), time.Now().Unix())
	key := fmt.Sprintf("%s/%s", subfolder, filename)

	fmt.Printf("Загружаем в S3: bucket=%s, key=%s\n", bucketName, key)

	// Загружаем в S3
	putInput := &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String("image/jpeg"),
		ACL:         aws.String("public-read"), // Делаем файл публично доступным
	}

	result, err := s3Client.PutObject(putInput)
	if err != nil {
		return "", fmt.Errorf("не удалось загрузить файл в S3: %v", err)
	}

	fmt.Printf("Файл успешно загружен в S3, ETag: %s\n", aws.StringValue(result.ETag))

	imageURL := fmt.Sprintf("%s/%s/%s", s3Endpoint, bucketName, key)
	fmt.Printf("Публичный URL: %s\n", imageURL)

	return imageURL, nil
}

// DeleteImageFromS3 удаляет изображение из S3 по URL
func DeleteImageFromS3(imageURL string) error {
	if s3Client == nil {
		return fmt.Errorf("S3 клиент не инициализирован")
	}

	key := extractKeyFromURL(imageURL)
	if key == "" {
		return fmt.Errorf("не удалось извлечь ключ из URL: %s", imageURL)
	}

	_, err := s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("не удалось удалить файл из S3: %v", err)
	}

	return nil
}

// extractKeyFromURL извлекает ключ файла из полного URL
func extractKeyFromURL(imageURL string) string {
	prefix := fmt.Sprintf("%s/%s/", s3Endpoint, bucketName)
	if strings.HasPrefix(imageURL, prefix) {
		return strings.TrimPrefix(imageURL, prefix)
	}
	return ""
}

func isValidImageType(ext string) bool {
	validTypes := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
		".gif":  true,
		".bmp":  true,
		".tiff": true,
		".tif":  true,
	}
	return validTypes[strings.ToLower(ext)]
}

// ValidateImageMIME валидация изображения по MIME типу
func ValidateImageMIME(header *multipart.FileHeader) error {
	file, err := header.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	mimeType := http.DetectContentType(buffer)
	validMIMEs := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
		"image/gif":  true,
		"image/bmp":  true,
		"image/tiff": true,
	}

	if !validMIMEs[mimeType] {
		return fmt.Errorf("неподдерживаемый MIME тип: %s", mimeType)
	}

	return nil
}
