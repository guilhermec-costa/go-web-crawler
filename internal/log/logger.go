package log

import (
	"fmt"
	"log"
)

func SetupLogger() {
	log.SetPrefix("[crawler] ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
}

func LogError(message string) {
	log.Println(GetLogErrorMsg(message))
}

func GetLogErrorMsg(message string) string {
	return fmt.Sprintf("[ERROR] %v", message)
}
