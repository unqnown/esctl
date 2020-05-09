package restore

import (
	"github.com/unqnown/esctl/config"
	"github.com/unqnown/esctl/pkg/bar"
	"github.com/unqnown/esctl/pkg/check"
	"github.com/unqnown/esctl/pkg/client"
	"github.com/unqnown/esctl/pkg/ctl"
	"github.com/unqnown/esctl/pkg/dump"
	"github.com/urfave/cli"
	"github.com/vbauerster/mpb"
	"io"
	"os"
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
		cli.IntFlag{
			Name:  "lpr",
			Usage: "Documents limit per request",
			Value: 1000,
		},
	},
}

func restore(_ config.Config, conn *client.Client, c *cli.Context) error {
	processor, err := conn.Bulk()
	check.Fatalf(err, "start bulk processor: %v", err)

	r, restoring, wait, err := factory(c.String("dump"))
	check.Fatalf(err, "open dump file: %v", err)
	defer r.Close()

	docs := make([]dump.Doc, c.Int("lpr"))

	index, present := c.Args().First(), c.Args().Present()

write:
	for {
		n, err := r.Read(docs)
		for _, doc := range docs[:n] {
			if present {
				doc.Index = index
			}
			processor.Save(doc)
		}

		switch err {
		case nil:
			// go ahead
		case io.EOF:
			break write
		default:
			check.Fatalf(err, "read dump: %v", err)
		}
	}

	err = processor.Flush()
	check.Fatalf(err, "flush save tasks: %v", err)

	restoring.SetTotal(0, true)

	wait()

	return nil
}

func factory(file string) (dump.ReadCloser, *mpb.Bar, func(), error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, nil, nil, err
	}
	s, err := f.Stat()
	if err != nil {
		return nil, nil, nil, err
	}
	b, wait := bar.Percent(s.Size(), "restoring")
	return dump.NewReadCloser(b.ProxyReader(f)), b, wait, nil
}
