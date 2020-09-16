package doc

import (
	"context"
	"log"

	"github.com/olivere/elastic/v7"
	"github.com/unqnown/esctl/internal/app"
	"github.com/unqnown/esctl/pkg/check"
	"github.com/unqnown/esctl/pkg/client"
	"github.com/unqnown/esctl/pkg/ctl"
	"github.com/urfave/cli"
)

var Command = cli.Command{
	Name:                   "document",
	Aliases:                []string{"documents", "docs", "doc"},
	Usage:                  "Deletes documents from index.",
	Description:            `For more information: https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-delete.html`,
	ArgsUsage:              "id...",
	Category:               "Beginner",
	Action:                 ctl.NewAction(document),
	UseShortOptionHandling: true,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:     "index, i",
			Required: true,
			Usage:    "Index documents delete from",
		},
	},
}

func document(_ app.Config, conn *client.Client, c *cli.Context) error {
	if !c.Args().Present() {
		log.Fatal("id not specified")
	}

	// TODO(d.andriichuk): rewrite to bulk usage.
	for _, id := range c.Args() {
		_, err := conn.Delete().Index(c.String("index")).Id(id).Do(context.Background())
		switch {
		case elastic.IsNotFound(err):
			log.Printf("%q not found", id)
			continue
		default:
			check.Fatal(err)
		}
		log.Printf("%q deleted", id)
	}

	return nil
}
