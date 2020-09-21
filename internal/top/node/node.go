package node

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/olivere/elastic/v7"
	"github.com/unqnown/esctl/internal/app"
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
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "sort",
			Usage: "Sort output in format column[:order].",
			Value: "name",
		},
		cli.BoolFlag{
			Name:  "master, m",
			Usage: "Show master nodes",
		},
		cli.BoolFlag{
			Name:  "data, d",
			Usage: "Show data nodes",
		},
		cli.BoolFlag{
			Name:  "ingest, i",
			Usage: "Show ingest nodes",
		},
		cli.BoolFlag{
			Name:  "transform, t",
			Usage: "Show transform nodes",
		},
	},
}

func nodes(_ app.Config, conn *client.Client, c *cli.Context) error {
	cat, err := conn.NodesStats().Do(context.Background())
	check.Fatalf(err, "cat nodes: %v", err)

	if len(cat.Nodes) == 0 {
		log.Printf("no nodes")

		return nil
	}

	filter := make(map[string]bool)

	if c.Bool("master") {
		filter["master"] = true
	}
	if c.Bool("data") {
		filter["data"] = true
	}
	if c.Bool("ingest") {
		filter["ingest"] = true
	}
	if c.Bool("transform") {
		filter["transform"] = true
	}

	nodes := make([]*elastic.NodesStatsNode, 0, len(cat.Nodes))
	for _, node := range cat.Nodes {
		if !show(filter, node.Roles...) {
			continue
		}
		nodes = append(nodes, node)
	}

	sortnode(nodes, c.String("sort"))

	t := table.New("name", "cpu", "mem", "jvm", "disk")

	for _, node := range nodes {
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
	if t.NumLines() == 0 {
		log.Printf("no such nodes")

		return nil
	}
	t.Render()

	return nil
}

const asc = "asc"

func sortnode(nodes []*elastic.NodesStatsNode, sorting string) {
	var by, dir string

	splits := strings.SplitN(sorting, ":", 2)
	switch len(splits) {
	case 1:
		by, dir = splits[0], "desc"
	case 2:
		by, dir = splits[0], splits[1]
	}

	switch by {
	case "name":
		sort.Slice(nodes, func(i, j int) bool { return order(dir == asc, nodes[i].Name < nodes[j].Name) })
	case "cpu":
		sort.Slice(nodes, func(i, j int) bool { return order(dir == asc, nodes[i].OS.CPU.Percent < nodes[j].OS.CPU.Percent) })
	case "mem":
		sort.Slice(nodes, func(i, j int) bool {
			return order(dir == asc, nodes[i].OS.Mem.UsedInBytes < nodes[j].OS.Mem.UsedInBytes)
		})
	case "jvm":
		sort.Slice(nodes, func(i, j int) bool {
			return order(dir == asc, nodes[i].JVM.Mem.HeapUsedInBytes < nodes[j].JVM.Mem.HeapUsedInBytes)
		})
	case "disk":
		sort.Slice(nodes, func(i, j int) bool {
			return order(dir == asc, nodes[i].FS.Total.TotalInBytes-nodes[i].FS.Total.FreeInBytes < nodes[j].FS.Total.TotalInBytes-nodes[j].FS.Total.FreeInBytes)
		})
	}
}

func order(asc bool, v bool) bool {
	if asc {
		return v
	}

	return !v
}

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

func allocation(_ app.Config, conn *client.Client, c *cli.Context) error {
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
