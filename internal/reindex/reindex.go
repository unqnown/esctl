package reindex

import (
	"context"
	"log"
	"time"

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
	ArgsUsage:              "index --destination index [-q]",
	Category:               "Intermediate",
	Action:                 ctl.NewAction(reindex),
	UseShortOptionHandling: true,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "quiet, q",
			Usage: "Skip progress tracking",
		},
		cli.StringFlag{
			Name:     "destination, d",
			Required: true,
			Usage:    "Destination index name",
		},
	},
}

func reindex(_ app.Config, conn *client.Client, c *cli.Context) error {
	if !c.Args().Present() {
		log.Fatal("index not specified")
	}

	src := c.Args().First()
	dst := c.String("destination")

	if c.Bool("quiet") {
		_, err := conn.Reindex().
			SourceIndex(src).
			DestinationIndex(dst).
			ProceedOnVersionConflict().
			WaitForCompletion(false).
			Do(context.Background())
		check.Fatal(err)

		return nil
	}

	total, err := conn.CatCount().Index(src).Do(context.Background())
	check.Fatalf(err, "cat total: %v", err)

	reindexing, wait := bar.Docs(int64(total[0].Count), "reindexing")

	started := time.Now()

	done := make(chan struct{})

	go func() {
		var prev int
	watch:
		for {
			// TODO: index out of range
			select {
			case <-time.NewTicker(time.Second).C:
				reindexed, _ := conn.CatCount().Index(dst).Do(context.Background())
				if len(reindexed) == 0 {
					continue
				}
				reindexing.IncrBy(reindexed[0].Count-prev, time.Since(started))
				prev = reindexed[0].Count
			case <-done:
				reindexed, _ := conn.CatCount().Index(dst).Do(context.Background())
				if len(reindexed) == 0 {
					reindexing.SetTotal(int64(prev), true)

					break watch
				}
				reindexing.SetTotal(int64(reindexed[0].Count), true)
			}
		}
	}()

	_, err = conn.Reindex().
		SourceIndex(src).
		DestinationIndex(dst).
		ProceedOnVersionConflict().
		Do(context.Background())
	check.Fatal(err)

	close(done)

	wait()

	return nil
}
