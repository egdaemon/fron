package testx

import (
	"log"
	"os"
)

// Logging enable logging to stderr with file and line information.
func Logging() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.LUTC)
	log.SetOutput(os.Stderr)
}
