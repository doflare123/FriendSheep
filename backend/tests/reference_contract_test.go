package tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"friendship/models/dto"
	"friendship/services/references"
)

func TestReferenceContractsAreDeclaredInDedicatedPackage(t *testing.T) {
	referencesDir := filepath.Join("..", "services", "references")
	combined := readCombinedGoSource(t, referencesDir)

	for _, snippet := range []string{
		"type ReferenceService interface",
		"type ReferenceStore interface",
		"GetReferences(ctx context.Context)",
		"SearchGenres(ctx context.Context",
		"type GenreSearchInput struct",
		"NewReferenceService(",
		"NewGORMReferenceStore(",
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("reference package is missing source snippet %q", snippet)
		}
	}
}

func TestReferenceCleanSourceDoesNotImportStorageLibraries(t *testing.T) {
	referencesDir := filepath.Join("..", "services", "references")
	entries, err := os.ReadDir(referencesDir)
	if err != nil {
		t.Fatalf("read reference service directory: %v", err)
	}

	var checkedCleanFile bool
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" {
			continue
		}
		if strings.Contains(strings.ToLower(entry.Name()), "gorm") {
			continue
		}

		path := filepath.Join(referencesDir, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read reference source %s: %v", path, err)
		}

		checkedCleanFile = true
		text := string(content)
		for _, forbidden := range []string{
			`"friendship/repository"`,
			`"friendship/models/events"`,
			`"friendship/models/groups"`,
			`"gorm.io/`,
			`"github.com/jackc/pgx/`,
			"repository.PostgresRepository",
			"gorm.DB",
			"eventmodels.",
			"groupmodels.",
			"pgconn.PgError",
		} {
			if strings.Contains(text, forbidden) {
				t.Fatalf("%s contains storage dependency %q", path, forbidden)
			}
		}
	}

	if !checkedCleanFile {
		t.Fatal("no clean reference source found outside GORM adapters")
	}
}

func TestLegacyEventsServiceAndReferenceMethodsAreRemoved(t *testing.T) {
	eventsSource := readCombinedGoSource(t, filepath.Join("..", "services", "events"))
	for _, legacy := range []string{
		"type EventsService interface",
		"func NewEventsService(",
		"GetAllGenres(",
		"GetAllReferences(",
	} {
		if strings.Contains(eventsSource, legacy) {
			t.Fatalf("event services still contain legacy declaration %q", legacy)
		}
	}

	handlerPath := filepath.Join("..", "handlers", "HandlerEvents.go")
	content, err := os.ReadFile(handlerPath)
	if err != nil {
		t.Fatalf("read event handler: %v", err)
	}
	handlerSource := string(content)
	for _, legacy := range []string{
		"GetAllGenres(",
		"GetAllReferences(",
		"GetReferences(",
		"SearchGenres(",
		"Events     events.EventsService",
		"srv        events.EventsService",
	} {
		if strings.Contains(handlerSource, legacy) {
			t.Fatalf("event handler still contains reference dependency or method %q", legacy)
		}
	}
}

func TestGeneralReferenceContractsExcludeGenres(t *testing.T) {
	referencesDTOType := reflect.TypeOf(dto.ReferencesDto{})
	if field, exists := referencesDTOType.FieldByName("Genres"); exists {
		t.Fatalf("ReferencesDto still exposes Genres with tag %q", field.Tag.Get("json"))
	}

	for index := 0; index < referencesDTOType.NumField(); index++ {
		field := referencesDTOType.Field(index)
		if strings.Split(field.Tag.Get("json"), ",")[0] == "genres" {
			t.Fatalf("ReferencesDto field %s still exposes json tag genres", field.Name)
		}
	}

	snapshotType := reflect.TypeOf(references.ReferenceSnapshot{})
	if field, exists := snapshotType.FieldByName("Genres"); exists {
		t.Fatalf("ReferenceSnapshot still exposes Genres with type %v", field.Type)
	}

	snapshotPayload, err := json.Marshal(references.ReferenceSnapshot{})
	if err != nil {
		t.Fatalf("marshal ReferenceSnapshot: %v", err)
	}
	var snapshotJSON map[string]json.RawMessage
	if err := json.Unmarshal(snapshotPayload, &snapshotJSON); err != nil {
		t.Fatalf("decode ReferenceSnapshot JSON: %v", err)
	}
	for key := range snapshotJSON {
		if strings.EqualFold(key, "genres") {
			t.Fatalf("general reference snapshot JSON still contains genres: %s", snapshotPayload)
		}
	}

	payload, err := json.Marshal(dto.ReferencesDto{})
	if err != nil {
		t.Fatalf("marshal ReferencesDto: %v", err)
	}
	var response map[string]json.RawMessage
	if err := json.Unmarshal(payload, &response); err != nil {
		t.Fatalf("decode ReferencesDto JSON: %v", err)
	}
	if _, exists := response["genres"]; exists {
		t.Fatalf("general references JSON still contains genres: %s", payload)
	}
}

func TestGenreSearchResponseContractRemainsDedicatedAndPaginated(t *testing.T) {
	responseType := reflect.TypeOf(dto.GenreSearchResponseDto{})
	wantFields := map[string]string{
		"Items":   "items",
		"Total":   "total",
		"Page":    "page",
		"Limit":   "limit",
		"HasMore": "hasMore",
	}

	if responseType.NumField() != len(wantFields) {
		t.Fatalf("GenreSearchResponseDto has %d fields, want %d", responseType.NumField(), len(wantFields))
	}
	for fieldName, wantJSONName := range wantFields {
		field, exists := responseType.FieldByName(fieldName)
		if !exists {
			t.Fatalf("GenreSearchResponseDto is missing %s", fieldName)
		}
		if got := strings.Split(field.Tag.Get("json"), ",")[0]; got != wantJSONName {
			t.Fatalf("%s json tag = %q, want %q", fieldName, got, wantJSONName)
		}
	}
}
