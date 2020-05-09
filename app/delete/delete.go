package delete

import (
	"context"
	"log"

	"github.com/unqnown/esctl/app/delete/doc"
	"github.com/unqnown/esctl/config"
	"github.com/unqnown/esctl/pkg/check"
	"github.com/unqnown/esctl/pkg/client"
	"github.com/unqnown/esctl/pkg/ctl"
	"github.com/urfave/cli"
)

var Command = cli.Command{
	Name:                   "delete",
	Usage:                  "Deletes an existing index.",
	Description:            `For more information: https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-delete-index.html`,
	ArgsUsage:              "indices...",
	Category:               "Beginner",
	Action:                 ctl.NewAction(delete),
	UseShortOptionHandling: true,
	Subcommands: []cli.Command{
		doc.Command,
	},
}

func delete(_ config.Config, conn *client.Client, c *cli.Context) error {
	if !c.Args().Present() {
		log.Fatal("indices not specified")
	}

	_, err := conn.DeleteIndex(c.Args()...).Do(context.Background())
	check.Fatal(err)

	for _, ind := range c.Args() {
		log.Printf("%q deleted", ind)
	}

	return nil
}
