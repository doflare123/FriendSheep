package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"friendship/models/dto"
	group "friendship/services/groups"

	"github.com/gin-gonic/gin"
)

type groupEnterpriseHandlerServiceStub struct {
	group.GroupsService

	createActorID uint
	createInput   group.CreateGroupInput
	updateActorID uint
	updateInput   group.GroupUpdateInput
}

func (s *groupEnterpriseHandlerServiceStub) CreateGroup(_ context.Context, actorID uint, input group.CreateGroupInput) (*dto.GroupFullDto, error) {
	s.createActorID = actorID
	s.createInput = input
	return &dto.GroupFullDto{Enterprise: false}, nil
}

func (s *groupEnterpriseHandlerServiceStub) UpdateGroup(_ context.Context, actorID uint, input group.GroupUpdateInput) (*dto.GroupFullDto, error) {
	s.updateActorID = actorID
	s.updateInput = input
	return &dto.GroupFullDto{Enterprise: true}, nil
}

func TestCreateGroupIgnoresEnterpriseAndReturnsSystemDefault(t *testing.T) {
	service := &groupEnterpriseHandlerServiceStub{}
	handler := NewGroupHandler(service)
	recorder, ctx := newGroupEnterpriseHandlerContext(t, http.MethodPost, `/api/v2/groups`, `{
		"name":"Board Game Club",
		"description":"Group for board game fans",
		"smallDescription":"Play together",
		"image":"https://example.com/group.png",
		"isPrivate":false,
		"enterprise":true,
		"categories":[1]
	}`)

	handler.CreateGroup(ctx)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body: %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}
	if service.createActorID != 7 {
		t.Fatalf("actor id = %d, want 7", service.createActorID)
	}
	assertGroupWriteInputHasNoEnterprise(t, CreateGroupRequest{})
	assertGroupWriteInputHasNoEnterprise(t, service.createInput)
	assertGroupEnterpriseResponse(t, recorder, false)
}

func TestUpdateGroupIgnoresEnterpriseAndReturnsStoredMarker(t *testing.T) {
	service := &groupEnterpriseHandlerServiceStub{}
	handler := NewGroupHandler(service)
	recorder, ctx := newGroupEnterpriseHandlerContext(t, http.MethodPut, `/api/v2/groups`, `{
		"groupId":42,
		"enterprise":false
	}`)

	handler.UpdateGroup(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if service.updateActorID != 7 {
		t.Fatalf("actor id = %d, want 7", service.updateActorID)
	}
	if service.updateInput.GroupID != 42 {
		t.Fatalf("group id = %d, want 42", service.updateInput.GroupID)
	}
	assertGroupWriteInputHasNoEnterprise(t, GroupUpdateRequest{})
	assertGroupWriteInputHasNoEnterprise(t, service.updateInput)
	assertGroupEnterpriseResponse(t, recorder, true)
}

func newGroupEnterpriseHandlerContext(t *testing.T, method string, target string, body string) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(method, target, bytes.NewBufferString(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Set("userID", uint(7))
	return recorder, ctx
}

func assertGroupEnterpriseResponse(t *testing.T, recorder *httptest.ResponseRecorder, want bool) {
	t.Helper()

	var payload map[string]json.RawMessage
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v; body: %s", err, recorder.Body.String())
	}

	wantJSON := json.RawMessage("false")
	if want {
		wantJSON = json.RawMessage("true")
	}
	if string(payload["enterprise"]) != string(wantJSON) {
		t.Fatalf("response enterprise = %s, want %s; body: %s", payload["enterprise"], wantJSON, recorder.Body.String())
	}
}

func assertGroupWriteInputHasNoEnterprise(t *testing.T, input interface{}) {
	t.Helper()

	inputType := reflect.TypeOf(input)
	if _, exists := inputType.FieldByName("Enterprise"); exists {
		t.Fatalf("%s exposes read-only Enterprise in a write input", inputType)
	}
}
