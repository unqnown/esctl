package dump

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/olivere/elastic/v7"
	"github.com/unqnown/esctl/internal/app"
	"github.com/unqnown/esctl/pkg/backup"
	"github.com/unqnown/esctl/pkg/bar"
	"github.com/unqnown/esctl/pkg/check"
	"github.com/unqnown/esctl/pkg/client"
	"github.com/unqnown/esctl/pkg/ctl"
	"github.com/urfave/cli"
)

var Command = cli.Command{
	Name:                   "dump",
	Aliases:                []string{"export"},
	Usage:                  "Exports index content.",
	Description:            `If no indices specified all indices will be exported.`,
	ArgsUsage:              "[indices...] --dump path/to/dump.json [--lpr 1000] [--query path/to/query.json]",
	Action:                 ctl.NewAction(dump),
	Category:               "Intermediate",
	UseShortOptionHandling: true,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "dump, d",
			Usage: "Dump `FILE`. By default exports to context's .backup dir.",
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
		cli.BoolFlag{
			Name:  "plain, p",
			Usage: "Dump in plain json",
		},
	},
}

func dump(conf app.Config, conn *client.Client, c *cli.Context) error {
	scroll := conn.Scroll(c.Args()...).Size(c.Int("lpr")).
		FetchSource(true)

	if query := c.String("query"); query != "" {
		q, err := openQuery(query)
		check.Fatalf(err, "open query: %v", err)
		scroll.Query(q)
	}

	file := c.String("dump")
	if file == "" {
		file = filepath.Join(conf.Home, app.BackupDir, fmt.Sprintf("%s.json", time.Now().Format("2006-01-02T15:04:05")))
	}

	f, err := newDumpFile(file)
	check.Fatalf(err, "create dump file: %v", err)
	defer f.Close()

	enc := json.NewEncoder(f)
	if !c.Bool("plain") {
		enc.SetIndent("", "	")
	}

	var (
		estimated int64
		processed int64
	)
	dumping, wait := bar.Docs(estimated, "dumping")

	started := time.Now()

	rsp, err := scroll.Do(context.Background())
	if err != nil {
		if errors.Is(err, io.EOF) {
			dumping.SetTotal(0, true)
			wait()

			return nil
		}
		check.Fatalf(err, "scroll: %v", err)
	}
	estimated = rsp.TotalHits()

	for {
		dumping.SetTotal(max(estimated, processed+int64(len(rsp.Hits.Hits))), false)

		for _, hit := range rsp.Hits.Hits {
			err = enc.Encode(backup.Document{
				ID:    hit.Id,
				Index: hit.Index,
				Body:  hit.Source,
			})
			check.Fatalf(err, "write dump page: %v", err)
			processed++

			dumping.IncrBy(1, time.Since(started))
		}

		rsp, err = scroll.Do(context.Background())
		if err != nil {
			if errors.Is(err, io.EOF) {
				dumping.SetTotal(max(estimated, processed), true)

				break
			}
			check.Fatalf(err, "scroll: %v", err)
		}
	}

	wait()

	return nil
}

func max(vs ...int64) int64 {
	if len(vs) == 0 {
		return 0
	}
	max := vs[0]
	for i := 1; i < len(vs); i++ {
		if vs[i] > max {
			max = vs[i]
		}
	}
	return max
}

func openQuery(path string) (elastic.Query, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return elastic.NewRawStringQuery(string(data)), nil
}

func newDumpFile(path string) (*os.File, error) {
	dir, _ := filepath.Split(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, err
	}

	return os.Create(path)
}
