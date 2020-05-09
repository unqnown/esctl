package alias

import (
	"context"
	"log"
	"os"

	"github.com/unqnown/esctl/config"
	"github.com/unqnown/esctl/pkg/check"
	"github.com/unqnown/esctl/pkg/client"
	"github.com/unqnown/esctl/pkg/ctl"
	"github.com/unqnown/esctl/pkg/io"
	"github.com/urfave/cli"
)

var Command = cli.Command{
	Name:                   "alias",
	Usage:                  "Adds or removes index aliases.",
	Description:            `For more information: https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-aliases.html`,
	ArgsUsage:              "indices... [-r] --alias alias",
	Category:               "Intermediate",
	Action:                 ctl.NewAction(alias),
	UseShortOptionHandling: true,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:     "alias, a",
			Required: true,
			Usage:    "Alias",
		},
		cli.BoolFlag{
			Name:  "remove, r",
			Usage: "Remove alias",
		},
	},
}

func alias(_ config.Config, conn *client.Client, c *cli.Context) error {
	if !c.Args().Present() {
		log.Fatal("indices not specified")
	}

	alias := conn.Alias()
	var w io.Buffer

	if a := c.String("alias"); c.Bool("remove") {
		for _, index := range c.Args() {
			alias.Remove(index, a)
			w.Writef("alias %q removed from %q\n", a, index)
		}
	} else {
		for _, index := range c.Args() {
			alias.Add(index, a)
			w.Writef("index %q aliased %q\n", index, a)
		}
	}

	_, err := alias.Do(context.Background())
	check.Fatal(err)

	_, _ = w.WriteTo(os.Stdout)

	return nil
}
