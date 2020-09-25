package reindex

import (
	"context"
	"log"
	"time"

	"github.com/olivere/elastic/v7"
	"github.com/unqnown/esctl/internal/app"
	"github.com/unqnown/esctl/pkg/bar"
	"github.com/unqnown/esctl/pkg/check"
	"github.com/unqnown/esctl/pkg/client"
	"github.com/unqnown/esctl/pkg/ctl"
	"github.com/urfave/cli"
)

var Command = cli.Command{
	Name:                   "reindex",
	Usage:                  "Copies documents from one index to another.",
	Description:            `For more information: https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-reindex.html`,
	ArgsUsage:              "--src source --dst destination [--remote old]",
	Category:               "Intermediate",
	Action:                 ctl.NewAction(reindex),
	UseShortOptionHandling: true,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:     "source, src, s",
			Required: true,
			Usage:    "Source index name. Means remote index if `--remote` flag presented",
		},
		cli.StringFlag{
			Name:     "destination, dst, d",
			Required: true,
			Usage:    "Destination index name in current context",
		},
		cli.StringFlag{
			Name:     "remote, r",
			Required: false,
			Usage:    "Remote cluster context",
		},
		cli.DurationFlag{
			Name:     "connection_timeout",
			Required: false,
			Usage:    "Connection timeout to connect with the remote cluster",
			Value:    10 * time.Second,
		},
		cli.DurationFlag{
			Name:     "socket_timeout",
			Required: false,
			Usage:    "Socket timeout to connect with the remote cluster.",
			Value:    1 * time.Minute,
		},
	},
}

func reindex(conf app.Config, conn *client.Client, c *cli.Context) error {
	sind := c.String("source")
	dind := c.String("destination")

	src := elastic.NewReindexSource().Index(sind)

	var total int64

	if remote := c.String("remote"); remote != "" {
		rctx, ok := conf.Contexts[remote]
		if !ok {
			log.Fatalf("remote cluster is not configured, please add it to esctl config.")
		}

		rcluster, ok := conf.Clusters[rctx.Cluster]
		if !ok {
			log.Fatalf("remote cluster is not configured, please add it to esctl config.")
		}

		var ruser app.User
		if rctx.User != nil {
			ruser = conf.Users[*rctx.User]
		}

		src.RemoteInfo(
			elastic.NewReindexRemoteInfo().
				Host(rcluster.Servers[0]).
				Username(ruser.Name).
				Password(ruser.Password).
				ConnectTimeout(c.Duration("connection_timeout").String()).
				SocketTimeout(c.Duration("socket_timeout").String()),
		)

		rconn, err := client.New(rcluster, ruser)
		check.Fatalf(err, "connect to %s cluster: %v", remote, err)

		cat, err := rconn.CatCount().Index(sind).Do(context.Background())
		check.Fatalf(err, "cat count: %v", err)

		total = int64(cat[0].Count)
	} else {
		cat, err := conn.CatCount().Index(sind).Do(context.Background())
		check.Fatalf(err, "cat count: %v", err)

		total = int64(cat[0].Count)
	}

	reindexing, wait := bar.Docs(total, "reindexing")

	started := time.Now()

	done := make(chan struct{})

	go func() {
		var prev int
	watch:
		for {
			// TODO: index out of range
			select {
			case <-time.NewTicker(time.Second).C:
				reindexed, _ := conn.CatCount().Index(dind).Do(context.Background())
				if len(reindexed) == 0 {
					continue
				}
				reindexing.IncrBy(reindexed[0].Count-prev, time.Since(started))
				prev = reindexed[0].Count
			case <-done:
				reindexed, _ := conn.CatCount().Index(dind).Do(context.Background())
				if len(reindexed) == 0 {
					reindexing.SetTotal(total, true)
				} else {
					reindexing.SetTotal(int64(reindexed[0].Count), true)
				}
				break watch
			}
		}
	}()

	_, err := conn.Reindex().
		Source(src).
		DestinationIndex(dind).
		ProceedOnVersionConflict().
		Do(context.Background())
	check.Fatal(err)

	close(done)

	wait()

	return nil
}
