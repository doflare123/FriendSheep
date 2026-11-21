package tests

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"friendship/models"
	"friendship/services/users"

	"gorm.io/gorm"
)

func TestUserEditPostgresConcurrency(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("FRIENDSHEEP_TEST_POSTGRES_DSN"))
	if dsn == "" {
		t.Skip("set FRIENDSHEEP_TEST_POSTGRES_DSN to run postgres integration tests")
	}

	db := newPostgresEventCommandDB(t, dsn)
	store := users.NewGORMEditStore(&testPostgresRepository{db: db})
	service := users.NewEditService(store, nil)

	t.Run("case insensitive us remains unique", func(t *testing.T) {
		first := createEditableUserPostgres(t, db, "first_user", "first@example.test")
		second := createEditableUserPostgres(t, db, "second_user", "second@example.test")

		start := make(chan struct{})
		errorsByUser := make(chan error, 2)
		var wait sync.WaitGroup
		claims := []struct {
			userID uint
			us     string
		}{
			{userID: first.ID, us: "Shared_Name"},
			{userID: second.ID, us: "shared_name"},
		}
		for _, claim := range claims {
			wait.Add(1)
			go func(id uint, value string) {
				defer wait.Done()
				<-start
				_, err := store.Patch(context.Background(), id, users.ProfilePatch{Us: &value})
				errorsByUser <- err
			}(claim.userID, claim.us)
		}
		close(start)
		wait.Wait()
		close(errorsByUser)

		var succeeded, conflicted int
		for err := range errorsByUser {
			switch {
			case err == nil:
				succeeded++
			case errors.Is(err, users.ErrUsTaken):
				conflicted++
			default:
				t.Fatalf("concurrent us patch returned unexpected error: %v", err)
			}
		}
		if succeeded != 1 || conflicted != 1 {
			t.Fatalf("concurrent us results success/conflict = %d/%d, want 1/1", succeeded, conflicted)
		}

		var count int64
		if err := db.Model(&models.User{}).Where("lower(us) = ?", "shared_name").Count(&count).Error; err != nil {
			t.Fatalf("count normalized us: %v", err)
		}
		if count != 1 {
			t.Fatalf("normalized us count = %d, want 1", count)
		}
	})

	t.Run("independent patches preserve both fields", func(t *testing.T) {
		user := createEditableUserPostgres(t, db, "independent_user", "independent@example.test")
		name := "Новое имя"
		status := "Новый статус"
		start := make(chan struct{})
		results := make(chan error, 2)

		go func() {
			<-start
			_, err := service.Patch(context.Background(), user.ID, users.ProfilePatch{Name: &name})
			results <- err
		}()
		go func() {
			<-start
			_, err := service.Patch(context.Background(), user.ID, users.ProfilePatch{Status: &status})
			results <- err
		}()
		close(start)
		for range 2 {
			if err := <-results; err != nil {
				t.Fatalf("independent concurrent patch: %v", err)
			}
		}

		profile, err := service.GetOwn(context.Background(), user.ID)
		if err != nil {
			t.Fatalf("load edited profile: %v", err)
		}
		if profile.Name != name || profile.Status != status {
			t.Fatalf("profile name/status = %q/%q, want %q/%q", profile.Name, profile.Status, name, status)
		}
	})

	t.Run("concurrent first birthday set permits exactly one value", func(t *testing.T) {
		user := createEditableUserPostgres(t, db, "birthday_race_user", "birthday-race@example.test")
		first := users.Date{Time: time.Date(1999, 1, 1, 0, 0, 0, 0, time.UTC)}
		second := users.Date{Time: time.Date(2000, 2, 2, 0, 0, 0, 0, time.UTC)}
		start := make(chan struct{})
		results := make(chan error, 2)
		for _, birthday := range []users.Date{first, second} {
			go func(value users.Date) {
				<-start
				_, err := store.Patch(context.Background(), user.ID, users.ProfilePatch{Birthday: users.OptionalDate{Set: true, Value: &value}})
				results <- err
			}(birthday)
		}
		close(start)

		var succeeded, denied int
		for range 2 {
			switch err := <-results; {
			case err == nil:
				succeeded++
			case errors.Is(err, users.ErrBirthdayChangeDenied):
				denied++
			default:
				t.Fatalf("конкурентная установка даты рождения: %v", err)
			}
		}
		if succeeded != 1 || denied != 1 {
			t.Fatalf("конкурентная установка успех/отказ = %d/%d, ожидалось 1/1", succeeded, denied)
		}
		profile, err := store.GetOwn(context.Background(), user.ID)
		if err != nil || profile.Birthday == nil || (!profile.Birthday.Time.Equal(first.Time) && !profile.Birthday.Time.Equal(second.Time)) {
			t.Fatalf("сохранённая дата рождения = %#v/%v", profile.Birthday, err)
		}
	})

	t.Run("image replacements serialize previous keys", func(t *testing.T) {
		user := createEditableUserPostgres(t, db, "image_user", "image@example.test")
		if err := db.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]any{
			"image": "https://cdn.example/initial.jpg", "image_object_key": "users/initial.jpg",
		}).Error; err != nil {
			t.Fatalf("seed initial image: %v", err)
		}

		type result struct {
			change users.ImageChange
			err    error
		}
		start := make(chan struct{})
		results := make(chan result, 2)
		for _, suffix := range []string{"first", "second"} {
			go func(value string) {
				<-start
				key := "users/" + value + ".jpg"
				change, err := store.ReplaceImage(context.Background(), user.ID, "https://cdn.example/"+value+".jpg", key)
				results <- result{change: change, err: err}
			}(suffix)
		}
		close(start)

		previous := make(map[string]bool, 2)
		for range 2 {
			outcome := <-results
			if outcome.err != nil {
				t.Fatalf("concurrent image replacement: %v", outcome.err)
			}
			previous[outcome.change.PreviousObjectKey] = true
		}

		var stored models.User
		if err := db.Select("image_object_key").First(&stored, user.ID).Error; err != nil {
			t.Fatalf("load final image key: %v", err)
		}
		if !previous["users/initial.jpg"] {
			t.Fatalf("previous keys = %#v, want initial key", previous)
		}
		if previous[stored.ImageObjectKey] {
			t.Fatalf("final key %q was returned as a previous key: %#v", stored.ImageObjectKey, previous)
		}
		if stored.ImageObjectKey != "users/first.jpg" && stored.ImageObjectKey != "users/second.jpg" {
			t.Fatalf("final image key = %q, want one replacement", stored.ImageObjectKey)
		}
	})
}

func createEditableUserPostgres(t *testing.T, db *gorm.DB, us, email string) models.User {
	t.Helper()
	user := models.User{Name: "Пользователь", Password: "Password123!", Us: us, Email: email}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create editable user %q: %v", us, err)
	}
	return user
}
