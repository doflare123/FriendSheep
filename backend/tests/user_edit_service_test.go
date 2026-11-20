package tests

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
	"time"

	"friendship/models"
	"friendship/services/users"
)

type editStoreStub struct {
	profile      users.EditableProfile
	patch        users.ProfilePatch
	patchErr     error
	replaceErr   error
	deleteErr    error
	replaceImage string
	replaceKey   string
	change       users.ImageChange
}

func (s *editStoreStub) GetOwn(_ context.Context, _ uint) (users.EditableProfile, error) {
	return s.profile, nil
}
func (s *editStoreStub) Patch(_ context.Context, _ uint, patch users.ProfilePatch) (users.EditableProfile, error) {
	s.patch = patch
	return s.profile, s.patchErr
}
func (s *editStoreStub) ReplaceImage(_ context.Context, _ uint, image, key string) (users.ImageChange, error) {
	s.replaceImage, s.replaceKey = image, key
	return s.change, s.replaceErr
}
func (s *editStoreStub) DeleteImage(_ context.Context, _ uint) (users.ImageChange, error) {
	return s.change, s.deleteErr
}

type imageStoreStub struct {
	url      string
	putErr   error
	deleted  []string
	putKey   string
	putType  string
	putBytes []byte
}

func (s *imageStoreStub) Put(_ context.Context, key string, body []byte, contentType string) (string, error) {
	s.putKey, s.putBytes, s.putType = key, append([]byte(nil), body...), contentType
	return s.url, s.putErr
}
func (s *imageStoreStub) Delete(_ context.Context, key string) error {
	s.deleted = append(s.deleted, key)
	return nil
}

func TestEditServiceNormalizesPatchAndRejectsInvalidValues(t *testing.T) {
	store := &editStoreStub{profile: users.EditableProfile{ID: 7}}
	service := users.NewEditService(store, nil)
	name, us, status := "  Алиса  ", " @Alice_2 ", "  "
	birthday := users.Date{Time: time.Date(2000, 2, 3, 0, 0, 0, 0, time.UTC)}
	cityID := uint(3)
	if _, err := service.Patch(context.Background(), 7, users.ProfilePatch{Name: &name, Us: &us, Status: &status, Birthday: users.OptionalDate{Set: true, Value: &birthday}, CityID: users.OptionalCityID{Set: true, Value: &cityID}}); err != nil {
		t.Fatal(err)
	}
	if store.patch.Name == nil || *store.patch.Name != "Алиса" || store.patch.Us == nil || *store.patch.Us != "alice_2" || store.patch.Status == nil || *store.patch.Status != "" {
		t.Fatalf("normalized patch = %#v", store.patch)
	}
	for _, patch := range []users.ProfilePatch{{}, {Us: stringPtr("Алиса")}, {CityID: users.OptionalCityID{Set: true, Value: uintPtr(0)}}, {Birthday: users.OptionalDate{Set: true, Value: &users.Date{Time: time.Now().UTC().AddDate(0, 0, 1)}}}} {
		if _, err := service.Patch(context.Background(), 7, patch); !errors.Is(err, users.ErrInvalidProfile) {
			t.Fatalf("Patch(%#v) error = %v, want ErrInvalidProfile", patch, err)
		}
	}
}

func TestEditServiceValidatesExactProfileTextBoundaries(t *testing.T) {
	store := &editStoreStub{profile: users.EditableProfile{ID: 7}}
	service := users.NewEditService(store, nil)
	for _, test := range []struct {
		name  string
		patch users.ProfilePatch
		valid bool
	}{
		{"name 100 unicode", users.ProfilePatch{Name: stringPtr(strings.Repeat("я", 100))}, true},
		{"name 101 unicode", users.ProfilePatch{Name: stringPtr(strings.Repeat("я", 101))}, false},
		{"blank name", users.ProfilePatch{Name: stringPtr(" \t ")}, false},
		{"status 500 unicode", users.ProfilePatch{Status: stringPtr(strings.Repeat("я", 500))}, true},
		{"status 501 unicode", users.ProfilePatch{Status: stringPtr(strings.Repeat("я", 501))}, false},
		{"empty status clears", users.ProfilePatch{Status: stringPtr("   ")}, true},
		{"us 3", users.ProfilePatch{Us: stringPtr("abc")}, true},
		{"us 32", users.ProfilePatch{Us: stringPtr(strings.Repeat("a", 32))}, true},
		{"us 2", users.ProfilePatch{Us: stringPtr("ab")}, false},
		{"us 33", users.ProfilePatch{Us: stringPtr(strings.Repeat("a", 33))}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := service.Patch(context.Background(), 7, test.patch)
			if test.valid && err != nil {
				t.Fatalf("Patch error = %v", err)
			}
			if !test.valid && !errors.Is(err, users.ErrInvalidProfile) {
				t.Fatalf("Patch error = %v, want invalid", err)
			}
		})
	}
}

func TestEditServiceImageCompensatesFailedDatabaseReplacement(t *testing.T) {
	store := &editStoreStub{replaceErr: users.ErrUserNotFound}
	images := &imageStoreStub{url: "https://cdn.example/users/7/a.png"}
	_, err := users.NewEditService(store, images).UploadImage(context.Background(), 7, bytes.NewReader(pngImage(t)))
	if !errors.Is(err, users.ErrUserNotFound) || len(images.deleted) != 1 || images.deleted[0] != images.putKey || images.putType != "image/png" {
		t.Fatalf("error/storage = %v/%#v", err, images)
	}
}

