package dashboard

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

var throttle <-chan time.Time

func init() {
	rate := time.Second / 30
	throttle = time.Tick(rate)
}

func get(url string, data interface{}) error {
	<-throttle
	log.Println("==> GET", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	return json.NewDecoder(resp.Body).Decode(data)
}

func getRetry(retries int, url string, data interface{}) error {
	var err error
	tries := 0
	for tries <= retries {
		tries += 1
		err = get(url, data)
		if err == nil {
			break
		}
	}
	return err
}
