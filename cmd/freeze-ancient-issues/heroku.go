// +build heroku

package main

import "log"
import _ "github.com/heroku/x/hmetrics/onload"

func init() {
	log.SetFlags(0)
}
