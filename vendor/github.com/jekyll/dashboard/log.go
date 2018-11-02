//+build heroku

package dashboard

import "log"

func init() {
	log.SetFlags(log.Lshortfile)
}
