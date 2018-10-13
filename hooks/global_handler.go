package hooks

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/ctx"
)

type EventHandlerMap map[EventType][]EventHandler

func (m EventHandlerMap) AddHandler(eventType EventType, handler EventHandler) {
	if m[eventType] == nil {
		m[eventType] = []EventHandler{}
	}

	m[eventType] = append(m[eventType], handler)
}

// GlobalHandler is a handy handler which can take in every event,
// choose which handlers to fire, and fires them.
type GlobalHandler struct {
	Context       *ctx.Context
	EventHandlers EventHandlerMap

	// secret is the secret used by GitHub to validate the integrity of the
	// request. It is given to GitHub in the webhook management interface.
	secret []byte
}

// ServeHTTP handles the incoming HTTP request, validates the payload and
// fires
func (h *GlobalHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	var payload []byte
	var err error
	if secret := h.getSecret(); len(secret) > 0 {
		payload, err = github.ValidatePayload(r, secret)
		if err != nil {
			log.Println("received invalid signature:", err)
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
	} else {
		err = json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			log.Println("received invalid json in body:", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	h.HandlePayload(w, r, payload)
}

// HandlePayload handles the actual unpacking of the payload and firing of the proper handlers.
// It will never respond with anything but a 200.
func (h *GlobalHandler) HandlePayload(w http.ResponseWriter, r *http.Request, payload []byte) {
	eventType := github.WebHookType(r)

	if eventType == "ping" {
		handlePingPayload(w, r, payload)
		return
	}

	if os.Getenv("AUTO_REPLY_DEBUG") == "true" {
		log.Printf("payload: %s %s", eventType, string(payload))
	}

	if handlers, ok := h.EventHandlers[EventType(eventType)]; ok {
		numHandlers := h.FireHandlers(handlers, eventType, payload)

		if EventType(eventType) == PullRequestEvent {
			if issueCommentHandlers, ok := h.EventHandlers[EventType(eventType)]; ok {
				numHandlers += h.FireHandlers(issueCommentHandlers, "issue_comment", payload)
			}
		}

		fmt.Fprintf(w, "fired %d handlers", numHandlers)
	} else {
		h.Context.IncrStat("handler.invalid", nil)
		errMessage := fmt.Sprintf("unhandled event type: %s", eventType)
		log.Printf("%s; handled events: %+v", errMessage, h.AcceptedEventTypes())
		http.Error(w, errMessage, 200)
	}

	return
}

func (h *GlobalHandler) FireHandlers(handlers []EventHandler, eventType string, payload []byte) int {
	h.Context.IncrStat("handler."+eventType, nil)
	event, err := github.ParseWebHook(eventType, payload)
	if err != nil {
		h.Context.NewError("FireHandlers: couldn't parse webhook: %+v", err)
		return 0
	}
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

func (h *GlobalHandler) getSecret() []byte {
	if len(h.secret) > 0 {
		return h.secret
	}

	// Fill the value of h.secret if one exists.
	if envVal := os.Getenv("GITHUB_WEBHOOK_SECRET"); envVal != "" {
		h.secret = []byte(envVal)
	}
	return h.secret
}

func handlePingPayload(w http.ResponseWriter, r *http.Request, payload []byte) {
	var ping pingEventPayload
	if err := json.Unmarshal(payload, &ping); err != nil {
		log.Println(string(payload))
		http.Error(w, "you sure that was a ping message?", 500)
		log.Printf("GlobalHandler.HandlePayload: couldn't handle ping payload: %+v", err)
		return
	}
	http.Error(w, ping.Zen, 200)
}
