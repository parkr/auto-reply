package hooks

import (
	"net/http"

	"github.com/parkr/auto-reply/ctx"
)

// HookHandler describes the interface for any type which can handle a webhook payload.
type HookHandler interface {
	HandlePayload(w http.ResponseWriter, r *http.Request, payload []byte)
}

// EventHandler is An event handler takes in a given event and operates on it.
type EventHandler func(context *ctx.Context, event interface{}) error
