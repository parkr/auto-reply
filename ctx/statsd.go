package ctx

import (
	"log"

	"github.com/DataDog/datadog-go/v5/statsd"
)

var (
	hostport  string  = "127.0.0.1:8125"
	namespace string  = "autoreply."
	countRate float64 = 1
)

func NewStatsd() *statsd.Client {
	client, err := statsd.New(
		hostport,
		statsd.WithNamespace(namespace),
	)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	return client
}

func (c *Context) IncrStat(name string, tags []string) {
	c.CountStat(name, 1, tags)
}

func (c *Context) CountStat(name string, value int64, tags []string) {
	if c.Statsd != nil {
		_ = c.Statsd.Count(name, value, tags, countRate)
	}
}
