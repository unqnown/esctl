package refresh

import (
	"context"
	"log"
	"strconv"

	"github.com/unqnown/esctl/internal/app"
	"github.com/unqnown/esctl/pkg/check"
	"github.com/unqnown/esctl/pkg/client"
	"github.com/unqnown/esctl/pkg/ctl"
	"github.com/unqnown/esctl/pkg/table"
	"github.com/urfave/cli"
)

var Command = cli.Command{
	Name:                   "refresh",
	Usage:                  "Refreshes one or more indices.",
	Description:            `For more information: https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-refresh.html`,
	Category:               "Intermediate",
	Action:                 ctl.NewAction(refresh),
	UseShortOptionHandling: true,
	Flags:                  []cli.Flag{},
}

func refresh(_ app.Config, conn *client.Client, c *cli.Context) error {
	if len(c.Args()) == 0 {
		log.Printf("nothing to refresh")

		return nil
	}

	done, err := conn.Refresh(c.Args()...).Do(context.Background())
	check.Fatal(err)

	t := table.New("total", "successful", "failed", "skipped")

	t.Append([]string{
		strconv.Itoa(done.Shards.Total),
		strconv.Itoa(done.Shards.Successful),
		strconv.Itoa(done.Shards.Failed),
		strconv.Itoa(done.Shards.Skipped),
	})
	t.Render()

	return nil
}
