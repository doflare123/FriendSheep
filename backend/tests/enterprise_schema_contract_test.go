package tests

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	friendshipdocs "friendship/docs"
	"friendship/models"
	"friendship/models/dto"
	groupmodels "friendship/models/groups"
	"friendship/services"
)

func TestUserContractsDoNotExposeEnterprise(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
	}{
		{name: "persistence model", value: models.User{}},
		{name: "user dto", value: dto.UserDto{}},
		{name: "user information response", value: services.InformationAboutUser{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valueType := reflect.TypeOf(tt.value)
			if _, exists := valueType.FieldByName("Enterprise"); exists {
				t.Fatalf("%s still declares Enterprise", valueType)
			}

			payload, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatalf("marshal %s: %v", valueType, err)
			}

			var fields map[string]json.RawMessage
			if err := json.Unmarshal(payload, &fields); err != nil {
				t.Fatalf("decode %s JSON: %v", valueType, err)
			}
			if _, exists := fields["enterprise"]; exists {
				t.Fatalf("%s JSON still exposes enterprise: %s", valueType, payload)
			}
		})
	}
}

func TestGroupContractsOwnEnterprise(t *testing.T) {
	field, exists := reflect.TypeOf(groupmodels.Group{}).FieldByName("Enterprise")
	if !exists {
		t.Fatal("groups.Group does not declare Enterprise")
	}
	if field.Type.Kind() != reflect.Bool {
		t.Fatalf("groups.Group Enterprise type = %s, want bool", field.Type)
	}
	if field.Tag.Get("json") != "enterprise" {
		t.Fatalf("groups.Group Enterprise json tag = %q, want enterprise", field.Tag.Get("json"))
	}
	if !strings.Contains(field.Tag.Get("gorm"), "default:false") {
		t.Fatalf("groups.Group Enterprise gorm tag = %q, want default:false", field.Tag.Get("gorm"))
	}

	payload, err := json.Marshal(dto.GroupFullDto{Enterprise: true})
	if err != nil {
		t.Fatalf("marshal group dto: %v", err)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(payload, &fields); err != nil {
		t.Fatalf("decode group dto JSON: %v", err)
	}
	if string(fields["enterprise"]) != "true" {
		t.Fatalf("group dto enterprise = %s, want true (payload: %s)", fields["enterprise"], payload)
	}
}

func TestPublicGroupSearchResponseSerializesEnterprise(t *testing.T) {
	enterprise := true
	regular := false
	payload, err := json.Marshal([]services.GetGroups{
		{Enterprise: &enterprise},
		{Enterprise: &regular},
	})
	if err != nil {
		t.Fatalf("marshal public group search items: %v", err)
	}

	var items []map[string]json.RawMessage
	if err := json.Unmarshal(payload, &items); err != nil {
		t.Fatalf("decode public group search items: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("items = %#v, want two items", items)
	}
	if string(items[0]["enterprise"]) != "true" || string(items[1]["enterprise"]) != "false" {
		t.Fatalf("enterprise values = %s/%s, want true/false; payload: %s", items[0]["enterprise"], items[1]["enterprise"], payload)
	}
}

func TestGroupSwaggerKeepsEnterpriseReadOnly(t *testing.T) {
	var document struct {
		Definitions map[string]struct {
			Properties map[string]json.RawMessage `json:"properties"`
		} `json:"definitions"`
	}

	if err := json.Unmarshal([]byte(friendshipdocs.SwaggerInfo.ReadDoc()), &document); err != nil {
		t.Fatalf("decode generated Swagger document: %v", err)
	}

	for _, definitionName := range []string{
		"handlers.CreateGroupRequest",
		"handlers.GroupUpdateRequest",
	} {
		definition, exists := document.Definitions[definitionName]
		if !exists {
			t.Fatalf("Swagger definition %q is missing", definitionName)
		}
		if _, exists := definition.Properties["enterprise"]; exists {
			t.Fatalf("Swagger write definition %q exposes read-only enterprise", definitionName)
		}
	}

	for _, definitionName := range []string{
		"dto.GroupFullDto",
		"dto.ManagedGroupItemDto",
		"dto.EventSearchGroupDto",
		"services.GetGroups",
	} {
		responseDefinition, exists := document.Definitions[definitionName]
		if !exists {
			t.Fatalf("Swagger response definition %q is missing", definitionName)
		}
		if _, exists := responseDefinition.Properties["enterprise"]; !exists {
			t.Fatalf("Swagger response definition %q does not expose enterprise", definitionName)
		}
	}
}
