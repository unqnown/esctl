package cmd

import (
	"log"
	"os"
	"path/filepath"

	"github.com/unqnown/esctl/internal/alias"
	"github.com/unqnown/esctl/internal/app"
	"github.com/unqnown/esctl/internal/close"
	"github.com/unqnown/esctl/internal/config"
	"github.com/unqnown/esctl/internal/create"
	"github.com/unqnown/esctl/internal/delete"
	"github.com/unqnown/esctl/internal/dump"
	"github.com/unqnown/esctl/internal/get"
	"github.com/unqnown/esctl/internal/open"
	"github.com/unqnown/esctl/internal/refresh"
	"github.com/unqnown/esctl/internal/reindex"
	"github.com/unqnown/esctl/internal/replace"
	"github.com/unqnown/esctl/internal/reroute"
	"github.com/unqnown/esctl/internal/restore"
	"github.com/unqnown/esctl/internal/top"
	"github.com/unqnown/esctl/internal/vacuum"
	"github.com/unqnown/esctl/pkg/check"
	"github.com/unqnown/semver"
	"github.com/urfave/cli"
)

func Run(ver string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	app := cli.NewApp()

	app.Name = "esctl"
	app.Version = ver
	app.Usage = "Elasticsearch cluster managing tool."
	app.Description = "To start using esctl immediately run `esctl init`."
	app.UseShortOptionHandling = true
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "config, conf",
			Usage:  "Path to config `FILE`",
			Value:  filepath.Join(home, ".esctl/config.yaml"),
			EnvVar: "ESCTLCONFIG",
		},
		cli.StringFlag{
			Name:  "context, ctx",
			Usage: "Context to apply",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:                   "init",
			Usage:                  "Initialize or reinitialize esctl.",
			Description:            "Creates default config file.",
			Action:                 _init,
			UseShortOptionHandling: true,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "force, f",
					Usage: "Force esctl reinitialization",
				},
			},
		},
		config.Command,
		get.Command,
		create.Command,
		top.Command,
		alias.Command,
		delete.Command,
		dump.Command,
		restore.Command,
		reindex.Command,
		vacuum.Command,
		replace.Command,
		open.Command,
		close.Command,
		reroute.Command,
		refresh.Command,
	}

	return app.Run(os.Args)
}

func _init(ctx *cli.Context) {
	path := ctx.GlobalString("config")

	switch _, err := os.Stat(path); {
	case err == nil:
		if !ctx.Bool("force") {
			log.Fatalf("config already exists")
		}
	case os.IsNotExist(err):
		// go ahead
	default:
		log.Fatalf("check config exists")
	}

	dir, _ := filepath.Split(path)

	err := os.MkdirAll(dir, os.ModePerm)
	check.Fatal(err)

	ver, err := semver.Parse(ctx.App.Version)
	check.Fatal(err)

	conf := app.NewConfig(ver, dir)

	err = conf.Save(path)
	check.Fatalf(err, "init config: %v", err)

	log.Printf("%q created", path)
}
