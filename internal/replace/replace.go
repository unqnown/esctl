package replace

import (
	"context"
	"github.com/unqnown/esctl/pkg/ctl"
	"io/ioutil"
	"log"
	"time"

	"github.com/olivere/elastic/v7"
	"github.com/unqnown/esctl/internal/app"
	"github.com/unqnown/esctl/pkg/bar"
	"github.com/unqnown/esctl/pkg/check"
	"github.com/unqnown/esctl/pkg/client"
	"github.com/urfave/cli"
)

var Command = cli.Command{
	Name:  "replace",
	Usage: "Replaces the index by another one.",
	Description: `Removes index and creates an alias to new index same as removed index name.
   Index swap performs atomically.

   For more information: https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-aliases.html`,
	ArgsUsage:              "index --by new [--create path/to/body.json]",
	Category:               "Advanced",
	Action:                 ctl.NewAction(replace),
	UseShortOptionHandling: true,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:     "by",
			Required: true,
			Usage:    "Index replace by",
		},
		cli.StringFlag{
			Name:  "create",
			Usage: "Index body `FILE`",
		},
		cli.BoolFlag{
			Name:  "reindex, r",
			Usage: "Reindex content of replaced index to new one",
		},
	},
}

func replace(_ app.Config, conn *client.Client, c *cli.Context) error {
	if !c.Args().Present() {
		log.Fatal("index not specified")
	}

	index := c.Args().First()
	by := c.String("by")

	if create := c.String("create"); create != "" {
		body, err := ioutil.ReadFile(create)
		check.Fatalf(err, "open body: %v", err)
		// TODO(d.andriichuk): copy index body instead of mapping load.

		_, err = conn.CreateIndex(by).Body(string(body)).Do(context.Background())
		check.Fatal(err)
		log.Printf("%q created", by)
	}

	if c.Bool("reindex") {
		total, err := conn.CatCount().Index(index).Do(context.Background())
		check.Fatalf(err, "cat total: %v", err)

		reindexing, wait := bar.Docs(int64(total[0].Count), "reindexing")

		started := time.Now()

		done := make(chan struct{})

		go func() {
			var prev int
		watch:
			for {
				select {
				case <-time.NewTicker(time.Second).C:
					reindexed, _ := conn.CatCount().Index(by).Do(context.Background())
					if len(reindexed) == 0 {
						continue
					}
					reindexing.IncrBy(reindexed[0].Count-prev, time.Since(started))
					prev = reindexed[0].Count
				case <-done:
					reindexed, _ := conn.CatCount().Index(by).Do(context.Background())
					if len(reindexed) == 0 {
						reindexing.SetTotal(int64(prev), true)
						break watch
					}
					reindexing.SetTotal(int64(reindexed[0].Count), true)
				}
			}
		}()

		_, err = conn.Reindex().
			SourceIndex(index).
			DestinationIndex(by).
			ProceedOnVersionConflict().
			Do(context.Background())
		check.Fatal(err)

		close(done)

		wait()
	}

	_, err := conn.Alias().
		Add(by, index).
		Action(elastic.NewAliasRemoveIndexAction(index)).
		Do(context.Background())
	check.Fatal(err)

	log.Printf("%q replaced by %q", index, by)

	return nil
}
