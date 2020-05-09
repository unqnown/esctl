package index

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-yaml/yaml"
	"github.com/olekukonko/tablewriter"
	"github.com/unqnown/esctl/config"
	"github.com/unqnown/esctl/pkg/check"
	"github.com/unqnown/esctl/pkg/client"
	"github.com/unqnown/esctl/pkg/ctl"
	"github.com/unqnown/esctl/pkg/highlighting"
	"github.com/unqnown/esctl/pkg/io"
	"github.com/unqnown/esctl/pkg/table"
	"github.com/urfave/cli"
)

var Command = cli.Command{
	Name:                   "index",
	Aliases:                []string{"indices", "indexes"},
	Usage:                  "Shows high-level information about indices in a cluster.",
	Description:            "For more information: https://www.elastic.co/guide/en/elasticsearch/reference/current/cat-indices.html",
	ArgsUsage:              "[indices...] [-a] [-o yaml|json]",
	Action:                 ctl.NewAction(indices),
	UseShortOptionHandling: true,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "all, a",
			Usage: "Show all indices",
		},
		cli.StringFlag{
			Name:  "output, o",
			Usage: "Output format",
		},
	},
}

func explicit(_ config.Config, conn *client.Client, c *cli.Context) error {
	indices, err := conn.IndexGet(c.Args()...).Do(context.Background())
	check.Fatal(err)

	var w io.Buffer
	switch c.String("output") {
	case "yaml", "yml":
		err = yaml.NewEncoder(&w).Encode(indices)
	default:
		enc := json.NewEncoder(&w)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "	")
		err = enc.Encode(indices)
	}
	check.Fatal(err)

	_, _ = w.WriteTo(os.Stdout)

	return nil
}

func indices(conf config.Config, conn *client.Client, c *cli.Context) error {
	if c.String("output") != "" {
		return cli.HandleAction(ctl.Call(explicit, conf, conn), c)
	}

	cat, err := conn.CatIndices().Do(context.Background())
	check.Fatalf(err, "cat indices: %v", err)

	if len(cat) == 0 {
		log.Printf("no indices")
		return nil
	}

	t := table.New("health", "status", "name", "p", "r", "docs", "deleted", "size", "primary.size")

	indices := make(map[string]struct{}, len(c.Args()))
	for _, ind := range c.Args() {
		indices[ind] = struct{}{}
	}
	filter := c.Args().Present()

	all := c.Bool("all")
	for _, ind := range cat {
		if !all && strings.HasPrefix(ind.Index, ".") {
			continue
		}
		if _, found := indices[ind.Index]; filter && !found {
			continue
		}
		t.Rich(
			[]string{
				ind.Health,
				ind.Status,
				ind.Index,
				fmt.Sprintf("%v", ind.Pri),
				fmt.Sprintf("%v", ind.Rep),
				fmt.Sprintf("%v", ind.DocsCount),
				fmt.Sprintf("%v", ind.DocsDeleted),
				ind.StoreSize,
				ind.PriStoreSize,
			},
			[]tablewriter.Colors{
				highlighting.Health[ind.Health],
			},
		)
	}
	t.Render()
	return nil
}
