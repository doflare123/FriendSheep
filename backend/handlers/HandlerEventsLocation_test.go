package handlers

import (
	"fmt"
	"net/http"
	"testing"

	"friendship/services/events"
)

func TestCreateEventMapsLocationErrorsToBadRequest(t *testing.T) {
	for _, err := range []error{
		fmt.Errorf("%w: для онлайн-мероприятия поле link обязательно", events.ErrInvalidEventUpdate),
		events.ErrEventLocationNotFound,
	} {
		t.Run(err.Error(), func(t *testing.T) {
			stub := &eventCommandHandlerStub{createErr: err}
			handler := NewEventsHandler(EventsHandlerDependencies{Commands: stub})
			body := []byte(`{"title":"Valid title","description":"Valid description","groupId":1,"eventTypeId":1,"locationId":1,"imageUrl":"https://example.com/a.png","startTime":"2035-01-02T10:00:00Z","duration":60,"maxUsers":10,"genres":[1],"ageLimit":1}`)
			recorder, ctx := newEventCommandHandlerContext(t, http.MethodPost, "/api/v2/admin/events", body)
			handler.CreateEvent(ctx)
			assertEventCommandHandlerError(t, recorder, http.StatusBadRequest, err.Error())
		})
	}
}
