package main

import (
	"log"

	"github.com/unqnown/esctl/cmd"
	"github.com/unqnown/esctl/pkg/check"
)

func init() {
	log.SetFlags(0)
}

func main() {
	check.Fatal(cmd.Run())
}
