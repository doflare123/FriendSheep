package validator

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

type ImageValidator struct {
	maxSize      int64
	allowedMIMEs map[string]bool
}

func NewImageValidator(maxSize int64) *ImageValidator {
	return &ImageValidator{
		maxSize: maxSize,
		allowedMIMEs: map[string]bool{
			"image/jpeg": true,
			"image/png":  true,
			"image/webp": true,
			"image/gif":  true,
			"image/bmp":  true,
			"image/tiff": true,
		},
	}
}

func (v *ImageValidator) Validate(header *multipart.FileHeader) error {
	if header.Size > v.maxSize {
		return fmt.Errorf("Превышен допустимый размер файла на %d байтов", v.maxSize)
	}

	file, err := header.Open()
	if err != nil {
		return fmt.Errorf("Ошибка при открытии файла: %w", err)
	}
	defer file.Close()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("Ошибка при чтении файла: %w", err)
	}

	mimeType := http.DetectContentType(buffer)
	if !v.allowedMIMEs[mimeType] {
		return fmt.Errorf("Не поддерживаемый формат: %s", mimeType)
	}

	return nil
}
