package middlewares

import (
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

func UploadImage(file io.Reader, originalFilename, subfolder string) (string, error) {
	uploadDir := path.Join("uploads", subfolder)
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			return "", fmt.Errorf("не удалось создать папку uploads: %v", err)
		}
	}

	ext := filepath.Ext(originalFilename)
	filename := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)
	filePath := path.Join(uploadDir, filename)

	// Валидируем тип изображения
	if !isValidImageType(ext) {
		return "", fmt.Errorf("неподдерживаемый формат изображения: %s", ext)
	}

	img, format, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("не удалось декодировать изображение: %v", err)
	}

	out, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("не удалось создать файл: %v", err)
	}
	defer out.Close()

	jpegFilename := strings.TrimSuffix(filename, ext) + ".jpg"
	jpegFilePath := path.Join(uploadDir, jpegFilename)

	jpegOut, err := os.Create(jpegFilePath)
	if err != nil {
		return "", fmt.Errorf("не удалось создать JPEG файл: %v", err)
	}
	defer jpegOut.Close()

	err = jpeg.Encode(jpegOut, img, &jpeg.Options{Quality: 85})
	if err != nil {
		return "", fmt.Errorf("не удалось сохранить JPEG: %v", err)
	}

	if format != "jpeg" {
		os.Remove(filePath)
	}

	imageURL := fmt.Sprintf("http://localhost:8080/uploads/%s/%s", subfolder, jpegFilename)
	return imageURL, nil
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

// Валидация изображения по MIME типу
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

func DeleteImage(filePath string) error {
	_ = os.Remove(filePath)
	return nil
}
