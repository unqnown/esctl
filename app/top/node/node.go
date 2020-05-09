package node

import (
	"context"
	"fmt"
	"log"

	"github.com/unqnown/esctl/config"
	"github.com/unqnown/esctl/pkg/check"
	"github.com/unqnown/esctl/pkg/client"
	"github.com/unqnown/esctl/pkg/ctl"
	"github.com/unqnown/esctl/pkg/table"
	"github.com/urfave/cli"
)

var Command = cli.Command{
	Name:                   "node",
	Aliases:                []string{"nodes"},
	Usage:                  "Shows high-level information about nodes in a cluster.",
	Description:            "For more information: https://www.elastic.co/guide/en/elasticsearch/reference/current/cluster-nodes-info.html",
	Action:                 ctl.NewAction(nodes),
	UseShortOptionHandling: true,
	Subcommands: []cli.Command{
		{
			Name:                   "allocation",
			Aliases:                []string{"alloc"},
			Usage:                  "Provides a snapshot of the number of shards allocated to each data node and their disk space.",
			Description:            "For more information: https://www.elastic.co/guide/en/elasticsearch/reference/current/cat-allocation.html",
			Action:                 ctl.NewAction(allocation),
			UseShortOptionHandling: true,
		},
	},
}

func nodes(_ config.Config, conn *client.Client, c *cli.Context) error {
	cat, err := conn.NodesStats().Do(context.Background())
	check.Fatalf(err, "cat nodes: %v", err)

	if len(cat.Nodes) == 0 {
		log.Printf("no nodes")
		return nil
	}

	t := table.New("name", "cpu", "mem", "jvm", "disk")

	for _, node := range cat.Nodes {
		t.Append(
			[]string{
				node.Name,
				fmt.Sprintf("%v%%", node.OS.CPU.Percent),
				fmt.Sprintf("%.2f GB", float64(node.OS.Mem.UsedInBytes)/(1024*1024*1024)),
				fmt.Sprintf("%.2f GB", float64(node.JVM.Mem.HeapUsedInBytes)/(1024*1024*1024)),
				fmt.Sprintf("%.2f GB", float64(node.FS.Total.TotalInBytes-node.FS.Total.FreeInBytes)/(1024*1024*1024)),
			},
		)
	}
	t.Render()

	return nil
}

func allocation(_ config.Config, conn *client.Client, c *cli.Context) error {
	alloc, err := conn.CatAllocation().Do(context.Background())
	check.Fatal(err)

	t := table.New("name", "total", "used", "left", "indices")

	for _, a := range alloc {
		t.Append(
			[]string{
				a.Node,
				a.DiskTotal,
				a.DiskUsed,
				a.DiskAvail,
				a.DiskIndices,
			},
		)
	}
	t.Render()

	return nil
}
