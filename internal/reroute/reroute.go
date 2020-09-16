package reroute

import (
	"context"
	"log"

	"github.com/unqnown/esctl/internal/app"
	"github.com/unqnown/esctl/pkg/check"
	"github.com/unqnown/esctl/pkg/client"
	"github.com/unqnown/esctl/pkg/ctl"
	"github.com/urfave/cli"
)

var Command = cli.Command{
	Name:                   "reroute",
	Usage:                  "Changes the allocation of shards in a cluster.",
	Description:            "For more information: https://www.elastic.co/guide/en/elasticsearch/reference/current/cluster-reroute.html",
	Category:               "Advanced",
	UseShortOptionHandling: true,
	Subcommands: []cli.Command{
		{
			Name:                   "failed",
			Usage:                  "Retries allocation of shards that are blocked due to too many subsequent allocation failures.",
			Action:                 ctl.NewAction(failed),
			UseShortOptionHandling: true,
		},
	},
}

func failed(_ app.Config, conn *client.Client, c *cli.Context) error {
	_, err := conn.ClusterReroute().RetryFailed(true).Do(context.Background())
	check.Fatal(err)

	log.Printf("acknowledged")

	return nil
}
