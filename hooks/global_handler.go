package hooks

import (
	"fmt"
	"log"
	"net/http"

	"github.com/parkr/auto-reply/ctx"
)

// GlobalHandler is a handy handler which can take in every event,
// choose which handlers to fire, and fires them.
type GlobalHandler struct {
	Context       *ctx.Context
	EventHandlers map[EventType][]EventHandler
}

// HandlePayload handles the actual unpacking of the payload and firing of the proper handlers.
// It will never respond with anything but a 200.
func (h *GlobalHandler) HandlePayload(w http.ResponseWriter, r *http.Request, payload []byte) {
	eventType := r.Header.Get("X-GitHub-Event")

	if eventType == "ping" {
		event := structFromPayload(eventType, payload)
		ping, ok := event.(*pingEventPayload)
		if !ok {
			log.Println(string(payload))
			http.Error(w, "you sure that was a ping message?", 200)
			return
		}
		http.Error(w, ping.Zen, 200)
		return
	}

	if handlers, ok := h.EventHandlers[EventType(eventType)]; ok {
		numHandlers := h.FireHandlers(handlers, eventType, payload)

		issueCommentHandlers, ok := h.EventHandlers[EventType(eventType)]
		if ok && EventType(eventType) == PullRequestEvent {
			numHandlers += h.FireHandlers(issueCommentHandlers, "issue_comment", payload)
		}

		fmt.Fprintf(w, "fired %d handlers", numHandlers)
	} else {
		h.Context.IncrStat("handler.invalid")
		errMessage := fmt.Sprintf("unhandled event type: %s", eventType)
		log.Printf("%s; handled events: %+v", errMessage, h.AcceptedEventTypes())
		http.Error(w, errMessage, 200)
	}

	return
}

func (h *GlobalHandler) FireHandlers(handlers []EventHandler, eventType string, payload []byte) int {
	h.Context.IncrStat("handler." + eventType)
	event := structFromPayload(eventType, payload)
	for _, handler := range handlers {
		go handler(h.Context, event)
	}
	return len(handlers)
}

// AcceptedEventTypes returns an array of all event types the GlobalHandler
// can accept.
func (h *GlobalHandler) AcceptedEventTypes() []EventType {
	keys := []EventType{}
	for k := range h.EventHandlers {
		keys = append(keys, k)
	}
	return keys
}
