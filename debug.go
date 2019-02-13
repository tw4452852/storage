package storage

import (
	"log"
)

var (
	debug = true
)

func dprintf(fmt string, v ...interface{}) {
	if debug {
		log.Printf(fmt, v...)
	}
}
