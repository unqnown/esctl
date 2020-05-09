package create

import (
	"context"
	"io/ioutil"
	"log"

	"github.com/unqnown/esctl/config"
	"github.com/unqnown/esctl/pkg/check"
	"github.com/unqnown/esctl/pkg/client"
	"github.com/unqnown/esctl/pkg/ctl"
	"github.com/urfave/cli"
)

var Command = cli.Command{
	Name:                   "create",
	Aliases:                []string{"new"},
	Usage:                  "Creates a new index.",
	Description:            "For more information: https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-create-index.html.",
	ArgsUsage:              "indices... --body path/to/body.json",
	Category:               "Beginner",
	Action:                 ctl.NewAction(create),
	UseShortOptionHandling: true,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:     "body, b",
			Required: true,
			Usage:    "Index body `FILE`",
		},
	},
}

func create(_ config.Config, conn *client.Client, c *cli.Context) error {
	if !c.Args().Present() {
		log.Fatal("indices not specified")
	}

	body, err := ioutil.ReadFile(c.String("body"))
	check.Fatalf(err, "open body: %v", err)

	for _, index := range c.Args() {
		_, err = conn.CreateIndex(index).
			Body(string(body)).
			Do(context.Background())
		check.Fatalf(err, "create %q: %v", index, err)
		log.Printf("%q created", index)
	}

	return nil
}