func TestEditServiceImageReplacesAndRemovesOnlyStoredKey(t *testing.T) {
	store := &editStoreStub{change: users.ImageChange{Profile: users.EditableProfile{ID: 7, Image: "https://cdn.example/new.png"}, PreviousObjectKey: "users/7/old.png"}}
	images := &imageStoreStub{url: "https://cdn.example/new.png"}
	profile, err := users.NewEditService(store, images).UploadImage(context.Background(), 7, bytes.NewReader(pngImage(t)))
	if err != nil || profile.Image == "" || len(images.deleted) != 1 || images.deleted[0] != "users/7/old.png" || store.replaceKey == "" {
		t.Fatalf("profile/error/storage/store = %#v/%v/%#v/%#v", profile, err, images, store)
	}
}

func TestEditServiceRejectsTruncatedImageAfterHeader(t *testing.T) {
	valid := pngImage(t)
	// Заголовок PNG содержит размеры, поэтому одной DecodeConfig недостаточно.
	truncated := valid[:len(valid)/2]
	store := &editStoreStub{}
	images := &imageStoreStub{url: "https://cdn.example/image.png"}
	_, err := users.NewEditService(store, images).UploadImage(context.Background(), 7, bytes.NewReader(truncated))
	if !errors.Is(err, users.ErrInvalidImage) || images.putKey != "" {
		t.Fatalf("truncated image error/storage = %v/%#v", err, images)
	}
}

func TestEditServiceRejectsOversizedInputAndUnavailableStorage(t *testing.T) {
	store := &editStoreStub{}
	overLimit := append(pngImage(t), bytes.Repeat([]byte{0}, (5<<20)+1)...)
	if _, err := users.NewEditService(store, &imageStoreStub{}).UploadImage(context.Background(), 7, bytes.NewReader(overLimit)); !errors.Is(err, users.ErrImageTooLarge) {
		t.Fatalf("oversized image error = %v", err)
	}
	if _, err := users.NewEditService(store, nil).UploadImage(context.Background(), 7, bytes.NewReader(pngImage(t))); !errors.Is(err, users.ErrImageStorageUnavailable) {
		t.Fatalf("unavailable storage error = %v", err)
	}
}

func TestEditServiceDeletesOnlyPersistedImageKey(t *testing.T) {
	store := &editStoreStub{change: users.ImageChange{Profile: users.EditableProfile{ID: 7}, PreviousObjectKey: "users/7/current.png"}}
	images := &imageStoreStub{}
	if _, err := users.NewEditService(store, images).DeleteImage(context.Background(), 7); err != nil {
		t.Fatal(err)
	}
	if len(images.deleted) != 1 || images.deleted[0] != "users/7/current.png" {
		t.Fatalf("deleted keys = %#v", images.deleted)
	}
}

func TestGORMEditStorePreservesFieldsAndRollsBackMissingCity(t *testing.T) {
	db := newProfileStoreDB(t)
	if err := db.AutoMigrate(&models.City{}); err != nil {
		t.Fatal(err)
	}
	birthday := time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC)
	city := models.City{Name: "Калининград", Region: "Калининградская область", FIASID: "city-edit-1"}
	if err := db.Create(&city).Error; err != nil {
		t.Fatal(err)
	}
	user := models.User{ID: 71, Name: "Alice", Password: "Password123!", Us: "alice", Email: "alice-edit@example.com", Birthday: &birthday, CityID: &city.ID, Status: "old"}
	other := models.User{ID: 72, Name: "Other", Password: "Password123!", Us: "taken", Email: "taken-edit@example.com"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&other).Error; err != nil {
		t.Fatal(err)
	}
	store := users.NewGORMEditStore(&testPostgresRepository{db: db})
	name := "Alice updated"
	if _, err := store.Patch(context.Background(), user.ID, users.ProfilePatch{Name: &name, Birthday: users.OptionalDate{Set: true}, CityID: users.OptionalCityID{Set: true}}); err != nil {
		t.Fatal(err)
	}
	got, err := store.GetOwn(context.Background(), user.ID)
	if err != nil || got.Name != name || got.Birthday != nil || got.City != nil || got.Status != "old" {
		t.Fatalf("cleared profile/error = %#v/%v", got, err)
	}
	missing := uint(999)
	changed := "must not persist"
	if _, err := store.Patch(context.Background(), user.ID, users.ProfilePatch{Name: &changed, CityID: users.OptionalCityID{Set: true, Value: &missing}}); !errors.Is(err, users.ErrCityNotFound) {
		t.Fatalf("missing city error = %v", err)
	}
	got, err = store.GetOwn(context.Background(), user.ID)
	if err != nil || got.Name != name {
		t.Fatalf("missing city rollback = %#v/%v", got, err)
	}
	conflict := "TAKEN"
	if _, err := store.Patch(context.Background(), user.ID, users.ProfilePatch{Us: &conflict}); !errors.Is(err, users.ErrUsTaken) {
		t.Fatalf("case-insensitive us conflict = %v", err)
	}
}

func pngImage(t *testing.T) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.NRGBA{R: 1, A: 255})
	var body bytes.Buffer
	if err := png.Encode(&body, img); err != nil {
		t.Fatal(err)
	}
	return body.Bytes()
}

func stringPtr(value string) *string { return &value }
func uintPtr(value uint) *uint       { return &value }

var _ users.EditStore = (*editStoreStub)(nil)
var _ users.ProfileImageStore = (*imageStoreStub)(nil)
