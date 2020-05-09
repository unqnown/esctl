package main

import (
	"github.com/unqnown/esctl/app"
	"github.com/unqnown/esctl/pkg/check"
	"log"
)

func main() {
	check.Fatal(app.Run())
}

func init() {
	log.SetFlags(0)
}
