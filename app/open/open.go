package open

import (
	"context"
	"log"

	"github.com/unqnown/esctl/config"
	"github.com/unqnown/esctl/pkg/check"
	"github.com/unqnown/esctl/pkg/client"
	"github.com/unqnown/esctl/pkg/ctl"
	"github.com/urfave/cli"
)

var Command = cli.Command{
	Name:                   "open",
	Usage:                  "Opens a closed index.",
	Action:                 ctl.NewAction(open),
	Category:               "Intermediate",
	UseShortOptionHandling: true,
}

func open(_ config.Config, conn *client.Client, c *cli.Context) error {
	if !c.Args().Present() {
		log.Fatal("index not specified")
	}

	index := c.Args().First()

	_, err := conn.OpenIndex(index).Do(context.Background())
	check.Fatal(err)

	log.Printf("%q opened", index)

	return nil
}
