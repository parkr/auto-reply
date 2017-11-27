package sentry

import (
	"errors"
	"net/http"
	"os"

	"github.com/getsentry/raven-go"
)

var sentryEnvVarName = "SENTRY_DSN"

func newRavenClient(tags map[string]string) (*raven.Client, error) {
	dsn := os.Getenv(sentryEnvVarName)
	if dsn == "" {
		return nil, errors.New("sentry: missing env var " + sentryEnvVarName)
	}
	return raven.NewWithTags(dsn, tags)
}

//
// Top-level SentryClient which should be lowest-level interface.
//

type SentryClient interface {
	Recover(func() error)
	GetSentry() *raven.Client
}

func NewClient(tags map[string]string) (SentryClient, error) {
	ravenClient, err := newRavenClient(tags)
	return &sentryClient{ravenClient: ravenClient}, err
}

type sentryClient struct {
	ravenClient *raven.Client
}

func (c *sentryClient) Recover(f func() error) {
	c.ravenClient.CapturePanicAndWait(func() {
		if err := f(); err != nil {
			panic(err)
		}
	}, nil)
}

func (c *sentryClient) GetSentry() *raven.Client {
	return c.ravenClient
}

//
// HTTP wrapper for Sentry
//

type sentryHTTPHandler struct {
	next http.Handler

	ravenClient SentryClient
}

func NewHTTPHandler(handler http.Handler, tags map[string]string) http.Handler {
	client, err := NewClient(tags)
	if err != nil {
		panic(err)
	}
	return &sentryHTTPHandler{
		next:        handler,
		ravenClient: client,
	}
}

func (h *sentryHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.ravenClient.Recover(func() error {
		h.next.ServeHTTP(w, r)
		return nil
	})
}
