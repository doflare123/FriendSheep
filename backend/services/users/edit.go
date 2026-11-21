package users

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	_ "golang.org/x/image/webp"
)

const (
	maxProfileImageBytes = 5 << 20
	maxProfileImageSide  = 4096
)

var (
	ErrInvalidProfile          = errors.New("некорректные данные профиля")
	ErrCityNotFound            = errors.New("город не найден")
	ErrUsTaken                 = errors.New("адрес пользователя уже занят")
	ErrBirthdayChangeDenied    = errors.New("дата рождения уже установлена")
	ErrInvalidImage            = errors.New("некорректное изображение профиля")
	ErrImageTooLarge           = errors.New("изображение профиля слишком большое")
	ErrImageStorageUnavailable = errors.New("хранилище изображений недоступно")

	profileUSPattern = regexp.MustCompile(`^[a-z0-9_]{3,32}$`)
)

// Date хранит календарную дату без часового пояса.
type Date struct{ time.Time }

func (d Date) MarshalJSON() ([]byte, error) { return json.Marshal(d.Format(time.DateOnly)) }

type OptionalDate struct {
	Set   bool
	Value *Date
}

type OptionalCityID struct {
	Set   bool
	Value *uint
}

type ProfilePatch struct {
	Name     *string        `json:"name,omitempty"`
	Us       *string        `json:"us,omitempty"`
	Status   *string        `json:"status,omitempty"`
	Birthday OptionalDate   `json:"birthday,omitempty" swaggertype:"string" format:"date" extensions:"x-nullable"`
	CityID   OptionalCityID `json:"cityId,omitempty" swaggertype:"integer" extensions:"x-nullable"`
}

func (p ProfilePatch) hasChangesBesidesBirthday() bool {
	return p.Name != nil || p.Us != nil || p.Status != nil || p.CityID.Set
}

func (p *ProfilePatch) UnmarshalJSON(data []byte) error {
	*p = ProfilePatch{}
	decoder := json.NewDecoder(bytes.NewReader(data))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return ErrInvalidProfile
	}
	seen := make(map[string]struct{}, 5)
	for decoder.More() {
		nameToken, err := decoder.Token()
		name, ok := nameToken.(string)
		if err != nil || !ok {
			return ErrInvalidProfile
		}
		if _, duplicate := seen[name]; duplicate {
			return ErrInvalidProfile
		}
		seen[name] = struct{}{}
		var raw json.RawMessage
		if err := decoder.Decode(&raw); err != nil {
			return ErrInvalidProfile
		}
		switch name {
		case "name":
			if err := decodeRequiredString(raw, &p.Name); err != nil {
				return err
			}
		case "us":
			if err := decodeRequiredString(raw, &p.Us); err != nil {
				return err
			}
		case "status":
			if err := decodeRequiredString(raw, &p.Status); err != nil {
				return err
			}
		case "birthday":
			p.Birthday.Set = true
			if !bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
				var value string
				if json.Unmarshal(raw, &value) != nil {
					return ErrInvalidProfile
				}
				parsed, err := time.Parse(time.DateOnly, value)
				if err != nil {
					return ErrInvalidProfile
				}
				p.Birthday.Value = &Date{Time: parsed}
			}
		case "cityId":
			p.CityID.Set = true
			if !bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
				var value uint
				if json.Unmarshal(raw, &value) != nil {
					return ErrInvalidProfile
				}
				p.CityID.Value = &value
			}
		default:
			return ErrInvalidProfile
		}
	}
	if len(seen) == 0 {
		return ErrInvalidProfile
	}
	if token, err = decoder.Token(); err != nil || token != json.Delim('}') {
		return ErrInvalidProfile
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return ErrInvalidProfile
	}
	return nil
}

func decodeRequiredString(raw json.RawMessage, target **string) error {
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return ErrInvalidProfile
	}
	var value string
	if json.Unmarshal(raw, &value) != nil {
		return ErrInvalidProfile
	}
	*target = &value
	return nil
}

type EditableCity struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Region string `json:"region"`
}

type EditableProfile struct {
	ID       uint          `json:"id"`
	Name     string        `json:"name"`
	Us       string        `json:"us"`
	Status   string        `json:"status"`
	Image    string        `json:"image"`
	Birthday *Date         `json:"birthday" swaggertype:"string" format:"date" extensions:"x-nullable"`
	City     *EditableCity `json:"city" extensions:"x-nullable"`
}

type ImageChange struct {
	Profile           EditableProfile
	PreviousObjectKey string
}

type EditStore interface {
	GetOwn(context.Context, uint) (EditableProfile, error)
	Patch(context.Context, uint, ProfilePatch) (EditableProfile, error)
	ReplaceImage(context.Context, uint, string, string) (ImageChange, error)
	DeleteImage(context.Context, uint) (ImageChange, error)
}

type ProfileImageStore interface {
	Put(context.Context, string, []byte, string) (string, error)
	Delete(context.Context, string) error
}

type EditService struct {
	store  EditStore
	images ProfileImageStore
}

func NewEditService(store EditStore, images ProfileImageStore) *EditService {
	return &EditService{store: store, images: images}
}

