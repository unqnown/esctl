package get

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/unqnown/esctl/internal/app"
	"github.com/unqnown/esctl/internal/get/cluster"
	"github.com/unqnown/esctl/internal/get/index"
	"github.com/unqnown/esctl/pkg/check"
	"github.com/unqnown/esctl/pkg/client"
	"github.com/unqnown/esctl/pkg/ctl"
	"github.com/unqnown/esctl/pkg/highlighting"
	"github.com/unqnown/esctl/pkg/table"
	"github.com/urfave/cli"
)

var Command = cli.Command{
	Name:                   "get",
	Aliases:                []string{"ls", "cat"},
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
					Usage: "Show all aliases",
				},
			},
		},
		{
			Name:                   "shard",
			Aliases:                []string{"shards"},
			Usage:                  "Shows high-level information about shards in a cluster.",
			Description:            "For more information: https://www.elastic.co/guide/en/elasticsearch/reference/current/cat-shards.html",
			Action:                 ctl.NewAction(shards),
			UseShortOptionHandling: true,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "colorless, c",
					Usage: "Colorless output",
				},
				cli.BoolFlag{
					Name:  "unassigned, u",
					Usage: "Reason the shard is unassigned",
				},
			},
		},
	},
}

func aliases(_ app.Config, conn *client.Client, c *cli.Context) error {
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

func nodes(_ app.Config, conn *client.Client, c *cli.Context) error {
	cat, err := conn.NodesInfo().Do(context.Background())
	check.Fatalf(err, "cat nodes: %v", err)

	if len(cat.Nodes) == 0 {
		log.Printf("no nodes found")
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

func shards(_ app.Config, conn *client.Client, c *cli.Context) error {
	columns := []string{"index", "shard", "prirep", "state"}
	filter := make(map[string]bool)

	unassigned := c.Bool("unassigned")
	if unassigned {
		columns = append(columns, "ur")
		filter["UNASSIGNED"] = true
	}

	cat, err := conn.CatShards().Index(c.Args()...).Columns(columns...).Do(context.Background())
	check.Fatalf(err, "cat shards: %v", err)

	if len(cat) == 0 {
		log.Printf("no shards found")
		return nil
	}

	t := table.New(columns...)
	colorless := c.Bool("colorless")

	for _, shard := range cat {
		if !show(filter, shard.State) {
			continue
		}

		var colors []tablewriter.Colors
		if !colorless {
			colors = append(colors,
				tablewriter.Colors{},
				tablewriter.Colors{},
				tablewriter.Colors{},
				highlighting.State[shard.State],
			)
		}

		row := []string{
			shard.Index,
			fmt.Sprint(shard.Shard),
			shard.Prirep,
			shard.State,
		}
		if unassigned {
			row = append(row, shard.UnassignedReason)
		}

		t.Rich(row, colors)
	}
	t.Render()

	return nil
}

// TODO(d.andriichuk): find in common and move to pkg.
func show(set map[string]bool, vs ...string) bool {
	if len(set) == 0 {
		return true
	}
	for _, v := range vs {
		if set[v] {
			return true
		}
	}
	return false
}
