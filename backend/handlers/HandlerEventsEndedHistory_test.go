package handlers

import (
	"encoding/json"
	"net/http"
	"testing"

	"friendship/models/dto"
	"friendship/services/events"

	"github.com/gin-gonic/gin"
)

func TestEndedEventCommandsMapTypedErrorToConflict(t *testing.T) {
	for _, command := range []string{"update", "delete", "leave", "kick"} {
		t.Run(command, func(t *testing.T) {
			handler := NewEventsHandler(EventsHandlerDependencies{
				Commands:   &eventCommandHandlerStub{updateErr: events.ErrEventAlreadyEnded, deleteErr: events.ErrEventAlreadyEnded},
				Membership: &eventMembershipHandlerStub{leaveErr: events.ErrEventAlreadyEnded},
				Admin:      &eventAdminHandlerStub{kickErr: events.ErrEventAlreadyEnded},
			})
			method := http.MethodDelete
			body := []byte(nil)
			if command == "update" {
				method = http.MethodPut
				body = []byte(`{"title":"Закрытая история"}`)
			}
			recorder, ctx := newEventCommandHandlerContext(t, method, "/api/v2/admin/events/17", body)
			ctx.Params = gin.Params{{Key: "eventId", Value: "17"}, {Key: "userId", Value: "14"}}
			switch command {
			case "update":
				handler.UpdateEvent(ctx)
			case "delete":
				handler.DeleteEvent(ctx)
			case "leave":
				handler.LeaveEvent(ctx)
			case "kick":
				handler.KickUserFromEvent(ctx)
			}
			if recorder.Code != http.StatusConflict {
				t.Fatalf("%s status/body=%d/%s", command, recorder.Code, recorder.Body.String())
			}
			var response dto.ErrorResponse
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if response.Error != events.ErrEventAlreadyEnded.Error() || response.Message != events.ErrEventAlreadyEnded.Error() {
				t.Fatalf("%s DTO=%#v", command, response)
			}
		})
	}
}
