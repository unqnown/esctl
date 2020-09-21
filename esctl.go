package main

import (
	"log"

	"github.com/unqnown/esctl/cmd"
	"github.com/unqnown/esctl/pkg/check"
)

// Version represents application version.
var Version = "v0.1.0"

func main() {
	log.SetFlags(0)

	check.Fatal(cmd.Run(Version))
}
