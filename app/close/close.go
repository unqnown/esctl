package close

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
	Name:                   "close",
	Usage:                  "Closes an index.",
	Description:            "For more information: https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-close.html",
	Action:                 ctl.NewAction(close),
	Category:               "Intermediate",
	UseShortOptionHandling: true,
}

func close(_ config.Config, conn *client.Client, c *cli.Context) error {
	if !c.Args().Present() {
		log.Fatal("index not specified")
	}

	index := c.Args().First()

	_, err := conn.CloseIndex(index).Do(context.Background())
	check.Fatal(err)

	log.Printf("%q closed", index)

	return nil
}
