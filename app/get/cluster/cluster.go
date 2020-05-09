package cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/unqnown/esctl/config"
	"github.com/unqnown/esctl/pkg/check"
	"github.com/unqnown/esctl/pkg/client"
	"github.com/unqnown/esctl/pkg/ctl"
	"github.com/unqnown/esctl/pkg/highlighting"
	"github.com/unqnown/esctl/pkg/pretty"
	"github.com/unqnown/esctl/pkg/table"
	"github.com/urfave/cli"
)

var Command = cli.Command{
	Name:                   "cluster",
	Usage:                  "Shows high-level information about a cluster.",
	Description:            "For more information: https://www.elastic.co/guide/en/elasticsearch/reference/current/cluster-stats.html",
	Action:                 ctl.NewAction(cluster),
	UseShortOptionHandling: true,
	Subcommands: []cli.Command{
		{
			Name:                   "health",
			Usage:                  "Shows the health status of a cluster.",
			Description:            "For more information: https://www.elastic.co/guide/en/elasticsearch/reference/current/cluster-health.html",
			Action:                 ctl.NewAction(health),
			UseShortOptionHandling: true,
		},
	},
}

func cluster(_ config.Config, conn *client.Client, c *cli.Context) error {
	stats, err := conn.ClusterStats().Do(context.Background())
	check.Fatal(err)

	t := table.New("name", "status", "age", "nodes", "indices", "p", "r", "docs", "store")

	uptime := time.Duration(stats.Nodes.JVM.MaxUptimeInMillis) * time.Millisecond
	t.Rich(
		[]string{
			stats.ClusterName,
			stats.Status,
			uptime.String(),
			fmt.Sprintf("%v", stats.Nodes.Count.Total),
			fmt.Sprintf("%v", stats.Indices.Count),
			fmt.Sprintf("%v", stats.Indices.Shards.Primaries),
			fmt.Sprintf("%v", stats.Indices.Shards.Total-stats.Indices.Shards.Primaries),
			fmt.Sprintf("%v", stats.Indices.Docs.Count),
			fmt.Sprintf("%0.1f GB", float64(stats.Indices.Store.SizeInBytes)/(1024*1024*1024)),
		},
		[]tablewriter.Colors{
			{},
			highlighting.Health[stats.Status],
		},
	)
	t.Render()

	return nil
}

func health(_ config.Config, conn *client.Client, c *cli.Context) error {
	stats, err := conn.ClusterHealth().Do(context.Background())
	check.Fatal(err)

	t := table.New("name", "")

	fmt.Printf("%s", pretty.String(stats))
	t.Append(
		[]string{
			stats.ClusterName,
		},
	)
	t.Render()

	return nil
}
