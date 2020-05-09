package dump

import (
	"context"
	"fmt"
	"github.com/unqnown/esctl/pkg/ctl"
	"io"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/olivere/elastic/v7"
	"github.com/unqnown/esctl/config"
	"github.com/unqnown/esctl/pkg/bar"
	"github.com/unqnown/esctl/pkg/check"
	"github.com/unqnown/esctl/pkg/client"
	"github.com/unqnown/esctl/pkg/dump"
	"github.com/urfave/cli"
)

var Command = cli.Command{
	Name:                   "dump",
	Aliases:                []string{"export"},
	Usage:                  "Exports index content.",
	Description:            `If no indices specified all indices will be exported.`,
	ArgsUsage:              "[indices...] --dump path/to/dump.json [--lpr 1000] [--query path/to/query.json]",
	Action:                 ctl.NewAction(_dump),
	Category:               "Intermediate",
	UseShortOptionHandling: true,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "dump, d",
			Usage: "Dump `FILE`. By default exports to context's .mappings dir.",
		},
		cli.IntFlag{
			Name:  "lpr",
			Usage: "Documents limit per request",
			Value: 1000,
		},
		cli.StringFlag{
			Name:  "query, q",
			Usage: "Query `FILE`",
		},
	},
}

func _dump(conf config.Config, conn *client.Client, c *cli.Context) error {
	scroll := conn.Scroll(c.Args()...).Size(c.Int("lpr")).
		FetchSource(true)

	if query := c.String("query"); query != "" {
		q, err := openQuery(query)
		check.Fatalf(err, "open query: %v", err)
		scroll.Query(q)
	}

	file := c.String("dump")
	if file == "" {
		loc, err := conf.Location()
		check.Fatal(err)
		file = filepath.Join(loc.Backups, fmt.Sprintf("%s.json", time.Now().Format("2006-01-02T15:04:05")))
	}

	w, err := dump.NewFileWriter(file)
	check.Fatalf(err, "create dump file: %v", err)
	defer w.Close()

	dumping, wait := bar.Docs(0, "dumping")

	started := time.Now()

read:
	for {
		rsp, err := scroll.Do(context.Background())
		switch err {
		case nil:
			// go ahead
		case io.EOF:
			break read
		default:
			check.Fatalf(err, "scroll: %v", err)
		}

		dumping.SetTotal(rsp.TotalHits(), false)

		size := len(rsp.Hits.Hits)
		docs := make([]dump.Doc, size)
		for i, hit := range rsp.Hits.Hits {
			docs[i].ID = hit.Id
			docs[i].Index = hit.Index
			docs[i].Body = hit.Source
		}

		_, err = w.Write(docs...)
		check.Fatalf(err, "write dump page: %v", err)

		dumping.IncrBy(size, time.Since(started))
	}

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
