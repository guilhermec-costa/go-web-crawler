package log

import (
	"log"
)

func SetupLogger() {
	log.SetPrefix("[crawler] ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
}