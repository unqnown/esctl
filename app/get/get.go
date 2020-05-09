package get

import (
	"context"
	"log"
	"strings"

	"github.com/unqnown/esctl/app/get/cluster"
	"github.com/unqnown/esctl/app/get/index"
	"github.com/unqnown/esctl/config"
	"github.com/unqnown/esctl/pkg/check"
	"github.com/unqnown/esctl/pkg/client"
	"github.com/unqnown/esctl/pkg/ctl"
	"github.com/unqnown/esctl/pkg/table"
	"github.com/urfave/cli"
)

var Command = cli.Command{
	Name:                   "get",
	Aliases:                []string{"ls"},
	Usage:                  "Shows one or many resources.",
	Category:               "Beginner",
	UseShortOptionHandling: true,
	Subcommands: []cli.Command{
		cluster.Command,
		index.Command,
		{
			Name:                   "node",
			Aliases:                []string{"nodes"},
			Usage:                  "Shows high-level information about nodes in a cluster.",
			Description:            "For more information: https://www.elastic.co/guide/en/elasticsearch/reference/current/cluster-nodes-info.html",
			Action:                 ctl.NewAction(nodes),
			UseShortOptionHandling: true,
		},
		{
			Name:                   "alias",
			Aliases:                []string{"aliases"},
			Usage:                  "Shows information about currently configured aliases to indices, including filter and routing information.",
			Description:            "For more information: https://www.elastic.co/guide/en/elasticsearch/reference/current/cat-alias.html",
			Action:                 ctl.NewAction(aliases),
			UseShortOptionHandling: true,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "all, a",
					Usage: "Show all indices",
				},
			},
		},
	},
}

func aliases(_ config.Config, conn *client.Client, c *cli.Context) error {
	aliases, err := conn.CatAliases().Do(context.Background())
	check.Fatal(err)

	t := table.New("index", "alias")

	all := c.Bool("all")
	for _, a := range aliases {
		if !all && strings.HasPrefix(a.Index, ".") {
			continue
		}
		t.Append(
			[]string{
				a.Index,
				a.Alias,
			},
		)
	}
	t.Render()

	return nil
}

func nodes(_ config.Config, conn *client.Client, c *cli.Context) error {
	cat, err := conn.NodesInfo().Do(context.Background())
	check.Fatalf(err, "cat nodes: %v", err)

	if len(cat.Nodes) == 0 {
		log.Printf("no nodes")
		return nil
	}

	t := table.New("id", "name", "host", "ip", "role", "version")

	for id, node := range cat.Nodes {
		t.Append(
			[]string{
				id,
				node.Name,
				node.Host,
				node.IP,
				decodeRoles(node.Roles),
				node.Version,
			},
		)
	}
	t.Render()

	return nil
}

var roles = map[string]string{
	"ingest": "i",
	"master": "m",
	"data":   "d",
}

func decodeRoles(rs []string) (s string) {
	for _, r := range rs {
		s += roles[r]
	}
	return s
}
