// +build heroku

package main

import "log"

func init() {
	log.SetFlags(log.Lshortfile)
}
