package dashboard

// +build heroku

import "log"

func init() {
	log.SetFlags(log.Lshortfile)
}
