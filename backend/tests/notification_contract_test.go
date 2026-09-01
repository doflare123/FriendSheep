package tests

import (
	"encoding/json"
	"testing"

	friendshipdocs "friendship/docs"
)

func TestNotificationInboxSwaggerContract(t *testing.T) {
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
			Security []map[string][]string `json:"security"`
		} `json:"paths"`
	}
	if err := json.Unmarshal([]byte(friendshipdocs.SwaggerInfo.ReadDoc()), &document); err != nil {
		t.Fatalf("decode generated Swagger document: %v", err)
	}

	tests := []struct {
		path          string
		method        string
		wantResponse  string
		wantParameter string
	}{
		{
			path:         "/api/v2/users/me/notifications",
			method:       "get",
			wantResponse: "#/definitions/notifications.InboxPage",
		},
		{
			path:         "/api/v2/users/me/notifications/unread-count",
			method:       "get",
			wantResponse: "#/definitions/notifications.UnreadCount",
		},
		{
			path:          "/api/v2/users/me/notifications/{notificationId}/read",
			method:        "patch",
			wantResponse:  "#/definitions/notifications.MarkReadResult",
			wantParameter: "notificationId",
		},
	}
	for _, test := range tests {
		operation, exists := document.Paths[test.path][test.method]
		if !exists {
			t.Fatalf("Swagger operation %s %s is missing", test.method, test.path)
		}
		if len(operation.Security) != 1 {
			t.Fatalf("%s %s security = %#v, want BearerAuth", test.method, test.path, operation.Security)
		}
		if _, exists := operation.Security[0]["BearerAuth"]; !exists {
			t.Fatalf("%s %s does not require BearerAuth", test.method, test.path)
		}

		response, exists := operation.Responses["200"]
		if !exists {
			t.Fatalf("%s %s response 200 is missing", test.method, test.path)
		}
		var schema struct {
			Ref string `json:"$ref"`
		}
		if err := json.Unmarshal(response.Schema, &schema); err != nil {
			t.Fatalf("decode %s %s response schema: %v", test.method, test.path, err)
		}
		if schema.Ref != test.wantResponse {
			t.Fatalf("%s %s response schema = %q, want %q", test.method, test.path, schema.Ref, test.wantResponse)
		}

		if test.wantParameter != "" {
			found := false
			for _, parameter := range operation.Parameters {
				if parameter.Name == test.wantParameter && parameter.In == "path" && parameter.Required {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("%s %s required path parameter %q is missing", test.method, test.path, test.wantParameter)
			}
		}
	}
}
