package restore

import (
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/unqnown/esctl/internal/app"
	"github.com/unqnown/esctl/pkg/backup"
	"github.com/unqnown/esctl/pkg/bar"
	"github.com/unqnown/esctl/pkg/check"
	"github.com/unqnown/esctl/pkg/client"
	"github.com/unqnown/esctl/pkg/ctl"
	"github.com/urfave/cli"
	"github.com/vbauerster/mpb"
)

var Command = cli.Command{
	Name:                   "restore",
	Aliases:                []string{"import"},
	Usage:                  "Imports content to index.",
	Description:            `If index not specified documents will be restored to their indices.`,
	ArgsUsage:              "[index] --dump path/to/dump.json [--lpr 1000]",
	Category:               "Intermediate",
	Action:                 ctl.NewAction(restore),
	UseShortOptionHandling: true,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:     "dump, d",
			Required: true,
			Usage:    "Dump `FILE`",
		},
	},
}

func restore(_ app.Config, conn *client.Client, c *cli.Context) error {
	processor, err := conn.Bulk()
	check.Fatalf(err, "start bulk processor: %v", err)

	r, restoring, wait, err := factory(c.String("dump"))
	check.Fatalf(err, "open dump file: %v", err)
	defer r.Close()

	dec := json.NewDecoder(r)

	index, present := c.Args().First(), c.Args().Present()

	for {
		var doc backup.Document

		if err := dec.Decode(&doc); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			check.Fatalf(err, "read dump: %v", err)
		}

		if present {
			doc.Index = index
		}

		processor.Save(doc)
	}

	err = processor.Close()
	check.Fatalf(err, "flush save tasks: %v", err)

	restoring.SetTotal(0, true)

	wait()

	return nil
}

func factory(file string) (io.ReadCloser, *mpb.Bar, func(), error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, nil, nil, err
	}
	s, err := f.Stat()
	if err != nil {
		return nil, nil, nil, err
	}
	b, wait := bar.Percent(s.Size(), "restoring")

	return b.ProxyReader(f), b, wait, nil
}
