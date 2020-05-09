package check

import (
	"log"
)

// Fatalf performs log.Fatal on non-nil error.
func Fatal(err error) {
	if err == nil {
		return
	}
	log.Fatal(err)
}

// Fatalf performs log.Fatalf on non-nil error.
func Fatalf(err error, f string, a ...interface{}) {
	if err == nil {
		return
	}
	log.Fatalf(f, a...)
}