func (s *EditService) GetOwn(ctx context.Context, userID uint) (EditableProfile, error) {
	if userID == 0 {
		return EditableProfile{}, ErrUnauthenticated
	}
	return s.store.GetOwn(ctx, userID)
}

func (s *EditService) Patch(ctx context.Context, userID uint, patch ProfilePatch) (EditableProfile, error) {
	if userID == 0 {
		return EditableProfile{}, ErrUnauthenticated
	}
	normalized, err := normalizeProfilePatch(patch, time.Now().UTC())
	if err != nil {
		return EditableProfile{}, err
	}
	return s.store.Patch(ctx, userID, normalized)
}

func (s *EditService) UploadImage(ctx context.Context, userID uint, body io.Reader) (EditableProfile, error) {
	if userID == 0 {
		return EditableProfile{}, ErrUnauthenticated
	}
	if s.images == nil || body == nil {
		return EditableProfile{}, ErrImageStorageUnavailable
	}
	data, err := io.ReadAll(io.LimitReader(body, maxProfileImageBytes+1))
	if err != nil {
		return EditableProfile{}, errors.Join(ErrInvalidImage, err)
	}
	if len(data) > maxProfileImageBytes {
		return EditableProfile{}, ErrImageTooLarge
	}
	contentType, extension, err := inspectProfileImage(data)
	if err != nil {
		return EditableProfile{}, err
	}
	if err := ctx.Err(); err != nil {
		return EditableProfile{}, err
	}
	key := fmt.Sprintf("users/%d/%s%s", userID, uuid.NewString(), extension)
	url, err := s.images.Put(ctx, key, data, contentType)
	if err != nil {
		return EditableProfile{}, errors.Join(ErrImageStorageUnavailable, err)
	}
	change, err := s.store.ReplaceImage(ctx, userID, url, key)
	if err != nil {
		s.deleteImageObject(ctx, key)
		return EditableProfile{}, err
	}
	if change.PreviousObjectKey != "" && change.PreviousObjectKey != key {
		s.deleteImageObject(ctx, change.PreviousObjectKey)
	}
	return change.Profile, nil
}

func (s *EditService) DeleteImage(ctx context.Context, userID uint) (EditableProfile, error) {
	if userID == 0 {
		return EditableProfile{}, ErrUnauthenticated
	}
	change, err := s.store.DeleteImage(ctx, userID)
	if err != nil {
		return EditableProfile{}, err
	}
	if change.PreviousObjectKey != "" && s.images != nil {
		s.deleteImageObject(ctx, change.PreviousObjectKey)
	}
	return change.Profile, nil
}

func (s *EditService) deleteImageObject(parent context.Context, key string) {
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(parent), 5*time.Second)
	defer cancel()
	_ = s.images.Delete(cleanupCtx, key)
}

func normalizeProfilePatch(patch ProfilePatch, now time.Time) (ProfilePatch, error) {
	if patch.Name == nil && patch.Us == nil && patch.Status == nil && !patch.Birthday.Set && !patch.CityID.Set {
		return ProfilePatch{}, ErrInvalidProfile
	}
	if patch.Name != nil {
		value := strings.TrimSpace(*patch.Name)
		if !utf8.ValidString(value) || utf8.RuneCountInString(value) < 1 || utf8.RuneCountInString(value) > 100 {
			return ProfilePatch{}, ErrInvalidProfile
		}
		patch.Name = &value
	}
	if patch.Us != nil {
		value := strings.ToLower(strings.TrimSpace(*patch.Us))
		value = strings.TrimPrefix(value, "@")
		if !profileUSPattern.MatchString(value) {
			return ProfilePatch{}, ErrInvalidProfile
		}
		patch.Us = &value
	}
	if patch.Status != nil {
		value := strings.TrimSpace(*patch.Status)
		if !utf8.ValidString(value) || utf8.RuneCountInString(value) > 500 {
			return ProfilePatch{}, ErrInvalidProfile
		}
		patch.Status = &value
	}
	if patch.Birthday.Value != nil {
		birthday := patch.Birthday.Value.Time
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		if birthday.Before(time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)) || birthday.After(today) {
			return ProfilePatch{}, ErrInvalidProfile
		}
	}
	if patch.CityID.Value != nil && *patch.CityID.Value == 0 {
		return ProfilePatch{}, ErrInvalidProfile
	}
	return patch, nil
}

func inspectProfileImage(data []byte) (string, string, error) {
	if len(data) == 0 {
		return "", "", ErrInvalidImage
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || config.Width <= 0 || config.Height <= 0 {
		return "", "", ErrInvalidImage
	}
	if config.Width > maxProfileImageSide || config.Height > maxProfileImageSide {
		return "", "", ErrImageTooLarge
	}
	decoded, decodedFormat, err := image.Decode(bytes.NewReader(data))
	if err != nil || decodedFormat != format || decoded.Bounds().Dx() != config.Width || decoded.Bounds().Dy() != config.Height {
		return "", "", ErrInvalidImage
	}
	switch format {
	case "jpeg":
		return "image/jpeg", ".jpg", nil
	case "png":
		return "image/png", ".png", nil
	case "webp":
		return "image/webp", ".webp", nil
	default:
		return "", "", ErrInvalidImage
	}
}
