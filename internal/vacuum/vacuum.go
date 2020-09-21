package vacuum

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
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
	Name:                   "vacuum",
	Usage:                  "Gently deletes documents from index.",
	ArgsUsage:              "[indices...] [--lpr 1000] [--query path/to/query.json]",
	Action:                 ctl.NewAction(vacuum),
	Category:               "Advanced",
	UseShortOptionHandling: true,
	Flags: []cli.Flag{
		cli.IntFlag{
			Name:  "lpr",
			Usage: "Limit per request",
			Value: 1000,
		},
		cli.StringFlag{
			Name:  "query, q",
			Usage: "Query `FILE`",
		},
	},
}

func vacuum(_ app.Config, conn *client.Client, c *cli.Context) error {
	if !c.Args().Present() {
		log.Fatal("indices not specified")
	}

	scroll := conn.Scroll(c.Args()...).
		Size(c.Int("lpr")).
		FetchSource(false)

	if query := c.String("query"); query != "" {
		q, err := openQuery(query)
		check.Fatalf(err, "open query: %v", err)
		scroll.Query(q)
	}

	started := time.Now()

	vacuum, wait := bar.Docs(0, "vacuum")

	bulk, err := conn.Bulk()
	check.Fatalf(err, "start bulk processor: %v", err)

	for {
		rsp, err := scroll.Do(context.Background())
		if err != nil {
			if errors.Is(err, io.EOF) {
				vacuum.SetTotal(rsp.TotalHits(), true)

				break
			}
			check.Fatalf(err, "scroll: %v", err)
		}

		vacuum.SetTotal(rsp.TotalHits(), false)

		for _, h := range rsp.Hits.Hits {
			bulk.Rm(h.Index, h.Id)
		}

		vacuum.IncrBy(len(rsp.Hits.Hits), time.Since(started))
	}

	err = bulk.Flush()
	check.Fatalf(err, "flush remove tasks: %v", err)

	wait()

	return nil
}

func openQuery(path string) (elastic.Query, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return elastic.NewRawStringQuery(string(data)), nil
}
