package middlewares

import (
	"fmt"
	"io"
	"os"
)

func UploadImage(file io.Reader, filename string) (string, error) {
	uploadDir := "uploads"

	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			return "", fmt.Errorf("не удалось создать папку uploads: %v", err)
		}
	}

	out, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("не удалось создать файл: %v", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		return "", fmt.Errorf("не удалось сохранить файл: %v", err)
	}

	imageURL := fmt.Sprintf("http://localhost:8080/%s", filename)
	return imageURL, nil
}

func DeleteImage(filePath string) error {
	_ = os.Remove(filePath)
	return nil
}
