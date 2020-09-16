package top

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/unqnown/esctl/internal/app"
	"github.com/unqnown/esctl/internal/top/node"
	"github.com/unqnown/esctl/pkg/check"
	"github.com/unqnown/esctl/pkg/client"
	"github.com/unqnown/esctl/pkg/ctl"
	"github.com/unqnown/esctl/pkg/table"
	"github.com/urfave/cli"
)

var Command = cli.Command{
	Name:                   "top",
	Usage:                  "Shows resources usage.",
	Category:               "Beginner",
	UseShortOptionHandling: true,
	Subcommands: []cli.Command{
		node.Command,
		{
			Name:                   "index",
			Aliases:                []string{"indices", "indexes"},
			Usage:                  "Shows high-level information about the resource consumption of indices.",
			Description:            "For more information: https://www.elastic.co/guide/en/elasticsearch/reference/current/cat-indices.html",
			ArgsUsage:              "[indices...] [-a]",
			Action:                 ctl.NewAction(indices),
			UseShortOptionHandling: true,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "all, a",
					Usage: "Show all indices",
				},
			},
		},
		{
			Name:                   "cluster",
			Usage:                  "Shows high-level information about a cluster.",
			Description:            "For more information: https://www.elastic.co/guide/en/elasticsearch/reference/current/cluster-stats.html",
			Action:                 ctl.NewAction(cluster),
			UseShortOptionHandling: true,
		},
	},
}

func cluster(_ app.Config, conn *client.Client, _ *cli.Context) error {
	stats, err := conn.ClusterStats().Do(context.Background())
	check.Fatal(err)

	t := table.New("name", "cpu", "mem", "jvm", "disk")

	t.Append(
		[]string{
			stats.ClusterName,
			fmt.Sprintf("%v", stats.Nodes.OS.AllocatedProcessors),
			fmt.Sprintf("%.2f GB", float64(stats.Nodes.OS.Mem.UsedInBytes)/(1024*1024*1024)),
			fmt.Sprintf("%.2f GB", float64(stats.Nodes.JVM.Mem.HeapUsedInBytes)/(1024*1024*1024)),
			fmt.Sprintf("%.2f GB", float64(stats.Nodes.FS.TotalInBytes-stats.Nodes.FS.FreeInBytes)/(1024*1024*1024)),
		},
	)
	t.Render()

	return nil
}

func indices(_ app.Config, conn *client.Client, c *cli.Context) error {
	cat, err := conn.CatIndices().Columns(
		"index",
		"indexing.delete_current",
		"indexing.delete_time",
		"indexing.index_current",
		"indexing.index_time",
	).Do(context.Background())
	check.Fatalf(err, "cat indices: %v", err)

	if len(cat) == 0 {
		log.Printf("no indices")
		return nil
	}

	t := table.New("name", "delete", "delete, s", "index", "index, s")

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
		t.Append(
			[]string{
				ind.Index,
				fmt.Sprintf("%v", ind.IndexingDeleteCurrent),
				fmt.Sprintf("%v", ind.IndexingDeleteTime),
				fmt.Sprintf("%v", ind.IndexingIndexCurrent),
				fmt.Sprintf("%v", ind.IndexingIndexTime),
			},
		)
	}
	t.Render()

	return nil
}
