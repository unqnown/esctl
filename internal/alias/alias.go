package alias

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/unqnown/esctl/internal/app"
	"github.com/unqnown/esctl/pkg/check"
	"github.com/unqnown/esctl/pkg/client"
	"github.com/unqnown/esctl/pkg/ctl"
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

func alias(_ app.Config, conn *client.Client, c *cli.Context) error {
	if !c.Args().Present() {
		log.Fatal("indices not specified")
	}

	alias := conn.Alias()
	var w bytes.Buffer

	if a := c.String("alias"); c.Bool("remove") {
		for _, index := range c.Args() {
			alias.Remove(index, a)
			fmt.Fprintf(&w, "alias %q removed from %q\n", a, index)
		}
	} else {
		for _, index := range c.Args() {
			alias.Add(index, a)
			fmt.Fprintf(&w, "index %q aliased as %q\n", index, a)
		}
	}

	_, err := alias.Do(context.Background())
	check.Fatal(err)

	_, _ = w.WriteTo(os.Stdout)

	return nil
}
