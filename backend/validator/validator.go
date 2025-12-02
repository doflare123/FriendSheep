package validator

import "friendship/config"

type Validator struct {
	Image *ImageValidator
}

func NewValidator(cfg *config.Config) *Validator {
	return &Validator{
		Image: NewImageValidator(int64(cfg.Upload.MaxImageSize)),
	}
}
