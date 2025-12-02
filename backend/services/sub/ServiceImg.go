package sub

import (
	"context"
	"fmt"
	storage "friendship/S3"
	"friendship/logger"
	"friendship/validator"
	"mime/multipart"
)

type ImgService interface {
	UploadImg(ctx context.Context, header *multipart.FileHeader, folder string) (string, error)
}

type imgService struct {
	logger    logger.Logger
	s3        storage.S3Storage
	validator *validator.ImageValidator
}

func NewImgService(
	logger logger.Logger,
	s3 storage.S3Storage,
	validator *validator.ImageValidator,
) ImgService {
	return &imgService{
		logger:    logger,
		s3:        s3,
		validator: validator,
	}
}

func (s *imgService) UploadImg(ctx context.Context, header *multipart.FileHeader, folder string) (string, error) {
	if err := s.validator.Validate(header); err != nil {
		s.logger.Error("При валидации изображения произошла ошибка", "error", err)
		return "", fmt.Errorf("При валидации изображения произошла ошибка: %w", err)
	}

	file, err := header.Open()
	if err != nil {
		s.logger.Error("Ошибка при открытии файла", "error", err)
		return "", fmt.Errorf("Ошибка при открытии файла: %w", err)
	}
	defer file.Close()

	url, err := s.s3.UploadFile(ctx, file, header, folder)
	if err != nil {
		s.logger.Error("Ошибка при сохранении", "error", err)
		return "", fmt.Errorf("Ошибка при сохранении: %w", err)
	}

	s.logger.Info("Файл загрузился", "url", url)
	return url, nil
}
