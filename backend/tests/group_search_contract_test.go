package tests

import (
	"encoding/json"
	"strings"
	"testing"

	friendshipdocs "friendship/docs"
)

func TestGroupSearchSwaggerContract(t *testing.T) {
	var document struct {
		Paths map[string]map[string]struct {
			Parameters []struct {
				Name     string `json:"name"`
				In       string `json:"in"`
				Required bool   `json:"required"`
			} `json:"parameters"`
			Responses map[string]struct {
				Schema json.RawMessage `json:"schema"`
			} `json:"responses"`
			Security json.RawMessage `json:"security"`
		} `json:"paths"`
		Definitions map[string]struct {
			Properties map[string]json.RawMessage `json:"properties"`
		} `json:"definitions"`
	}

	if err := json.Unmarshal([]byte(friendshipdocs.SwaggerInfo.ReadDoc()), &document); err != nil {
		t.Fatalf("decode generated Swagger document: %v", err)
	}

	path, exists := document.Paths["/api/v2/groups/search"]
	if !exists {
		t.Fatal("Swagger path /api/v2/groups/search is missing")
	}
	operation, exists := path["get"]
	if !exists {
		t.Fatal("Swagger GET /api/v2/groups/search operation is missing")
	}
	if len(operation.Security) > 0 && string(operation.Security) != "null" && string(operation.Security) != "[]" {
		t.Fatalf("public group search unexpectedly declares security: %s", operation.Security)
	}

	wantParameters := map[string]bool{
		"q": false, "categoryIds": false, "isPrivate": false, "city": false,
		"sortBy": false, "sortOrder": false, "page": false, "limit": false,
	}
	if len(operation.Parameters) != len(wantParameters) {
		t.Fatalf("query parameters = %#v, want exactly %v", operation.Parameters, groupSearchMapKeys(wantParameters))
	}
	for _, parameter := range operation.Parameters {
		if parameter.In != "query" || parameter.Required {
			t.Fatalf("parameter %#v must be an optional query parameter", parameter)
		}
		if _, exists := wantParameters[parameter.Name]; !exists {
			t.Fatalf("unexpected query parameter %q", parameter.Name)
		}
		wantParameters[parameter.Name] = true
	}
	for name, found := range wantParameters {
		if !found {
			t.Fatalf("Swagger operation is missing query parameter %q", name)
		}
	}

	response, exists := operation.Responses["200"]
	if !exists {
		t.Fatal("Swagger group search response 200 is missing")
	}
	responseDefinition := swaggerDefinitionFromRef(t, response.Schema, document.Definitions)
	assertExactSwaggerProperties(t, "group search response", responseDefinition.Properties, []string{
		"items", "total", "page", "limit", "totalPages", "hasMore",
	})

	var itemsProperty struct {
		Items json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(responseDefinition.Properties["items"], &itemsProperty); err != nil {
		t.Fatalf("decode group search items schema: %v", err)
	}
	itemDefinition := swaggerDefinitionFromRef(t, itemsProperty.Items, document.Definitions)
	assertExactSwaggerProperties(t, "group search item", itemDefinition.Properties, []string{
		"id", "name", "categories", "memberCount", "image", "createdAt", "smallDescription", "isPrivate", "enterprise", "isSubscribed",
	})
	for _, status := range []string{"400", "401", "500", "503"} {
		if _, exists := operation.Responses[status]; !exists {
			t.Fatalf("Swagger group search response %s is missing", status)
		}
	}
}

func swaggerDefinitionFromRef[T any](t *testing.T, raw json.RawMessage, definitions map[string]T) T {
	t.Helper()

	var schema struct {
		Ref string `json:"$ref"`
	}
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatalf("decode Swagger schema reference: %v", err)
	}
	const prefix = "#/definitions/"
	if !strings.HasPrefix(schema.Ref, prefix) {
		t.Fatalf("Swagger schema ref = %q, want definitions reference", schema.Ref)
	}
	name := strings.TrimPrefix(schema.Ref, prefix)
	definition, exists := definitions[name]
	if !exists {
		t.Fatalf("Swagger definition %q is missing", name)
	}
	return definition
}

func assertExactSwaggerProperties(t *testing.T, subject string, properties map[string]json.RawMessage, want []string) {
	t.Helper()

	if len(properties) != len(want) {
		t.Fatalf("%s properties = %v, want exactly %v", subject, groupSearchMapKeys(properties), want)
	}
	for _, name := range want {
		if _, exists := properties[name]; !exists {
			t.Fatalf("%s is missing property %q", subject, name)
		}
	}
}

func groupSearchMapKeys[V any](items map[string]V) []string {
	result := make([]string, 0, len(items))
	for key := range items {
		result = append(result, key)
	}
	return result
}
