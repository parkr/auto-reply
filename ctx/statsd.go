package ctx

import (
	"log"

	"github.com/DataDog/datadog-go/statsd"
)

var (
	countRate float64 = 1
	noTags            = []string{}
)

func NewStatsd() *statsd.Client {
	client, err := statsd.New("127.0.0.1:8125")
	if err != nil {
		log.Fatal(err)
		return nil
	}
	client.Namespace = "autoreply."
	return client
}

func (c *Context) IncrStat(name string) {
	c.CountStat(name, 1)
}

func (c *Context) CountStat(name string, value int64) {
	if c.Statsd != nil {
		c.Statsd.Count(name, value, noTags, countRate)
	}
}
